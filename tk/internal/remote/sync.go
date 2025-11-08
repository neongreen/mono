package remote

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/segment"
)

// PushResult contains the result of a push operation
type PushResult struct {
	SegmentsPushed int
}

// PushParams captures the inputs required for a push operation.
type PushParams struct {
	RemoteName   string
	RemoteConfig config.RemoteConfig
	Space        string
	StateDir     string
	SyncConfig   config.SyncConfig
	ExportAll    bool
}

// PullResult contains the result of a pull operation
type PullResult struct {
	SegmentsPulled int
}

// Push pushes local segments to a remote
func Push(db *database.DB, params PushParams) (*PushResult, error) {
	log := slog.With("remote", params.RemoteName, "space", params.Space)
	log.Debug("push: starting", "remote_path", params.RemoteConfig.Path)

	// Get current node ID
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return nil, fmt.Errorf("failed to get node ID: %w", err)
	}
	log.Debug("push: got node ID", "node_id", nodeID)

	// First export any local events to segments
	exportResult, err := Export(db, ExportParams{
		RemoteName:   params.RemoteName,
		RemoteConfig: params.RemoteConfig,
		Space:        params.Space,
		ExportAll:    params.ExportAll,
		StateDir:     params.StateDir,
		SyncConfig:   params.SyncConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to export events before push: %w", err)
	}
	log.Info("push: export completed", "segments_written", exportResult.SegmentsWritten, "events_exported", exportResult.EventsExported)

	// Load local index mirror
	localIndexPath := filepath.Join(params.StateDir, "remotes", params.RemoteName, params.Space, "index.json")
	log.Debug("push: loading local index", "path", localIndexPath)
	localIndex, err := LoadIndexFile(localIndexPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load local index: %w", err)
	}
	if localIndex == nil {
		log.Debug("push: no local index found, nothing to push")
		return &PushResult{SegmentsPushed: 0}, nil
	}
	log.Info("push: local index loaded", "segments", len(localIndex.Segments))

	// Try to restore missing segments from cache (for current node only)
	// This helps if the remote was accidentally cleared but we have cached copies
	restoredCount := 0
	for _, seg := range localIndex.Segments {
		if !segment.SegmentBelongsToNode(seg.Rel, nodeID) {
			continue // Other nodes' segments aren't our responsibility
		}
		wasRestored, err := RestoreSegmentFromCache(params.StateDir, params.RemoteName, params.RemoteConfig.Path, seg)
		if err != nil {
			return nil, err
		}
		if wasRestored {
			restoredCount++
		}
	}
	if restoredCount > 0 {
		log.Info("push: restored segments from cache", "count", restoredCount)
	}

	// Check if remote index exists
	remoteIndexPath := filepath.Join(params.RemoteConfig.Path, params.Space, "index.json")
	log.Debug("push: loading remote index", "path", remoteIndexPath)
	remoteIndex, err := LoadIndexFile(remoteIndexPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load remote index: %w", err)
	}

	// If remote index doesn't exist, create it
	if remoteIndex == nil {
		log.Debug("push: remote index not found, creating new index")
		remoteIndex = &IndexFile{
			Schema:   "tk.index.v1",
			Space:    params.Space,
			Segments: []SegmentInfo{},
		}
	} else {
		log.Info("push: remote index loaded", "segments", len(remoteIndex.Segments))
	}

	// Find segments that are in local but not in remote
	remoteSegmentPaths := make(map[string]bool)
	for _, seg := range remoteIndex.Segments {
		remoteSegmentPaths[seg.Rel] = true
	}

	var segmentsToPush []SegmentInfo
	for _, seg := range localIndex.Segments {
		if !remoteSegmentPaths[seg.Rel] {
			segmentsToPush = append(segmentsToPush, seg)
		}
	}

	log.Info("push: compared indices", "local_segments", len(localIndex.Segments), "remote_segments", len(remoteIndex.Segments), "to_push", len(segmentsToPush))

	if len(segmentsToPush) == 0 {
		log.Debug("push: no new segments to push")
		return &PushResult{SegmentsPushed: 0}, nil
	}

	// Verify each segment exists before adding to remote index
	// Only add segments that actually exist on the remote filesystem
	var verifiedSegments []SegmentInfo
	for _, seg := range segmentsToPush {
		segFullPath := filepath.Join(params.RemoteConfig.Path, seg.Rel)
		if stat, err := os.Stat(segFullPath); err == nil {
			log.Debug("push: segment verified", "rel_path", seg.Rel, "size", stat.Size())
			verifiedSegments = append(verifiedSegments, seg)
		} else {
			log.Debug("push: segment not found at remote, skipping", "rel_path", seg.Rel)
		}
	}

	if len(verifiedSegments) == 0 {
		log.Debug("push: no verified segments to add to remote index")
		return &PushResult{SegmentsPushed: 0}, nil
	}

	log.Debug("push: adding verified segments to remote index", "count", len(verifiedSegments))

	// Add verified segments to remote index
	remoteIndex.Segments = append(remoteIndex.Segments, verifiedSegments...)
	log.Debug("push: added segments to remote index", "new_total", len(remoteIndex.Segments))

	// Save remote index
	log.Debug("push: saving remote index", "path", remoteIndexPath)
	if err := SaveIndexFile(remoteIndexPath, remoteIndex); err != nil {
		return nil, fmt.Errorf("failed to save remote index: %w", err)
	}

	log.Info("push: completed", "segments_pushed", len(verifiedSegments))
	return &PushResult{SegmentsPushed: len(verifiedSegments)}, nil
}

// Pull pulls segments from a remote
func Pull(remoteName string, remote config.RemoteConfig, space string, stateDir string) (*PullResult, error) {
	log := slog.With("remote", remoteName, "space", space)
	log.Debug("pull: starting", "remote_path", remote.Path)

	// Load remote index
	remoteIndexPath := filepath.Join(remote.Path, space, "index.json")
	log.Debug("pull: loading remote index", "path", remoteIndexPath)
	remoteIndex, err := LoadIndexFile(remoteIndexPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Debug("pull: remote index not found, attempting to reconstruct")
			// Try to reconstruct index by scanning segments
			remoteIndex, err = ReconstructIndex(remote.Path, space)
			if err != nil {
				return nil, fmt.Errorf("failed to reconstruct index: %w", err)
			}
			log.Info("pull: reconstructed index", "segments", len(remoteIndex.Segments))
			// Save reconstructed index
			if err := SaveIndexFile(remoteIndexPath, remoteIndex); err != nil {
				// Non-fatal, just log
				log.Warn("pull: failed to save reconstructed index", "error", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load remote index: %w", err)
		}
	} else {
		log.Info("pull: remote index loaded", "segments", len(remoteIndex.Segments))
	}

	if remoteIndex == nil || len(remoteIndex.Segments) == 0 {
		log.Debug("pull: no segments in remote index")
		return &PullResult{SegmentsPulled: 0}, nil
	}

	// Verify segments exist - if any are missing, regenerate index from actual files
	log.Debug("pull: verifying segment files exist")
	missingSegments := 0
	for _, seg := range remoteIndex.Segments {
		segPath := filepath.Join(remote.Path, seg.Rel)
		if _, err := os.Stat(segPath); err != nil {
			missingSegments++
			log.Debug("pull: segment file missing", "path", segPath, "error", err)
		}
	}

	if missingSegments > 0 {
		log.Warn("pull: found missing segment files", "count", missingSegments)
	}

	// If any segments are missing, regenerate index from actual files
	if missingSegments > 0 {
		log.Debug("pull: regenerating index from actual files")
		remoteIndex, err = ReconstructIndex(remote.Path, space)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct index: %w", err)
		}
		log.Info("pull: regenerated index", "segments", len(remoteIndex.Segments))
		// Save regenerated index to remote
		if err := SaveIndexFile(remoteIndexPath, remoteIndex); err != nil {
			// Non-fatal, just log
			log.Warn("pull: failed to save regenerated index", "error", err)
		}
	}

	// For folder remotes, segments are already accessible locally
	// Count segments that actually exist
	log.Debug("pull: counting segments that exist")
	segmentCount := 0
	for _, seg := range remoteIndex.Segments {
		segPath := filepath.Join(remote.Path, seg.Rel)
		if _, err := os.Stat(segPath); err == nil {
			segmentCount++
			log.Debug("pull: segment exists", "path", segPath)
		}
	}

	// Save local index mirror
	localIndexPath := filepath.Join(stateDir, "remotes", remoteName, space, "index.json")
	log.Debug("pull: saving local index mirror", "path", localIndexPath)
	if err := SaveIndexFile(localIndexPath, remoteIndex); err != nil {
		return nil, fmt.Errorf("failed to save local index mirror: %w", err)
	}

	log.Info("pull: completed", "segments_pulled", segmentCount)
	return &PullResult{SegmentsPulled: segmentCount}, nil
}

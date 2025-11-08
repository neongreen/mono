package remote

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/neongreen/mono/tk/internal/config"
)

// PushResult contains the result of a push operation
type PushResult struct {
	SegmentsPushed int
}

// PullResult contains the result of a pull operation
type PullResult struct {
	SegmentsPulled int
}

// Push pushes local segments to a remote
func Push(remoteName string, remote config.RemoteConfig, space string, stateDir string) (*PushResult, error) {
	log := slog.With("remote", remoteName, "space", space)
	log.Debug("push: starting", "remote_path", remote.Path)

	// Load local index mirror
	localIndexPath := filepath.Join(stateDir, "remotes", remoteName, space, "index.json")
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

	// Check if remote index exists
	remoteIndexPath := filepath.Join(remote.Path, space, "index.json")
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
			Space:    space,
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

	// Log each segment to push with details
	for i, seg := range segmentsToPush {
		segFullPath := filepath.Join(remote.Path, seg.Rel)
		segExists := false
		if stat, err := os.Stat(segFullPath); err == nil {
			segExists = true
			log.Debug("push: segment to push", "index", i, "rel_path", seg.Rel, "full_path", segFullPath, "exists", segExists, "size", stat.Size())
		} else {
			log.Debug("push: segment to push", "index", i, "rel_path", seg.Rel, "full_path", segFullPath, "exists", segExists, "error", err)
		}
	}

	// Copy segment files to remote (they should already exist locally in the remote path)
	// Since we're using folder remotes, the segments are already written to the remote path
	// We just need to update the index
	log.Debug("push: segments should already exist at remote path, only updating index")

	// Add new segments to remote index
	remoteIndex.Segments = append(remoteIndex.Segments, segmentsToPush...)
	log.Debug("push: added segments to remote index", "new_total", len(remoteIndex.Segments))

	// Save remote index
	log.Debug("push: saving remote index", "path", remoteIndexPath)
	if err := SaveIndexFile(remoteIndexPath, remoteIndex); err != nil {
		return nil, fmt.Errorf("failed to save remote index: %w", err)
	}

	log.Info("push: completed", "segments_pushed", len(segmentsToPush))
	return &PushResult{SegmentsPushed: len(segmentsToPush)}, nil
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

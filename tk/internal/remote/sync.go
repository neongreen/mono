package remote

import (
	"fmt"
	"os"
	"path/filepath"
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
	// Load local index mirror
	localIndexPath := filepath.Join(stateDir, "remotes", remoteName, space, "index.json")
	localIndex, err := LoadIndexFile(localIndexPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load local index: %w", err)
	}
	if localIndex == nil {
		return &PushResult{SegmentsPushed: 0}, nil
	}

	// Check if remote index exists
	remoteIndexPath := filepath.Join(remote.Path, space, "index.json")
	remoteIndex, err := LoadIndexFile(remoteIndexPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load remote index: %w", err)
	}

	// If remote index doesn't exist, create it
	if remoteIndex == nil {
		remoteIndex = &IndexFile{
			Schema:   "tk.index.v1",
			Space:    space,
			Segments: []SegmentInfo{},
		}
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

	if len(segmentsToPush) == 0 {
		return &PushResult{SegmentsPushed: 0}, nil
	}

	// Copy segment files to remote (they should already exist locally in the remote path)
	// Since we're using folder remotes, the segments are already written to the remote path
	// We just need to update the index

	// Add new segments to remote index
	remoteIndex.Segments = append(remoteIndex.Segments, segmentsToPush...)

	// Save remote index
	if err := SaveIndexFile(remoteIndexPath, remoteIndex); err != nil {
		return nil, fmt.Errorf("failed to save remote index: %w", err)
	}

	return &PushResult{SegmentsPushed: len(segmentsToPush)}, nil
}

// Pull pulls segments from a remote
func Pull(remoteName string, remote config.RemoteConfig, space string, stateDir string) (*PullResult, error) {
	// Load remote index
	remoteIndexPath := filepath.Join(remote.Path, space, "index.json")
	remoteIndex, err := LoadIndexFile(remoteIndexPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Try to reconstruct index by scanning segments
			remoteIndex, err = ReconstructIndex(remote.Path, space)
			if err != nil {
				return nil, fmt.Errorf("failed to reconstruct index: %w", err)
			}
			// Save reconstructed index
			if err := SaveIndexFile(remoteIndexPath, remoteIndex); err != nil {
				// Non-fatal, just log
				fmt.Printf("Warning: failed to save reconstructed index: %v\n", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load remote index: %w", err)
		}
	}

	if remoteIndex == nil || len(remoteIndex.Segments) == 0 {
		return &PullResult{SegmentsPulled: 0}, nil
	}

	// Verify segments exist - if any are missing, regenerate index from actual files
	missingSegments := 0
	for _, seg := range remoteIndex.Segments {
		segPath := filepath.Join(remote.Path, seg.Rel)
		if _, err := os.Stat(segPath); err != nil {
			missingSegments++
		}
	}

	// If any segments are missing, regenerate index from actual files
	if missingSegments > 0 {
		remoteIndex, err = ReconstructIndex(remote.Path, space)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct index: %w", err)
		}
		// Save regenerated index to remote
		if err := SaveIndexFile(remoteIndexPath, remoteIndex); err != nil {
			// Non-fatal, just log
			fmt.Printf("Warning: failed to save regenerated index: %v\n", err)
		}
	}

	// For folder remotes, segments are already accessible locally
	// Count segments that actually exist
	segmentCount := 0
	for _, seg := range remoteIndex.Segments {
		segPath := filepath.Join(remote.Path, seg.Rel)
		if _, err := os.Stat(segPath); err == nil {
			segmentCount++
		}
	}

	// Save local index mirror
	localIndexPath := filepath.Join(stateDir, "remotes", remoteName, space, "index.json")
	if err := SaveIndexFile(localIndexPath, remoteIndex); err != nil {
		return nil, fmt.Errorf("failed to save local index mirror: %w", err)
	}

	return &PullResult{SegmentsPulled: segmentCount}, nil
}

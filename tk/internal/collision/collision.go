package collision

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/segment"
	"github.com/neongreen/mono/tk/internal/sync"
)

// NodeCollisionChecker checks for node ID collisions
type NodeCollisionChecker struct {
	localNodeID string
	seenNodes   map[string]bool
}

// NewNodeCollisionChecker creates a new collision checker
func NewNodeCollisionChecker(localNodeID string) *NodeCollisionChecker {
	return &NodeCollisionChecker{
		localNodeID: localNodeID,
		seenNodes:   make(map[string]bool),
	}
}

// extractNodeIDFromSegmentFilename extracts the node ID from a segment filename.
// Segment filename format: YYYY-MM-DDThh-mm-ssZ_<node>_v1_s<segment_seq>.jsonl.zst
func extractNodeIDFromSegmentFilename(filename string) string {
	// Remove directory path if present
	base := filepath.Base(filename)
	// Format: YYYY-MM-DDThh-mm-ssZ_<node>_v1_s<segment_seq>.jsonl.zst
	// Find the parts: timestamp_nodeID_v1_s...
	parts := strings.Split(base, "_")
	if len(parts) < 3 {
		return ""
	}
	// parts[1] should be the node ID (between timestamp and "v1")
	return parts[1]
}

// CheckSegment checks a segment file for node ID collisions
// It skips segments created by the local node (identified by node ID in filename)
func (ncc *NodeCollisionChecker) CheckSegment(segmentPath string) error {
	// Extract node ID from filename
	segmentNodeID := extractNodeIDFromSegmentFilename(segmentPath)

	// Skip segments created by the local node
	// Both are 6-character node IDs, so compare directly
	if segmentNodeID == ncc.localNodeID {
		// This segment was created by the local node, skip it
		return nil
	}

	reader := segment.NewSegmentReader(segmentPath)
	events, err := reader.ReadEvents()
	if err != nil {
		return fmt.Errorf("failed to read segment: %w", err)
	}

	for _, event := range events {
		ncc.seenNodes[event.Node] = true
	}

	return nil
}

// CheckForCollisions checks if there are any collisions
func (ncc *NodeCollisionChecker) CheckForCollisions() error {
	// Check if our local node ID appears in the remote
	// This indicates a potential collision - the same node ID is being used by multiple machines
	if ncc.seenNodes[ncc.localNodeID] {
		return fmt.Errorf("node ID collision detected: local node ID %s also appears in remote events", ncc.localNodeID)
	}

	// Note: A more sophisticated collision detection would use machine fingerprints
	// to detect if different machines are generating events with the same node ID.
	// For now, we can only detect if we see our own node ID in remote events,
	// which is a clear indicator of collision.
	return nil
}

// GetSeenNodes returns all seen node IDs
func (ncc *NodeCollisionChecker) GetSeenNodes() []string {
	nodes := make([]string, 0, len(ncc.seenNodes))
	for node := range ncc.seenNodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// CheckNodeCollision checks for node ID collisions in a remote
func CheckNodeCollision(db *database.DB, remoteName string, remote sync.RemoteConfig) error {
	// Get local node ID
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return fmt.Errorf("failed to get node ID: %w", err)
	}

	checker := NewNodeCollisionChecker(nodeID)

	// Scan all segments in the remote
	for _, space := range remote.Spaces {
		segmentsDir := filepath.Join(remote.Path, space, "segments")
		if _, err := os.Stat(segmentsDir); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(segmentsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".zst" {
				if err := checker.CheckSegment(path); err != nil {
					fmt.Printf("Warning: failed to check segment %s: %v\n", path, err)
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to scan segments: %w", err)
		}
	}

	// Check for collisions
	if err := checker.CheckForCollisions(); err != nil {
		return err
	}

	// Report seen nodes
	seenNodes := checker.GetSeenNodes()
	if len(seenNodes) > 0 {
		fmt.Printf("Info: Found %d other node(s) in remote '%s': %v\n",
			len(seenNodes), remoteName, seenNodes)
	}

	return nil
}

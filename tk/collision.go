package main

import (
	"fmt"
	"os"
	"path/filepath"
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

// CheckSegment checks a segment file for node ID collisions
func (ncc *NodeCollisionChecker) CheckSegment(segmentPath string) error {
	reader := NewSegmentReader(segmentPath)
	events, err := reader.ReadEvents()
	if err != nil {
		return fmt.Errorf("failed to read segment: %w", err)
	}

	for _, event := range events {
		if event.Node != ncc.localNodeID {
			ncc.seenNodes[event.Node] = true
		}
	}

	return nil
}

// CheckForCollisions checks if there are any collisions
func (ncc *NodeCollisionChecker) CheckForCollisions() error {
	// For now, we just track seen nodes
	// A real collision would require tracking machine fingerprints
	// which is beyond the scope of v1

	// In v1, we assume no collisions unless the same node ID appears
	// from different sources (which we can't detect without fingerprints)
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

// checkNodeCollision checks for node ID collisions in a remote
func checkNodeCollision(db *DB, remoteName string, remote RemoteConfig) error {
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

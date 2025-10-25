package main

import (
	"fmt"
)

// GenerateTaskID generates a task ID in the format tk-<number>-<node>
func GenerateTaskID(db *DB) (string, error) {
	// Get the node ID
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return "", fmt.Errorf("failed to get node ID: %w", err)
	}

	// Get the next task number
	taskNum, err := db.GetNextTaskNumber()
	if err != nil {
		return "", fmt.Errorf("failed to get next task number: %w", err)
	}

	return fmt.Sprintf("tk-%d-%s", taskNum, nodeID), nil
}

// GenerateEventID generates an event ID in the format ev-<number>-<node>
func GenerateEventID(db *DB) (string, error) {
	// Get the node ID
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return "", fmt.Errorf("failed to get node ID: %w", err)
	}

	// Get the next event number
	eventNum, err := db.GetNextEventNumber()
	if err != nil {
		return "", fmt.Errorf("failed to get next event number: %w", err)
	}

	return fmt.Sprintf("ev-%d-%s", eventNum, nodeID), nil
}

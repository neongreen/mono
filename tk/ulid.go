package main

import (
	"fmt"
	"strings"
)

// GenerateTaskUUID generates a unique task UUID in the format task-<ulid>
func GenerateTaskUUID() string {
	return "task-" + strings.ToLower(generateULID())
}

// GenerateTaskID generates a task ID in the format <prefix>-<number>-<node>
func GenerateTaskID(db *DB, prefix string) (string, error) {
	// Get the node ID
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return "", fmt.Errorf("failed to get node ID: %w", err)
	}

	// Get the next task number for this prefix
	taskNum, err := db.GetNextTaskNumberForPrefix(prefix)
	if err != nil {
		return "", fmt.Errorf("failed to get next task number for prefix %q: %w", prefix, err)
	}

	return fmt.Sprintf("%s-%d-%s", prefix, taskNum, nodeID), nil
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

func generateULID() string {
	// Use the same node ID generation logic for now (6 chars)
	// In production, this would be a proper ULID
	return generateNodeID(20) // 20 chars for UUID
}

// splitEventID splits an event ID into its components
// Format: ev-<number>-<node> or other formats
func splitEventID(eventID string) []string {
	return strings.Split(eventID, "-")
}

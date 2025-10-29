package main

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

// GenerateTaskUUID generates a unique task UUID in the format task-<ulid>
func GenerateTaskUUID() (string, error) {
	ulid, err := generateULID()
	if err != nil {
		return "", err
	}
	return "task-" + strings.ToLower(ulid), nil
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

func generateULID() (string, error) {
	// Generate a proper ULID using the oklog/ulid library
	// ULIDs are lexicographically sortable and contain a timestamp
	entropy := ulid.Monotonic(rand.Reader, 0)
	id, err := ulid.New(ulid.Timestamp(time.Now()), entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate ULID: %w", err)
	}
	return id.String(), nil
}

// splitEventID splits an event ID into its components
// Format: ev-<number>-<node> or other formats
func splitEventID(eventID string) []string {
	return strings.Split(eventID, "-")
}

package database

import "fmt"

// GenerateEventID generates an event ID in the format ev-<number>-<node>
func GenerateEventID(db *DB) (string, error) {

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return "", fmt.Errorf("failed to get node ID: %w", err)
	}

	eventNum, err := db.GetNextEventNumber()
	if err != nil {
		return "", fmt.Errorf("failed to get next event number: %w", err)
	}

	return fmt.Sprintf("ev-%d-%s", eventNum, nodeID), nil
}

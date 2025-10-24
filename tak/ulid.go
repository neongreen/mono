package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateTaskID generates a task ID in the format tak-<number>-<suffix>
func GenerateTaskID(db *DB) (string, error) {
	// Get the installation suffix
	suffix, err := db.GetOrCreateInstallationSuffix()
	if err != nil {
		return "", fmt.Errorf("failed to get installation suffix: %w", err)
	}

	// Get the next task number
	taskNum, err := db.GetNextTaskNumber()
	if err != nil {
		return "", fmt.Errorf("failed to get next task number: %w", err)
	}

	return fmt.Sprintf("tak-%d-%s", taskNum, suffix), nil
}

// GenerateEventID generates an event ID (random alphanumeric string)
func GenerateEventID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 16
	b := make([]byte, length)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		b[i] = charset[idx.Int64()]
	}
	return string(b)
}

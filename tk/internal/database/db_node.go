package database

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
)

// GetOrCreateNodeID gets the node ID or creates one if it doesn't exist
func (d *DB) GetOrCreateNodeID() (string, error) {
	// Try to get existing node ID
	var nodeID string
	err := d.Db.QueryRow("SELECT value FROM metadata WHERE key = 'node_id'").Scan(&nodeID)
	if err == nil {
		return nodeID, nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to query node ID: %w", err)
	}

	nodeID, err = generateNodeID(6)
	if err != nil {
		return "", fmt.Errorf("failed to generate node ID: %w", err)
	}

	_, err = d.Db.Exec("INSERT INTO metadata (key, value) VALUES ('node_id', ?)", nodeID)
	if err != nil {
		return "", fmt.Errorf("failed to store node ID: %w", err)
	}

	return nodeID, nil
}

// RegenerateNodeID generates a new node ID and updates the metadata
func (d *DB) RegenerateNodeID() (string, error) {
	newNodeID, err := generateNodeID(6)
	if err != nil {
		return "", fmt.Errorf("failed to generate node ID: %w", err)
	}

	_, err = d.Db.Exec("UPDATE metadata SET value = ? WHERE key = 'node_id'", newNodeID)
	if err != nil {
		return "", fmt.Errorf("failed to update node ID: %w", err)
	}

	return newNodeID, nil
}

// generateNodeID generates a random alphanumeric node ID with mixed case
func generateNodeID(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}

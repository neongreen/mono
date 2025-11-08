package beads

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

// CreateProjectForImport creates a new project for beads import
// Always creates a new project (never reuses existing)
func CreateProjectForImport(db *database.DB, prefix string, alias string, actor string) (string, error) {
	// Get current node
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return "", fmt.Errorf("failed to get node ID: %w", err)
	}

	// Always create NEW project (never reuse existing)
	projectUID := types.NewProjectUID()

	payload := types.ProjectCreatedPayload{
		ProjectUID:  projectUID.String(),
		Type:        "local",
		Name:        prefix,
		Description: "Imported from beads (prefix: " + prefix + ", alias: " + alias + ")",
		CreatedBy:   actor,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal project payload: %w", err)
	}

	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return "", fmt.Errorf("failed to generate event ID: %w", err)
	}

	ts, err := db.GetNextLamportTS()
	if err != nil {
		return "", fmt.Errorf("failed to get lamport timestamp: %w", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindProjectCreated),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return "", fmt.Errorf("failed to insert project event: %w", err)
	}

	if err := db.ProjectProjectCreatedEvent(event); err != nil {
		return "", fmt.Errorf("failed to project project creation: %w", err)
	}

	// Add alias (single alias: <aliasPrefix><beadsPrefix>)
	aliasPayload := types.ProjectAliasAddPayload{
		ProjectUID: projectUID.String(),
		Alias:      alias,
		Node:       nodeID,
		AddedBy:    actor,
	}

	aliasPayloadJSON, err := json.Marshal(aliasPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal alias payload: %w", err)
	}

	aliasEventID, err := database.GenerateEventID(db)
	if err != nil {
		return "", fmt.Errorf("failed to generate alias event ID: %w", err)
	}

	aliasTS, err := db.GetNextLamportTS()
	if err != nil {
		return "", fmt.Errorf("failed to get alias lamport timestamp: %w", err)
	}

	aliasEvent := types.Event{
		ID:        aliasEventID,
		TS:        aliasTS,
		CreatedAt: time.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindProjectAliasAdd),
		Payload:   aliasPayloadJSON,
	}

	if err := db.InsertEvent(aliasEvent); err != nil {
		return "", fmt.Errorf("failed to insert alias event: %w", err)
	}

	if err := db.ProjectProjectAliasAddEvent(aliasEvent); err != nil {
		return "", fmt.Errorf("failed to project alias addition: %w", err)
	}

	return projectUID.String(), nil
}

// CheckAliasClash checks if an alias already exists for the current node
func CheckAliasClash(db *database.DB, alias string, nodeID string) (bool, error) {
	var exists int
	err := db.Db.QueryRow(`
		SELECT COUNT(*) FROM project_aliases
		WHERE alias = ? AND node = ?
	`, alias, nodeID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check alias clash: %w", err)
	}
	return exists > 0, nil
}

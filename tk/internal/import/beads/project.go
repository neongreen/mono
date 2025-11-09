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
	// Validate project name
	if err := types.ValidateProjectName(prefix); err != nil {
		return "", fmt.Errorf("invalid project name for import: %w", err)
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

	// Note: alias parameter is kept for backward compatibility but no longer used
	// (aliases have been removed from tk)

	return projectUID.String(), nil
}

// CheckAliasClash is deprecated - aliases have been removed.
// Kept for backward compatibility but always returns false.
// deprecated:v5 remove-after:v5-migration
func CheckAliasClash(db *database.DB, alias string, nodeID string) (bool, error) {
	return false, nil
}

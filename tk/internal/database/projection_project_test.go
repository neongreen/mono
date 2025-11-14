package database

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

func TestProjectProjectCreatedEvent(t *testing.T) {
	db := openTempDB(t)

	projectUID := string(types.NewProjectUID())
	payload := types.ProjectCreatedPayload{
		ProjectUID:  projectUID,
		Type:        "local",
		Name:        "test-project",
		Description: "Test description",
		CreatedBy:   "tester",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindProjectCreated),
		Payload:   payloadJSON,
	}

	if err := db.ProjectProjectCreatedEvent(event); err != nil {
		t.Fatalf("ProjectProjectCreatedEvent() error = %v", err)
	}

	// Verify project was created in database
	var count int
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM projects WHERE project_uid = ?`, projectUID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query projects: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 project, got %d", count)
	}
}

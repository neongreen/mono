package database

// Test event factory functions
//
// NOTE: This duplicates code from testutil/events.go due to import cycle issues.
// The testutil package imports database, so database tests can't import testutil.
//
// FUTURE: Consider restructuring to avoid this duplication, perhaps by:
// - Moving event factories to a separate testevents package
// - Or extending the reducer DSL to work across all test packages

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

// createProjectEvent creates a project.created event
func createProjectEvent(ts int64, projectUID types.ProjectUID, name string) types.Event {
	payload := types.ProjectCreatedPayload{
		ProjectUID:  projectUID,
		Type:        types.ProjectTypeLocal,
		Name:        name,
		Description: "Test project: " + name,
		CreatedBy:   "test",
	}
	payloadJSON, _ := json.Marshal(payload)

	return types.Event{
		ID:        types.NewEventID().String(),
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindProjectCreated),
		Payload:   payloadJSON,
	}
}

// createTaskEvent creates a task.created event
func createTaskEvent(ts int64, taskUID types.TaskUID, projectUID types.ProjectUID, title string) types.Event {
	payload := types.TaskCreatedPayload{
		TaskUID:        string(taskUID),
		ProjectUID:     string(projectUID),
		ProposedNumber: 1,
		CreatedNode:    string(types.NewNodeID()),
		Title:          title,
		CreatedBy:      "test",
	}
	payloadJSON, _ := json.Marshal(payload)

	return types.Event{
		ID:        types.NewEventID().String(),
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   payloadJSON,
	}
}

// setTaskNumberEvent creates a task.number.set event
func setTaskNumberEvent(ts int64, taskUID types.TaskUID, projectUID types.ProjectUID, number int64) types.Event {
	payload := types.TaskNumberSetPayload{
		TaskUID:    taskUID,
		ProjectUID: projectUID,
		Number:     number,
		Reason:     "test",
	}
	payloadJSON, _ := json.Marshal(payload)

	return types.Event{
		ID:        types.NewEventID().String(),
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindTaskNumberSet),
		Payload:   payloadJSON,
	}
}

// Additional helper signatures for compatibility with edge_cases_test.go

func createProjectCreatedEvent(projectUID, name, description, createdBy, nodeID string) types.Event {
	payload := types.ProjectCreatedPayload{
		ProjectUID:  types.ProjectUID(projectUID),
		Type:        types.ProjectTypeLocal,
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
	}
	payloadJSON, _ := json.Marshal(payload)

	return types.Event{
		ID:        types.NewEventID().String(),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     createdBy,
		Role:      "human",
		Kind:      string(types.EventKindProjectCreated),
		Payload:   payloadJSON,
	}
}

func createTaskCreatedEvent(taskUID, projectUID string, proposedNumber int64, nodeID, title, createdBy string) types.Event {
	payload := types.TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: proposedNumber,
		CreatedNode:    nodeID,
		Title:          title,
		CreatedBy:      createdBy,
	}
	payloadJSON, _ := json.Marshal(payload)

	return types.Event{
		ID:        types.NewEventID().String(),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     createdBy,
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   payloadJSON,
	}
}

func createTaskNumberSetEvent(taskUID, projectUID string, number int64, reason string) types.Event {
	payload := types.TaskNumberSetPayload{
		TaskUID:    types.TaskUID(taskUID),
		ProjectUID: types.ProjectUID(projectUID),
		Number:     number,
		Reason:     reason,
	}
	payloadJSON, _ := json.Marshal(payload)

	return types.Event{
		ID:        types.NewEventID().String(),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindTaskNumberSet),
		Payload:   payloadJSON,
	}
}

// createTempDB creates a temporary test database
func createTempDB(t *testing.T) *DB {
	t.Helper()
	// Inline implementation to avoid import cycle with testutil
	tempDir := t.TempDir()
	dbPath := tempDir + "/tk.db"

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open temp db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	if err := db.SetDBVersion(8); err != nil {
		t.Fatalf("failed to set version: %v", err)
	}

	return db
}

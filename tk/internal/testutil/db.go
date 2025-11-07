package testutil

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

// OpenTempDB creates a temporary test database
func OpenTempDB(t *testing.T) *database.DB {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "tk.db")

	db, err := database.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open temp db: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialise db: %v", err)
	}
	if err := db.SetDBVersion(4); err != nil {
		t.Fatalf("failed to set database version: %v", err)
	}
	if _, err := db.Db.Exec(`
		INSERT OR REPLACE INTO metadata (key, value)
		VALUES ('remote_subdir', ?)
	`, "v4"); err != nil {
		t.Fatalf("failed to set remote_subdir: %v", err)
	}
	return db
}

// SeedProject creates a test project with an alias
func SeedProject(t *testing.T, db *database.DB, alias string) string {
	t.Helper()
	projectUID := string(types.NewProjectUID())
	now := time.Now()

	projectPayload := types.ProjectCreatedPayload{
		ProjectUID:  projectUID,
		Type:        "local",
		Name:        alias,
		Description: alias + " project",
		CreatedBy:   "tester",
	}

	payloadJSON := MustJSON(t, projectPayload)
	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: now,
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindProjectCreated),
		Payload:   payloadJSON,
	}
	if err := db.InsertEvent(event); err != nil {
		t.Fatalf("failed to insert project.created: %v", err)
	}
	if err := db.ProjectProjectCreatedEvent(event); err != nil {
		t.Fatalf("failed to project project.created: %v", err)
	}

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node id: %v", err)
	}

	aliasPayload := types.ProjectAliasAddPayload{
		ProjectUID: projectUID,
		Alias:      alias,
		Node:       nodeID,
		AddedBy:    "tester",
	}
	aliasJSON := MustJSON(t, aliasPayload)
	aliasEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: now,
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindProjectAliasAdd),
		Payload:   aliasJSON,
	}
	if err := db.InsertEvent(aliasEvent); err != nil {
		t.Fatalf("failed to insert project.alias.add: %v", err)
	}
	if err := db.ProjectProjectAliasAddEvent(aliasEvent); err != nil {
		t.Fatalf("failed to project project.alias.add: %v", err)
	}

	return projectUID
}

// SeedTask creates a test task with the default node ID
func SeedTask(t *testing.T, db *database.DB, projectUID string, title string, number int64) string {
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node id: %v", err)
	}
	return SeedTaskWithNode(t, db, projectUID, title, number, nodeID)
}

// SeedTaskWithNode creates a test task with a specific node ID
func SeedTaskWithNode(t *testing.T, db *database.DB, projectUID string, title string, number int64, nodeID string) string {
	t.Helper()
	taskUID := string(types.NewTaskUID())
	now := time.Now()

	taskPayload := types.TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: number,
		CreatedNode:    nodeID,
		Title:          title,
		CreatedBy:      "tester",
	}

	taskJSON := MustJSON(t, taskPayload)
	taskEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: now,
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   taskJSON,
	}
	if err := db.InsertEvent(taskEvent); err != nil {
		t.Fatalf("failed to insert task.created: %v", err)
	}
	if err := db.ProjectTaskCreatedEvent(taskEvent); err != nil {
		t.Fatalf("failed to project task.created: %v", err)
	}

	numberPayload := types.TaskNumberSetPayload{
		TaskUID:    taskUID,
		ProjectUID: projectUID,
		Number:     number,
		Reason:     "seed",
	}
	numberJSON := MustJSON(t, numberPayload)
	numberEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: now,
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindTaskNumberSet),
		Payload:   numberJSON,
	}
	if err := db.InsertEvent(numberEvent); err != nil {
		t.Fatalf("failed to insert task.number.set: %v", err)
	}
	if err := db.ProjectTaskNumberSetEvent(numberEvent); err != nil {
		t.Fatalf("failed to project task.number.set: %v", err)
	}

	return taskUID
}

// MustJSON marshals a value to JSON or fails the test
func MustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	return json.RawMessage(data)
}

// MarshalPayload marshals a value to JSON without test context
func MarshalPayload(v any) json.RawMessage {
	data, _ := json.Marshal(v)
	return json.RawMessage(data)
}

package database

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

func openTempDB(t *testing.T) *DB {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "tk.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open temp db: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialise db: %v", err)
	}
	if err := db.SetDBVersion(8); err != nil {
		t.Fatalf("failed to set database version: %v", err)
	}
	if _, err := db.Db.Exec(`
		INSERT OR REPLACE INTO metadata (key, value)
		VALUES ('remote_subdir', ?)
	`, "v4"); err != nil {
		t.Fatalf("failed to set remote_subdir: %v", err)
	}

	// Run migrations to get to latest schema
	if err := db.RunMigrationsIfNeeded(); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

func seedProject(t *testing.T, db *DB, name string) string {
	t.Helper()
	projectUID := string(types.NewProjectUID())
	now := time.Now()

	projectPayload := types.ProjectCreatedPayload{
		ProjectUID:  types.ProjectUID(projectUID),
		Type:        types.ProjectTypeLocal,
		Name:        name,
		Description: name + " project",
		CreatedBy:   "tester",
	}

	payloadJSON := mustJSON(t, projectPayload)
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
	if err := db.RebuildProjections(); err != nil {
		t.Fatalf("failed to project project.created: %v", err)
	}

	return projectUID
}

func seedProjectWithoutAlias(t *testing.T, db *DB, name string) string {
	t.Helper()
	projectUID := string(types.NewProjectUID())
	now := time.Now()

	projectPayload := types.ProjectCreatedPayload{
		ProjectUID:  types.ProjectUID(projectUID),
		Type:        types.ProjectTypeLocal,
		Name:        name,
		Description: name + " project",
		CreatedBy:   "tester",
	}

	payloadJSON := mustJSON(t, projectPayload)
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
	if err := db.RebuildProjections(); err != nil {
		t.Fatalf("failed to project project.created: %v", err)
	}

	// No alias event is created, so the project has no alias
	return projectUID
}

func seedTask(t *testing.T, db *DB, projectUID string, title string, number int64) string {
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node id: %v", err)
	}
	return seedTaskWithNode(t, db, projectUID, title, number, nodeID)
}

func seedTaskWithNode(t *testing.T, db *DB, projectUID string, title string, number int64, nodeID string) string {
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

	taskJSON := mustJSON(t, taskPayload)
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
	if err := db.RebuildProjections(); err != nil {
		t.Fatalf("failed to project task.created: %v", err)
	}

	numberPayload := types.TaskNumberSetPayload{
		TaskUID:    types.TaskUID(taskUID),
		ProjectUID: types.ProjectUID(projectUID),
		Number:     number,
		Reason:     "seed",
	}
	numberJSON := mustJSON(t, numberPayload)
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
	if err := db.RebuildProjections(); err != nil {
		t.Fatalf("failed to project task.number.set: %v", err)
	}

	return taskUID
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	return json.RawMessage(data)
}

// Helper functions for creating events without testing context

// createProjectAliasAddEvent removed - aliases no longer supported


func seedContainerKindAndInstance(t *testing.T, db *DB, kindName string, primitive types.ContainerPrimitive, containerID string, containerName string) {
	t.Helper()

	// Define kind
	definePayload := types.DefineContainerKindPayload{
		Name:        kindName,
		Primitive:   primitive,
		Description: "Test container",
		CreatedBy:   "tester",
	}
	definePayloadJSON, _ := json.Marshal(definePayload)
	defineEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerKindDefine),
		Payload:   definePayloadJSON,
	}
	if err := db.InsertEvent(defineEvent); err != nil {
		t.Fatalf("failed to insert define event: %v", err)
	}

	// Create container
	createPayload := types.CreateContainerPayload{
		ID:        containerID,
		Primitive: primitive,
		Kind:      kindName,
		Name:      containerName,
		CreatedBy: "tester",
	}
	createPayloadJSON, _ := json.Marshal(createPayload)
	createEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerCreate),
		Payload:   createPayloadJSON,
	}
	if err := db.InsertEvent(createEvent); err != nil {
		t.Fatalf("failed to insert create event: %v", err)
	}

	// Rebuild projections from events
	if err := db.RebuildProjections(); err != nil {
		t.Fatalf("failed to rebuild: %v", err)
	}
}

func seedQueueItems(t *testing.T, db *DB, containerID string, itemIDs []string) {
	t.Helper()

	for i, itemID := range itemIDs {
		payload := types.QueuePushPayload{
			ContainerID: containerID,
			ItemID:      types.TaskUID(itemID),
		}
		payloadJSON, _ := json.Marshal(payload)
		event := types.Event{
			ID:        string(types.NewEventID()),
			TS:        int64(i),
			CreatedAt: time.Now(),
			Actor:     "tester",
			Role:      "human",
			Kind:      string(types.EventKindQueuePush),
			Payload:   payloadJSON,
		}
		if err := db.InsertEvent(event); err != nil {
			t.Fatalf("failed to insert queue push event: %v", err)
		}
	}

	// Rebuild projections after all events inserted
	if err := db.RebuildProjections(); err != nil {
		t.Fatalf("failed to rebuild projections: %v", err)
	}
}

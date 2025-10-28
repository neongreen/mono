package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

// TestV4MigrationWithUnseenTask tests that migration can handle task.status.set
// events for tasks that haven't been encountered yet in the migration process.
// This can happen if events are out of order or if a task.created event comes
// after other events referencing the same task.
func TestV4MigrationWithUnseenTask(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Initialize v1/v2 database
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// Create a prefix
	if err := db.CreatePrefix("tak", "Test tasks", "alice"); err != nil {
		t.Fatalf("failed to create prefix: %v", err)
	}

	// Create task ID
	taskID := "tak-1-16uq1v"
	taskUUID := string(NewTaskUID())

	// Insert task.status.set event BEFORE task.created event
	// This simulates the problematic scenario
	statusPayload := TaskStatusSetPayload{
		TaskUUID: "",
		TaskID:   taskID,
		Axis:     "completion",
		State:    "done",
	}
	statusPayloadJSON, err := json.Marshal(statusPayload)
	if err != nil {
		t.Fatalf("failed to marshal status payload: %v", err)
	}

	statusEvent := Event{
		ID:        string(NewEventID()),
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(EventKindTaskStatusSet),
		Payload:   statusPayloadJSON,
	}

	if err := db.InsertEvent(statusEvent); err != nil {
		t.Fatalf("failed to insert status event: %v", err)
	}

	// Now insert task.created event AFTER the status event
	taskPayload := TaskCreatedPayload{
		TaskUUID:  taskUUID,
		TaskID:    taskID,
		Title:     "Test task",
		CreatedBy: "alice",
	}
	taskPayloadJSON, err := json.Marshal(taskPayload)
	if err != nil {
		t.Fatalf("failed to marshal task payload: %v", err)
	}

	taskEvent := Event{
		ID:        string(NewEventID()),
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(EventKindTaskCreated),
		Payload:   taskPayloadJSON,
	}

	if err := db.InsertEvent(taskEvent); err != nil {
		t.Fatalf("failed to insert task event: %v", err)
	}

	db.Close()

	// Now try to migrate - this should not fail
	db, err = OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to reinit: %v", err)
	}

	// This should succeed without "unknown task reference" error
	if err := db.MigrateToV4(dbPath); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Verify the task was migrated
	var count int
	err = db.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count tasks: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 task after migration, got %d", count)
	}
}

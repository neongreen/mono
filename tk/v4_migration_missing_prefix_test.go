package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

// TestV4MigrationWithMissingPrefix tests that migration handles tasks with
// prefixes that don't exist in the prefixes table (e.g., removed prefixes).
// This reproduces the error: "no project found for prefix tak"
func TestV4MigrationWithMissingPrefix(t *testing.T) {
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

	// Create a task with prefix "tak" but DON'T create the prefix in the prefixes table
	// This simulates a scenario where:
	// - A prefix was removed (removed = 1)
	// - A prefix was never created but a task exists
	taskID := "tak-1-16uq1v"
	taskUUID := string(NewTaskUID())

	taskPayload := TaskCreatedPayload{
		TaskUUID:  taskUUID,
		TaskID:    taskID,
		Title:     "Test task with missing prefix",
		CreatedBy: "alice",
	}
	taskPayloadJSON, err := json.Marshal(taskPayload)
	if err != nil {
		t.Fatalf("failed to marshal task payload: %v", err)
	}

	taskEvent := Event{
		ID:        string(NewEventID()),
		TS:        1,
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

	// Now try to migrate - this should NOT fail even though prefix "tak" doesn't exist
	db, err = OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to reinit: %v", err)
	}

	// This should succeed by creating a project on-demand for the missing prefix
	if err := db.MigrateToV4(dbPath); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Verify the task was migrated
	var taskCount int
	err = db.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&taskCount)
	if err != nil {
		t.Fatalf("failed to count tasks: %v", err)
	}
	if taskCount != 1 {
		t.Errorf("expected 1 task after migration, got %d", taskCount)
	}

	// Verify a project was created for the missing prefix
	// Note: there may be 2 projects - one for "tak" and one for the default "tk" prefix
	var projectCount int
	err = db.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&projectCount)
	if err != nil {
		t.Fatalf("failed to count projects: %v", err)
	}
	if projectCount < 1 {
		t.Errorf("expected at least 1 project after migration, got %d", projectCount)
	}

	// Verify the project has the correct alias for "tak"
	var alias string
	err = db.db.QueryRow("SELECT alias FROM project_aliases WHERE alias = 'tak'").Scan(&alias)
	if err != nil {
		t.Fatalf("failed to get project alias for 'tak': %v", err)
	}
	if alias != "tak" {
		t.Errorf("expected alias 'tak', got %s", alias)
	}
}

// TestV4MigrationWithRemovedPrefix tests that migration handles tasks with
// prefixes that exist but are marked as removed.
func TestV4MigrationWithRemovedPrefix(t *testing.T) {
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

	// Create a prefix and then remove it
	if err := db.CreatePrefix("removed", "Removed prefix", "alice"); err != nil {
		t.Fatalf("failed to create prefix: %v", err)
	}

	// Manually mark the prefix as removed
	_, err = db.db.Exec("UPDATE prefixes SET removed = 1 WHERE prefix = 'removed'")
	if err != nil {
		t.Fatalf("failed to mark prefix as removed: %v", err)
	}

	// Create a task with the removed prefix
	taskID := "removed-42-node1"
	taskUUID := string(NewTaskUID())

	taskPayload := TaskCreatedPayload{
		TaskUUID:  taskUUID,
		TaskID:    taskID,
		Title:     "Task with removed prefix",
		CreatedBy: "alice",
	}
	taskPayloadJSON, err := json.Marshal(taskPayload)
	if err != nil {
		t.Fatalf("failed to marshal task payload: %v", err)
	}

	taskEvent := Event{
		ID:        string(NewEventID()),
		TS:        1,
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

	// Now try to migrate
	db, err = OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to reinit: %v", err)
	}

	// This should succeed by creating a project on-demand for the removed prefix
	if err := db.MigrateToV4(dbPath); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Verify the task was migrated
	var taskCount int
	err = db.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&taskCount)
	if err != nil {
		t.Fatalf("failed to count tasks: %v", err)
	}
	if taskCount != 1 {
		t.Errorf("expected 1 task after migration, got %d", taskCount)
	}

	// Verify a project was created for the removed prefix
	var projectCount int
	err = db.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&projectCount)
	if err != nil {
		t.Fatalf("failed to count projects: %v", err)
	}
	if projectCount < 1 {
		t.Errorf("expected at least 1 project after migration, got %d", projectCount)
	}

	// Verify the project has the correct alias
	var alias string
	err = db.db.QueryRow("SELECT alias FROM project_aliases WHERE alias = 'removed'").Scan(&alias)
	if err != nil {
		t.Fatalf("failed to get project alias: %v", err)
	}
	if alias != "removed" {
		t.Errorf("expected alias 'removed', got %s", alias)
	}
}

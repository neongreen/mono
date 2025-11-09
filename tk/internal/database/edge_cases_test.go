package database

import (
	"path/filepath"
	"testing"

	"github.com/neongreen/mono/tk/internal/types"
)

// TestEventProjectionIdempotency tests that projecting the same event twice is safe
func TestEventProjectionIdempotency(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	if err := db.CreateProjectTables(); err != nil {
		t.Fatalf("failed to create project tables: %v", err)
	}
	if err := db.SetDBVersion(4); err != nil {
		t.Fatalf("failed to set version: %v", err)
	}

	nodeA, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node: %v", err)
	}

	// Create a project event
	projectUID := string(types.NewProjectUID())
	projectEvent := createProjectCreatedEvent(projectUID, "Test Project", "A test", "alice", nodeA)

	// Insert and project once
	if err := db.InsertEvent(projectEvent); err != nil {
		t.Fatalf("failed to insert project event: %v", err)
	}
	if err := db.ProjectProjectCreatedEvent(projectEvent); err != nil {
		t.Fatalf("failed to project project first time: %v", err)
	}

	// Project again (should be idempotent - no error, no duplicate)
	if err := db.ProjectProjectCreatedEvent(projectEvent); err != nil {
		t.Fatalf("failed to project project second time: %v", err)
	}

	// Verify only one project exists
	var count int
	err = db.Db.QueryRow("SELECT COUNT(*) FROM projects WHERE project_uid = ?", projectUID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query projects: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 project after double projection, got %d", count)
	}
}

// TestTaskNumberCollisionHandling tests collision display logic
func TestTaskNumberCollisionHandling(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	if err := db.CreateProjectTables(); err != nil {
		t.Fatalf("failed to create project tables: %v", err)
	}
	if err := db.SetDBVersion(4); err != nil {
		t.Fatalf("failed to set version: %v", err)
	}

	nodeA, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node: %v", err)
	}

	// Create a project
	projectUID := string(types.NewProjectUID())
	projectEvent := createProjectCreatedEvent(projectUID, "Test", "Test", "alice", nodeA)
	if err := db.InsertEvent(projectEvent); err != nil {
		t.Fatalf("failed to insert project: %v", err)
	}
	if err := db.ProjectProjectCreatedEvent(projectEvent); err != nil {
		t.Fatalf("failed to project project: %v", err)
	}

	// Note: alias creation removed (tk-246)

	// Create two tasks with the same number (collision)
	task1UID := string(types.NewTaskUID())
	task1Event := createTaskCreatedEvent(task1UID, projectUID, 1, nodeA, "Task 1", "alice")
	if err := db.InsertEvent(task1Event); err != nil {
		t.Fatalf("failed to insert task1: %v", err)
	}
	if err := db.ProjectTaskCreatedEvent(task1Event); err != nil {
		t.Fatalf("failed to project task1: %v", err)
	}

	number1Event := createTaskNumberSetEvent(task1UID, projectUID, 1, "initial")
	if err := db.InsertEvent(number1Event); err != nil {
		t.Fatalf("failed to insert number1: %v", err)
	}
	if err := db.ProjectTaskNumberSetEvent(number1Event); err != nil {
		t.Fatalf("failed to project number1: %v", err)
	}

	// Simulate a different node creating another task with number 1
	nodeB := "DifferentNode"
	task2UID := string(types.NewTaskUID())
	task2Event := createTaskCreatedEvent(task2UID, projectUID, 1, nodeB, "Task 2", "bob")
	if err := db.InsertEvent(task2Event); err != nil {
		t.Fatalf("failed to insert task2: %v", err)
	}
	if err := db.ProjectTaskCreatedEvent(task2Event); err != nil {
		t.Fatalf("failed to project task2: %v", err)
	}

	number2Event := createTaskNumberSetEvent(task2UID, projectUID, 1, "initial")
	if err := db.InsertEvent(number2Event); err != nil {
		t.Fatalf("failed to insert number2: %v", err)
	}
	if err := db.ProjectTaskNumberSetEvent(number2Event); err != nil {
		t.Fatalf("failed to project number2: %v", err)
	}

	// Verify both tasks exist with number 1
	var count int
	err = db.Db.QueryRow("SELECT COUNT(*) FROM task_numbers WHERE project_uid = ? AND number = 1", projectUID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count task numbers: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 tasks with number 1, got %d", count)
	}

	// Test display ID rendering for both tasks
	displayID1, err := RenderTaskDisplayID(db, task1UID)
	if err != nil {
		t.Fatalf("failed to render display ID 1: %v", err)
	}

	displayID2, err := RenderTaskDisplayID(db, task2UID)
	if err != nil {
		t.Fatalf("failed to render display ID 2: %v", err)
	}

	// Both should include node hints due to collision
	if displayID1 == "test-1" || displayID2 == "test-1" {
		t.Errorf("expected disambiguated IDs, got: %s, %s", displayID1, displayID2)
	}

	// They should be different
	if displayID1 == displayID2 {
		t.Errorf("display IDs should be different for colliding tasks: %s", displayID1)
	}
}

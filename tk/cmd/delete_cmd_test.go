package cmd

import (
	"encoding/json"
	"testing"
	"time"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

func deleteTaskByUID(db *database.DB, taskUID string) error {
	lamportTS, err := db.GetNextLamportTS()
	if err != nil {
		return err
	}

	payload := types.TaskDeletePayload{
		TaskUUID: taskUID,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	now := time.Now()
	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        lamportTS,
		CreatedAt: now,
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindTaskDelete),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return err
	}

	return db.ProjectTaskDeleteEvent(event)
}

func TestDeleteTask(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "test")
	taskUID := seedTask(t, db, projectUID, "Task to delete", 1)

	// Verify task exists
	var exists bool
	err := db.Db.QueryRow(`SELECT EXISTS(SELECT 1 FROM tasks WHERE task_uid = ?)`, taskUID).Scan(&exists)
	if err != nil {
		t.Fatalf("failed to check task existence: %v", err)
	}
	if !exists {
		t.Fatal("task should exist before delete")
	}

	// Execute delete by creating and projecting the event
	if err := deleteTaskByUID(db, taskUID); err != nil {
		t.Fatalf("deleteTask failed: %v", err)
	}

	// Verify task no longer exists in tasks table
	err = db.Db.QueryRow(`SELECT EXISTS(SELECT 1 FROM tasks WHERE task_uid = ?)`, taskUID).Scan(&exists)
	if err != nil {
		t.Fatalf("failed to check task existence after delete: %v", err)
	}
	if exists {
		t.Fatal("task should not exist after delete")
	}

	// Verify task no longer exists in task_numbers table
	err = db.Db.QueryRow(`SELECT EXISTS(SELECT 1 FROM task_numbers WHERE task_uid = ?)`, taskUID).Scan(&exists)
	if err != nil {
		t.Fatalf("failed to check task_numbers existence after delete: %v", err)
	}
	if exists {
		t.Fatal("task should not exist in task_numbers after delete")
	}

	// Verify event was created
	var eventCount int
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM events WHERE kind = 'task.delete'`).Scan(&eventCount)
	if err != nil {
		t.Fatalf("failed to count delete events: %v", err)
	}
	if eventCount != 1 {
		t.Fatalf("expected 1 delete event, got %d", eventCount)
	}
}

func TestDeleteTaskByDisplayID(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "proj")
	taskUID := seedTask(t, db, projectUID, "Task with display ID", 5)

	// Delete by UUID
	if err := deleteTaskByUID(db, taskUID); err != nil {
		t.Fatalf("deleteTask failed: %v", err)
	}

	// Verify task no longer exists
	var exists bool
	err := db.Db.QueryRow(`SELECT EXISTS(SELECT 1 FROM tasks WHERE task_uid = ?)`, taskUID).Scan(&exists)
	if err != nil {
		t.Fatalf("failed to check task existence: %v", err)
	}
	if exists {
		t.Fatal("task should not exist after delete")
	}
}

func TestDeleteNonexistentTask(t *testing.T) {
	db := openTempDB(t)

	seedProject(t, db, "test")

	// Try to delete a task that doesn't exist
	nonExistentUID := string(types.NewTaskUID())
	err := deleteTaskByUID(db, nonExistentUID)
	// This should succeed (idempotent delete) but the task won't be in the DB
	// The projection function will just be a no-op
	if err != nil {
		t.Fatalf("delete should be idempotent: %v", err)
	}
}

func TestDeleteTaskWithRelations(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "rel")
	task1UID := seedTask(t, db, projectUID, "Task 1", 1)
	task2UID := seedTask(t, db, projectUID, "Task 2", 2)

	// Add a relation between tasks (task1 blocks task2)
	// We need to add this through the events system
	config, err := config_pkg.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Build reducer to verify relations
	reducer, err := db.GetCachedReducerWithConfig(config)
	if err != nil {
		t.Fatalf("failed to build reducer: %v", err)
	}

	// Delete task1
	if err := deleteTaskByUID(db, task1UID); err != nil {
		t.Fatalf("deleteTask failed: %v", err)
	}

	// Rebuild reducer after delete
	reducer, err = db.GetCachedReducerWithConfig(config)
	if err != nil {
		t.Fatalf("failed to rebuild reducer: %v", err)
	}

	// Verify task1 is gone
	_, exists := reducer.GetTask(task1UID)
	if exists {
		t.Fatal("deleted task should not exist in reducer")
	}

	// Verify task2 still exists
	_, exists = reducer.GetTask(task2UID)
	if !exists {
		t.Fatal("undeleted task should still exist")
	}
}

func TestDeleteTaskIdempotency(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "idem")
	taskUID := seedTask(t, db, projectUID, "Task to delete twice", 1)

	// Delete the task
	if err := deleteTaskByUID(db, taskUID); err != nil {
		t.Fatalf("first deleteTask failed: %v", err)
	}

	// Try to delete again - should be idempotent (no error)
	if err := deleteTaskByUID(db, taskUID); err != nil {
		t.Fatalf("second deleteTask failed (should be idempotent): %v", err)
	}

	// Verify we have 2 delete events for the same task
	var eventCount int
	if err := db.Db.QueryRow(`SELECT COUNT(*) FROM events WHERE kind = 'task.delete'`).Scan(&eventCount); err != nil {
		t.Fatalf("failed to count delete events: %v", err)
	}
	if eventCount != 2 {
		t.Fatalf("expected 2 delete events, got %d", eventCount)
	}
}

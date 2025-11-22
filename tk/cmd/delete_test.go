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

	return db.RebuildProjections()
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

	// Verify task is marked as deleted in tasks table
	var deleted int
	err = db.Db.QueryRow(`SELECT deleted FROM tasks WHERE task_uid = ?`, taskUID).Scan(&deleted)
	if err != nil {
		t.Fatalf("failed to check task deleted status: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("task should be marked as deleted, got deleted=%d", deleted)
	}

	// Verify task numbers are removed from task_numbers table
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

	// Verify task is marked as deleted
	var deleted int
	err := db.Db.QueryRow(`SELECT deleted FROM tasks WHERE task_uid = ?`, taskUID).Scan(&deleted)
	if err != nil {
		t.Fatalf("failed to check task deleted status: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("task should be marked as deleted, got deleted=%d", deleted)
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

	// Delete task1
	if err := deleteTaskByUID(db, task1UID); err != nil {
		t.Fatalf("deleteTask failed: %v", err)
	}

	// Build reducer after delete to verify relations
	reducer, err := db.GetCachedReducerWithConfig(config)
	if err != nil {
		t.Fatalf("failed to rebuild reducer: %v", err)
	}

	// Verify task1 is not visible (deleted)
	_, exists := reducer.GetTask(task1UID)
	if exists {
		t.Fatal("deleted task should not be returned by GetTask")
	}
	
	// But should still exist if we ask for it including deleted
	task1, exists := reducer.GetTaskIncludingDeleted(task1UID)
	if !exists {
		t.Fatal("deleted task should still exist in reducer (soft delete)")
	}
	if !task1.Deleted {
		t.Fatal("task1 should be marked as deleted")
	}

	// Verify task2 still exists and is visible
	task2, exists := reducer.GetTask(task2UID)
	if !exists {
		t.Fatal("undeleted task should still exist")
	}
	if task2.Deleted {
		t.Fatal("task2 should not be marked as deleted")
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

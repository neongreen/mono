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

	// Verify the status was preserved
	var eventCount int
	err = db.db.QueryRow("SELECT COUNT(*) FROM events WHERE kind = ?", string(EventKindTaskStatusSet)).Scan(&eventCount)
	if err != nil {
		t.Fatalf("failed to count status events: %v", err)
	}
	if eventCount < 1 {
		t.Errorf("expected at least 1 status event after migration, got %d", eventCount)
	}
}

// TestV4MigrationWithMultipleUnseenTasks tests migration with multiple tasks
// where events reference tasks before they are created
func TestV4MigrationWithMultipleUnseenTasks(t *testing.T) {
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
	if err := db.CreatePrefix("test", "Test tasks", "alice"); err != nil {
		t.Fatalf("failed to create prefix: %v", err)
	}

	// Create multiple tasks with events out of order
	tasks := []struct {
		id   string
		uuid string
	}{
		{"test-1-abc", string(NewTaskUID())},
		{"test-2-abc", string(NewTaskUID())},
		{"test-3-abc", string(NewTaskUID())},
	}

	ts := int64(1)

	// Insert status events for all tasks BEFORE their creation events
	for _, task := range tasks {
		statusPayload := TaskStatusSetPayload{
			TaskID: task.id,
			Axis:   "completion",
			State:  "todo",
		}
		statusPayloadJSON, err := json.Marshal(statusPayload)
		if err != nil {
			t.Fatalf("failed to marshal status payload: %v", err)
		}

		statusEvent := Event{
			ID:        string(NewEventID()),
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     "alice",
			Role:      "human",
			Kind:      string(EventKindTaskStatusSet),
			Payload:   statusPayloadJSON,
		}
		ts++

		if err := db.InsertEvent(statusEvent); err != nil {
			t.Fatalf("failed to insert status event: %v", err)
		}
	}

	// Now insert creation events
	for _, task := range tasks {
		taskPayload := TaskCreatedPayload{
			TaskUUID:  task.uuid,
			TaskID:    task.id,
			Title:     "Task " + task.id,
			CreatedBy: "alice",
		}
		taskPayloadJSON, err := json.Marshal(taskPayload)
		if err != nil {
			t.Fatalf("failed to marshal task payload: %v", err)
		}

		taskEvent := Event{
			ID:        string(NewEventID()),
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     "alice",
			Role:      "human",
			Kind:      string(EventKindTaskCreated),
			Payload:   taskPayloadJSON,
		}
		ts++

		if err := db.InsertEvent(taskEvent); err != nil {
			t.Fatalf("failed to insert task event: %v", err)
		}
	}

	db.Close()

	// Migrate to v4
	db, err = OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to reinit: %v", err)
	}

	if err := db.MigrateToV4(dbPath); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Verify all tasks were migrated
	var count int
	err = db.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count tasks: %v", err)
	}
	if count != len(tasks) {
		t.Errorf("expected %d tasks after migration, got %d", len(tasks), count)
	}
}

// TestV4MigrationWithRelationBeforeTask tests migration when a relation.add
// event references tasks that haven't been seen yet
func TestV4MigrationWithRelationBeforeTask(t *testing.T) {
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
	if err := db.CreatePrefix("test", "Test tasks", "alice"); err != nil {
		t.Fatalf("failed to create prefix: %v", err)
	}

	// Create two tasks
	task1ID := "test-1-abc"
	task1UUID := string(NewTaskUID())
	task2ID := "test-2-abc"
	task2UUID := string(NewTaskUID())

	ts := int64(1)

	// Insert relation.add event BEFORE task creation events
	relationPayload := RelationAddPayload{
		Src:  task1UUID,
		Type: "blocks",
		Dst:  task2UUID,
		Note: "",
	}
	relationPayloadJSON, err := json.Marshal(relationPayload)
	if err != nil {
		t.Fatalf("failed to marshal relation payload: %v", err)
	}

	relationEvent := Event{
		ID:        string(NewEventID()),
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(EventKindRelationAdd),
		Payload:   relationPayloadJSON,
	}
	ts++

	if err := db.InsertEvent(relationEvent); err != nil {
		t.Fatalf("failed to insert relation event: %v", err)
	}

	// Now insert task creation events
	task1Payload := TaskCreatedPayload{
		TaskUUID:  task1UUID,
		TaskID:    task1ID,
		Title:     "Task 1",
		CreatedBy: "alice",
	}
	task1PayloadJSON, err := json.Marshal(task1Payload)
	if err != nil {
		t.Fatalf("failed to marshal task1 payload: %v", err)
	}

	task1Event := Event{
		ID:        string(NewEventID()),
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(EventKindTaskCreated),
		Payload:   task1PayloadJSON,
	}
	ts++

	if err := db.InsertEvent(task1Event); err != nil {
		t.Fatalf("failed to insert task1 event: %v", err)
	}

	task2Payload := TaskCreatedPayload{
		TaskUUID:  task2UUID,
		TaskID:    task2ID,
		Title:     "Task 2",
		CreatedBy: "alice",
	}
	task2PayloadJSON, err := json.Marshal(task2Payload)
	if err != nil {
		t.Fatalf("failed to marshal task2 payload: %v", err)
	}

	task2Event := Event{
		ID:        string(NewEventID()),
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(EventKindTaskCreated),
		Payload:   task2PayloadJSON,
	}
	ts++

	if err := db.InsertEvent(task2Event); err != nil {
		t.Fatalf("failed to insert task2 event: %v", err)
	}

	db.Close()

	// Migrate to v4
	db, err = OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to reinit: %v", err)
	}

	if err := db.MigrateToV4(dbPath); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Verify both tasks were migrated
	var count int
	err = db.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count tasks: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 tasks after migration, got %d", count)
	}
}

// TestV4MigrationWithTaskIDInUUIDField tests migration when TaskUUID field
// contains a task ID instead of a UUID (legacy data issue)
// This reproduces the error: unknown task reference (uuid="tk-30-wiWhKW" id="tk-30")
func TestV4MigrationWithTaskIDInUUIDField(t *testing.T) {
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
	if err := db.CreatePrefix("test", "Test tasks", "alice"); err != nil {
		t.Fatalf("failed to create prefix: %v", err)
	}

	// Create a task where TaskUUID contains a task ID (legacy format)
	taskID := "test-30-wiWhKW"

	// Insert task.created event with task ID in TaskUUID field (legacy format)
	taskPayload := TaskCreatedPayload{
		TaskUUID:  taskID, // Using task ID in UUID field (legacy format)
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

	// Insert task.status.set event that references the task
	statusPayload := TaskStatusSetPayload{
		TaskUUID: taskID, // Referencing by the task ID in UUID field
		TaskID:   "test-30",
		Axis:     "completion",
		State:    "done",
	}
	statusPayloadJSON, err := json.Marshal(statusPayload)
	if err != nil {
		t.Fatalf("failed to marshal status payload: %v", err)
	}

	statusEvent := Event{
		ID:        string(NewEventID()),
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(EventKindTaskStatusSet),
		Payload:   statusPayloadJSON,
	}

	if err := db.InsertEvent(statusEvent); err != nil {
		t.Fatalf("failed to insert status event: %v", err)
	}

	db.Close()

	// Now try to migrate - this should NOT fail
	db, err = OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to reinit: %v", err)
	}

	// This should succeed even though TaskUUID contains a task ID
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

	// Verify the status event was migrated
	var statusCount int
	err = db.db.QueryRow("SELECT COUNT(*) FROM events WHERE kind = ?", string(EventKindTaskStatusSet)).Scan(&statusCount)
	if err != nil {
		t.Fatalf("failed to count status events: %v", err)
	}
	if statusCount < 1 {
		t.Errorf("expected at least 1 status event after migration, got %d", statusCount)
	}
}

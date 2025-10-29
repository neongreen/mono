package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestV4TypeValidation tests v4 type validation
func TestV4TypeValidation(t *testing.T) {
	// Test ProjectUID
	t.Run("ProjectUID validation", func(t *testing.T) {
		validUID := NewProjectUID()
		if err := validUID.Validate(); err != nil {
			t.Errorf("valid ProjectUID failed validation: %v", err)
		}

		invalidUID := ProjectUID("invalid")
		if err := invalidUID.Validate(); err == nil {
			t.Error("invalid ProjectUID passed validation")
		}
	})

	// Test TaskUID
	t.Run("TaskUID validation", func(t *testing.T) {
		validUID := NewTaskUID()
		if err := validUID.Validate(); err != nil {
			t.Errorf("valid TaskUID failed validation: %v", err)
		}

		invalidUID := TaskUID("invalid")
		if err := invalidUID.Validate(); err == nil {
			t.Error("invalid TaskUID passed validation")
		}
	})

	// Test Alias
	t.Run("Alias validation", func(t *testing.T) {
		validAlias := Alias("tk")
		if err := validAlias.Validate(); err != nil {
			t.Errorf("valid alias failed validation: %v", err)
		}

		tooShort := Alias("x")
		if err := tooShort.Validate(); err == nil {
			t.Error("too short alias passed validation")
		}

		tooLong := Alias("this-is-way-too-long-for-an-alias")
		if err := tooLong.Validate(); err == nil {
			t.Error("too long alias passed validation")
		}
	})

	// Test TaskNumber
	t.Run("TaskNumber validation", func(t *testing.T) {
		validNum := TaskNumber(1)
		if err := validNum.Validate(); err != nil {
			t.Errorf("valid TaskNumber failed validation: %v", err)
		}

		invalidNum := TaskNumber(0)
		if err := invalidNum.Validate(); err == nil {
			t.Error("zero TaskNumber passed validation")
		}

		negativeNum := TaskNumber(-1)
		if err := negativeNum.Validate(); err == nil {
			t.Error("negative TaskNumber passed validation")
		}
	})
}

// TestV4Migration tests the migration from v1/v2 to v4
func TestV4Migration(t *testing.T) {
	// Create a temporary database
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

	// Create a prefix (v1/v2 style)
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node ID: %v", err)
	}

	_, err = db.db.Exec(`
		INSERT INTO prefixes (prefix, node, description, created_at, created_by, removed)
		VALUES (?, ?, ?, ?, ?, 0)
	`, "test", nodeID, "Test prefix", time.Now().Unix(), "testuser")
	if err != nil {
		t.Fatalf("failed to create prefix: %v", err)
	}

	// Create a task (v1/v2 style)
	taskUUID, err := GenerateTaskUUID()
	if err != nil {
		t.Fatalf("failed to generate task UUID: %v", err)
	}
	taskID := "test-1-" + nodeID

	payload := TaskCreatedPayload{
		TaskUUID:  taskUUID,
		TaskID:    taskID,
		Title:     "Test task",
		CreatedBy: "testuser",
	}
	payloadJSON, _ := json.Marshal(payload)

	event := Event{
		ID:        "ev-1-" + nodeID,
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "testuser",
		Role:      "human",
		Kind:      "task.created",
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	db.Close()

	// Reopen and trigger migration
	db, err = OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to reinitialize database: %v", err)
	}

	// Check if migration is needed
	needsMigration, err := db.NeedsMigrationToV4()
	if err != nil {
		t.Fatalf("failed to check migration status: %v", err)
	}

	if !needsMigration {
		t.Fatal("expected database to need migration")
	}

	// Perform migration
	if err := db.MigrateToV4(dbPath); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Check version after migration
	version, err := db.GetDBVersion()
	if err != nil {
		t.Fatalf("failed to get version: %v", err)
	}

	if version != v4SpecVersion {
		t.Errorf("expected version %d, got %d", v4SpecVersion, version)
	}

	// Verify v4 tables exist
	var count int
	err = db.db.QueryRow("SELECT COUNT(*) FROM projects").Scan(&count)
	if err != nil {
		t.Errorf("projects table doesn't exist: %v", err)
	}

	err = db.db.QueryRow("SELECT COUNT(*) FROM project_aliases").Scan(&count)
	if err != nil {
		t.Errorf("project_aliases table doesn't exist: %v", err)
	}

	err = db.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count)
	if err != nil {
		t.Errorf("tasks table doesn't exist: %v", err)
	}

	err = db.db.QueryRow("SELECT COUNT(*) FROM task_numbers").Scan(&count)
	if err != nil {
		t.Errorf("task_numbers table doesn't exist: %v", err)
	}

	// Verify project was created from prefix
	err = db.db.QueryRow(`
		SELECT COUNT(*) FROM projects WHERE name = 'test'
	`).Scan(&count)
	if err != nil {
		t.Errorf("failed to query projects: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 project, got %d", count)
	}

	// Verify alias was created
	err = db.db.QueryRow(`
		SELECT COUNT(*) FROM project_aliases WHERE alias = 'test'
	`).Scan(&count)
	if err != nil {
		t.Errorf("failed to query project_aliases: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 alias, got %d", count)
	}

	// Verify task was migrated
	err = db.db.QueryRow(`
		SELECT COUNT(*) FROM tasks
	`).Scan(&count)
	if err != nil {
		t.Errorf("failed to query tasks: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 task, got %d", count)
	}
}

// TestV4Events tests v4 event handling in reducer
func TestV4Events(t *testing.T) {
	r := NewReducer()

	// Test project.created event
	t.Run("project.created", func(t *testing.T) {
		payload := ProjectCreatedPayload{
			ProjectUID:  string(NewProjectUID()),
			Type:        "local",
			Name:        "Test Project",
			Description: "A test project",
			CreatedBy:   "testuser",
		}
		payloadJSON, _ := json.Marshal(payload)

		event := Event{
			ID:        string(NewEventID()),
			TS:        1,
			CreatedAt: time.Now(),
			Actor:     "testuser",
			Role:      "human",
			Kind:      string(EventKindProjectCreated),
			Payload:   payloadJSON,
		}

		if err := r.Apply(event); err != nil {
			t.Errorf("failed to apply project.created event: %v", err)
		}
	})

	// Test task.created (v4) event
	t.Run("task.created (v4)", func(t *testing.T) {
		taskUID := NewTaskUID()
		projectUID := NewProjectUID()

		payload := TaskCreatedV4Payload{
			TaskUID:        string(taskUID),
			ProjectUID:     string(projectUID),
			ProposedNumber: 1,
			CreatedNode:    string(NewNodeID()),
			Title:          "Test Task",
			CreatedBy:      "testuser",
		}
		payloadJSON, _ := json.Marshal(payload)

		event := Event{
			ID:        string(NewEventID()),
			TS:        2,
			CreatedAt: time.Now(),
			Actor:     "testuser",
			Role:      "human",
			Kind:      string(EventKindTaskCreated),
			Payload:   payloadJSON,
		}

		if err := r.Apply(event); err != nil {
			t.Errorf("failed to apply task.created (v4) event: %v", err)
		}

		// Verify task was created
		if _, exists := r.tasks[string(taskUID)]; !exists {
			t.Error("task was not created in reducer")
		}
	})

	// Test task.title.set event
	t.Run("task.title.set", func(t *testing.T) {
		taskUID := NewTaskUID()

		// First create the task
		createPayload := TaskCreatedV4Payload{
			TaskUID:        string(taskUID),
			ProjectUID:     string(NewProjectUID()),
			ProposedNumber: 1,
			CreatedNode:    string(NewNodeID()),
			Title:          "Original Title",
			CreatedBy:      "testuser",
		}
		createPayloadJSON, _ := json.Marshal(createPayload)

		createEvent := Event{
			ID:        string(NewEventID()),
			TS:        3,
			CreatedAt: time.Now(),
			Actor:     "testuser",
			Role:      "human",
			Kind:      string(EventKindTaskCreated),
			Payload:   createPayloadJSON,
		}

		if err := r.Apply(createEvent); err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		// Now change the title
		titlePayload := TaskTitleSetPayload{
			TaskUID: string(taskUID),
			Title:   "Updated Title",
		}
		titlePayloadJSON, _ := json.Marshal(titlePayload)

		titleEvent := Event{
			ID:        string(NewEventID()),
			TS:        4,
			CreatedAt: time.Now(),
			Actor:     "testuser",
			Role:      "human",
			Kind:      string(EventKindTaskTitleSet),
			Payload:   titlePayloadJSON,
		}

		if err := r.Apply(titleEvent); err != nil {
			t.Errorf("failed to apply task.title.set event: %v", err)
		}

		// Verify title was updated
		task := r.tasks[string(taskUID)]
		if task.Title != "Updated Title" {
			t.Errorf("expected title 'Updated Title', got '%s'", task.Title)
		}
	})
}

func TestV4MigrationPreservesStatusAndNotes(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node ID: %v", err)
	}

	_, err = db.db.Exec(`
		INSERT INTO prefixes (prefix, node, description, created_at, created_by, removed)
		VALUES (?, ?, ?, ?, ?, 0)
	`, "test", nodeID, "Test prefix", time.Now().Unix(), "tester")
	if err != nil {
		t.Fatalf("failed to insert prefix: %v", err)
	}

	legacyTaskUUID, err := GenerateTaskUUID()
	if err != nil {
		t.Fatalf("failed to generate task UUID: %v", err)
	}
	legacyTaskID := fmt.Sprintf("test-1-%s", nodeID)
	baseTime := time.Now()

	insertEvent := func(kind string, payload any, ts int64) {
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload for %s: %v", kind, err)
		}

		event := Event{
			ID:        fmt.Sprintf("ev-%d-%s", ts, nodeID),
			TS:        ts,
			CreatedAt: baseTime.Add(time.Duration(ts) * time.Second),
			Actor:     "tester",
			Role:      "human",
			Kind:      kind,
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			t.Fatalf("failed to insert %s event: %v", kind, err)
		}
	}

	insertEvent("task.created", TaskCreatedPayload{
		TaskUUID:  legacyTaskUUID,
		TaskID:    legacyTaskID,
		Title:     "Legacy Task",
		CreatedBy: "tester",
	}, 1)

	insertEvent("task.status.set", TaskStatusSetPayload{
		TaskUUID: legacyTaskUUID,
		TaskID:   legacyTaskID,
		Axis:     "generic",
		State:    "in_progress",
		Role:     "human",
	}, 2)

	insertEvent("task.note.add", TaskNoteAddPayload{
		TaskUUID: legacyTaskUUID,
		TaskID:   legacyTaskID,
		Markdown: "note from legacy system",
	}, 3)

	db.Close()

	db, err = OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	if err := db.MigrateToV4(dbPath); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	var migratedTaskUID string
	if err := db.db.QueryRow(`SELECT task_uid FROM tasks`).Scan(&migratedTaskUID); err != nil {
		t.Fatalf("failed to query migrated task UID: %v", err)
	}

	if !strings.HasPrefix(migratedTaskUID, "tsk_") {
		t.Fatalf("expected migrated task UID to use tsk_ prefix, got %s", migratedTaskUID)
	}

	var statusCount int
	if err := db.db.QueryRow(`
		SELECT COUNT(*) FROM events
		WHERE kind = 'task.status.set'
		  AND json_extract(payload, '$.task_uuid') = ?
	`, migratedTaskUID).Scan(&statusCount); err != nil {
		t.Fatalf("failed to count migrated status events: %v", err)
	}

	if statusCount != 1 {
		t.Fatalf("expected 1 migrated status event for task %s, got %d", migratedTaskUID, statusCount)
	}

	var noteCount int
	if err := db.db.QueryRow(`
		SELECT COUNT(*) FROM events
		WHERE kind = 'task.note.add'
		  AND json_extract(payload, '$.task_uuid') = ?
	`, migratedTaskUID).Scan(&noteCount); err != nil {
		t.Fatalf("failed to count migrated note events: %v", err)
	}

	if noteCount != 1 {
		t.Fatalf("expected 1 migrated note event for task %s, got %d", migratedTaskUID, noteCount)
	}
}

func TestV4MigrationConvertsReprefixToRelocate(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node ID: %v", err)
	}

	for _, prefix := range []string{"src", "dst"} {
		_, err = db.db.Exec(`
			INSERT INTO prefixes (prefix, node, description, created_at, created_by, removed)
			VALUES (?, ?, ?, ?, ?, 0)
		`, prefix, nodeID, fmt.Sprintf("%s prefix", prefix), time.Now().Unix(), "tester")
		if err != nil {
			t.Fatalf("failed to insert prefix %s: %v", prefix, err)
		}
	}

	legacyTaskUUID, err := GenerateTaskUUID()
	if err != nil {
		t.Fatalf("failed to generate task UUID: %v", err)
	}
	taskID := fmt.Sprintf("src-1-%s", nodeID)
	baseTime := time.Now()

	insertEvent := func(kind string, payload any, ts int64) {
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload for %s: %v", kind, err)
		}

		event := Event{
			ID:        fmt.Sprintf("ev-%d-%s", ts, nodeID),
			TS:        ts,
			CreatedAt: baseTime.Add(time.Duration(ts) * time.Second),
			Actor:     "tester",
			Role:      "human",
			Kind:      kind,
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			t.Fatalf("failed to insert %s event: %v", kind, err)
		}
	}

	insertEvent("task.created", TaskCreatedPayload{
		TaskUUID:  legacyTaskUUID,
		TaskID:    taskID,
		Title:     "Legacy Task",
		CreatedBy: "tester",
	}, 1)

	insertEvent("task.reprefix", TaskReprefixPayload{
		TaskUUID:  legacyTaskUUID,
		OldPrefix: "src",
		NewPrefix: "dst",
		OldNumber: 1,
		NewNumber: 42,
		OldNode:   nodeID,
		Reason:    "move to dst",
	}, 2)

	db.Close()

	db, err = OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	if err := db.MigrateToV4(dbPath); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	var taskUID string
	if err := db.db.QueryRow(`SELECT task_uid FROM tasks`).Scan(&taskUID); err != nil {
		t.Fatalf("failed to query migrated task UID: %v", err)
	}

	var dstProjectUID string
	if err := db.db.QueryRow(`SELECT project_uid FROM project_aliases WHERE alias = ? LIMIT 1`, "dst").Scan(&dstProjectUID); err != nil {
		t.Fatalf("failed to resolve dst project UID: %v", err)
	}

	var projectUID string
	if err := db.db.QueryRow(`SELECT project_uid FROM tasks WHERE task_uid = ?`, taskUID).Scan(&projectUID); err != nil {
		t.Fatalf("failed to query migrated task project: %v", err)
	}

	if projectUID != dstProjectUID {
		t.Fatalf("expected task project %s, got %s", dstProjectUID, projectUID)
	}

	var numberProjectUID string
	var numberValue int64
	if err := db.db.QueryRow(`
		SELECT project_uid, number FROM task_numbers WHERE task_uid = ?
	`, taskUID).Scan(&numberProjectUID, &numberValue); err != nil {
		t.Fatalf("failed to query task number: %v", err)
	}

	if numberProjectUID != dstProjectUID {
		t.Fatalf("expected task_numbers project %s, got %s", dstProjectUID, numberProjectUID)
	}
	if numberValue != 42 {
		t.Fatalf("expected task number 42, got %d", numberValue)
	}

	var relocateCount int
	if err := db.db.QueryRow(`
		SELECT COUNT(*) FROM events
		WHERE kind = 'task.relocate'
		  AND json_extract(payload, '$.task_uid') = ?
	`, taskUID).Scan(&relocateCount); err != nil {
		t.Fatalf("failed to count task.relocate events: %v", err)
	}

	if relocateCount != 1 {
		t.Fatalf("expected 1 task.relocate event for %s, got %d", taskUID, relocateCount)
	}

	var relocatedNumber float64
	if err := db.db.QueryRow(`
		SELECT json_extract(payload, '$.number_policy.number')
		FROM events
		WHERE kind = 'task.relocate'
		  AND json_extract(payload, '$.task_uid') = ?
	`, taskUID).Scan(&relocatedNumber); err != nil {
		t.Fatalf("failed to inspect task.relocate payload: %v", err)
	}

	if int64(relocatedNumber) != 42 {
		t.Fatalf("expected relocate payload number 42, got %d", int64(relocatedNumber))
	}
}

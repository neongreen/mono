package cmd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/testutil"
	"github.com/neongreen/mono/tk/internal/types"
)

func TestRecreateProject(t *testing.T) {
	db := testutil.OpenTempDB(t)

	// Create a synthetic project by inserting directly into database
	_, err := db.Db.Exec(`
		INSERT INTO projects (project_uid, name, type, is_synthetic, description, created_at, created_by)
		VALUES ('abc', 'abc', 'local', 1, 'Synthetic project', unixepoch(), 'system')
	`)
	if err != nil {
		t.Fatalf("failed to create synthetic project: %v", err)
	}

	// Create a task in the synthetic project
	// Note: We bypass validation by inserting events and projections directly,
	// because this tests legacy migration functionality with old-format UIDs
	taskUID := string(types.NewTaskUID())
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node ID: %v", err)
	}

	// Insert task.created event directly (bypassing validation)
	taskPayload := map[string]any{
		"task_uid":        taskUID,
		"project_uid":     "abc",
		"proposed_number": 5,
		"created_node":    nodeID,
		"title":           "test task",
		"created_by":      "tester",
	}
	taskJSON, _ := json.Marshal(taskPayload)
	_, err = db.Db.Exec(`
		INSERT INTO events (id, ts, created_at, actor, role, kind, payload)
		VALUES (?, 1, ?, 'tester', 'human', 'task.created', ?)
	`, types.NewEventID().String(), time.Now().Unix(), taskJSON)
	if err != nil {
		t.Fatalf("failed to insert task.created event: %v", err)
	}

	// Insert task into tasks table directly
	_, err = db.Db.Exec(`
		INSERT INTO tasks (task_uid, project_uid, created_node, title, created_at, created_by)
		VALUES (?, 'abc', ?, 'test task', unixepoch(), 'tester')
	`, taskUID, nodeID)
	if err != nil {
		t.Fatalf("failed to insert task: %v", err)
	}

	// Insert task.number.set event directly (bypassing validation)
	numberPayload := map[string]any{
		"task_uid":    taskUID,
		"project_uid": "abc",
		"number":      5,
		"reason":      "seed",
	}
	numberJSON, _ := json.Marshal(numberPayload)
	_, err = db.Db.Exec(`
		INSERT INTO events (id, ts, created_at, actor, role, kind, payload)
		VALUES (?, 2, ?, 'tester', 'human', 'task.number.set', ?)
	`, types.NewEventID().String(), time.Now().Unix(), numberJSON)
	if err != nil {
		t.Fatalf("failed to insert task.number.set event: %v", err)
	}

	// Insert task number directly
	_, err = db.Db.Exec(`
		INSERT INTO task_numbers (task_uid, project_uid, number)
		VALUES (?, 'abc', 5)
	`, taskUID)
	if err != nil {
		t.Fatalf("failed to insert task number: %v", err)
	}

	// Verify synthetic project exists
	var isSynthetic int
	err = db.Db.QueryRow(`SELECT is_synthetic FROM projects WHERE project_uid = 'abc'`).Scan(&isSynthetic)
	if err != nil {
		t.Fatalf("failed to query synthetic project: %v", err)
	}
	if isSynthetic != 1 {
		t.Fatalf("expected synthetic project, got is_synthetic=%d", isSynthetic)
	}

	// Recreate the project
	err = recreateProject(db, "abc", "abc")
	if err != nil {
		t.Fatalf("recreateProject failed: %v", err)
	}

	// Verify a real project now exists (not synthetic)
	err = db.Db.QueryRow(`SELECT is_synthetic FROM projects WHERE name = 'abc' AND is_synthetic = 0`).Scan(&isSynthetic)
	if err != nil {
		t.Fatalf("failed to find real project: %v", err)
	}
	if isSynthetic != 0 {
		t.Fatalf("expected real project, got is_synthetic=%d", isSynthetic)
	}

	// Verify task was moved to new project
	var taskProjectUID string
	err = db.Db.QueryRow(`SELECT project_uid FROM tasks WHERE task_uid = ?`, taskUID).Scan(&taskProjectUID)
	if err != nil {
		t.Fatalf("failed to query task: %v", err)
	}

	// Should be moved to a proper project UID (not "abc")
	if taskProjectUID == "abc" {
		t.Fatalf("task still in synthetic project")
	}
	if taskProjectUID[:4] != "prj_" {
		t.Fatalf("task project_uid should be a proper UUID, got %s", taskProjectUID)
	}

	// Verify task kept its number
	var taskNumber int64
	err = db.Db.QueryRow(`SELECT number FROM task_numbers WHERE task_uid = ?`, taskUID).Scan(&taskNumber)
	if err != nil {
		t.Fatalf("failed to query task number: %v", err)
	}
	if taskNumber != 5 {
		t.Fatalf("expected task to keep number 5, got %d", taskNumber)
	}

	// Verify synthetic project was deleted
	var count int
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM projects WHERE project_uid = 'abc'`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query projects: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected synthetic project to be deleted, got count=%d", count)
	}
}

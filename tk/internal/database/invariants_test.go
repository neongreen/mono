package database

import (
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

// TestCheckInvariants verifies the invariant checker works
func TestCheckInvariants_Valid(t *testing.T) {
	db := createTempDB(t)
	defer db.Close()

	// Create a valid event sequence
	projectUID := types.NewProjectUID()
	taskUID := types.NewTaskUID()

	events := []types.Event{
		createProjectEvent(1, projectUID, "test"),
		addProjectAliasEvent(2, projectUID, "test", "node1"),
		createTaskEvent(3, taskUID, projectUID, "Test task"),
		setTaskNumberEvent(4, taskUID, projectUID, 1),
	}

	for _, e := range events {
		if err := db.InsertEvent(e); err != nil {
			t.Fatalf("failed to insert event: %v", err)
		}
		if err := db.ProjectEvent(e); err != nil {
			t.Fatalf("failed to project event: %v", err)
		}
	}

	// Invariants should be satisfied
	if err := db.CheckInvariants(); err != nil {
		t.Fatalf("CheckInvariants failed on valid database: %v", err)
	}

	t.Log("✓ CheckInvariants passed on valid database")
}

// TestCheckInvariants_OrphanedTask detects orphaned projection data
func TestCheckInvariants_OrphanedTask(t *testing.T) {
	db := createTempDB(t)
	defer db.Close()

	projectUID := types.NewProjectUID()

	// Create project properly first
	events := []types.Event{
		createProjectEvent(1, projectUID, "test"),
		addProjectAliasEvent(2, projectUID, "test", "node1"),
	}

	for _, e := range events {
		db.InsertEvent(e)
		db.ProjectEvent(e)
	}

	// Create a task in projections WITHOUT a corresponding event
	orphanedUID := "task_ORPHAN123"

	_, err := db.Db.Exec(`INSERT INTO tasks (task_uid, project_uid, title, created_by, created_at, created_node)
		VALUES (?, ?, ?, ?, ?, ?)`, orphanedUID, projectUID.String(), "Orphaned task", "test", time.Now().UnixNano(), "node1")
	if err != nil {
		t.Fatalf("failed to insert orphaned task: %v", err)
	}

	// CheckInvariants should detect this
	err = db.CheckInvariants()
	if err == nil {
		t.Fatal("Expected CheckInvariants to detect orphaned task, but it passed")
	}

	t.Logf("Got expected error: %v", err)
	t.Log("✓ CheckInvariants correctly detected orphaned task")
}

// TestCheckInvariants_MissingProjection detects missing projections
func TestCheckInvariants_MissingProjection(t *testing.T) {
	db := createTempDB(t)
	defer db.Close()

	projectUID := types.NewProjectUID()
	taskUID := types.NewTaskUID()

	events := []types.Event{
		createProjectEvent(1, projectUID, "test"),
		addProjectAliasEvent(2, projectUID, "test", "node1"),
		createTaskEvent(3, taskUID, projectUID, "Test task"),
		// Note: no number.set event
	}

	// Insert events
	for _, e := range events {
		if err := db.InsertEvent(e); err != nil {
			t.Fatalf("failed to insert event: %v", err)
		}
	}

	// Project only some events (skip the task.created)
	if err := db.ProjectEvent(events[0]); err != nil {
		t.Fatalf("failed to project event: %v", err)
	}
	if err := db.ProjectEvent(events[1]); err != nil {
		t.Fatalf("failed to project event: %v", err)
	}
	// Skip events[2] - task.created not projected!

	// CheckInvariants should detect task exists in rebuilt but not in projections
	err := db.CheckInvariants()
	if err == nil {
		t.Fatal("Expected CheckInvariants to detect missing projection, but it passed")
	}

	t.Logf("Got expected error: %v", err)
	t.Log("✓ CheckInvariants correctly detected missing projection")
}

// TestProjectionCompleteness verifies all event types are projected
func TestProjectionCompleteness(t *testing.T) {
	db := createTempDB(t)
	defer db.Close()

	projectUID := types.NewProjectUID()
	taskUID := types.NewTaskUID()

	// Create a comprehensive set of events
	events := []types.Event{
		createProjectEvent(1, projectUID, "test"),
		addProjectAliasEvent(2, projectUID, "test", "node1"),
		createTaskEvent(3, taskUID, projectUID, "Test task"),
		setTaskNumberEvent(4, taskUID, projectUID, 1),
	}

	// Project all events
	for _, e := range events {
		if err := db.InsertEvent(e); err != nil {
			t.Fatalf("failed to insert event: %v", err)
		}
		if err := db.ProjectEvent(e); err != nil {
			t.Fatalf("failed to project event %s: %v", e.Kind, err)
		}
	}

	// Verify all projections exist
	var projectCount int
	if err := db.Db.QueryRow(`SELECT COUNT(*) FROM projects WHERE project_uid = ?`, projectUID.String()).Scan(&projectCount); err != nil {
		t.Fatalf("failed to query projects: %v", err)
	}
	if projectCount != 1 {
		t.Fatalf("expected 1 project in projections, got %d", projectCount)
	}

	var taskCount int
	if err := db.Db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE task_uid = ?`, taskUID.String()).Scan(&taskCount); err != nil {
		t.Fatalf("failed to query tasks: %v", err)
	}
	if taskCount != 1 {
		t.Fatalf("expected 1 task in projections, got %d", taskCount)
	}

	var numberCount int
	if err := db.Db.QueryRow(`SELECT COUNT(*) FROM task_numbers WHERE task_uid = ?`, taskUID.String()).Scan(&numberCount); err != nil {
		t.Fatalf("failed to query task_numbers: %v", err)
	}
	if numberCount != 1 {
		t.Fatalf("expected 1 task_number in projections, got %d", numberCount)
	}

	// CheckInvariants should pass
	if err := db.CheckInvariants(); err != nil {
		t.Fatalf("CheckInvariants failed: %v", err)
	}

	t.Log("✓ All events projected correctly")
	t.Log("✓ Projection completeness verified")
}

// TestNoDoubleProjection verifies events aren't projected multiple times
func TestNoDoubleProjection(t *testing.T) {
	db := createTempDB(t)
	defer db.Close()

	projectUID := types.NewProjectUID()
	taskUID := types.NewTaskUID()

	event := createTaskEvent(1, taskUID, projectUID, "Test task")

	// Insert event once
	if err := db.InsertEvent(event); err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	// Project it once
	if err := db.ProjectTaskCreatedEvent(event); err != nil {
		t.Fatalf("failed to project event: %v", err)
	}

	// Verify exactly 1 task in projection
	var count int
	if err := db.Db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE task_uid = ?`, taskUID.String()).Scan(&count); err != nil {
		t.Fatalf("failed to query tasks: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 task after first projection, got %d", count)
	}

	// Project the SAME event again (simulating bug #3/#4/#5 from audit)
	if err := db.ProjectTaskCreatedEvent(event); err != nil {
		t.Fatalf("failed to project event second time: %v", err)
	}

	// Verify still only 1 task (idempotent projection)
	if err := db.Db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE task_uid = ?`, taskUID.String()).Scan(&count); err != nil {
		t.Fatalf("failed to query tasks: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 task after double projection (should be idempotent), got %d", count)
	}

	// CheckInvariants should pass
	if err := db.CheckInvariants(); err != nil {
		t.Fatalf("CheckInvariants failed after double projection: %v", err)
	}

	t.Log("✓ Double projection handled idempotently (no duplicate tasks)")
	t.Log("✓ Projection is idempotent")
}

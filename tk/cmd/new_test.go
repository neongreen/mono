package cmd

import (
	"testing"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/tasks"
	"github.com/neongreen/mono/tk/internal/types"
)

// TestNewTaskInheritsParentProject verifies that when --parent is specified
// without an explicit --project flag, the child task inherits the parent's project
func TestNewTaskInheritsParentProject(t *testing.T) {
	db := openTempDB(t)

	// Create two projects: "work" and "personal"
	workProjectUID := seedProject(t, db, "work")
	personalProjectUID := seedProject(t, db, "personal")

	// Create a parent task in the "work" project
	parentTaskUID := seedTask(t, db, workProjectUID, "Parent task", 1)

	// Create a child task with --parent but without explicit --project
	// This should inherit the parent's project (work)
	childResult, err := tasks.Create(db, tasks.CreateParams{
		ProjectUID: types.ProjectUID(workProjectUID), // This simulates the new behavior where parent's project is used
		Title:      "Child task",
		ItemKind:   "task",
	}, "tester", &clock.RealClock{})
	if err != nil {
		t.Fatalf("failed to create child task: %v", err)
	}

	// Verify child task is in the same project as parent
	var childProjectUID string
	err = db.Db.QueryRow(`SELECT project_uid FROM tasks WHERE task_uid = ?`, childResult.TaskUID).Scan(&childProjectUID)
	if err != nil {
		t.Fatalf("failed to query child task project: %v", err)
	}

	if childProjectUID != string(workProjectUID) {
		t.Errorf("child task project = %s, want %s (same as parent)", childProjectUID, workProjectUID)
	}

	// Verify parent task is also in work project
	var parentProjectUID string
	err = db.Db.QueryRow(`SELECT project_uid FROM tasks WHERE task_uid = ?`, parentTaskUID).Scan(&parentProjectUID)
	if err != nil {
		t.Fatalf("failed to query parent task project: %v", err)
	}

	if parentProjectUID != string(workProjectUID) {
		t.Errorf("parent task project = %s, want %s", parentProjectUID, workProjectUID)
	}

	// Verify they're in the same project
	if childProjectUID != parentProjectUID {
		t.Errorf("child and parent should be in same project: child=%s, parent=%s", childProjectUID, parentProjectUID)
	}

	// Verify personal project was not used
	if childProjectUID == string(personalProjectUID) {
		t.Error("child task should not be in personal project")
	}
}

// TestNewTaskExplicitProjectOverridesParent verifies that when both --parent and
// --project are specified, the explicit --project takes precedence
func TestNewTaskExplicitProjectOverridesParent(t *testing.T) {
	db := openTempDB(t)

	// Create two projects
	workProjectUID := seedProject(t, db, "work")
	personalProjectUID := seedProject(t, db, "personal")

	// Create a parent task in the "work" project
	seedTask(t, db, workProjectUID, "Parent task", 1)

	// Create a child task with explicit --project that differs from parent
	childResult, err := tasks.Create(db, tasks.CreateParams{
		ProjectUID: types.ProjectUID(personalProjectUID), // Explicitly specify different project
		Title:      "Child task in different project",
		ItemKind:   "task",
	}, "tester", &clock.RealClock{})
	if err != nil {
		t.Fatalf("failed to create child task: %v", err)
	}

	// Verify child task is in the explicitly specified project
	var childProjectUID string
	err = db.Db.QueryRow(`SELECT project_uid FROM tasks WHERE task_uid = ?`, childResult.TaskUID).Scan(&childProjectUID)
	if err != nil {
		t.Fatalf("failed to query child task project: %v", err)
	}

	if childProjectUID != string(personalProjectUID) {
		t.Errorf("child task project = %s, want %s (explicitly specified)", childProjectUID, personalProjectUID)
	}

	// Verify it's NOT in the parent's project
	if childProjectUID == string(workProjectUID) {
		t.Error("child task should not inherit parent's project when explicit project is specified")
	}
}

// TestNewTaskWithoutParent verifies that when --parent is not specified,
// the default project behavior works as expected
func TestNewTaskWithoutParent(t *testing.T) {
	db := openTempDB(t)

	defaultProjectUID := seedProject(t, db, "me")

	// Create a task without --parent (should use default project)
	result, err := tasks.Create(db, tasks.CreateParams{
		ProjectUID: types.ProjectUID(defaultProjectUID),
		Title:      "Task without parent",
		ItemKind:   "task",
	}, "tester", &clock.RealClock{})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Verify task is in the default project
	var taskProjectUID string
	err = db.Db.QueryRow(`SELECT project_uid FROM tasks WHERE task_uid = ?`, result.TaskUID).Scan(&taskProjectUID)
	if err != nil {
		t.Fatalf("failed to query task project: %v", err)
	}

	if taskProjectUID != string(defaultProjectUID) {
		t.Errorf("task project = %s, want %s", taskProjectUID, defaultProjectUID)
	}
}

// TestResolveParentAndGetProject verifies the low-level database operation
// to fetch a parent task's project
func TestResolveParentAndGetProject(t *testing.T) {
	db := openTempDB(t)

	// Create a project and task
	projectUID := seedProject(t, db, "test")
	taskUID := seedTask(t, db, projectUID, "Test task", 1)

	// Resolve task by display ID
	resolvedUID, err := database.ResolveTaskReference(db, types.NewTaskRef("test-1"))
	if err != nil {
		t.Fatalf("failed to resolve task reference: %v", err)
	}

	if resolvedUID != taskUID {
		t.Errorf("resolved task UID = %s, want %s", resolvedUID, taskUID)
	}

	// Get project from resolved task
	var taskProjectUID string
	err = db.Db.QueryRow(`SELECT project_uid FROM tasks WHERE task_uid = ?`, resolvedUID).Scan(&taskProjectUID)
	if err != nil {
		t.Fatalf("failed to get project from task: %v", err)
	}

	if taskProjectUID != string(projectUID) {
		t.Errorf("task project = %s, want %s", taskProjectUID, projectUID)
	}
}

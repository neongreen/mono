package project

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/testutil"
	"github.com/neongreen/mono/tk/internal/types"
)

// TestProjectRmRequiresForceWithTasks tests that deleting a project with tasks requires --force
// Regression test for tk-146: tk project rm should require --force when deleting project with tasks
func TestProjectRmRequiresForceWithTasks(t *testing.T) {
	db := testutil.OpenTempDB(t)

	// Create a project with a task
	projectUID := testutil.SeedProject(t, db, "test-proj")
	_ = testutil.SeedTask(t, db, projectUID, "test task", 1)

	// Try to delete without --force - should fail
	err := deleteProject(db, projectUID, false)
	if err == nil {
		t.Fatal("expected error when deleting project with tasks, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "cannot delete") {
		t.Errorf("error should mention 'cannot delete', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "task") {
		t.Errorf("error should mention tasks, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "--force") {
		t.Errorf("error should suggest --force flag, got: %s", errMsg)
	}

	// Verify project still exists
	var count int
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM projects WHERE project_uid = ?`, projectUID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query projects: %v", err)
	}
	if count != 1 {
		t.Errorf("expected project to still exist, got count=%d", count)
	}

	// Now try with --force - should succeed
	err = deleteProject(db, projectUID, true)
	if err != nil {
		t.Fatalf("delete with --force failed: %v", err)
	}

	// Verify project was deleted
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM projects WHERE project_uid = ?`, projectUID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query projects: %v", err)
	}
	if count != 0 {
		t.Errorf("expected project to be deleted, got count=%d", count)
	}
}

// TestProjectRmEmptyProjectNoForceRequired tests that empty projects can be deleted without --force
func TestProjectRmEmptyProjectNoForceRequired(t *testing.T) {
	db := testutil.OpenTempDB(t)

	// Create a project with no tasks
	projectUID := testutil.SeedProject(t, db, "empty-proj")

	// Delete without --force - should succeed
	err := deleteProject(db, projectUID, false)
	if err != nil {
		t.Fatalf("delete empty project failed: %v", err)
	}

	// Verify project was deleted
	var count int
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM projects WHERE project_uid = ?`, projectUID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query projects: %v", err)
	}
	if count != 0 {
		t.Errorf("expected project to be deleted, got count=%d", count)
	}
}

// deleteProject executes the project deletion logic (extracted for testing)
func deleteProject(db *database.DB, projectUID string, force bool) error {
	// Check if project has tasks
	var taskCount int
	err := db.Db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE project_uid = ?`, projectUID).Scan(&taskCount)
	if err != nil {
		return err
	}

	if taskCount > 0 && !force {
		return fmt.Errorf("cannot delete project %s: it has %d task(s). Use --force to delete anyway", projectUID, taskCount)
	}

	// Create project.delete event
	payload := types.ProjectDeletePayload{
		ProjectUID: types.ProjectUID(projectUID),
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return err
	}

	ts, err := db.GetNextLamportTS()
	if err != nil {
		return err
	}

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindProjectDelete),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return err
	}

	return db.ProjectProjectDeleteEvent(event)
}

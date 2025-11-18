package cmd

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

	// Verify project still exists and is not marked as deleted
	var deleted int
	err = db.Db.QueryRow(`SELECT deleted FROM projects WHERE project_uid = ?`, projectUID).Scan(&deleted)
	if err != nil {
		t.Fatalf("failed to query projects: %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected project to not be deleted, got deleted=%d", deleted)
	}

	// Now try with --force - should succeed
	err = deleteProject(db, projectUID, true)
	if err != nil {
		t.Fatalf("delete with --force failed: %v", err)
	}

	// Verify project was marked as deleted
	err = db.Db.QueryRow(`SELECT deleted FROM projects WHERE project_uid = ?`, projectUID).Scan(&deleted)
	if err != nil {
		t.Fatalf("failed to query projects: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected project to be marked as deleted, got deleted=%d", deleted)
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

	// Verify project was marked as deleted
	var deleted int
	err = db.Db.QueryRow(`SELECT deleted FROM projects WHERE project_uid = ?`, projectUID).Scan(&deleted)
	if err != nil {
		t.Fatalf("failed to query projects: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected project to be marked as deleted, got deleted=%d", deleted)
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

	return db.RebuildProjections()
}

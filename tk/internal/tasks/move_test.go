package tasks

import (
	"strings"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/testutil"
)

func TestMove_ValidatesDestinationProjectUID(t *testing.T) {
	db := testutil.OpenTempDB(t)

	projectA := testutil.SeedProject(t, db, "proj-a")
	projectB := testutil.SeedProject(t, db, "proj-b")
	taskUID := testutil.SeedTask(t, db, projectA, "task", 1)

	clk := clock.NewVirtualClock(time.Unix(100, 0))

	// Valid move should work
	err := Move(db, taskUID, projectB, MoveOptions{Mode: "auto"}, "tester", clk)
	if err != nil {
		t.Fatalf("valid move failed: %v", err)
	}

	// Move with invalid project UID format should fail
	err = Move(db, taskUID, "lovable", MoveOptions{Mode: "auto"}, "tester", clk)
	if err == nil {
		t.Fatal("expected error when moving to invalid project UID, got nil")
	}
	if !strings.Contains(err.Error(), "invalid destination project UID") {
		t.Errorf("expected 'invalid destination project UID' error, got: %v", err)
	}

	// Move with valid format but non-existent project should fail
	nonExistentProject := "prj_01HJKKZ4W7MJNE8KTXZZZZZZZZ" // Valid ULID format but doesn't exist
	err = Move(db, taskUID, nonExistentProject, MoveOptions{Mode: "auto"}, "tester", clk)
	if err == nil {
		t.Fatal("expected error when moving to non-existent project, got nil")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected 'does not exist' error, got: %v", err)
	}
}

func TestMove_ValidatesTaskUID(t *testing.T) {
	db := testutil.OpenTempDB(t)

	_ = testutil.SeedProject(t, db, "proj-a")
	projectB := testutil.SeedProject(t, db, "proj-b")

	clk := clock.NewVirtualClock(time.Unix(100, 0))

	// Move with invalid task UID format should fail
	err := Move(db, "task-invalid", projectB, MoveOptions{Mode: "auto"}, "tester", clk)
	if err == nil {
		t.Fatal("expected error when moving invalid task UID, got nil")
	}
	if !strings.Contains(err.Error(), "invalid task UID") {
		t.Errorf("expected 'invalid task UID' error, got: %v", err)
	}

	// This also tests that we validate before trying to look up the task
	// (otherwise we'd get a different error about task not found)
}

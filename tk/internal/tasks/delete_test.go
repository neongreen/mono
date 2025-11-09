package tasks

import (
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/testutil"
	"github.com/neongreen/mono/tk/internal/types"
)

func TestDelete(t *testing.T) {
	db := testutil.OpenTempDB(t)

	projectUID := testutil.SeedProject(t, db, "test")

	clk := clock.NewVirtualClock(time.Unix(100, 0))
	result, err := Create(db, CreateParams{ProjectUID: types.ProjectUID(projectUID), Title: "Test task"}, "tester", clk)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete the task
	clk.Advance(time.Second)
	err = Delete(db, string(result.TaskUID), "tester", clk)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify task was deleted (removed from table)
	var count int
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE task_uid = ?`, result.TaskUID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query task: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected task to be deleted, but still found %d rows", count)
	}

	// Verify invariants are satisfied after delete
	db.CheckInvariantsT(t)
}

func TestDeleteIdempotent(t *testing.T) {
	db := testutil.OpenTempDB(t)

	projectUID := testutil.SeedProject(t, db, "test")

	clk := clock.NewVirtualClock(time.Unix(200, 0))
	result, err := Create(db, CreateParams{ProjectUID: types.ProjectUID(projectUID), Title: "Test task"}, "tester", clk)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete twice should not error
	clk.Advance(time.Second)
	err = Delete(db, string(result.TaskUID), "tester", clk)
	if err != nil {
		t.Fatalf("First delete failed: %v", err)
	}

	clk.Advance(time.Second)
	err = Delete(db, string(result.TaskUID), "tester", clk)
	if err != nil {
		t.Fatalf("Second delete failed (should be idempotent): %v", err)
	}

	// Verify invariants are satisfied after idempotent deletes
	db.CheckInvariantsT(t)
}

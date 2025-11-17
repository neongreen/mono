package tasks

import (
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/testutil"
	"github.com/neongreen/mono/tk/internal/types"
)

func TestCreate(t *testing.T) {
	db := testutil.OpenTempDB(t)

	projectUID := testutil.SeedProject(t, db, "test")

	clk := clock.NewVirtualClock(time.Unix(300, 0))
	result, err := Create(db, CreateParams{
		ProjectUID: types.ProjectUID(projectUID),
		Title:      "Test task",
	}, "tester", clk)

	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if result.TaskUID == "" {
		t.Fatal("expected task UID, got empty")
	}

	if result.DisplayID == "" {
		t.Fatal("expected display ID, got empty")
	}

	// Verify task was created in database
	var count int
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE task_uid = ?`, result.TaskUID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query task: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 task, got %d", count)
	}

	// Verify title was set
	var title string
	err = db.Db.QueryRow(`SELECT title FROM tasks WHERE task_uid = ?`, result.TaskUID).Scan(&title)
	if err != nil {
		t.Fatalf("failed to query title: %v", err)
	}
	if title != "Test task" {
		t.Fatalf("expected 'Test task', got %q", title)
	}

	// Verify invariants are satisfied
	db.CheckInvariantsT(t)
}

func TestCreateMultipleTasks(t *testing.T) {
	db := testutil.OpenTempDB(t)

	projectUID := testutil.SeedProject(t, db, "test")

	clk := clock.NewVirtualClock(time.Unix(400, 0))

	// Create multiple tasks
	result1, err := Create(db, CreateParams{ProjectUID: types.ProjectUID(projectUID), Title: "Task 1"}, "tester", clk)
	if err != nil {
		t.Fatalf("Create task 1 failed: %v", err)
	}

	clk.Advance(time.Second)
	result2, err := Create(db, CreateParams{ProjectUID: types.ProjectUID(projectUID), Title: "Task 2"}, "tester", clk)
	if err != nil {
		t.Fatalf("Create task 2 failed: %v", err)
	}

	// Verify different UIDs
	if result1.TaskUID == result2.TaskUID {
		t.Fatal("expected different task UIDs")
	}

	// Verify sequential numbers
	var num1, num2 int64
	err = db.Db.QueryRow(`SELECT number FROM task_numbers WHERE task_uid = ?`, result1.TaskUID).Scan(&num1)
	if err != nil {
		t.Fatalf("failed to query number for task 1: %v", err)
	}
	err = db.Db.QueryRow(`SELECT number FROM task_numbers WHERE task_uid = ?`, result2.TaskUID).Scan(&num2)
	if err != nil {
		t.Fatalf("failed to query number for task 2: %v", err)
	}

	if num2 != num1+1 {
		t.Fatalf("expected sequential numbers, got %d and %d", num1, num2)
	}

	// Verify invariants are satisfied
	db.CheckInvariantsT(t)
}

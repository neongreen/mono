package cmd

import (
	"testing"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/tasks"
)

func TestEditNumber(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "edit")
	taskUID := seedTask(t, db, projectUID, "edit task", 1)

	if err := tasks.EditNumber(db, taskUID, "5", "test-user", &clock.RealClock{}); err != nil {
		t.Fatalf("EditNumber failed: %v", err)
	}

	var number int64
	if err := db.Db.QueryRow(`SELECT number FROM task_numbers WHERE task_uid = ?`, taskUID).Scan(&number); err != nil {
		t.Fatalf("failed to load task number: %v", err)
	}
	if number != 5 {
		t.Fatalf("expected number 5, got %d", number)
	}
}

func TestEditTitle(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "title")
	taskUID := seedTask(t, db, projectUID, "Old Title", 3)

	if err := tasks.EditTitle(db, taskUID, "New Title", "test-user", &clock.RealClock{}); err != nil {
		t.Fatalf("EditTitle failed: %v", err)
	}

	var title string
	if err := db.Db.QueryRow(`SELECT title FROM tasks WHERE task_uid = ?`, taskUID).Scan(&title); err != nil {
		t.Fatalf("failed to load task title: %v", err)
	}
	if title != "New Title" {
		t.Fatalf("expected title to be updated, got %q", title)
	}
}

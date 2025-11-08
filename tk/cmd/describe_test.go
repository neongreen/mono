package cmd

import (
	"testing"

	"github.com/neongreen/mono/tk/internal/tasks"
)

func TestDescribeCommand(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "desc")
	taskUID := seedTask(t, db, projectUID, "Old Title", 1)

	if err := tasks.EditTitle(db, taskUID, "New Title", "test-user"); err != nil {
		t.Fatalf("EditTitle failed: %v", err)
	}

	var title string
	if err := db.Db.QueryRow(`SELECT title FROM tasks WHERE task_uid = ?`, taskUID).Scan(&title); err != nil {
		t.Fatalf("failed to load task title: %v", err)
	}
	if title != "New Title" {
		t.Fatalf("expected title to be 'New Title', got %q", title)
	}
}

func TestDescribeByTaskRef(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "myproject")
	taskUID := seedTask(t, db, projectUID, "Initial Task", 5)

	// Edit the task title directly using the task UID
	if err := tasks.EditTitle(db, taskUID, "Updated Task Title", "test-user"); err != nil {
		t.Fatalf("EditTitle failed: %v", err)
	}

	var title string
	if err := db.Db.QueryRow(`SELECT title FROM tasks WHERE task_uid = ?`, taskUID).Scan(&title); err != nil {
		t.Fatalf("failed to load task title: %v", err)
	}
	if title != "Updated Task Title" {
		t.Fatalf("expected title to be 'Updated Task Title', got %q", title)
	}
}

func TestDescribeEmptyTitle(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "test")
	taskUID := seedTask(t, db, projectUID, "Some Title", 1)

	// EditTitle should reject empty titles
	err := tasks.EditTitle(db, taskUID, "", "test-user")
	if err == nil {
		t.Fatal("expected error for empty title, got nil")
	}
}

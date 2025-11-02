package main

import "testing"

func TestEditNumber(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "edit")
	taskUID := seedTask(t, db, projectUID, "edit task", 1)

	if err := editTask(db, "edit-1", "number", "5"); err != nil {
		t.Fatalf("editTask number failed: %v", err)
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

	if err := editTask(db, "title-3", "title", "New Title"); err != nil {
		t.Fatalf("editTask title failed: %v", err)
	}

	var title string
	if err := db.Db.QueryRow(`SELECT title FROM tasks WHERE task_uid = ?`, taskUID).Scan(&title); err != nil {
		t.Fatalf("failed to load task title: %v", err)
	}
	if title != "New Title" {
		t.Fatalf("expected title to be updated, got %q", title)
	}
}

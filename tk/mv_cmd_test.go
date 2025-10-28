package main

import (
	"testing"
)

func TestMoveTaskAutoAssign(t *testing.T) {
	db := openTempDB(t)

	srcProject := seedProject(t, db, "src")
	dstProject := seedProject(t, db, "dst")

	// Existing task in destination to force auto numbering to pick next slot.
	_ = seedTask(t, db, dstProject, "existing", 1)

	taskUID := seedTask(t, db, srcProject, "to move", 1)

	opts := moveOptions{
		Mode:        "auto",
		OnCollision: "fail",
	}
	if err := moveTask(db, "src-1", "dst", opts); err != nil {
		t.Fatalf("moveTask failed: %v", err)
	}

	var projectUID string
	var number int64
	if err := db.db.QueryRow(`SELECT project_uid FROM tasks WHERE task_uid = ?`, taskUID).Scan(&projectUID); err != nil {
		t.Fatalf("failed to load task project: %v", err)
	}
	if projectUID != dstProject {
		t.Fatalf("expected project %s, got %s", dstProject, projectUID)
	}

	if err := db.db.QueryRow(`SELECT number FROM task_numbers WHERE task_uid = ?`, taskUID).Scan(&number); err != nil {
		t.Fatalf("failed to load task number: %v", err)
	}
	if number != 2 {
		t.Fatalf("expected auto-assigned number 2, got %d", number)
	}
}

func TestMoveTaskKeepCollisionFails(t *testing.T) {
	db := openTempDB(t)

	srcProject := seedProject(t, db, "src")
	dstProject := seedProject(t, db, "dst")

	_ = seedTask(t, db, dstProject, "existing", 1)
	_ = seedTask(t, db, srcProject, "to move", 1)

	err := moveTask(db, "src-1", "dst", moveOptions{
		Mode:        "keep",
		OnCollision: "fail",
	})
	if err == nil {
		t.Fatalf("expected collision error, got nil")
	}
}

func TestMoveTaskKeepCollisionAutoFallback(t *testing.T) {
	db := openTempDB(t)

	srcProject := seedProject(t, db, "src")
	dstProject := seedProject(t, db, "dst")

	_ = seedTask(t, db, dstProject, "existing", 1)
	taskUID := seedTask(t, db, srcProject, "to move", 1)

	if err := moveTask(db, "src-1", "dst", moveOptions{
		Mode:        "keep",
		OnCollision: "auto",
	}); err != nil {
		t.Fatalf("moveTask with keep+auto failed: %v", err)
	}

	var number int64
	if err := db.db.QueryRow(`SELECT number FROM task_numbers WHERE task_uid = ?`, taskUID).Scan(&number); err != nil {
		t.Fatalf("failed to load task number: %v", err)
	}
	if number != 2 {
		t.Fatalf("expected fallback number 2, got %d", number)
	}
}

package cmd

import (
	"testing"

	"github.com/neongreen/mono/tk/internal/tasks"
)

func TestEditStatus(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "alpha")
	taskUID := seedTask(t, db, projectUID, "Task", 1)

	if err := tasks.EditStatus(db, taskUID, "done", "test-user"); err != nil {
		t.Fatalf("EditStatus failed: %v", err)
	}

	var state string
	err := db.Db.QueryRow(`
		SELECT json_extract(payload, '$.state')
		FROM events
		WHERE kind = 'task.status.set'
		ORDER BY ts DESC LIMIT 1
	`).Scan(&state)
	if err != nil {
		t.Fatalf("failed to query status event: %v", err)
	}
	if state != "done" {
		t.Fatalf("expected status done, got %s", state)
	}

	// Ensure status applies to task
	reducer := buildReducerFromDB(t, db)
	task, ok := reducer.GetTask(taskUID)
	if !ok {
		t.Fatalf("task not found in reducer")
	}
	axis := task.Axes["generic"]
	if axis.Effective != "done" {
		t.Fatalf("expected effective status done, got %s", axis.Effective)
	}
}

package cmd

import (
	"testing"
)

func TestEditStatus(t *testing.T) {
	t.Skip("TODO: Rewrite to use internal/tasks.EditField")
	/*
		db := openTempDB(t)

		projectUID := seedProject(t, db, "alpha")
		taskUID := seedTask(t, db, projectUID, "Task", 1)

		if err := editTask(db, "alpha-1", "status", "done"); err != nil {
			t.Fatalf("editTask status failed: %v", err)
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

		// ensure status applies to task
		reducer := buildReducerFromDB(t, db)
		task, ok := reducer.GetTask(taskUID)
		if !ok {
			t.Fatalf("task not found in reducer")
		}
		axis := task.Axes["generic"]
		if axis.Effective != "done" {
			t.Fatalf("expected effective status done, got %s", axis.Effective)
		}
	*/
}

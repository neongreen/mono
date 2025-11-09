package tasks

import (
	"fmt"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/testutil"
	"github.com/neongreen/mono/tk/internal/types"
)

func TestAddNote(t *testing.T) {
	db := testutil.OpenTempDB(t)

	projectUID := testutil.SeedProject(t, db, "test")

	clk := clock.NewVirtualClock(time.Unix(500, 0))
	result, err := Create(db, CreateParams{ProjectUID: types.ProjectUID(projectUID), Title: "Test task"}, "tester", clk)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Add a note
	clk.Advance(time.Second)
	err = AddNote(db, string(result.TaskUID), "This is a test note", "tester", clk)
	if err != nil {
		t.Fatalf("AddNote failed: %v", err)
	}

	// Verify note was added by checking events
	var count int
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM events WHERE kind = 'task.note.add' AND json_extract(payload, '$.task_uuid') = ?`,
		result.TaskUID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query events: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 note event, got %d", count)
	}

	// Verify invariants are satisfied
	db.CheckInvariantsT(t)
}

func TestAddMultipleNotes(t *testing.T) {
	db := testutil.OpenTempDB(t)

	projectUID := testutil.SeedProject(t, db, "test")

	clk := clock.NewVirtualClock(time.Unix(600, 0))
	result, err := Create(db, CreateParams{ProjectUID: types.ProjectUID(projectUID), Title: "Test task"}, "tester", clk)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Add multiple notes
	for i := 1; i <= 3; i++ {
		clk.Advance(time.Second)
		err = AddNote(db, string(result.TaskUID), fmt.Sprintf("Note %d", i), "tester", clk)
		if err != nil {
			t.Fatalf("AddNote %d failed: %v", i, err)
		}
	}

	// Verify all notes were added
	var count int
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM events WHERE kind = 'task.note.add' AND json_extract(payload, '$.task_uuid') = ?`,
		result.TaskUID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query events: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 note events, got %d", count)
	}

	// Verify invariants are satisfied
	db.CheckInvariantsT(t)
}

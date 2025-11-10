package debug

import (
	"strings"
	"testing"

	"github.com/neongreen/mono/tk/internal/testutil"
)

// Test helper aliases
var (
	openTempDB       = testutil.OpenTempDB
	seedProject      = testutil.SeedProject
	seedTask         = testutil.SeedTask
	seedTaskWithNode = testutil.SeedTaskWithNode
)

func TestDoctorHealthy(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "proj")
	seedTask(t, db, projectUID, "task", 1)

	report, err := RunDoctor(db)
	if err != nil {
		t.Fatalf("runDoctor failed: %v", err)
	}
	if report.ProblemCount() != 0 {
		t.Fatalf("expected no issues, got %d", report.ProblemCount())
	}
}

func TestDoctorDetectsOrphanTask(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "proj")
	taskUID := seedTask(t, db, projectUID, "task", 1)

	if _, err := db.Db.Exec(`DELETE FROM projects WHERE project_uid = ?`, projectUID); err != nil {
		t.Fatalf("failed to remove project: %v", err)
	}

	report, err := RunDoctor(db)
	if err != nil {
		t.Fatalf("runDoctor failed: %v", err)
	}
	if report.ProblemCount() == 0 {
		t.Fatalf("expected issues, got none")
	}
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue, taskUID) || strings.Contains(issue, "proj") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected orphan task issue, got %v", report.Issues)
	}
}

func TestDoctorDetectsCollisions(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "proj")
	nodeA := "NODE_A"
	nodeB := "NODE_B"
	if _, err := db.Db.Exec(`INSERT OR REPLACE INTO metadata (key, value) VALUES ('node_id', ?)`, nodeA); err != nil {
		t.Fatalf("failed to override node id: %v", err)
	}
	seedTaskWithNode(t, db, projectUID, "first", 3, nodeA)
	seedTaskWithNode(t, db, projectUID, "second", 3, nodeB)

	report, err := RunDoctor(db)
	if err != nil {
		t.Fatalf("runDoctor failed: %v", err)
	}
	if len(report.Collisions) == 0 {
		t.Fatalf("expected collision to be reported")
	}
}

func TestDoctorDetectsInvalidEventReferences(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "proj")
	taskUID := seedTask(t, db, projectUID, "task", 1)

	// Insert a task.relocate event with invalid project UID (the bug from tk-281)
	badEvent := []byte(`{
		"task_uid": "` + taskUID + `",
		"from_project_uid": "` + projectUID + `",
		"to_project_uid": "abc",
		"number_policy": {"mode": "auto"}
	}`)

	ts, err := db.GetNextLamportTS()
	if err != nil {
		t.Fatalf("failed to get lamport ts: %v", err)
	}

	_, err = db.Db.Exec(`
		INSERT INTO events (id, ts, created_at, actor, role, kind, payload)
		VALUES ('evt_test_123', ?, unixepoch(), 'test', 'human', 'task.relocate', ?)
	`, ts, badEvent)
	if err != nil {
		t.Fatalf("failed to insert bad event: %v", err)
	}

	report, err := RunDoctor(db)
	if err != nil {
		t.Fatalf("runDoctor failed: %v", err)
	}

	// Should detect the invalid to_project_uid
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue, "to_project_uid") && strings.Contains(issue, "abc") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected invalid reference issue for 'abc', got issues: %v", report.Issues)
	}
}

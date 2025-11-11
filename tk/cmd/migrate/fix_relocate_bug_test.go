package migrate

import (
	"testing"

	"github.com/neongreen/mono/tk/internal/testutil"
)

func TestRecreateProject(t *testing.T) {
	db := testutil.OpenTempDB(t)

	// Create a synthetic project by inserting directly into database
	_, err := db.Db.Exec(`
		INSERT INTO projects (project_uid, name, type, is_synthetic, description, created_at, created_by)
		VALUES ('lovable', 'lovable', 'local', 1, 'Synthetic project', unixepoch(), 'system')
	`)
	if err != nil {
		t.Fatalf("failed to create synthetic project: %v", err)
	}

	// Create a task in the synthetic project
	taskUID := testutil.SeedTask(t, db, "lovable", "test task", 5)

	// Verify synthetic project exists
	var isSynthetic int
	err = db.Db.QueryRow(`SELECT is_synthetic FROM projects WHERE project_uid = 'lovable'`).Scan(&isSynthetic)
	if err != nil {
		t.Fatalf("failed to query synthetic project: %v", err)
	}
	if isSynthetic != 1 {
		t.Fatalf("expected synthetic project, got is_synthetic=%d", isSynthetic)
	}

	// Recreate the project
	err = recreateProject(db, "lovable", "lovable")
	if err != nil {
		t.Fatalf("recreateProject failed: %v", err)
	}

	// Verify a real project now exists (not synthetic)
	err = db.Db.QueryRow(`SELECT is_synthetic FROM projects WHERE name = 'lovable' AND is_synthetic = 0`).Scan(&isSynthetic)
	if err != nil {
		t.Fatalf("failed to find real project: %v", err)
	}
	if isSynthetic != 0 {
		t.Fatalf("expected real project, got is_synthetic=%d", isSynthetic)
	}

	// Verify task was moved to new project
	var taskProjectUID string
	err = db.Db.QueryRow(`SELECT project_uid FROM tasks WHERE task_uid = ?`, taskUID).Scan(&taskProjectUID)
	if err != nil {
		t.Fatalf("failed to query task: %v", err)
	}

	// Should be moved to a proper project UID (not "lovable")
	if taskProjectUID == "lovable" {
		t.Fatalf("task still in synthetic project")
	}
	if taskProjectUID[:4] != "prj_" {
		t.Fatalf("task project_uid should be a proper UUID, got %s", taskProjectUID)
	}

	// Verify task kept its number
	var taskNumber int64
	err = db.Db.QueryRow(`SELECT number FROM task_numbers WHERE task_uid = ?`, taskUID).Scan(&taskNumber)
	if err != nil {
		t.Fatalf("failed to query task number: %v", err)
	}
	if taskNumber != 5 {
		t.Fatalf("expected task to keep number 5, got %d", taskNumber)
	}

	// Verify synthetic project was deleted
	var count int
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM projects WHERE project_uid = 'lovable'`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query projects: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected synthetic project to be deleted, got count=%d", count)
	}
}

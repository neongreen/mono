package conflicts

import (
	"testing"

	debug_pkg "github.com/neongreen/mono/tk/cmd/debug"
	"github.com/neongreen/mono/tk/internal/testutil"
)

func TestGetNumberCollisionsAll(t *testing.T) {
	db := testutil.OpenTempDB(t)

	projectA := testutil.SeedProject(t, db, "a")
	projectB := testutil.SeedProject(t, db, "b")

	node := "NODE_X"
	if _, err := db.Db.Exec(`INSERT OR REPLACE INTO metadata (key, value) VALUES ('node_id', ?)`, node); err != nil {
		t.Fatalf("failed to override node id: %v", err)
	}
	testutil.SeedTaskWithNode(t, db, projectA, "task1", 1, node)
	testutil.SeedTaskWithNode(t, db, projectA, "task2", 1, "NODE_Y")
	testutil.SeedTaskWithNode(t, db, projectB, "task3", 2, node)
	testutil.SeedTaskWithNode(t, db, projectB, "task4", 2, "NODE_Z")

	collisions, err := debug_pkg.GetNumberCollisions(db, "")
	if err != nil {
		t.Fatalf("getNumberCollisions failed: %v", err)
	}
	if len(collisions) != 2 {
		t.Fatalf("expected 2 collisions, got %d", len(collisions))
	}
}

func TestGetNumberCollisionsFilter(t *testing.T) {
	db := testutil.OpenTempDB(t)

	project := testutil.SeedProject(t, db, "a")
	other := testutil.SeedProject(t, db, "b")

	if _, err := db.Db.Exec(`INSERT OR REPLACE INTO metadata (key, value) VALUES ('node_id', 'NODE_X')`); err != nil {
		t.Fatalf("failed to override node id: %v", err)
	}
	testutil.SeedTaskWithNode(t, db, project, "task1", 1, "NODE_X")
	testutil.SeedTaskWithNode(t, db, project, "task2", 1, "NODE_Y")
	testutil.SeedTaskWithNode(t, db, other, "task3", 2, "NODE_X")

	collisions, err := debug_pkg.GetNumberCollisions(db, project)
	if err != nil {
		t.Fatalf("getNumberCollisions failed: %v", err)
	}
	if len(collisions) != 1 {
		t.Fatalf("expected 1 collision, got %d", len(collisions))
	}
	if collisions[0].ProjectUID != project {
		t.Fatalf("expected project %s, got %s", project, collisions[0].ProjectUID)
	}
}

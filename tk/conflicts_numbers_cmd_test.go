package main

import "testing"

func TestGetNumberCollisionsAll(t *testing.T) {
	db := openTempDB(t)

	projectA := seedProject(t, db, "a")
	projectB := seedProject(t, db, "b")

	node := "NODE_X"
	if _, err := db.Db.Exec(`INSERT OR REPLACE INTO metadata (key, value) VALUES ('node_id', ?)`, node); err != nil {
		t.Fatalf("failed to override node id: %v", err)
	}
	seedTaskWithNode(t, db, projectA, "task1", 1, node)
	seedTaskWithNode(t, db, projectA, "task2", 1, "NODE_Y")
	seedTaskWithNode(t, db, projectB, "task3", 2, node)
	seedTaskWithNode(t, db, projectB, "task4", 2, "NODE_Z")

	collisions, err := getNumberCollisions(db, "")
	if err != nil {
		t.Fatalf("getNumberCollisions failed: %v", err)
	}
	if len(collisions) != 2 {
		t.Fatalf("expected 2 collisions, got %d", len(collisions))
	}
}

func TestGetNumberCollisionsFilter(t *testing.T) {
	db := openTempDB(t)

	project := seedProject(t, db, "a")
	other := seedProject(t, db, "b")

	if _, err := db.Db.Exec(`INSERT OR REPLACE INTO metadata (key, value) VALUES ('node_id', 'NODE_X')`); err != nil {
		t.Fatalf("failed to override node id: %v", err)
	}
	seedTaskWithNode(t, db, project, "task1", 1, "NODE_X")
	seedTaskWithNode(t, db, project, "task2", 1, "NODE_Y")
	seedTaskWithNode(t, db, other, "task3", 2, "NODE_X")

	collisions, err := getNumberCollisions(db, project)
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

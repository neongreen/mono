package main

import "testing"

func TestResolveTaskReferenceByAlias(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "proj")
	taskUID := seedTask(t, db, projectUID, "task", 7)

	resolved, err := ResolveTaskReference(db, "proj-7")
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if resolved != taskUID {
		t.Fatalf("expected %s, got %s", taskUID, resolved)
	}
}

func TestResolveTaskReferenceCollisionRequiresHint(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "proj")
	nodeA := "NODE_A"
	nodeB := "NODE_B"
	if _, err := db.db.Exec(`INSERT OR REPLACE INTO metadata (key, value) VALUES ('node_id', ?)`, nodeA); err != nil {
		t.Fatalf("failed to override node id: %v", err)
	}
	taskA := seedTaskWithNode(t, db, projectUID, "first", 5, nodeA)
	taskB := seedTaskWithNode(t, db, projectUID, "second", 5, nodeB)

	if _, err := ResolveTaskReference(db, "proj-5"); err == nil {
		t.Fatalf("expected ambiguity error, got none")
	}

	hintB := NodeID(nodeB).Short()
	resolvedB, err := ResolveTaskReference(db, "proj-5-"+hintB)
	if err != nil {
		t.Fatalf("resolve with hint failed: %v", err)
	}
	if resolvedB != taskB {
		t.Fatalf("expected %s, got %s", taskB, resolvedB)
	}

	hintA := NodeID(nodeA).Short()
	resolvedA, err := ResolveTaskReference(db, "proj-5-"+hintA)
	if err != nil {
		t.Fatalf("resolve with hint failed: %v", err)
	}
	if resolvedA != taskA {
		t.Fatalf("expected %s, got %s", taskA, resolvedA)
	}
}

func TestRenderTaskDisplayIDWithCollision(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "proj")
	nodeA := "NODE_A"
	nodeB := "NODE_B"
	if _, err := db.db.Exec(`INSERT OR REPLACE INTO metadata (key, value) VALUES ('node_id', ?)`, nodeA); err != nil {
		t.Fatalf("failed to override node id: %v", err)
	}
	taskA := seedTaskWithNode(t, db, projectUID, "first", 2, nodeA)
	taskB := seedTaskWithNode(t, db, projectUID, "second", 2, nodeB)

	displayA, err := RenderTaskDisplayID(db, taskA)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	displayB, err := RenderTaskDisplayID(db, taskB)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	hintA := NodeID(nodeA).Short()
	hintB := NodeID(nodeB).Short()

	if displayA != "proj-2-"+hintA {
		t.Fatalf("expected display with hint %s, got %s", hintA, displayA)
	}
	if displayB != "proj-2-"+hintB {
		t.Fatalf("expected display with hint %s, got %s", hintB, displayB)
	}
}

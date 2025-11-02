package database

import (
	"testing"

	"github.com/neongreen/mono/tk/internal/types"
)

func TestResolveTaskReferenceByAlias(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "proj")
	taskUID := seedTask(t, db, projectUID, "task", 7)

	resolved, err := ResolveTaskReference(db, types.NewTaskRef("proj-7"))
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
	if _, err := db.Db.Exec(`INSERT OR REPLACE INTO metadata (key, value) VALUES ('node_id', ?)`, nodeA); err != nil {
		t.Fatalf("failed to override node id: %v", err)
	}
	taskA := seedTaskWithNode(t, db, projectUID, "first", 5, nodeA)
	taskB := seedTaskWithNode(t, db, projectUID, "second", 5, nodeB)

	if _, err := ResolveTaskReference(db, types.NewTaskRef("proj-5")); err == nil {
		t.Fatalf("expected ambiguity error, got none")
	}

	hintB := types.NodeID(nodeB).Short()
	resolvedB, err := ResolveTaskReference(db, types.NewTaskRef("proj-5-"+hintB))
	if err != nil {
		t.Fatalf("resolve with hint failed: %v", err)
	}
	if resolvedB != taskB {
		t.Fatalf("expected %s, got %s", taskB, resolvedB)
	}

	hintA := types.NodeID(nodeA).Short()
	resolvedA, err := ResolveTaskReference(db, types.NewTaskRef("proj-5-"+hintA))
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
	if _, err := db.Db.Exec(`INSERT OR REPLACE INTO metadata (key, value) VALUES ('node_id', ?)`, nodeA); err != nil {
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

	hintA := types.NodeID(nodeA).Short()
	hintB := types.NodeID(nodeB).Short()

	if displayA != "proj-2-"+hintA {
		t.Fatalf("expected display with hint %s, got %s", hintA, displayA)
	}
	if displayB != "proj-2-"+hintB {
		t.Fatalf("expected display with hint %s, got %s", hintB, displayB)
	}
}

func TestRenderTaskDisplayIDWithoutAlias(t *testing.T) {
	db := openTempDB(t)

	// Create a project without an alias
	projectUID := seedProjectWithoutAlias(t, db, "My Project")
	taskUID := seedTask(t, db, projectUID, "task", 1)

	displayID, err := RenderTaskDisplayID(db, taskUID)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	// Should use project name, not projectUID
	expected := "My Project-1"
	if displayID != expected {
		t.Fatalf("expected %s, got %s", expected, displayID)
	}
}

package database

import (
	"strings"
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

// TestResolveTaskReferenceEmptyString tests that empty references are rejected
func TestResolveTaskReferenceEmptyString(t *testing.T) {
	db := openTempDB(t)

	_, err := ResolveTaskReference(db, types.NewTaskRef(""))
	if err == nil {
		t.Fatal("expected error for empty reference, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("expected 'cannot be empty' error, got: %v", err)
	}
}

// TestResolveTaskReferenceWhitespaceOnly tests that whitespace-only references are rejected
func TestResolveTaskReferenceWhitespaceOnly(t *testing.T) {
	db := openTempDB(t)

	_, err := ResolveTaskReference(db, types.NewTaskRef("   \t\n  "))
	if err == nil {
		t.Fatal("expected error for whitespace-only reference, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("expected 'cannot be empty' error, got: %v", err)
	}
}

// TestResolveTaskReferencePureNumeric tests that pure numbers are rejected as ambiguous
func TestResolveTaskReferencePureNumeric(t *testing.T) {
	db := openTempDB(t)

	_, err := ResolveTaskReference(db, types.NewTaskRef("123"))
	if err == nil {
		t.Fatal("expected error for numeric reference, got nil")
	}
	if !strings.Contains(err.Error(), "ambiguous") || !strings.Contains(err.Error(), "numeric") {
		t.Errorf("expected ambiguous numeric error, got: %v", err)
	}
}

// TestResolveTaskReferenceNotFound tests that non-existent tasks return appropriate error
func TestResolveTaskReferenceNotFound(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "proj")
	// Create task 1 but not task 99
	_ = seedTask(t, db, projectUID, "task", 1)

	_, err := ResolveTaskReference(db, types.NewTaskRef("proj-99"))
	if err == nil {
		t.Fatal("expected error for non-existent task, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// TestResolveTaskReferenceInvalidNodeHint tests that invalid node hints return appropriate error
func TestResolveTaskReferenceInvalidNodeHint(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "proj")
	nodeA := "NODE_A"
	nodeB := "NODE_B"
	if _, err := db.Db.Exec(`INSERT OR REPLACE INTO metadata (key, value) VALUES ('node_id', ?)`, nodeA); err != nil {
		t.Fatalf("failed to override node id: %v", err)
	}
	_ = seedTaskWithNode(t, db, projectUID, "first", 5, nodeA)
	_ = seedTaskWithNode(t, db, projectUID, "second", 5, nodeB)

	// Use a hint that doesn't match either task
	invalidHint := "xyz123"
	_, err := ResolveTaskReference(db, types.NewTaskRef("proj-5-"+invalidHint))
	if err == nil {
		t.Fatal("expected error for invalid node hint, got nil")
	}
	if !strings.Contains(err.Error(), "node hint") && !strings.Contains(err.Error(), "does not match") {
		t.Errorf("expected node hint error, got: %v", err)
	}
}

// TestResolveTaskReferenceDirectUIDNotFound tests that non-existent UIDs return appropriate error
func TestResolveTaskReferenceDirectUIDNotFound(t *testing.T) {
	db := openTempDB(t)

	_, err := ResolveTaskReference(db, types.NewTaskRef("tsk_01NONEXISTENT12345"))
	if err == nil {
		t.Fatal("expected error for non-existent UID, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// TestResolveTaskReferenceWithWhitespace tests that references with leading/trailing whitespace work
func TestResolveTaskReferenceWithWhitespace(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "proj")
	taskUID := seedTask(t, db, projectUID, "task", 7)

	resolved, err := ResolveTaskReference(db, types.NewTaskRef("  proj-7  \t"))
	if err != nil {
		t.Fatalf("resolve with whitespace failed: %v", err)
	}
	if resolved != taskUID {
		t.Fatalf("expected %s, got %s", taskUID, resolved)
	}
}

// TestResolveTaskReferenceNonexistentProject tests that references to non-existent projects fail
func TestResolveTaskReferenceNonexistentProject(t *testing.T) {
	db := openTempDB(t)

	_, err := ResolveTaskReference(db, types.NewTaskRef("nonexistent-1"))
	if err == nil {
		t.Fatal("expected error for non-existent project, got nil")
	}
	// Should mention the project wasn't found (not just the task)
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention project name, got: %v", err)
	}
}

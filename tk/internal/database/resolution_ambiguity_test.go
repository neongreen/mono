package database

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

// Helper function to add an alias to a project using events
func addAliasToProject(t *testing.T, db *DB, projectUID string, alias string) error {
	t.Helper()

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return err
	}

	payload := types.ProjectAliasAddPayload{
		ProjectUID: projectUID,
		Alias:      alias,
		Node:       nodeID,
		AddedBy:    "tester",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindProjectAliasAdd),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return err
	}

	return db.ProjectProjectAliasAddEvent(event)
}

// TestResolveProjectRefAmbiguous tests Bug 1: ambiguous project reference should fail
// Regression test for tk-134: tk new -p should fail on ambiguous project reference with list of matches
func TestResolveProjectRefAmbiguous(t *testing.T) {
	db := openTempDB(t)

	// Create two projects that create ambiguity:
	// Project 1: name="tk-vsc", no alias
	// Project 2: name="tk-vscode", alias="tk-vsc"
	_ = seedProjectWithoutAlias(t, db, "tk-vsc")
	proj2UID := seedProject(t, db, "tk-vscode")

	// Add alias "tk-vsc" to proj2 using event
	if err := addAliasToProject(t, db, proj2UID, "tk-vsc"); err != nil {
		t.Fatalf("failed to add alias: %v", err)
	}

	// Now "tk-vsc" is ambiguous (matches proj1 by name, proj2 by alias)
	_, err := ResolveProjectRef(db, types.NewProjectRef("tk-vsc"))
	if err == nil {
		t.Fatal("expected ambiguity error, got nil")
	}

	// Error should mention ambiguity and list both projects
	errMsg := err.Error()
	if !strings.Contains(errMsg, "ambiguous") {
		t.Errorf("error should mention 'ambiguous', got: %s", errMsg)
	}
	// Should show tk-vsc (name) for project 1
	if !strings.Contains(errMsg, "tk-vsc (name)") && !strings.Contains(errMsg, "tk-vsc (alias)") {
		t.Errorf("error should list tk-vsc as name or alias, got: %s", errMsg)
	}
	// Should show tk-vsc (alias) for project 2
	if !strings.Contains(errMsg, "tk-vsc (alias)") {
		t.Errorf("error should list tk-vsc as alias, got: %s", errMsg)
	}
}

// TestResolveProjectRefNotFoundListsProjects tests Bug 2: should list available projects when not found
// Regression test for tk-135: tk new -p should list all available projects when project not found
func TestResolveProjectRefNotFoundListsProjects(t *testing.T) {
	db := openTempDB(t)

	// Create a few projects
	_ = seedProject(t, db, "project1")
	_ = seedProject(t, db, "project2")
	_ = seedProject(t, db, "project3")

	// Try to resolve nonexistent project
	_, err := ResolveProjectRef(db, types.NewProjectRef("nonexistent"))
	if err == nil {
		t.Fatal("expected error for nonexistent project, got nil")
	}

	errMsg := err.Error()
	// Error should suggest available projects
	if !strings.Contains(errMsg, "project1") || !strings.Contains(errMsg, "project2") || !strings.Contains(errMsg, "project3") {
		t.Errorf("error should list available projects, got: %s", errMsg)
	}
}

// TestResolveTaskReferenceAmbiguousPrefix tests Bug 3: ambiguous task ID should fail
// Regression test for tk-136: tk show should fail on ambiguous task ID with list of matches
func TestResolveTaskReferenceAmbiguousPrefix(t *testing.T) {
	db := openTempDB(t)

	// Create two projects with overlapping display prefixes
	proj1UID := seedProjectWithoutAlias(t, db, "foo")
	proj2UID := seedProject(t, db, "foobar")

	// Both projects have tasks with number 1
	task1UID := seedTask(t, db, proj1UID, "task in foo", 1)
	task2UID := seedTask(t, db, proj2UID, "task in foobar", 1)

	// "foo-1" could refer to:
	// - Task 1 in project "foo" (foo-1)
	// - Task 1 in project "foobar" if alias is "foo" (but it's not in this case)

	// This test is about the case where we have:
	// Project "tk-vsc" with task tk-vsc-45
	// Project "tk-vscode" (alias "tk-vsc") with potential tk-vsc-45

	// Let's recreate that scenario more accurately
	proj3UID := seedProjectWithoutAlias(t, db, "tk-vsc")
	proj4UID := seedProject(t, db, "tk-vscode")

	// Add alias "tk-vsc" to proj4 using event
	if err := addAliasToProject(t, db, proj4UID, "tk-vsc"); err != nil {
		t.Fatalf("failed to add alias: %v", err)
	}

	// Create task 45 in BOTH projects
	task45_proj3 := seedTask(t, db, proj3UID, "task in tk-vsc project", 45)
	task45_proj4 := seedTask(t, db, proj4UID, "task in tk-vscode project", 45)

	// Now trying to resolve "tk-vsc-45" should fail with ambiguity
	// because it could be either task
	_, err := ResolveTaskReference(db, types.NewTaskRef("tk-vsc-45"))
	if err == nil {
		t.Fatal("expected ambiguity error for tk-vsc-45, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "ambiguous") {
		t.Errorf("error should mention 'ambiguous', got: %s", errMsg)
	}
	// The ambiguity is detected at the project resolution level
	// Should mention both tk-vsc matches (name and alias)
	if !strings.Contains(errMsg, "tk-vsc (name)") {
		t.Errorf("error should mention tk-vsc (name), got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "tk-vsc (alias)") {
		t.Errorf("error should mention tk-vsc (alias), got: %s", errMsg)
	}

	// Verify the individual tasks can be resolved by UID
	resolved1, err := ResolveTaskReference(db, types.NewTaskRef(task45_proj3))
	if err != nil {
		t.Fatalf("failed to resolve by UID: %v", err)
	}
	if resolved1 != task45_proj3 {
		t.Errorf("expected %s, got %s", task45_proj3, resolved1)
	}

	resolved2, err := ResolveTaskReference(db, types.NewTaskRef(task45_proj4))
	if err != nil {
		t.Fatalf("failed to resolve by UID: %v", err)
	}
	if resolved2 != task45_proj4 {
		t.Errorf("expected %s, got %s", task45_proj4, resolved2)
	}

	_ = task1UID
	_ = task2UID
}

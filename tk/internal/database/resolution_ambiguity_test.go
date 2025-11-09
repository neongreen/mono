package database

import (
	"strings"
	"testing"

	"github.com/neongreen/mono/tk/internal/types"
)

// addAliasToProject removed - aliases no longer supported (tk-246)

// TestResolveProjectRefAmbiguous tests Bug 1: ambiguous project reference should fail
// Regression test for tk-134: tk new -p should fail on ambiguous project reference with list of matches
// Updated for tk-246: Now tests multiple projects with same name (aliases removed)
func TestResolveProjectRefAmbiguous(t *testing.T) {
	db := openTempDB(t)

	// Create two projects with the SAME NAME (ambiguous)
	_ = seedProjectWithoutAlias(t, db, "tk")
	_ = seedProjectWithoutAlias(t, db, "tk")

	// Now "tk" is ambiguous (two different projects with same name)
	_, err := ResolveProjectRef(db, types.NewProjectRef("tk"))
	if err == nil {
		t.Fatal("expected ambiguity error, got nil")
	}

	// Error should mention ambiguity and list both project UIDs
	errMsg := err.Error()
	if !strings.Contains(errMsg, "ambiguous") {
		t.Errorf("error should mention 'ambiguous', got: %s", errMsg)
	}
	// Should show project UIDs to disambiguate
	if !strings.Contains(errMsg, "prj_") {
		t.Errorf("error should list project UIDs, got: %s", errMsg)
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
// Updated for tk-246: Now tests ambiguity from duplicate project names
func TestResolveTaskReferenceAmbiguousPrefix(t *testing.T) {
	db := openTempDB(t)

	// Create two projects with THE SAME NAME (ambiguous)
	proj1UID := seedProjectWithoutAlias(t, db, "tk")
	proj2UID := seedProjectWithoutAlias(t, db, "tk")

	// Both projects have task with number 45
	_ = seedTask(t, db, proj1UID, "task in first tk project", 45)
	_ = seedTask(t, db, proj2UID, "task in second tk project", 45)

	// Now trying to resolve "tk-45" should fail with ambiguity
	// because the project name "tk" matches two different projects
	_, err := ResolveTaskReference(db, types.NewTaskRef("tk-45"))
	if err == nil {
		t.Fatal("expected ambiguity error for tk-45, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "ambiguous") {
		t.Errorf("error should mention 'ambiguous', got: %s", errMsg)
	}
	// The ambiguity is detected at the project resolution level
	// Should mention project UIDs to disambiguate
	if !strings.Contains(errMsg, "prj_") {
		t.Errorf("error should list project UIDs for disambiguation, got: %s", errMsg)
	}
}

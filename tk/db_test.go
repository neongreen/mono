package main

import (
	"testing"
)

func TestDB_GetOrCreateNodeID(t *testing.T) {
	db := openTempDB(t)

	nodeID1, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("GetOrCreateNodeID() error = %v", err)
	}

	if nodeID1 == "" {
		t.Error("GetOrCreateNodeID() returned empty nodeID")
	}

	// Second call should return the same node ID
	nodeID2, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("GetOrCreateNodeID() second call error = %v", err)
	}

	if nodeID1 != nodeID2 {
		t.Errorf("GetOrCreateNodeID() returned different IDs: %v != %v", nodeID1, nodeID2)
	}
}

func TestDB_RegenerateNodeID(t *testing.T) {
	db := openTempDB(t)

	// Get initial node ID
	nodeID1, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("GetOrCreateNodeID() error = %v", err)
	}

	// Regenerate
	_, err = db.RegenerateNodeID()
	if err != nil {
		t.Fatalf("RegenerateNodeID() error = %v", err)
	}

	// Get new node ID
	nodeID2, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("GetOrCreateNodeID() after regenerate error = %v", err)
	}

	if nodeID1 == nodeID2 {
		t.Errorf("RegenerateNodeID() did not change node ID: %v == %v", nodeID1, nodeID2)
	}
}

func TestDB_GetNextEventNumber(t *testing.T) {
	db := openTempDB(t)

	num1, err := db.GetNextEventNumber()
	if err != nil {
		t.Fatalf("GetNextEventNumber() error = %v", err)
	}

	if num1 <= 0 {
		t.Errorf("GetNextEventNumber() = %v, want > 0", num1)
	}

	// Second call should increment
	num2, err := db.GetNextEventNumber()
	if err != nil {
		t.Fatalf("GetNextEventNumber() second call error = %v", err)
	}

	if num2 <= num1 {
		t.Errorf("GetNextEventNumber() = %v, want > %v", num2, num1)
	}
}

func TestDB_GetNextLamportTS_and_BumpLamport(t *testing.T) {
	db := openTempDB(t)

	// Get initial lamport
	ts1, err := db.GetNextLamportTS()
	if err != nil {
		t.Fatalf("GetNextLamportTS() error = %v", err)
	}

	if ts1 <= 0 {
		t.Errorf("GetNextLamportTS() = %v, want > 0", ts1)
	}

	// Bump lamport to a specific value
	newTS := ts1 + 100
	if err := db.BumpLamport(newTS); err != nil {
		t.Fatalf("BumpLamport() error = %v", err)
	}

	// Get next lamport - should be at least newTS + 1
	ts2, err := db.GetNextLamportTS()
	if err != nil {
		t.Fatalf("GetNextLamportTS() after bump error = %v", err)
	}

	if ts2 <= newTS {
		t.Errorf("GetNextLamportTS() = %v, want > %v", ts2, newTS)
	}
}

func TestDB_GetAllTaskIDs(t *testing.T) {
	db := openTempDB(t)

	// Create a project with tasks
	projectUID := seedProject(t, db, "test")
	taskUID1 := seedTask(t, db, projectUID, "Task 1", 1)
	taskUID2 := seedTask(t, db, projectUID, "Task 2", 2)

	taskUIDs, err := db.GetAllTaskIDs()
	if err != nil {
		t.Fatalf("GetAllTaskIDs() error = %v", err)
	}

	if len(taskUIDs) < 2 {
		t.Errorf("GetAllTaskIDs() returned %d task UIDs, want at least 2", len(taskUIDs))
	}

	// Check that our task UIDs are in the result
	foundTask1 := false
	foundTask2 := false
	for _, uid := range taskUIDs {
		if uid == taskUID1 {
			foundTask1 = true
		}
		if uid == taskUID2 {
			foundTask2 = true
		}
	}

	if !foundTask1 {
		t.Errorf("GetAllTaskIDs() did not return task1 UID: %v", taskUID1)
	}
	if !foundTask2 {
		t.Errorf("GetAllTaskIDs() did not return task2 UID: %v", taskUID2)
	}
}

func TestDB_GetTaskIDsByProjects(t *testing.T) {
	db := openTempDB(t)

	// Create projects with tasks
	projectUID1 := seedProject(t, db, "proj1")
	projectUID2 := seedProject(t, db, "proj2")
	taskUID1 := seedTask(t, db, projectUID1, "Task 1", 1)
	taskUID2 := seedTask(t, db, projectUID2, "Task 2", 1)

	// Get tasks for proj1
	taskUIDs, err := db.GetTaskIDsByProjects([]string{"proj1"})
	if err != nil {
		t.Fatalf("GetTaskIDsByProjects() error = %v", err)
	}

	if len(taskUIDs) == 0 {
		t.Error("GetTaskIDsByProjects() returned no task UIDs")
	}

	// Should include task1 but not task2
	foundTask1 := false
	foundTask2 := false
	for _, uid := range taskUIDs {
		if uid == taskUID1 {
			foundTask1 = true
		}
		if uid == taskUID2 {
			foundTask2 = true
		}
	}

	if !foundTask1 {
		t.Errorf("GetTaskIDsByProjects() did not return task from proj1: %v", taskUID1)
	}
	if foundTask2 {
		t.Errorf("GetTaskIDsByProjects() unexpectedly returned task from proj2: %v", taskUID2)
	}
}

func TestDB_GetTaskIDsByProjects_AcceptsProjectUID(t *testing.T) {
	db := openTempDB(t)

	projectUID := seedProject(t, db, "proj1")
	taskUID := seedTask(t, db, projectUID, "Task 1", 1)
	seedTask(t, db, seedProject(t, db, "proj2"), "Task 2", 1)

	taskUIDs, err := db.GetTaskIDsByProjects([]string{projectUID})
	if err != nil {
		t.Fatalf("GetTaskIDsByProjects() error = %v", err)
	}

	found := false
	for _, uid := range taskUIDs {
		if uid == taskUID {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("GetTaskIDsByProjects() did not return task for project UID %s", projectUID)
	}
}

// TestFormatTaskID_NoCollision tests that FormatTaskID returns the short form when no collision exists
func TestFormatTaskID_NoCollision(t *testing.T) {
	db := openTempDB(t)
	projectUID := seedProject(t, db, "proj")

	// Create a task with number 1
	taskUID := seedTask(t, db, projectUID, "Task 1", 1)

	// Get the task's display ID
	displayID, err := RenderTaskDisplayID(db, taskUID)
	if err != nil {
		t.Fatalf("failed to get display ID: %v", err)
	}

	// FormatTaskID should return short form since no collision
	formatted := FormatTaskID(db, displayID)
	expected := "proj-1"
	if formatted != expected {
		t.Errorf("FormatTaskID() = %q, want %q", formatted, expected)
	}
}

// TestFormatTaskID_WithCollision tests that FormatTaskID returns the full ID when collision exists
func TestFormatTaskID_WithCollision(t *testing.T) {
	db := openTempDB(t)
	projectUID := seedProject(t, db, "proj")

	nodeID1, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node id: %v", err)
	}

	// Create first task with number 1 on node1
	taskUID1 := seedTaskWithNode(t, db, projectUID, "Task 1 from node1", 1, nodeID1)

	// Simulate a second node by using a different node ID
	nodeID2 := "node02"

	// Create second task with same number 1 on node2 (collision)
	_ = seedTaskWithNode(t, db, projectUID, "Task 1 from node2", 1, nodeID2)

	// Get the first task's display ID
	displayID1, err := RenderTaskDisplayID(db, taskUID1)
	if err != nil {
		t.Fatalf("failed to get display ID: %v", err)
	}

	// FormatTaskID should return full ID since there's a collision
	formatted := FormatTaskID(db, displayID1)
	if formatted != displayID1 {
		t.Errorf("FormatTaskID() = %q, want full ID %q (due to collision)", formatted, displayID1)
	}

	// Verify the formatted ID is longer than short form
	if formatted == "proj-1" {
		t.Error("FormatTaskID() returned short form despite collision")
	}
}

// TestFormatTaskID_MalformedID tests that FormatTaskID returns malformed IDs unchanged
func TestFormatTaskID_MalformedID(t *testing.T) {
	db := openTempDB(t)

	testCases := []struct {
		name string
		id   string
	}{
		{"empty", ""},
		{"no dashes", "invalid"},
		{"single part", "proj"},
		{"only two parts", "proj-1"},
		{"invalid number", "proj-abc-xyz"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatted := FormatTaskID(db, tc.id)
			if formatted != tc.id {
				t.Errorf("FormatTaskID(%q) = %q, want %q (unchanged)", tc.id, formatted, tc.id)
			}
		})
	}
}

// TestFormatTaskID_UnknownAlias tests that FormatTaskID returns full ID when alias cannot be resolved
func TestFormatTaskID_UnknownAlias(t *testing.T) {
	db := openTempDB(t)

	// Create a display ID with unknown alias
	fullID := "unknown-1-abc"

	formatted := FormatTaskID(db, fullID)
	if formatted != fullID {
		t.Errorf("FormatTaskID(%q) = %q, want %q (unknown alias)", fullID, formatted, fullID)
	}
}

// TestFormatTaskID_MultipleTasksSameProject tests formatting with multiple tasks in same project
func TestFormatTaskID_MultipleTasksSameProject(t *testing.T) {
	db := openTempDB(t)
	projectUID := seedProject(t, db, "proj")

	// Create multiple tasks with different numbers (no collisions)
	taskUID1 := seedTask(t, db, projectUID, "Task 1", 1)
	taskUID2 := seedTask(t, db, projectUID, "Task 2", 2)
	taskUID3 := seedTask(t, db, projectUID, "Task 3", 3)

	// Get display IDs
	displayID1, _ := RenderTaskDisplayID(db, taskUID1)
	displayID2, _ := RenderTaskDisplayID(db, taskUID2)
	displayID3, _ := RenderTaskDisplayID(db, taskUID3)

	// All should return short form
	if formatted := FormatTaskID(db, displayID1); formatted != "proj-1" {
		t.Errorf("FormatTaskID(task1) = %q, want %q", formatted, "proj-1")
	}
	if formatted := FormatTaskID(db, displayID2); formatted != "proj-2" {
		t.Errorf("FormatTaskID(task2) = %q, want %q", formatted, "proj-2")
	}
	if formatted := FormatTaskID(db, displayID3); formatted != "proj-3" {
		t.Errorf("FormatTaskID(task3) = %q, want %q", formatted, "proj-3")
	}
}

// TestFormatTaskID_DifferentProjects tests formatting with tasks from different projects
func TestFormatTaskID_DifferentProjects(t *testing.T) {
	db := openTempDB(t)

	// Create two projects
	projectUID1 := seedProject(t, db, "proj1")
	projectUID2 := seedProject(t, db, "proj2")

	// Create tasks with same number in different projects (no collision within each project)
	taskUID1 := seedTask(t, db, projectUID1, "Task in proj1", 1)
	taskUID2 := seedTask(t, db, projectUID2, "Task in proj2", 1)

	// Get display IDs
	displayID1, _ := RenderTaskDisplayID(db, taskUID1)
	displayID2, _ := RenderTaskDisplayID(db, taskUID2)

	// Both should return short form since collisions are per-project
	if formatted := FormatTaskID(db, displayID1); formatted != "proj1-1" {
		t.Errorf("FormatTaskID(proj1 task) = %q, want %q", formatted, "proj1-1")
	}
	if formatted := FormatTaskID(db, displayID2); formatted != "proj2-1" {
		t.Errorf("FormatTaskID(proj2 task) = %q, want %q", formatted, "proj2-1")
	}
}

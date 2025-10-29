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

func TestDB_GetAllPrefixes(t *testing.T) {
	db := openTempDB(t)

	// GetAllPrefixes returns legacy v1/v2 prefixes, not v4 projects
	// Just test that it doesn't error
	prefixes, err := db.GetAllPrefixes()
	if err != nil {
		t.Fatalf("GetAllPrefixes() error = %v", err)
	}

	// Should return empty list for new v4 database
	_ = prefixes
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

func TestDB_GetTaskIDsByPrefixes(t *testing.T) {
	db := openTempDB(t)

	// Create projects with tasks
	projectUID1 := seedProject(t, db, "proj1")
	projectUID2 := seedProject(t, db, "proj2")
	taskUID1 := seedTask(t, db, projectUID1, "Task 1", 1)
	taskUID2 := seedTask(t, db, projectUID2, "Task 2", 1)

	// Get tasks for proj1
	taskUIDs, err := db.GetTaskIDsByPrefixes([]string{"proj1"})
	if err != nil {
		t.Fatalf("GetTaskIDsByPrefixes() error = %v", err)
	}

	if len(taskUIDs) == 0 {
		t.Error("GetTaskIDsByPrefixes() returned no task UIDs")
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
		t.Errorf("GetTaskIDsByPrefixes() did not return task from proj1: %v", taskUID1)
	}
	if foundTask2 {
		t.Errorf("GetTaskIDsByPrefixes() unexpectedly returned task from proj2: %v", taskUID2)
	}
}

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestTaskMove(t *testing.T) {
	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "tk-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "tk.db")
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Create foo prefix (tk is auto-created during InitDB)
	if err := db.CreatePrefix("foo", "Foo prefix", "test"); err != nil {
		t.Fatalf("Failed to create foo prefix: %v", err)
	}

	// Get events and build reducer
	events, err := db.GetEvents()
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	reducer, err := BuildFromEvents(events)
	if err != nil {
		t.Fatalf("Failed to build reducer: %v", err)
	}

	// Verify prefixes were created
	if len(reducer.GetAllTasks()) != 0 {
		t.Fatalf("Expected no tasks, got %d", len(reducer.GetAllTasks()))
	}

	// Create a task in tk prefix
	taskUUID := GenerateTaskUUID()
	taskID, err := GenerateTaskID(db, "tk")
	if err != nil {
		t.Fatalf("Failed to generate task ID: %v", err)
	}
	eventID, err := GenerateEventID(db)
	if err != nil {
		t.Fatalf("Failed to generate event ID: %v", err)
	}
	lamportTS, err := db.GetNextLamportTS()
	if err != nil {
		t.Fatalf("Failed to get lamport timestamp: %v", err)
	}

	// Insert task.created event
	taskCreatedPayload := TaskCreatedPayload{
		TaskUUID:  taskUUID,
		TaskID:    taskID,
		Title:     "Test task",
		CreatedBy: "test",
	}
	taskCreatedEvent := Event{
		ID:      eventID,
		TS:      lamportTS,
		Actor:   "test",
		Role:    "human",
		Kind:    "task.created",
		Payload: mustMarshal(taskCreatedPayload),
	}
	if err := db.InsertEvent(taskCreatedEvent); err != nil {
		t.Fatalf("Failed to insert task.created event: %v", err)
	}

	// Get events and rebuild reducer
	events, err = db.GetEvents()
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}
	reducer, err = BuildFromEvents(events)
	if err != nil {
		t.Fatalf("Failed to build reducer: %v", err)
	}

	// Verify task was created
	task, ok := reducer.GetTask(taskUUID)
	if !ok {
		t.Fatalf("Task not found by UUID")
	}
	if task.TaskID != taskID {
		t.Fatalf("Expected task ID %s, got %s", taskID, task.TaskID)
	}
	if len(task.Aliases) != 0 {
		t.Fatalf("Expected no aliases, got %d", len(task.Aliases))
	}

	// Move the task to foo prefix
	eventID2, err := GenerateEventID(db)
	if err != nil {
		t.Fatalf("Failed to generate event ID: %v", err)
	}
	lamportTS2, err := db.GetNextLamportTS()
	if err != nil {
		t.Fatalf("Failed to get lamport timestamp: %v", err)
	}

	// Parse the old task ID to get prefix, number, and node
	parts := splitTaskID(taskID)
	if len(parts) != 3 {
		t.Fatalf("Invalid task ID format: %s", taskID)
	}
	oldPrefix := parts[0]
	oldNumber := parts[1]
	oldNode := parts[2]

	// Create task.reprefix event
	reprefixPayload := TaskReprefixPayload{
		TaskUUID:  taskUUID,
		OldPrefix: oldPrefix,
		NewPrefix: "foo",
		OldNumber: parseInt64(oldNumber),
		NewNumber: 1,
		OldNode:   oldNode,
		Reason:    "test move",
	}
	reprefixEvent := Event{
		ID:      eventID2,
		TS:      lamportTS2,
		Actor:   "test",
		Role:    "human",
		Kind:    "task.reprefix",
		Payload: mustMarshal(reprefixPayload),
	}
	if err := db.InsertEvent(reprefixEvent); err != nil {
		t.Fatalf("Failed to insert task.reprefix event: %v", err)
	}

	// Create task.alias.added event
	eventID3, err := GenerateEventID(db)
	if err != nil {
		t.Fatalf("Failed to generate event ID: %v", err)
	}
	lamportTS3, err := db.GetNextLamportTS()
	if err != nil {
		t.Fatalf("Failed to get lamport timestamp: %v", err)
	}

	aliasPayload := TaskAliasAddedPayload{
		TaskUUID: taskUUID,
		AliasID:  taskID, // Old task ID becomes an alias
	}
	aliasEvent := Event{
		ID:      eventID3,
		TS:      lamportTS3,
		Actor:   "test",
		Role:    "human",
		Kind:    "task.alias.added",
		Payload: mustMarshal(aliasPayload),
	}
	if err := db.InsertEvent(aliasEvent); err != nil {
		t.Fatalf("Failed to insert task.alias.added event: %v", err)
	}

	// Get events and rebuild reducer
	events, err = db.GetEvents()
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}
	reducer, err = BuildFromEvents(events)
	if err != nil {
		t.Fatalf("Failed to build reducer: %v", err)
	}

	// Verify task was moved
	task, ok = reducer.GetTask(taskUUID)
	if !ok {
		t.Fatalf("Task not found by UUID after move")
	}

	newTaskID := "foo-1-" + oldNode
	if task.TaskID != newTaskID {
		t.Fatalf("Expected task ID %s, got %s", newTaskID, task.TaskID)
	}

	if len(task.Aliases) != 1 {
		t.Fatalf("Expected 1 alias, got %d", len(task.Aliases))
	}
	if task.Aliases[0] != taskID {
		t.Fatalf("Expected alias %s, got %s", taskID, task.Aliases[0])
	}

	// Verify we can look up the task by the old ID (alias)
	taskByAlias, ok := reducer.GetTask(taskID)
	if !ok {
		t.Fatalf("Task not found by alias")
	}
	if taskByAlias.TaskUUID != taskUUID {
		t.Fatalf("Expected task UUID %s, got %s", taskUUID, taskByAlias.TaskUUID)
	}

	// Verify we can look up the task by the new ID
	taskByNewID, ok := reducer.GetTask(newTaskID)
	if !ok {
		t.Fatalf("Task not found by new ID")
	}
	if taskByNewID.TaskUUID != taskUUID {
		t.Fatalf("Expected task UUID %s, got %s", taskUUID, taskByNewID.TaskUUID)
	}
}

func TestPrefixRemove(t *testing.T) {
	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "tk-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "tk.db")
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Create a prefix
	if err := db.CreatePrefix("test", "Test prefix", "test"); err != nil {
		t.Fatalf("Failed to create prefix: %v", err)
	}

	// Verify prefix exists and is not removed
	prefixes, err := db.GetPrefixes()
	if err != nil {
		t.Fatalf("Failed to get prefixes: %v", err)
	}

	// Find the test prefix (skip default tk prefix)
	var testPrefix *Prefix
	for i := range prefixes {
		if prefixes[i].Prefix == "test" {
			testPrefix = &prefixes[i]
			break
		}
	}
	if testPrefix == nil {
		t.Fatalf("Test prefix not found")
	}
	if testPrefix.Removed {
		t.Fatalf("Expected prefix to not be removed")
	}

	// Remove the prefix
	if err := db.RemovePrefix("test", "test"); err != nil {
		t.Fatalf("Failed to remove prefix: %v", err)
	}

	// Verify prefix is marked as removed
	prefixes, err = db.GetPrefixes()
	if err != nil {
		t.Fatalf("Failed to get prefixes: %v", err)
	}

	testPrefix = nil
	for i := range prefixes {
		if prefixes[i].Prefix == "test" {
			testPrefix = &prefixes[i]
			break
		}
	}
	if testPrefix == nil {
		t.Fatalf("Test prefix not found after removal")
	}
	if !testPrefix.Removed {
		t.Fatalf("Expected prefix to be removed")
	}
}

// Helper functions

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func splitTaskID(taskID string) []string {
	return strings.Split(taskID, "-")
}

func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

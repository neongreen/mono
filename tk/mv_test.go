package main

import (
	"encoding/json"
	"fmt"
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

func TestMovingByAlias(t *testing.T) {
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

	// Create prefixes
	if err := db.CreatePrefix("foo", "Foo prefix", "test"); err != nil {
		t.Fatalf("Failed to create foo prefix: %v", err)
	}
	if err := db.CreatePrefix("bar", "Bar prefix", "test"); err != nil {
		t.Fatalf("Failed to create bar prefix: %v", err)
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

	// Move task to foo prefix (tk-1 -> foo-1)
	parts := splitTaskID(taskID)
	oldNode := parts[2]

	eventID2, _ := GenerateEventID(db)
	lamportTS2, _ := db.GetNextLamportTS()
	reprefixPayload := TaskReprefixPayload{
		TaskUUID:  taskUUID,
		OldPrefix: "tk",
		NewPrefix: "foo",
		OldNumber: 1,
		NewNumber: 1,
		OldNode:   oldNode,
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

	// Add alias
	eventID3, _ := GenerateEventID(db)
	lamportTS3, _ := db.GetNextLamportTS()
	aliasPayload := TaskAliasAddedPayload{
		TaskUUID: taskUUID,
		AliasID:  taskID,
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

	// Verify we can resolve by alias
	resolvedUUID, err := db.ResolveTaskIDToUUID(taskID)
	if err != nil {
		t.Fatalf("Failed to resolve alias to UUID: %v", err)
	}
	if resolvedUUID != taskUUID {
		t.Fatalf("Expected UUID %s, got %s", taskUUID, resolvedUUID)
	}

	// Verify we can also resolve by current ID
	newTaskID := "foo-1-" + oldNode
	resolvedUUID2, err := db.ResolveTaskIDToUUID(newTaskID)
	if err != nil {
		t.Fatalf("Failed to resolve current ID to UUID: %v", err)
	}
	if resolvedUUID2 != taskUUID {
		t.Fatalf("Expected UUID %s, got %s", taskUUID, resolvedUUID2)
	}
}

func TestDuplicateTaskCreated(t *testing.T) {
	reducer := NewReducer()

	taskUUID := "task-test123"
	taskID := "tk-1-node123"

	// First task.created
	payload1 := TaskCreatedPayload{
		TaskUUID:  taskUUID,
		TaskID:    taskID,
		Title:     "First title",
		CreatedBy: "test",
	}
	event1 := Event{
		ID:      "ev-1-node123",
		TS:      1,
		Kind:    "task.created",
		Payload: mustMarshal(payload1),
	}
	if err := reducer.Apply(event1); err != nil {
		t.Fatalf("Failed to apply first task.created: %v", err)
	}

	// Duplicate task.created with same UUID should be ignored
	payload2 := TaskCreatedPayload{
		TaskUUID:  taskUUID,
		TaskID:    taskID,
		Title:     "Second title (should be ignored)",
		CreatedBy: "test",
	}
	event2 := Event{
		ID:      "ev-2-node123",
		TS:      2,
		Kind:    "task.created",
		Payload: mustMarshal(payload2),
	}
	if err := reducer.Apply(event2); err != nil {
		t.Fatalf("Failed to apply duplicate task.created: %v", err)
	}

	// Verify the title is still from the first event
	task, ok := reducer.GetTask(taskUUID)
	if !ok {
		t.Fatalf("Task not found")
	}
	if task.Title != "First title" {
		t.Fatalf("Expected title 'First title', got '%s'", task.Title)
	}
}

func TestDuplicateAliasAdded(t *testing.T) {
	reducer := NewReducer()

	taskUUID := "task-test123"
	taskID := "tk-1-node123"
	aliasID := "old-1-node123"

	// Create task
	payload := TaskCreatedPayload{
		TaskUUID:  taskUUID,
		TaskID:    taskID,
		Title:     "Test task",
		CreatedBy: "test",
	}
	event := Event{
		ID:      "ev-1-node123",
		TS:      1,
		Kind:    "task.created",
		Payload: mustMarshal(payload),
	}
	if err := reducer.Apply(event); err != nil {
		t.Fatalf("Failed to apply task.created: %v", err)
	}

	// Add alias
	aliasPayload1 := TaskAliasAddedPayload{
		TaskUUID: taskUUID,
		AliasID:  aliasID,
	}
	aliasEvent1 := Event{
		ID:      "ev-2-node123",
		TS:      2,
		Kind:    "task.alias.added",
		Payload: mustMarshal(aliasPayload1),
	}
	if err := reducer.Apply(aliasEvent1); err != nil {
		t.Fatalf("Failed to apply alias.added: %v", err)
	}

	// Add duplicate alias (e.g., from sync arriving out of order)
	aliasPayload2 := TaskAliasAddedPayload{
		TaskUUID: taskUUID,
		AliasID:  aliasID,
	}
	aliasEvent2 := Event{
		ID:      "ev-3-node123",
		TS:      3,
		Kind:    "task.alias.added",
		Payload: mustMarshal(aliasPayload2),
	}
	if err := reducer.Apply(aliasEvent2); err != nil {
		t.Fatalf("Failed to apply duplicate alias.added: %v", err)
	}

	// Verify alias appears only once
	task, ok := reducer.GetTask(taskUUID)
	if !ok {
		t.Fatalf("Task not found")
	}
	if len(task.Aliases) != 1 {
		t.Fatalf("Expected 1 alias, got %d", len(task.Aliases))
	}
	if task.Aliases[0] != aliasID {
		t.Fatalf("Expected alias %s, got %s", aliasID, task.Aliases[0])
	}
}

func TestAutoNumberAvoidsCollision(t *testing.T) {
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

	// Create tak prefix
	if err := db.CreatePrefix("tak", "Tak prefix", "test"); err != nil {
		t.Fatalf("Failed to create tak prefix: %v", err)
	}

	// Simulate a task from another node (e.g., from sync)
	// Create tk-10-nodeB (from a different node)
	otherTaskUUID := GenerateTaskUUID()
	otherTaskID := "tk-10-nodeB"
	otherEventID, _ := GenerateEventID(db)
	otherLamportTS, _ := db.GetNextLamportTS()

	otherTaskCreatedPayload := TaskCreatedPayload{
		TaskUUID:  otherTaskUUID,
		TaskID:    otherTaskID,
		Title:     "Task from other node",
		CreatedBy: "other",
	}
	otherTaskCreatedEvent := Event{
		ID:      otherEventID,
		TS:      otherLamportTS,
		Actor:   "other",
		Role:    "human",
		Kind:    "task.created",
		Payload: mustMarshal(otherTaskCreatedPayload),
	}
	if err := db.InsertEvent(otherTaskCreatedEvent); err != nil {
		t.Fatalf("Failed to insert other task.created event: %v", err)
	}

	// Create some local tk tasks to advance the counter to 9
	// This way, GetNextTaskNumberForPrefix will return 10
	for i := 1; i <= 9; i++ {
		localTkTaskUUID := GenerateTaskUUID()
		localTkTaskID, err := GenerateTaskID(db, "tk")
		if err != nil {
			t.Fatalf("Failed to generate tk task ID: %v", err)
		}
		localTkEventID, _ := GenerateEventID(db)
		localTkLamportTS, _ := db.GetNextLamportTS()

		localTkTaskCreatedPayload := TaskCreatedPayload{
			TaskUUID:  localTkTaskUUID,
			TaskID:    localTkTaskID,
			Title:     fmt.Sprintf("Local tk task %d", i),
			CreatedBy: "test",
		}
		localTkTaskCreatedEvent := Event{
			ID:      localTkEventID,
			TS:      localTkLamportTS,
			Actor:   "test",
			Role:    "human",
			Kind:    "task.created",
			Payload: mustMarshal(localTkTaskCreatedPayload),
		}
		if err := db.InsertEvent(localTkTaskCreatedEvent); err != nil {
			t.Fatalf("Failed to insert local tk task.created event: %v", err)
		}
	}

	// Create our local task tak-10 from local node
	localTaskUUID := GenerateTaskUUID()
	localTaskID, err := GenerateTaskID(db, "tak")
	if err != nil {
		t.Fatalf("Failed to generate task ID: %v", err)
	}
	// Manually set it to tak-10 to match the scenario
	parts := splitTaskID(localTaskID)
	localNode := parts[2]
	localTaskID = "tak-10-" + localNode

	localEventID, _ := GenerateEventID(db)
	localLamportTS, _ := db.GetNextLamportTS()

	localTaskCreatedPayload := TaskCreatedPayload{
		TaskUUID:  localTaskUUID,
		TaskID:    localTaskID,
		Title:     "Local task to move",
		CreatedBy: "test",
	}
	localTaskCreatedEvent := Event{
		ID:      localEventID,
		TS:      localLamportTS,
		Actor:   "test",
		Role:    "human",
		Kind:    "task.created",
		Payload: mustMarshal(localTaskCreatedPayload),
	}
	if err := db.InsertEvent(localTaskCreatedEvent); err != nil {
		t.Fatalf("Failed to insert local task.created event: %v", err)
	}

	// Build reducer to get current state
	events, err := db.GetEvents()
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}
	reducer, err := BuildFromEvents(events)
	if err != nil {
		t.Fatalf("Failed to build reducer: %v", err)
	}

	// Verify both tasks exist
	_, ok := reducer.GetTask(otherTaskUUID)
	if !ok {
		t.Fatalf("Other task not found")
	}
	_, ok = reducer.GetTask(localTaskUUID)
	if !ok {
		t.Fatalf("Local task not found")
	}

	// Now simulate: tk mv tak-10 tk (with auto-number)
	// This should choose a number that doesn't collide with tk-10-nodeB
	spec := moveSpec{
		oldID:      localTaskID,
		newPrefix:  "tk",
		autoNumber: true,
		addAlias:   true,
	}

	reserved := make(map[string]struct{})
	entry, err := planMove(db, reducer, localNode, spec, reserved)
	if err != nil {
		t.Fatalf("Failed to plan move: %v", err)
	}

	t.Logf("Old task ID: %s", entry.oldID)
	t.Logf("New task ID: %s", entry.newID)
	t.Logf("New number: %d", entry.newNumber)
	t.Logf("Other task ID: %s", otherTaskID)

	// The new number should NOT be 10 (because that would collide with tk-10-nodeB)
	// It should be a different number that avoids collision
	if entry.newNumber == 10 {
		t.Errorf("Auto-number chose 10, which collides with tk-10-nodeB. Expected a different number.")
	}

	// Verify the new ID doesn't cause a display collision
	// When formatted with FormatTaskID, it should show without node suffix
	allTaskIDs := []string{otherTaskID, entry.newID}
	formattedNewID := FormatTaskID(entry.newID, allTaskIDs)
	formattedOtherID := FormatTaskID(otherTaskID, allTaskIDs)

	// Both should be displayed without node suffix since they don't collide
	if strings.Contains(formattedNewID, localNode) {
		t.Errorf("New task ID %s is displayed with node suffix %s, indicating collision", formattedNewID, localNode)
	}
	if strings.Contains(formattedOtherID, "nodeB") {
		t.Errorf("Other task ID %s is displayed with node suffix, indicating collision", formattedOtherID)
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

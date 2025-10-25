package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidatePrefixName(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		wantErr bool
		errMsg  string
	}{
		{"valid lowercase", "backend", false, ""},
		{"valid with digits", "backend2", false, ""},
		{"valid with underscore", "backend_api", false, ""},
		{"too short", "a", true, "2-20 characters"},
		{"too long", "verylongprefixnameabcd", true, "2-20 characters"},
		{"starts with number", "2backend", true, "must start with a lowercase letter"},
		{"starts with uppercase", "Backend", true, "must start with a lowercase letter"},
		{"contains hyphen", "back-end", true, "lowercase letters, digits, and underscores"},
		{"contains space", "back end", true, "lowercase letters, digits, and underscores"},
		{"reserved ev", "ev", true, "reserved"},
		{"reserved event", "event", true, "reserved"},
		{"reserved task", "task", true, "reserved"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePrefixName(tt.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePrefixName(%q) error = %v, wantErr %v", tt.prefix, err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidatePrefixName(%q) error = %v, want error containing %q", tt.prefix, err, tt.errMsg)
			}
		})
	}
}

func TestPrefixCreatedEvent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tk.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// Create a prefix
	err = db.CreatePrefix("backend", "Backend tasks", "test-user")
	if err != nil {
		t.Fatalf("failed to create prefix: %v", err)
	}

	// Check that prefix.created event was created
	events, err := db.GetEvents()
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}

	var foundPrefixCreated bool
	for _, e := range events {
		if e.Kind == "prefix.created" {
			foundPrefixCreated = true

			// Validate payload
			var payload PrefixCreatedPayload
			if err := json.Unmarshal(e.Payload, &payload); err != nil {
				t.Fatalf("failed to unmarshal payload: %v", err)
			}

			if payload.Prefix != "backend" {
				t.Errorf("expected prefix 'backend', got %q", payload.Prefix)
			}
			if payload.Description != "Backend tasks" {
				t.Errorf("expected description 'Backend tasks', got %q", payload.Description)
			}
			if payload.CreatedBy != "test-user" {
				t.Errorf("expected created_by 'test-user', got %q", payload.CreatedBy)
			}
			break
		}
	}

	if !foundPrefixCreated {
		t.Error("prefix.created event was not created")
	}
}

func TestPrefixNormalization(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tk.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// Create a prefix with uppercase letters
	err = db.CreatePrefix("Backend", "Backend tasks", "test-user")
	if err != nil {
		t.Fatalf("failed to create prefix: %v", err)
	}

	// Check that it was normalized to lowercase
	prefixes, err := db.GetPrefixes()
	if err != nil {
		t.Fatalf("failed to get prefixes: %v", err)
	}

	found := false
	for _, p := range prefixes {
		if p.Prefix == "backend" {
			found = true
			break
		}
	}

	if !found {
		t.Error("prefix was not normalized to lowercase")
	}
}

func TestProjectPrefixCreatedEvent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tk.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node ID: %v", err)
	}

	// Create a prefix.created event
	payload := PrefixCreatedPayload{
		Prefix:      "frontend",
		Description: "Frontend tasks",
		CreatedBy:   "remote-user",
	}
	payloadJSON, _ := json.Marshal(payload)

	event := Event{
		ID:        "ev-1-" + nodeID,
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "remote-user",
		Role:      "human",
		Kind:      "prefix.created",
		Payload:   payloadJSON,
	}

	// Insert the event
	if err := db.InsertEvent(event); err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	// Project it
	if err := db.ProjectPrefixCreatedEvent(event); err != nil {
		t.Fatalf("failed to project event: %v", err)
	}

	// Check that the prefix was created
	prefixes, err := db.GetPrefixes()
	if err != nil {
		t.Fatalf("failed to get prefixes: %v", err)
	}

	found := false
	for _, p := range prefixes {
		if p.Prefix == "frontend" && p.Description == "Frontend tasks" {
			found = true
			break
		}
	}

	if !found {
		t.Error("prefix was not projected from event")
	}
}

func TestPrefixCreation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tk.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// Default "tk" prefix should exist after initialization
	prefixes, err := db.GetPrefixes()
	if err != nil {
		t.Fatalf("failed to get prefixes: %v", err)
	}

	if len(prefixes) != 1 {
		t.Fatalf("expected 1 default prefix, got %d", len(prefixes))
	}

	if prefixes[0].Prefix != "tk" {
		t.Errorf("expected default prefix 'tk', got %q", prefixes[0].Prefix)
	}

	// Create a new prefix
	err = db.CreatePrefix("foo", "Test prefix", "test-user")
	if err != nil {
		t.Fatalf("failed to create prefix: %v", err)
	}

	// Should now have 2 prefixes
	prefixes, err = db.GetPrefixes()
	if err != nil {
		t.Fatalf("failed to get prefixes: %v", err)
	}

	if len(prefixes) != 2 {
		t.Fatalf("expected 2 prefixes, got %d", len(prefixes))
	}

	// Try to create duplicate prefix
	err = db.CreatePrefix("foo", "Duplicate", "test-user")
	if err == nil {
		t.Error("expected error when creating duplicate prefix, got nil")
	}
}

func TestPrefixExists(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tk.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// Check default prefix exists
	exists, err := db.PrefixExists("tk")
	if err != nil {
		t.Fatalf("failed to check prefix existence: %v", err)
	}
	if !exists {
		t.Error("expected 'tk' prefix to exist")
	}

	// Check non-existent prefix
	exists, err = db.PrefixExists("nonexistent")
	if err != nil {
		t.Fatalf("failed to check prefix existence: %v", err)
	}
	if exists {
		t.Error("expected 'nonexistent' prefix to not exist")
	}
}

func TestPrefixCounters(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tk.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// Create a new prefix
	err = db.CreatePrefix("foo", "Test prefix", "test-user")
	if err != nil {
		t.Fatalf("failed to create prefix: %v", err)
	}

	// Get next task number for "tk" prefix
	tkNum1, err := db.GetNextTaskNumberForPrefix("tk")
	if err != nil {
		t.Fatalf("failed to get next task number for 'tk': %v", err)
	}

	tkNum2, err := db.GetNextTaskNumberForPrefix("tk")
	if err != nil {
		t.Fatalf("failed to get next task number for 'tk': %v", err)
	}

	if tkNum2 != tkNum1+1 {
		t.Errorf("expected tk counter to increment, got %d and %d", tkNum1, tkNum2)
	}

	// Get next task number for "foo" prefix
	fooNum1, err := db.GetNextTaskNumberForPrefix("foo")
	if err != nil {
		t.Fatalf("failed to get next task number for 'foo': %v", err)
	}

	if fooNum1 != 1 {
		t.Errorf("expected first foo task number to be 1, got %d", fooNum1)
	}

	// Verify "tk" counter is independent
	tkNum3, err := db.GetNextTaskNumberForPrefix("tk")
	if err != nil {
		t.Fatalf("failed to get next task number for 'tk': %v", err)
	}

	if tkNum3 != tkNum2+1 {
		t.Errorf("expected tk counter to continue independently, got %d", tkNum3)
	}
}

func TestGenerateTaskIDWithPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tk.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// Create a prefix
	err = db.CreatePrefix("test", "Test prefix", "test-user")
	if err != nil {
		t.Fatalf("failed to create prefix: %v", err)
	}

	// Generate task ID with "test" prefix
	taskID, err := GenerateTaskID(db, "test")
	if err != nil {
		t.Fatalf("failed to generate task ID: %v", err)
	}

	// Check format: test-<number>-<node>
	if len(taskID) < 10 {
		t.Errorf("task ID too short: %s", taskID)
	}

	// Should start with "test-"
	if taskID[:5] != "test-" {
		t.Errorf("expected task ID to start with 'test-', got %s", taskID)
	}
}

func TestResolveTaskIDWithPrefixes(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tk.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// Create prefixes
	err = db.CreatePrefix("foo", "Foo prefix", "test-user")
	if err != nil {
		t.Fatalf("failed to create foo prefix: %v", err)
	}

	err = db.CreatePrefix("bar", "Bar prefix", "test-user")
	if err != nil {
		t.Fatalf("failed to create bar prefix: %v", err)
	}

	// Generate task IDs
	tkTaskID, err := GenerateTaskID(db, "tk")
	if err != nil {
		t.Fatalf("failed to generate tk task ID: %v", err)
	}

	fooTaskID, err := GenerateTaskID(db, "foo")
	if err != nil {
		t.Fatalf("failed to generate foo task ID: %v", err)
	}

	// Insert task.created events
	lamportTS, _ := db.GetNextLamportTS()
	eventID1, _ := GenerateEventID(db)
	event1 := Event{
		ID:        eventID1,
		TS:        lamportTS,
		CreatedAt: testTime(),
		Actor:     "test-user",
		Role:      "human",
		Kind:      "task.created",
		Payload:   []byte(`{"task_id":"` + tkTaskID + `","title":"tk task","created_by":"test-user"}`),
	}
	db.InsertEvent(event1)

	lamportTS, _ = db.GetNextLamportTS()
	eventID2, _ := GenerateEventID(db)
	event2 := Event{
		ID:        eventID2,
		TS:        lamportTS,
		CreatedAt: testTime(),
		Actor:     "test-user",
		Role:      "human",
		Kind:      "task.created",
		Payload:   []byte(`{"task_id":"` + fooTaskID + `","title":"foo task","created_by":"test-user"}`),
	}
	db.InsertEvent(event2)

	// Test resolving with full ID
	resolved, err := db.ResolveTaskID(tkTaskID)
	if err != nil {
		t.Fatalf("failed to resolve full task ID: %v", err)
	}
	if resolved != tkTaskID {
		t.Errorf("expected %s, got %s", tkTaskID, resolved)
	}

	// Test resolving with prefix-number format
	resolved, err = db.ResolveTaskID("foo-1")
	if err != nil {
		t.Fatalf("failed to resolve foo-1: %v", err)
	}
	if resolved != fooTaskID {
		t.Errorf("expected %s, got %s", fooTaskID, resolved)
	}

	// Test resolving with just number (should work if unambiguous)
	resolved, err = db.ResolveTaskID("1")
	if err == nil {
		// If it resolves, it should match one of the tasks
		if resolved != tkTaskID && resolved != fooTaskID {
			t.Errorf("unexpected resolution for '1': %s", resolved)
		}
	}
}

func TestGetTaskIDsByPrefixes(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tk.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// Create prefixes
	err = db.CreatePrefix("foo", "Foo prefix", "test-user")
	if err != nil {
		t.Fatalf("failed to create foo prefix: %v", err)
	}

	err = db.CreatePrefix("bar", "Bar prefix", "test-user")
	if err != nil {
		t.Fatalf("failed to create bar prefix: %v", err)
	}

	// Generate task IDs
	tkTaskID, _ := GenerateTaskID(db, "tk")
	fooTaskID, _ := GenerateTaskID(db, "foo")
	barTaskID, _ := GenerateTaskID(db, "bar")

	// Insert task.created events
	for i, taskID := range []string{tkTaskID, fooTaskID, barTaskID} {
		lamportTS, _ := db.GetNextLamportTS()
		eventID, _ := GenerateEventID(db)
		event := Event{
			ID:        eventID,
			TS:        lamportTS,
			CreatedAt: testTime().Add(time.Duration(i) * time.Second),
			Actor:     "test-user",
			Role:      "human",
			Kind:      "task.created",
			Payload:   []byte(`{"task_id":"` + taskID + `","title":"task","created_by":"test-user"}`),
		}
		db.InsertEvent(event)
	}

	// Test getting all tasks (no prefix filter)
	allTasks, err := db.GetTaskIDsByPrefixes([]string{})
	if err != nil {
		t.Fatalf("failed to get all tasks: %v", err)
	}
	if len(allTasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(allTasks))
	}

	// Test filtering by single prefix
	fooTasks, err := db.GetTaskIDsByPrefixes([]string{"foo"})
	if err != nil {
		t.Fatalf("failed to get foo tasks: %v", err)
	}
	if len(fooTasks) != 1 {
		t.Errorf("expected 1 foo task, got %d", len(fooTasks))
	}

	// Test filtering by multiple prefixes
	tkFooTasks, err := db.GetTaskIDsByPrefixes([]string{"tk", "foo"})
	if err != nil {
		t.Fatalf("failed to get tk and foo tasks: %v", err)
	}
	if len(tkFooTasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tkFooTasks))
	}
}

func TestResolveTaskIDMultiNodeAmbiguity(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "tk.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// Create a prefix
	err = db.CreatePrefix("foo", "Foo prefix", "test-user")
	if err != nil {
		t.Fatalf("failed to create foo prefix: %v", err)
	}

	// Simulate two different nodes creating tasks with same prefix-number
	// This would happen after syncing tasks from another node
	lamportTS1, _ := db.GetNextLamportTS()
	eventID1, _ := GenerateEventID(db)
	event1 := Event{
		ID:        eventID1,
		TS:        lamportTS1,
		CreatedAt: testTime(),
		Actor:     "test-user",
		Role:      "human",
		Kind:      "task.created",
		Payload:   []byte(`{"task_id":"foo-1-nodeA","title":"task from node A","created_by":"test-user"}`),
	}
	db.InsertEvent(event1)

	lamportTS2, _ := db.GetNextLamportTS()
	eventID2, _ := GenerateEventID(db)
	event2 := Event{
		ID:        eventID2,
		TS:        lamportTS2,
		CreatedAt: testTime(),
		Actor:     "test-user",
		Role:      "human",
		Kind:      "task.created",
		Payload:   []byte(`{"task_id":"foo-1-nodeB","title":"task from node B","created_by":"test-user"}`),
	}
	db.InsertEvent(event2)

	// Try to resolve with just number "1" - should error about ambiguity
	_, err = db.ResolveTaskID("1")
	if err == nil {
		t.Error("expected error for ambiguous numeric ID, got nil")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("expected 'ambiguous' in error, got: %v", err)
	}

	// Try to resolve with "foo-1" - should error about multiple nodes
	_, err = db.ResolveTaskID("foo-1")
	if err == nil {
		t.Error("expected error for multi-node ambiguity, got nil")
	}
	if !strings.Contains(err.Error(), "ambiguous") || !strings.Contains(err.Error(), "multiple nodes") {
		t.Errorf("expected error about multiple nodes, got: %v", err)
	}

	// Full ID should work
	resolved, err := db.ResolveTaskID("foo-1-nodeA")
	if err != nil {
		t.Fatalf("failed to resolve full ID: %v", err)
	}
	if resolved != "foo-1-nodeA" {
		t.Errorf("expected foo-1-nodeA, got %s", resolved)
	}
}

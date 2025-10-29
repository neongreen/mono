package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

// TestFullSyncWorkflow tests the complete sync workflow between two machines
// using a directory remote, simulating the real-world use case.
func TestFullSyncWorkflow(t *testing.T) {
	tmpDir := t.TempDir()

	// Setup machine A
	dbPathA := filepath.Join(tmpDir, "machine-a", "tk.db")
	dbA, err := OpenDB(dbPathA)
	if err != nil {
		t.Fatalf("failed to open database A: %v", err)
	}
	defer dbA.Close()

	if err := dbA.InitDB(); err != nil {
		t.Fatalf("failed to initialize database A: %v", err)
	}

	nodeA, err := dbA.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node ID for machine A: %v", err)
	}

	// Setup machine B
	dbPathB := filepath.Join(tmpDir, "machine-b", "tk.db")
	dbB, err := OpenDB(dbPathB)
	if err != nil {
		t.Fatalf("failed to open database B: %v", err)
	}
	defer dbB.Close()

	if err := dbB.InitDB(); err != nil {
		t.Fatalf("failed to initialize database B: %v", err)
	}

	nodeB, err := dbB.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node ID for machine B: %v", err)
	}

	// Ensure nodes are different
	if nodeA == nodeB {
		t.Fatal("machine A and B should have different node IDs")
	}

	// Setup shared remote directory (simulating iCloud, Dropbox, etc.)
	remotePath := filepath.Join(tmpDir, "remote")

	// Machine A: Create prefixes
	t.Log("Machine A: Creating prefixes...")
	if err := dbA.CreatePrefix("work", "Work-related tasks", "alice"); err != nil {
		t.Fatalf("failed to create prefix 'work' on machine A: %v", err)
	}

	if err := dbA.CreatePrefix("personal", "Personal tasks", "alice"); err != nil {
		t.Fatalf("failed to create prefix 'personal' on machine A: %v", err)
	}

	// Machine A: Create some tasks
	t.Log("Machine A: Creating tasks...")
	taskUUID1, err := GenerateTaskUUID()
	if err != nil {
		t.Fatalf("failed to generate task UUID: %v", err)
	}
	taskID1, err := GenerateTaskID(dbA, "work")
	if err != nil {
		t.Fatalf("failed to generate task ID: %v", err)
	}

	eventID1, err := GenerateEventID(dbA)
	if err != nil {
		t.Fatalf("failed to generate event ID: %v", err)
	}

	lamport1, err := dbA.GetNextLamportTS()
	if err != nil {
		t.Fatalf("failed to get lamport timestamp: %v", err)
	}

	payload1 := TaskCreatedPayload{
		TaskUUID:  taskUUID1,
		TaskID:    taskID1,
		Title:     "Implement feature X",
		CreatedBy: "alice",
	}
	payloadJSON1, _ := json.Marshal(payload1)

	event1 := Event{
		ID:        eventID1,
		TS:        lamport1,
		CreatedAt: testTime(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.created",
		Payload:   payloadJSON1,
	}

	if err := dbA.InsertEvent(event1); err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	// Machine A: Export all events to remote
	t.Log("Machine A: Exporting events to remote...")
	remoteConfig := RemoteConfig{
		Type: "folder",
		Path: remotePath,
		Pull: true,
		Push: true,
	}

	// Get all events from machine A
	eventsA, err := dbA.GetEvents()
	if err != nil {
		t.Fatalf("failed to get events from machine A: %v", err)
	}

	space := "personal"
	writer := NewSegmentWriter(remotePath, space, nodeA, 1, 2_000_000, 120)
	for _, event := range eventsA {
		segEvent, err := eventToSegmentEvent(event, space, nodeA)
		if err != nil {
			t.Fatalf("failed to convert event to segment event: %v", err)
		}
		writer.AddEvent(segEvent)
	}

	segInfo, err := writer.WriteSegment()
	if err != nil {
		t.Fatalf("failed to write segment: %v", err)
	}

	if segInfo == nil {
		t.Fatal("expected segment info, got nil")
	}

	// Machine B: Ingest from remote (simulating `tk sync`)
	t.Log("Machine B: Ingesting events from remote...")
	if err := ingestRemote(dbB, "test-remote", remoteConfig); err != nil {
		t.Fatalf("failed to ingest on machine B: %v", err)
	}

	// Machine B: Verify prefixes are available with full metadata
	t.Log("Machine B: Verifying prefixes...")
	prefixesB, err := dbB.GetAllPrefixes()
	if err != nil {
		t.Fatalf("failed to get prefixes on machine B: %v", err)
	}

	foundWork := false
	foundPersonal := false
	for _, p := range prefixesB {
		t.Logf("Found prefix on machine B: %s (node: %s, desc: %s, created_at: %v)",
			p.Prefix, p.Node, p.Description, p.CreatedAt)

		if p.Prefix == "work" && p.Node == nodeA {
			foundWork = true
			if p.Description != "Work-related tasks" {
				t.Errorf("expected description 'Work-related tasks', got %s", p.Description)
			}
			if p.CreatedBy != "alice" {
				t.Errorf("expected created_by 'alice', got %s", p.CreatedBy)
			}
			if p.CreatedAt.IsZero() {
				t.Error("expected non-zero created_at for synced prefix 'work'")
			}
		}

		if p.Prefix == "personal" && p.Node == nodeA {
			foundPersonal = true
			if p.Description != "Personal tasks" {
				t.Errorf("expected description 'Personal tasks', got %s", p.Description)
			}
			if p.CreatedAt.IsZero() {
				t.Error("expected non-zero created_at for synced prefix 'personal'")
			}
		}
	}

	if !foundWork {
		t.Error("prefix 'work' from machine A not found on machine B after sync")
	}
	if !foundPersonal {
		t.Error("prefix 'personal' from machine A not found on machine B after sync")
	}

	// Machine B: Verify that GetPrefixes() only returns local node prefixes
	localPrefixesB, err := dbB.GetPrefixes()
	if err != nil {
		t.Fatalf("failed to get local prefixes on machine B: %v", err)
	}

	// Should NOT find work/personal in local prefixes since they're from a different node
	for _, p := range localPrefixesB {
		if (p.Prefix == "work" || p.Prefix == "personal") && p.Node == nodeA {
			t.Errorf("prefix %s from node A should not appear in local prefix list on machine B", p.Prefix)
		}
	}

	// Machine B: Verify tasks are also synced
	t.Log("Machine B: Verifying tasks...")
	eventsB, err := dbB.GetEvents()
	if err != nil {
		t.Fatalf("failed to get events from machine B: %v", err)
	}

	reducer, err := BuildFromEvents(eventsB)
	if err != nil {
		t.Fatalf("failed to build reducer: %v", err)
	}

	task, ok := reducer.GetTask(taskUUID1)
	if !ok {
		t.Errorf("task %s not found on machine B", taskUUID1)
	} else {
		if task.Title != "Implement feature X" {
			t.Errorf("expected title 'Implement feature X', got %s", task.Title)
		}
		if task.TaskID != taskID1 {
			t.Errorf("expected task ID %s, got %s", taskID1, task.TaskID)
		}
	}

	// Machine B: Create its own prefix and task
	t.Log("Machine B: Creating local prefix and task...")
	if err := dbB.CreatePrefix("bugs", "Bug tracking", "bob"); err != nil {
		t.Fatalf("failed to create prefix 'bugs' on machine B: %v", err)
	}

	taskUUID2, err := GenerateTaskUUID()
	if err != nil {
		t.Fatalf("failed to generate task UUID on machine B: %v", err)
	}
	taskID2, err := GenerateTaskID(dbB, "bugs")
	if err != nil {
		t.Fatalf("failed to generate task ID on machine B: %v", err)
	}

	eventID2, err := GenerateEventID(dbB)
	if err != nil {
		t.Fatalf("failed to generate event ID on machine B: %v", err)
	}

	lamport2, err := dbB.GetNextLamportTS()
	if err != nil {
		t.Fatalf("failed to get lamport timestamp on machine B: %v", err)
	}

	payload2 := TaskCreatedPayload{
		TaskUUID:  taskUUID2,
		TaskID:    taskID2,
		Title:     "Fix critical bug",
		CreatedBy: "bob",
	}
	payloadJSON2, _ := json.Marshal(payload2)

	event2 := Event{
		ID:        eventID2,
		TS:        lamport2,
		CreatedAt: testTime(),
		Actor:     "bob",
		Role:      "human",
		Kind:      "task.created",
		Payload:   payloadJSON2,
	}

	if err := dbB.InsertEvent(event2); err != nil {
		t.Fatalf("failed to insert event on machine B: %v", err)
	}

	// Machine B: Export its events to remote
	t.Log("Machine B: Exporting events to remote...")
	eventsBNew, err := dbB.GetEvents()
	if err != nil {
		t.Fatalf("failed to get events from machine B: %v", err)
	}

	// Find new events (those not already in the remote)
	existingEventIDs := make(map[string]bool)
	for _, e := range eventsA {
		existingEventIDs[e.ID] = true
	}

	writerB := NewSegmentWriter(remotePath, space, nodeB, 1, 2_000_000, 120)
	for _, event := range eventsBNew {
		if !existingEventIDs[event.ID] {
			segEvent, err := eventToSegmentEvent(event, space, nodeB)
			if err != nil {
				t.Fatalf("failed to convert event to segment event: %v", err)
			}
			writerB.AddEvent(segEvent)
		}
	}

	segInfoB, err := writerB.WriteSegment()
	if err != nil {
		t.Fatalf("failed to write segment from machine B: %v", err)
	}

	if segInfoB == nil {
		t.Fatal("expected segment info from machine B, got nil")
	}

	// Machine A: Ingest from remote (pull machine B's changes)
	t.Log("Machine A: Ingesting new events from remote...")
	if err := ingestRemote(dbA, "test-remote", remoteConfig); err != nil {
		t.Fatalf("failed to ingest on machine A: %v", err)
	}

	// Machine A: Verify it received machine B's prefix and task
	t.Log("Machine A: Verifying synced data from machine B...")
	prefixesA, err := dbA.GetAllPrefixes()
	if err != nil {
		t.Fatalf("failed to get prefixes on machine A: %v", err)
	}

	foundBugs := false
	for _, p := range prefixesA {
		if p.Prefix == "bugs" && p.Node == nodeB {
			foundBugs = true
			if p.Description != "Bug tracking" {
				t.Errorf("expected description 'Bug tracking', got %s", p.Description)
			}
			if p.CreatedBy != "bob" {
				t.Errorf("expected created_by 'bob', got %s", p.CreatedBy)
			}
			if p.CreatedAt.IsZero() {
				t.Error("expected non-zero created_at for synced prefix 'bugs'")
			}
		}
	}

	if !foundBugs {
		t.Error("prefix 'bugs' from machine B not found on machine A after sync")
	}

	// Machine A: Verify it received machine B's task
	eventsANew, err := dbA.GetEvents()
	if err != nil {
		t.Fatalf("failed to get events from machine A: %v", err)
	}

	reducerA, err := BuildFromEvents(eventsANew)
	if err != nil {
		t.Fatalf("failed to build reducer on machine A: %v", err)
	}

	taskFromB, ok := reducerA.GetTask(taskUUID2)
	if !ok {
		t.Errorf("task %s from machine B not found on machine A", taskUUID2)
	} else {
		if taskFromB.Title != "Fix critical bug" {
			t.Errorf("expected title 'Fix critical bug', got %s", taskFromB.Title)
		}
		if taskFromB.TaskID != taskID2 {
			t.Errorf("expected task ID %s, got %s", taskID2, taskFromB.TaskID)
		}
	}

	t.Log("Full sync workflow completed successfully")
}

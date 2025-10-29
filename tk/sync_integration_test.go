package main

import (
	"encoding/json"
	"fmt"
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

	// Machine A: Create projects for testing
	t.Log("Machine A: Creating projects...")
	projectUIDWork := seedProject(t, dbA, "work")
	_ = seedProject(t, dbA, "personal")

	// Machine A: Create project and task using v4 format
	t.Log("Machine A: Creating tasks...")
	projectUIDA := projectUIDWork
	taskUUID1 := seedTask(t, dbA, projectUIDA, "Implement feature X", 1)

	// Create legacy format event for sync testing
	eventID1, err := GenerateEventID(dbA)
	if err != nil {
		t.Fatalf("failed to generate event ID: %v", err)
	}

	lamport1, err := dbA.GetNextLamportTS()
	if err != nil {
		t.Fatalf("failed to get lamport timestamp: %v", err)
	}

	nodeIDA, err := dbA.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node ID: %v", err)
	}
	taskID1 := fmt.Sprintf("work-1-%s", nodeIDA)

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

	// Machine B: Verify that events were ingested (prefix verification removed - prefix functionality deprecated)
	t.Log("Machine B: Verifying ingested events...")

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

	// Machine B: Create its own project and task using v4 format
	t.Log("Machine B: Creating local project and task...")
	projectUIDB := seedProject(t, dbB, "bugs")
	taskUUID2 := seedTask(t, dbB, projectUIDB, "Fix critical bug", 1)

	// Create legacy format event for sync testing
	taskUUID2Str := taskUUID2
	eventID2, err := GenerateEventID(dbB)
	if err != nil {
		t.Fatalf("failed to generate event ID on machine B: %v", err)
	}

	lamport2, err := dbB.GetNextLamportTS()
	if err != nil {
		t.Fatalf("failed to get lamport timestamp on machine B: %v", err)
	}

	nodeIDB, err := dbB.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node ID on machine B: %v", err)
	}
	taskID2 := fmt.Sprintf("bugs-1-%s", nodeIDB)

	payload2 := TaskCreatedPayload{
		TaskUUID:  taskUUID2Str,
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

	// Machine A: Verify it received machine B's project and task (prefix verification removed - prefix functionality deprecated)
	t.Log("Machine A: Verifying synced data from machine B...")
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

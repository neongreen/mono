package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testTime() time.Time {
	return time.Date(2025, 10, 24, 12, 0, 0, 0, time.UTC)
}

func TestSegmentRoundTrip(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	remotePath := filepath.Join(tmpDir, "remote")
	dbPath := filepath.Join(tmpDir, "tk.db")

	// Create and initialize database
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// Get node ID
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node ID: %v", err)
	}

	// Create a test event
	lamportTS, err := db.GetNextLamportTS()
	if err != nil {
		t.Fatalf("failed to get lamport timestamp: %v", err)
	}

	eventID, err := GenerateEventID(db)
	if err != nil {
		t.Fatalf("failed to generate event ID: %v", err)
	}

	taskID, err := GenerateTaskID(db, "tk")
	if err != nil {
		t.Fatalf("failed to generate task ID: %v", err)
	}

	event := Event{
		ID:        eventID,
		TS:        lamportTS,
		CreatedAt: testTime(),
		Actor:     "test-user",
		Role:      "human",
		Kind:      "task.created",
		Payload:   []byte(`{"task_id":"` + taskID + `","title":"test task","created_by":"test-user"}`),
	}

	if err := db.InsertEvent(event); err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	// Export to segment
	space := "personal"
	writer := NewSegmentWriter(remotePath, space, nodeID, 1, 2_000_000, 120)

	segEvent, err := eventToSegmentEvent(event, space, nodeID)
	if err != nil {
		t.Fatalf("failed to convert event to segment event: %v", err)
	}

	writer.AddEvent(segEvent)

	segInfo, err := writer.WriteSegment()
	if err != nil {
		t.Fatalf("failed to write segment: %v", err)
	}

	if segInfo == nil {
		t.Fatal("expected segment info, got nil")
	}

	// Verify segment file exists
	segPath := filepath.Join(remotePath, segInfo.Rel)
	if _, err := os.Stat(segPath); err != nil {
		t.Fatalf("segment file not found: %v", err)
	}

	// Read segment back
	reader := NewSegmentReader(segPath)
	events, err := reader.ReadEvents()
	if err != nil {
		t.Fatalf("failed to read segment: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].ID != eventID {
		t.Errorf("expected event ID %s, got %s", eventID, events[0].ID)
	}

	if events[0].Lamport != lamportTS {
		t.Errorf("expected lamport %d, got %d", lamportTS, events[0].Lamport)
	}
}

func TestDuplicateIngest(t *testing.T) {
	// Create temporary database
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

	// Create a test event
	lamportTS, err := db.GetNextLamportTS()
	if err != nil {
		t.Fatalf("failed to get lamport timestamp: %v", err)
	}

	eventID, err := GenerateEventID(db)
	if err != nil {
		t.Fatalf("failed to generate event ID: %v", err)
	}

	taskID, err := GenerateTaskID(db, "tk")
	if err != nil {
		t.Fatalf("failed to generate task ID: %v", err)
	}

	event := Event{
		ID:        eventID,
		TS:        lamportTS,
		CreatedAt: testTime(),
		Actor:     "test-user",
		Role:      "human",
		Kind:      "task.created",
		Payload:   []byte(`{"task_id":"` + taskID + `","title":"test task","created_by":"test-user"}`),
	}

	// Insert event first time
	if err := db.InsertEvent(event); err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	// Try to insert same event again (should fail with duplicate error)
	err = db.InsertEvent(event)
	if err == nil {
		t.Fatal("expected duplicate error, got nil")
	}

	if !isDuplicateError(err) {
		t.Errorf("expected duplicate error, got: %v", err)
	}
}

func TestLamportBump(t *testing.T) {
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

	// Get initial lamport
	ts1, err := db.GetNextLamportTS()
	if err != nil {
		t.Fatalf("failed to get lamport timestamp: %v", err)
	}

	if ts1 != 1 {
		t.Errorf("expected initial lamport 1, got %d", ts1)
	}

	// Bump to higher value
	if err := db.BumpLamport(100); err != nil {
		t.Fatalf("failed to bump lamport: %v", err)
	}

	// Get next should be 101
	ts2, err := db.GetNextLamportTS()
	if err != nil {
		t.Fatalf("failed to get lamport timestamp: %v", err)
	}

	if ts2 != 101 {
		t.Errorf("expected lamport 101, got %d", ts2)
	}

	// Bump to lower value (should have no effect)
	if err := db.BumpLamport(50); err != nil {
		t.Fatalf("failed to bump lamport: %v", err)
	}

	// Get next should be 102 (not affected by lower bump)
	ts3, err := db.GetNextLamportTS()
	if err != nil {
		t.Fatalf("failed to get lamport timestamp: %v", err)
	}

	if ts3 != 102 {
		t.Errorf("expected lamport 102, got %d", ts3)
	}
}

func TestNodeID(t *testing.T) {
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

	// Get or create node ID
	nodeID1, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node ID: %v", err)
	}

	// Should be 6 characters
	if len(nodeID1) != 6 {
		t.Errorf("expected node ID length 6, got %d", len(nodeID1))
	}

	// Getting again should return same ID
	nodeID2, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node ID: %v", err)
	}

	if nodeID1 != nodeID2 {
		t.Errorf("expected same node ID, got %s and %s", nodeID1, nodeID2)
	}

	// Regenerate should give different ID
	nodeID3, err := db.RegenerateNodeID()
	if err != nil {
		t.Fatalf("failed to regenerate node ID: %v", err)
	}

	if nodeID1 == nodeID3 {
		t.Errorf("expected different node ID after regeneration")
	}

	// Getting after regeneration should return new ID
	nodeID4, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node ID: %v", err)
	}

	if nodeID3 != nodeID4 {
		t.Errorf("expected regenerated node ID %s, got %s", nodeID3, nodeID4)
	}
}

func TestEventAndTaskIDFormats(t *testing.T) {
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

	// Generate task ID
	taskID, err := GenerateTaskID(db, "tk")
	if err != nil {
		t.Fatalf("failed to generate task ID: %v", err)
	}

	// Should match format tk-<number>-<node>
	if len(taskID) < 8 {
		t.Errorf("task ID too short: %s", taskID)
	}

	if taskID[:3] != "tk-" {
		t.Errorf("task ID should start with 'tk-', got: %s", taskID)
	}

	// Generate event ID
	eventID, err := GenerateEventID(db)
	if err != nil {
		t.Fatalf("failed to generate event ID: %v", err)
	}

	// Should match format ev-<number>-<node>
	if len(eventID) < 8 {
		t.Errorf("event ID too short: %s", eventID)
	}

	if eventID[:3] != "ev-" {
		t.Errorf("event ID should start with 'ev-', got: %s", eventID)
	}
}

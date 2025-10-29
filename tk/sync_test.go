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

// TestPrefixSync tests that prefix.created events are properly synced between machines
func TestPrefixSync(t *testing.T) {
	tmpDir := t.TempDir()

	// Machine A setup
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

	// Machine B setup
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

	// Create prefixes on machine A
	if err := dbA.CreatePrefix("foo", "Tasks for foo project", "alice"); err != nil {
		t.Fatalf("failed to create prefix 'foo' on machine A: %v", err)
	}

	if err := dbA.CreatePrefix("bar", "Tasks for bar project", "alice"); err != nil {
		t.Fatalf("failed to create prefix 'bar' on machine A: %v", err)
	}

	// Verify prefixes exist on machine A
	prefixesA, err := dbA.GetAllPrefixes()
	if err != nil {
		t.Fatalf("failed to get prefixes on machine A: %v", err)
	}

	foundFoo := false
	foundBar := false
	for _, p := range prefixesA {
		if p.Prefix == "foo" && p.Node == nodeA {
			foundFoo = true
			if p.Description != "Tasks for foo project" {
				t.Errorf("expected description 'Tasks for foo project', got %s", p.Description)
			}
			if p.CreatedBy != "alice" {
				t.Errorf("expected created_by 'alice', got %s", p.CreatedBy)
			}
		}
		if p.Prefix == "bar" && p.Node == nodeA {
			foundBar = true
			if p.Description != "Tasks for bar project" {
				t.Errorf("expected description 'Tasks for bar project', got %s", p.Description)
			}
		}
	}

	if !foundFoo {
		t.Error("prefix 'foo' not found on machine A")
	}
	if !foundBar {
		t.Error("prefix 'bar' not found on machine A")
	}

	// Export events from machine A to shared remote
	remotePath := filepath.Join(tmpDir, "remote")
	space := "personal"

	// Get all events from machine A
	eventsA, err := dbA.GetEvents()
	if err != nil {
		t.Fatalf("failed to get events from machine A: %v", err)
	}

	// Write events to segment file
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

	// Ingest on machine B using ingestRemote (simulating 'tk sync')
	remoteConfig := RemoteConfig{
		Type: "folder",
		Path: remotePath,
		Pull: true,
		Push: true,
	}

	if err := ingestRemote(dbB, "test-remote", remoteConfig, false); err != nil {
		t.Fatalf("failed to ingest on machine B: %v", err)
	}

	// Verify prefixes are now available on machine B with full metadata
	prefixesB, err := dbB.GetAllPrefixes()
	if err != nil {
		t.Fatalf("failed to get prefixes on machine B: %v", err)
	}

	foundFooB := false
	foundBarB := false
	for _, p := range prefixesB {
		if p.Prefix == "foo" && p.Node == nodeA {
			foundFooB = true
			// Check that we have full metadata, not just "discovered"
			if p.Description != "Tasks for foo project" {
				t.Errorf("expected description 'Tasks for foo project', got %s", p.Description)
			}
			if p.CreatedBy != "alice" {
				t.Errorf("expected created_by 'alice', got %s", p.CreatedBy)
			}
			if p.CreatedAt.IsZero() {
				t.Error("expected non-zero created_at for synced prefix 'foo'")
			}
		}
		if p.Prefix == "bar" && p.Node == nodeA {
			foundBarB = true
			if p.Description != "Tasks for bar project" {
				t.Errorf("expected description 'Tasks for bar project', got %s", p.Description)
			}
			if p.CreatedAt.IsZero() {
				t.Error("expected non-zero created_at for synced prefix 'bar'")
			}
		}
	}

	if !foundFooB {
		t.Error("prefix 'foo' from machine A not found on machine B after sync")
	}
	if !foundBarB {
		t.Error("prefix 'bar' from machine A not found on machine B after sync")
	}

	// Also verify that GetPrefixes() returns only local node prefixes
	// while GetAllPrefixes() returns all prefixes including remote ones
	localPrefixesB, err := dbB.GetPrefixes()
	if err != nil {
		t.Fatalf("failed to get local prefixes on machine B: %v", err)
	}

	// Should NOT find foo/bar in local prefixes since they're from a different node
	for _, p := range localPrefixesB {
		if (p.Prefix == "foo" || p.Prefix == "bar") && p.Node == nodeA {
			t.Errorf("prefix %s from node A should not appear in local prefix list on machine B", p.Prefix)
		}
	}
}

// TestPrefixRemovedSync tests that prefix.removed events are properly synced between machines
func TestPrefixRemovedSync(t *testing.T) {
	tmpDir := t.TempDir()

	// Machine A setup
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

	// Create and then remove a prefix on machine A
	if err := dbA.CreatePrefix("temp", "Temporary prefix", "alice"); err != nil {
		t.Fatalf("failed to create prefix 'temp' on machine A: %v", err)
	}

	if err := dbA.RemovePrefix("temp", "alice"); err != nil {
		t.Fatalf("failed to remove prefix 'temp' on machine A: %v", err)
	}

	// Verify prefix is marked as removed on machine A
	prefixesA, err := dbA.GetAllPrefixes()
	if err != nil {
		t.Fatalf("failed to get prefixes on machine A: %v", err)
	}

	foundTemp := false
	for _, p := range prefixesA {
		if p.Prefix == "temp" && p.Node == nodeA {
			foundTemp = true
			if !p.Removed {
				t.Error("prefix 'temp' should be marked as removed on machine A")
			}
		}
	}
	if !foundTemp {
		t.Error("prefix 'temp' not found on machine A")
	}

	// Machine B setup
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

	if nodeA == nodeB {
		t.Fatal("machine A and B should have different node IDs")
	}

	// Export events from machine A to shared remote
	remotePath := filepath.Join(tmpDir, "remote")
	space := "personal"

	eventsA, err := dbA.GetEvents()
	if err != nil {
		t.Fatalf("failed to get events from machine A: %v", err)
	}

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

	// Ingest on machine B using ingestRemote
	remoteConfig := RemoteConfig{
		Type: "folder",
		Path: remotePath,
		Pull: true,
		Push: true,
	}

	if err := ingestRemote(dbB, "test-remote", remoteConfig, false); err != nil {
		t.Fatalf("failed to ingest on machine B: %v", err)
	}

	// Verify prefix.removed was properly projected on machine B
	prefixesB, err := dbB.GetAllPrefixes()
	if err != nil {
		t.Fatalf("failed to get prefixes on machine B: %v", err)
	}

	foundTempB := false
	for _, p := range prefixesB {
		if p.Prefix == "temp" && p.Node == nodeA {
			foundTempB = true
			// This is the key check - the removed flag should be synced
			if !p.Removed {
				t.Error("prefix 'temp' should be marked as removed on machine B after sync")
			}
		}
	}

	if !foundTempB {
		t.Error("prefix 'temp' from machine A not found on machine B after sync")
	}
}

package database

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := OpenDB(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})

	err = db.InitDB()
	require.NoError(t, err)

	return db
}

func makeTaskCreatedPayload(taskUID string) json.RawMessage {
	payload := map[string]any{
		"task_uid":     taskUID,
		"project_uid":  "prj_test",
		"title":        "Test Task",
		"created_by":   "test",
		"created_node": "nodeA",
	}
	data, _ := json.Marshal(payload)
	return json.RawMessage(data)
}

func makeStatusSetPayload(taskUID string, state string) json.RawMessage {
	payload := map[string]any{
		"task_uuid": taskUID,
		"task_id":   "",
		"axis":      "generic",
		"state":     state,
		"role":      "human",
	}
	data, _ := json.Marshal(payload)
	return json.RawMessage(data)
}

// TestEventOrdering_MultiMachineSync tests that events are ordered by
// Lamport timestamp, not wall clock, to handle multi-machine sync correctly.
//
// This simulates a scenario where two machines with different wall clocks
// create events. The machine with the clock running behind creates an event
// with a LATER Lamport timestamp but an EARLIER wall clock timestamp.
//
// Correct ordering must use Lamport timestamps to preserve causal order.
func TestEventOrdering_MultiMachineSync(t *testing.T) {
	db := setupTestDB(t)

	// Simulate two machines with different wall clocks
	// Machine A has wall clock at current time
	// Machine B has wall clock 1 hour behind

	// Machine A creates event at TS=1, created_at = now
	eventA := types.Event{
		ID:        "ev-1-nodeA",
		TS:        1, // Lamport TS = 1 (happens first logically)
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   makeTaskCreatedPayload("tsk_test1"),
	}
	err := db.InsertEvent(eventA)
	require.NoError(t, err, "failed to insert eventA")

	// Machine B creates event at TS=2, created_at = 1 hour ago
	// (Machine B's clock is behind, but this event happened AFTER eventA logically)
	eventB := types.Event{
		ID:        "ev-2-nodeB",
		TS:        2,                              // Lamport TS = 2 (happens second logically)
		CreatedAt: time.Now().Add(-1 * time.Hour), // Wall clock shows 1 hour ago!
		Actor:     "bob",
		Role:      "human",
		Kind:      string(types.EventKindTaskStatusSet),
		Payload:   makeStatusSetPayload("tsk_test1", "done"),
	}
	err = db.InsertEvent(eventB)
	require.NoError(t, err, "failed to insert eventB")

	// Get events from database
	events, err := db.GetEvents()
	require.NoError(t, err)
	require.Len(t, events, 2, "should have 2 events")

	// CRITICAL: Events must be ordered by Lamport TS, not wall clock
	// eventA (TS=1) should come before eventB (TS=2)
	// even though eventB has an older created_at timestamp

	assert.Equal(t, "ev-1-nodeA", events[0].ID,
		"First event should be ev-1-nodeA (TS=1), got %s", events[0].ID)
	assert.Equal(t, "ev-2-nodeB", events[1].ID,
		"Second event should be ev-2-nodeB (TS=2), got %s", events[1].ID)

	// Verify Lamport timestamps are in correct order
	assert.Equal(t, int64(1), events[0].TS, "First event should have TS=1")
	assert.Equal(t, int64(2), events[1].TS, "Second event should have TS=2")
	assert.Less(t, events[0].TS, events[1].TS, "Events should be ordered by TS")

	// Show why this matters: wall clocks are in reverse order
	assert.True(t, events[1].CreatedAt.Before(events[0].CreatedAt),
		"eventB has earlier wall clock but later Lamport timestamp - this is the multi-machine scenario")
}

// TestEventOrdering_SameLamportTimestamp tests that when multiple events
// have the same Lamport timestamp (concurrent events), they are ordered by event ID
func TestEventOrdering_SameLamportTimestamp(t *testing.T) {
	db := setupTestDB(t)

	// Two concurrent events with same Lamport timestamp
	event1 := types.Event{
		ID:        "ev-1-nodeA",
		TS:        5,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   makeTaskCreatedPayload("tsk_test1"),
	}
	db.InsertEvent(event1)

	event2 := types.Event{
		ID:        "ev-2-nodeB",
		TS:        5,                                // Same Lamport timestamp (concurrent)
		CreatedAt: time.Now().Add(-1 * time.Minute), // Different wall clock
		Actor:     "bob",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   makeTaskCreatedPayload("tsk_test2"),
	}
	db.InsertEvent(event2)

	// Get events
	events, err := db.GetEvents()
	require.NoError(t, err)
	require.Len(t, events, 2)

	// When TS is the same, order by event ID
	// ev-1-nodeA < ev-2-nodeB lexicographically
	assert.Equal(t, "ev-1-nodeA", events[0].ID, "Should sort by ID when TS is equal")
	assert.Equal(t, "ev-2-nodeB", events[1].ID, "Should sort by ID when TS is equal")
}

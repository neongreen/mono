package database

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

func TestContainerRebuildFromEvents(t *testing.T) {
	db := openTempDB(t)

	// Create a sequence of events
	events := []types.Event{}

	// 1. Define a queue kind
	definePayload := types.DefineContainerKindPayload{
		Name:        "sprint",
		Primitive:   types.PrimitiveQueue,
		Description: "Sprint queue",
		CreatedBy:   "tester",
	}
	definePayloadJSON, _ := json.Marshal(definePayload)
	events = append(events, types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerKindDefine),
		Payload:   definePayloadJSON,
	})

	// 2. Create a queue
	createPayload := types.CreateContainerPayload{
		ID:        "q-1",
		Primitive: types.PrimitiveQueue,
		Kind:      "sprint",
		Name:      "Nov Sprint",
		CreatedBy: "tester",
	}
	createPayloadJSON, _ := json.Marshal(createPayload)
	events = append(events, types.Event{
		ID:        string(types.NewEventID()),
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerCreate),
		Payload:   createPayloadJSON,
	})

	// 3. Push 5 items
	for i := range 5 {
		pushPayload := types.QueuePushPayload{
			ContainerID: "q-1",
			ItemID:      types.TaskUID(fmt.Sprintf("tk-%d", i+1)),
		}
		pushPayloadJSON, _ := json.Marshal(pushPayload)
		events = append(events, types.Event{
			ID:        string(types.NewEventID()),
			TS:        int64(i + 2),
			CreatedAt: time.Now(),
			Actor:     "tester",
			Role:      "human",
			Kind:      string(types.EventKindQueuePush),
			Payload:   pushPayloadJSON,
		})
	}

	// 4. Pop 2 items
	for i := range 2 {
		popPayload := types.QueuePopPayload{
			ContainerID: "q-1",
			ItemID:      types.TaskUID(fmt.Sprintf("tk-%d", i+1)),
		}
		popPayloadJSON, _ := json.Marshal(popPayload)
		events = append(events, types.Event{
			ID:        string(types.NewEventID()),
			TS:        int64(i + 7),
			CreatedAt: time.Now(),
			Actor:     "tester",
			Role:      "human",
			Kind:      string(types.EventKindQueuePop),
			Payload:   popPayloadJSON,
		})
	}

	// Replay all events
	for _, e := range events {
		if err := db.ProjectEvent(e); err != nil {
			t.Fatalf("ProjectEvent() error = %v", err)
		}
	}

	// Drop tables and rebuild
	db.Db.Exec(`DROP TABLE container_kinds`)
	db.Db.Exec(`DROP TABLE containers`)
	db.Db.Exec(`DROP TABLE container_members`)

	// Recreate tables
	if err := db.CreateContainerTables(); err != nil {
		t.Fatalf("failed to recreate tables: %v", err)
	}

	// Replay all events again
	for _, e := range events {
		if err := db.ProjectEvent(e); err != nil {
			t.Fatalf("ProjectEvent() after rebuild error = %v", err)
		}
	}

	// Verify state matches
	var rebuiltMembers []string
	rows, err := db.Db.Query(`
		SELECT item_id FROM container_members
		WHERE container_id = 'q-1' AND removed = 0
		ORDER BY position
	`)
	if err != nil {
		t.Fatalf("failed to query members after rebuild: %v", err)
	}
	for rows.Next() {
		var itemID string
		if err := rows.Scan(&itemID); err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}
		rebuiltMembers = append(rebuiltMembers, itemID)
	}
	rows.Close()

	// Should have tk-3, tk-4, tk-5 (pushed 5, popped 2 from head)
	if len(rebuiltMembers) != 3 {
		t.Errorf("after rebuild: got %d members, want 3", len(rebuiltMembers))
	}

	expected := []string{"tk-3", "tk-4", "tk-5"}
	for i, itemID := range rebuiltMembers {
		if itemID != expected[i] {
			t.Errorf("member[%d] = %q, want %q", i, itemID, expected[i])
		}
	}

	t.Logf("✓ Rebuild from events produced identical state")
}

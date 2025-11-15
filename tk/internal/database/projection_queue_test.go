package database

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)
func TestProjectQueuePushEvent_PositionAssignment(t *testing.T) {
	db := openTempDB(t)

	// Create kind and container
	seedContainerKindAndInstance(t, db, "sprint", types.PrimitiveQueue, "q-1", "Nov Sprint")

	// Push three items
	for i, itemID := range []string{"tk-1", "tk-2", "tk-3"} {
		payload := types.QueuePushPayload{
			ContainerID: "q-1",
			ItemID:      types.TaskUID(itemID),
		}
		payloadJSON, _ := json.Marshal(payload)
		event := types.Event{
			ID:        string(types.NewEventID()),
			TS:        int64(i),
			CreatedAt: time.Now(),
			Actor:     "tester",
			Role:      "human",
			Kind:      string(types.EventKindQueuePush),
			Payload:   payloadJSON,
		}
		if err := db.ProjectQueuePushEvent(event); err != nil {
			t.Fatalf("ProjectQueuePushEvent() error = %v", err)
		}
	}

	// Verify positions are 1, 2, 3
	rows, err := db.Db.Query(`
		SELECT item_id, position
		FROM container_members
		WHERE container_id = 'q-1' AND removed = 0
		ORDER BY position
	`)
	if err != nil {
		t.Fatalf("failed to query members: %v", err)
	}
	defer rows.Close()

	expectedPositions := map[string]int64{
		"tk-1": 1,
		"tk-2": 2,
		"tk-3": 3,
	}

	for rows.Next() {
		var itemID string
		var position int64
		if err := rows.Scan(&itemID, &position); err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}
		if expectedPositions[itemID] != position {
			t.Errorf("item %s: position = %d, want %d", itemID, position, expectedPositions[itemID])
		}
	}
}

func TestProjectQueuePopEvent(t *testing.T) {
	db := openTempDB(t)

	// Create kind, container, and push items
	seedContainerKindAndInstance(t, db, "sprint", types.PrimitiveQueue, "q-1", "Nov Sprint")
	seedQueueItems(t, db, "q-1", []string{"tk-1", "tk-2", "tk-3"})

	// Pop the head item (tk-1)
	popPayload := types.QueuePopPayload{
		ContainerID: "q-1",
		ItemID:      "tk-1",
	}
	popPayloadJSON, _ := json.Marshal(popPayload)
	popEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        10,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindQueuePop),
		Payload:   popPayloadJSON,
	}
	if err := db.ProjectQueuePopEvent(popEvent); err != nil {
		t.Fatalf("ProjectQueuePopEvent() error = %v", err)
	}

	// Verify tk-1 is removed, tk-2 and tk-3 remain
	var count int
	err := db.Db.QueryRow(`
		SELECT COUNT(*) FROM container_members
		WHERE container_id = 'q-1' AND removed = 0
	`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count members: %v", err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	// Verify tk-1 is marked as removed
	var removed int
	err = db.Db.QueryRow(`
		SELECT removed FROM container_members
		WHERE container_id = 'q-1' AND item_id = 'tk-1'
	`).Scan(&removed)
	if err != nil {
		t.Fatalf("failed to query tk-1: %v", err)
	}
	if removed != 1 {
		t.Errorf("tk-1 removed = %d, want 1", removed)
	}
}

// Comprehensive edge case tests

func TestQueuePush_PositionGapsHandled(t *testing.T) {
	db := openTempDB(t)
	seedContainerKindAndInstance(t, db, "sprint", types.PrimitiveQueue, "q-1", "Sprint")

	// Manually insert items with gaps in positions (1, 2, 5, 6)
	db.Db.Exec(`
		INSERT INTO container_members (container_id, item_id, position, removed)
		VALUES ('q-1', 'tk-1', 1, 0),
		       ('q-1', 'tk-2', 2, 0),
		       ('q-1', 'tk-5', 5, 0),
		       ('q-1', 'tk-6', 6, 0)
	`)

	// Push a new item - should get position 7 (max+1)
	payload := types.QueuePushPayload{
		ContainerID: "q-1",
		ItemID:      "tk-new",
	}
	payloadJSON, _ := json.Marshal(payload)
	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        10,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindQueuePush),
		Payload:   payloadJSON,
	}

	if err := db.ProjectQueuePushEvent(event); err != nil {
		t.Fatalf("ProjectQueuePushEvent() error = %v", err)
	}

	// Verify new item has position 7
	var position int64
	err := db.Db.QueryRow(`
		SELECT position FROM container_members
		WHERE container_id = 'q-1' AND item_id = 'tk-new'
	`).Scan(&position)
	if err != nil {
		t.Fatalf("failed to query new item: %v", err)
	}

	if position != 7 {
		t.Errorf("position = %d, want 7 (max of 1,2,5,6 is 6, next is 7)", position)
	}
}

func TestQueuePush_RemovedItemsIgnored(t *testing.T) {
	db := openTempDB(t)
	seedContainerKindAndInstance(t, db, "sprint", types.PrimitiveQueue, "q-1", "Sprint")

	// Insert items where some are removed (positions: 1, 2-removed, 3, 4-removed)
	db.Db.Exec(`
		INSERT INTO container_members (container_id, item_id, position, removed)
		VALUES ('q-1', 'tk-1', 1, 0),
		       ('q-1', 'tk-2', 2, 1),
		       ('q-1', 'tk-3', 3, 0),
		       ('q-1', 'tk-4', 4, 1)
	`)

	// Push new item - should use max of non-removed (3) and assign position 4
	payload := types.QueuePushPayload{
		ContainerID: "q-1",
		ItemID:      "tk-new",
	}
	payloadJSON, _ := json.Marshal(payload)
	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        10,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindQueuePush),
		Payload:   payloadJSON,
	}

	if err := db.ProjectQueuePushEvent(event); err != nil {
		t.Fatalf("ProjectQueuePushEvent() error = %v", err)
	}

	// Verify new item has position 4 (max of active items 1,3 is 3, next is 4)
	var position int64
	err := db.Db.QueryRow(`
		SELECT position FROM container_members
		WHERE container_id = 'q-1' AND item_id = 'tk-new'
	`).Scan(&position)
	if err != nil {
		t.Fatalf("failed to query new item: %v", err)
	}

	if position != 4 {
		t.Errorf("position = %d, want 4 (should ignore removed items at pos 2,4)", position)
	}
}

func TestQueuePush_EmptyContainer(t *testing.T) {
	db := openTempDB(t)
	seedContainerKindAndInstance(t, db, "sprint", types.PrimitiveQueue, "q-1", "Sprint")

	// Push to empty container - should get position 1
	payload := types.QueuePushPayload{
		ContainerID: "q-1",
		ItemID:      "tk-first",
	}
	payloadJSON, _ := json.Marshal(payload)
	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindQueuePush),
		Payload:   payloadJSON,
	}

	if err := db.ProjectQueuePushEvent(event); err != nil {
		t.Fatalf("ProjectQueuePushEvent() error = %v", err)
	}

	// Verify position is 1
	var position int64
	err := db.Db.QueryRow(`
		SELECT position FROM container_members
		WHERE container_id = 'q-1' AND item_id = 'tk-first'
	`).Scan(&position)
	if err != nil {
		t.Fatalf("failed to query item: %v", err)
	}

	if position != 1 {
		t.Errorf("position = %d, want 1 (first item in empty container)", position)
	}
}

func TestQueuePush_DuplicateItem_Idempotent(t *testing.T) {
	db := openTempDB(t)
	seedContainerKindAndInstance(t, db, "sprint", types.PrimitiveQueue, "q-1", "Sprint")

	// Push item twice
	payload := types.QueuePushPayload{
		ContainerID: "q-1",
		ItemID:      "tk-1",
	}
	payloadJSON, _ := json.Marshal(payload)
	event1 := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindQueuePush),
		Payload:   payloadJSON,
	}

	if err := db.ProjectQueuePushEvent(event1); err != nil {
		t.Fatalf("first ProjectQueuePushEvent() error = %v", err)
	}

	// Push same item again (different event, simulating replay)
	event2 := types.Event{
		ID:        string(types.NewEventID()),
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindQueuePush),
		Payload:   payloadJSON,
	}

	if err := db.ProjectQueuePushEvent(event2); err != nil {
		t.Fatalf("second ProjectQueuePushEvent() error = %v", err)
	}

	// Should only have one item with position 2 (gets updated)
	var count int
	err := db.Db.QueryRow(`
		SELECT COUNT(*) FROM container_members
		WHERE container_id = 'q-1' AND item_id = 'tk-1'
	`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count items: %v", err)
	}

	if count != 1 {
		t.Errorf("count = %d, want 1 (duplicate push should update, not duplicate)", count)
	}

	// Verify it has the newer position
	var position int64
	err = db.Db.QueryRow(`
		SELECT position FROM container_members
		WHERE container_id = 'q-1' AND item_id = 'tk-1'
	`).Scan(&position)
	if err != nil {
		t.Fatalf("failed to query position: %v", err)
	}

	if position != 2 {
		t.Errorf("position = %d, want 2 (second push updates position)", position)
	}
}

func TestQueuePush_RepushRemovedItem(t *testing.T) {
	db := openTempDB(t)
	seedContainerKindAndInstance(t, db, "sprint", types.PrimitiveQueue, "q-1", "Sprint")

	// Push item
	seedQueueItems(t, db, "q-1", []string{"tk-1"})

	// Pop it
	popPayload := types.QueuePopPayload{
		ContainerID: "q-1",
		ItemID:      "tk-1",
	}
	popPayloadJSON, _ := json.Marshal(popPayload)
	popEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        10,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindQueuePop),
		Payload:   popPayloadJSON,
	}
	db.ProjectQueuePopEvent(popEvent)

	// Push it again
	pushPayload := types.QueuePushPayload{
		ContainerID: "q-1",
		ItemID:      "tk-1",
	}
	pushPayloadJSON, _ := json.Marshal(pushPayload)
	pushEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        11,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindQueuePush),
		Payload:   pushPayloadJSON,
	}
	if err := db.ProjectQueuePushEvent(pushEvent); err != nil {
		t.Fatalf("ProjectQueuePushEvent() error = %v", err)
	}

	// Verify item is back with removed=0 and new position
	var removed int
	var position int64
	err := db.Db.QueryRow(`
		SELECT removed, position FROM container_members
		WHERE container_id = 'q-1' AND item_id = 'tk-1'
	`).Scan(&removed, &position)
	if err != nil {
		t.Fatalf("failed to query item: %v", err)
	}

	if removed != 0 {
		t.Errorf("removed = %d, want 0 (repush should unmark as removed)", removed)
	}
	// When repushing the only item (which was removed), max of non-removed items is NULL,
	// so we start fresh at position 1
	if position != 1 {
		t.Errorf("position = %d, want 1 (repush to empty queue starts at 1)", position)
	}
}


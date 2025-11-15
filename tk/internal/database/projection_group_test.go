package database

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)
func TestProjectGroupAddEvent(t *testing.T) {
	db := openTempDB(t)

	// Create kind and container
	seedContainerKindAndInstance(t, db, "today", types.PrimitiveGroup, "g-1", "Today's Tasks")

	// Add items to group
	for i, itemID := range []string{"tk-1", "tk-2", "tk-3"} {
		payload := types.GroupAddPayload{
			ContainerID: "g-1",
			ItemID:      types.TaskUID(itemID),
		}
		payloadJSON, _ := json.Marshal(payload)
		event := types.Event{
			ID:        string(types.NewEventID()),
			TS:        int64(i),
			CreatedAt: time.Now(),
			Actor:     "tester",
			Role:      "human",
			Kind:      string(types.EventKindGroupAdd),
			Payload:   payloadJSON,
		}
		if err := db.ProjectGroupAddEvent(event); err != nil {
			t.Fatalf("ProjectGroupAddEvent() error = %v", err)
		}
	}

	// Verify all items have NULL position
	rows, err := db.Db.Query(`
		SELECT item_id, position
		FROM container_members
		WHERE container_id = 'g-1' AND removed = 0
	`)
	if err != nil {
		t.Fatalf("failed to query members: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var itemID string
		var position sql.NullInt64
		if err := rows.Scan(&itemID, &position); err != nil {
			t.Fatalf("failed to scan row: %v", err)
		}
		if position.Valid {
			t.Errorf("item %s: position should be NULL, got %d", itemID, position.Int64)
		}
		count++
	}

	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}
func TestGroupAdd_Idempotent(t *testing.T) {
	db := openTempDB(t)
	seedContainerKindAndInstance(t, db, "today", types.PrimitiveGroup, "g-1", "Today")

	// Add item twice
	payload := types.GroupAddPayload{
		ContainerID: "g-1",
		ItemID:      "tk-1",
	}
	payloadJSON, _ := json.Marshal(payload)

	for i := range 2 {
		event := types.Event{
			ID:        string(types.NewEventID()),
			TS:        int64(i),
			CreatedAt: time.Now(),
			Actor:     "tester",
			Role:      "human",
			Kind:      string(types.EventKindGroupAdd),
			Payload:   payloadJSON,
		}
		if err := db.ProjectGroupAddEvent(event); err != nil {
			t.Fatalf("ProjectGroupAddEvent() #%d error = %v", i+1, err)
		}
	}

	// Should only have one copy
	var count int
	err := db.Db.QueryRow(`
		SELECT COUNT(*) FROM container_members
		WHERE container_id = 'g-1' AND item_id = 'tk-1'
	`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}

	if count != 1 {
		t.Errorf("count = %d, want 1 (duplicate add should be idempotent)", count)
	}
}

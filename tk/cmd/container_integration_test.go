package cmd

import (
	"database/sql"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/cmd/schema"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

// TestContainerIntegration tests the complete container workflow end-to-end
func TestContainerIntegration(t *testing.T) {
	db := openTempDB(t)

	// 1. Define container kinds
	defineKind(t, db, "sprint", types.PrimitiveQueue, "Timeboxed work period")
	defineKind(t, db, "focus", types.PrimitiveGroup, "Current focus items")

	// 2. Create containers
	createContainer(t, db, "q-1", types.PrimitiveQueue, "sprint", "November Sprint")
	createContainer(t, db, "g-1", types.PrimitiveGroup, "focus", "Focus Group")

	// 3. Add items to queue
	pushToQueue(t, db, "q-1", "tk-1")
	pushToQueue(t, db, "q-1", "tk-2")
	pushToQueue(t, db, "q-1", "tk-3")

	// 4. Verify queue has 3 items in correct order
	members := getQueueMembers(t, db, "q-1")
	if len(members) != 3 {
		t.Fatalf("expected 3 queue members, got %d", len(members))
	}
	if members[0] != "tk-1" || members[1] != "tk-2" || members[2] != "tk-3" {
		t.Errorf("queue order incorrect: %v", members)
	}

	// 5. Pop from queue (should remove tk-1 from head)
	popFromQueue(t, db, "q-1", "tk-1")
	members = getQueueMembers(t, db, "q-1")
	if len(members) != 2 {
		t.Fatalf("after pop: expected 2 queue members, got %d", len(members))
	}
	if members[0] != "tk-2" {
		t.Errorf("after pop: queue head should be tk-2, got %s", members[0])
	}

	// 6. Add items to group
	addToGroup(t, db, "g-1", "tk-10")
	addToGroup(t, db, "g-1", "tk-11")
	addToGroup(t, db, "g-1", "tk-12")

	groupMembers := getGroupMembers(t, db, "g-1")
	if len(groupMembers) != 3 {
		t.Fatalf("expected 3 group members, got %d", len(groupMembers))
	}

	// 7. Remove from group
	removeFromGroup(t, db, "g-1", "tk-11")
	groupMembers = getGroupMembers(t, db, "g-1")
	if len(groupMembers) != 2 {
		t.Fatalf("after remove: expected 2 group members, got %d", len(groupMembers))
	}

	// 8. Test schema export
	export := exportSchema(t, db)
	if len(export.QueueKinds) != 1 {
		t.Errorf("expected 1 queue kind, got %d", len(export.QueueKinds))
	}
	if export.QueueKinds[0].Name != "sprint" {
		t.Errorf("queue kind name = %q, want \"sprint\"", export.QueueKinds[0].Name)
	}

	t.Log("✓ Complete container workflow tested successfully")
}

// Helper functions

func defineKind(t *testing.T, db *database.DB, name string, primitive types.ContainerPrimitive, description string) {
	payload := types.DefineContainerKindPayload{
		Name:        name,
		Primitive:   primitive,
		Description: description,
		CreatedBy:   "test",
	}

	eventID, _ := database.GenerateEventID(db)
	ts, _ := db.GetNextLamportTS()

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindContainerKindDefine),
		Payload:   mustJSON(t, payload),
	}

	if err := db.InsertEvent(event); err != nil {
		t.Fatalf("failed to insert define kind event: %v", err)
	}
	if err := db.ProjectContainerKindDefineEvent(event); err != nil {
		t.Fatalf("failed to project define kind event: %v", err)
	}
}

func createContainer(t *testing.T, db *database.DB, id string, primitive types.ContainerPrimitive, kind string, name string) {
	payload := types.CreateContainerPayload{
		ID:        id,
		Primitive: primitive,
		Kind:      kind,
		Name:      name,
		CreatedBy: "test",
	}

	eventID, _ := database.GenerateEventID(db)
	ts, _ := db.GetNextLamportTS()

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindContainerCreate),
		Payload:   mustJSON(t, payload),
	}

	if err := db.InsertEvent(event); err != nil {
		t.Fatalf("failed to insert create container event: %v", err)
	}
	if err := db.ProjectContainerCreateEvent(event); err != nil {
		t.Fatalf("failed to project create container event: %v", err)
	}
}

func pushToQueue(t *testing.T, db *database.DB, queueID string, itemID string) {
	payload := types.QueuePushPayload{
		ContainerID: queueID,
		ItemID:      itemID,
	}

	eventID, _ := database.GenerateEventID(db)
	ts, _ := db.GetNextLamportTS()

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindQueuePush),
		Payload:   mustJSON(t, payload),
	}

	if err := db.InsertEvent(event); err != nil {
		t.Fatalf("failed to insert queue push event: %v", err)
	}
	if err := db.ProjectQueuePushEvent(event); err != nil {
		t.Fatalf("failed to project queue push event: %v", err)
	}
}

func popFromQueue(t *testing.T, db *database.DB, queueID string, itemID string) {
	payload := types.QueuePopPayload{
		ContainerID: queueID,
		ItemID:      itemID,
	}

	eventID, _ := database.GenerateEventID(db)
	ts, _ := db.GetNextLamportTS()

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindQueuePop),
		Payload:   mustJSON(t, payload),
	}

	if err := db.InsertEvent(event); err != nil {
		t.Fatalf("failed to insert queue pop event: %v", err)
	}
	if err := db.ProjectQueuePopEvent(event); err != nil {
		t.Fatalf("failed to project queue pop event: %v", err)
	}
}

func addToGroup(t *testing.T, db *database.DB, groupID string, itemID string) {
	payload := types.GroupAddPayload{
		ContainerID: groupID,
		ItemID:      itemID,
	}

	eventID, _ := database.GenerateEventID(db)
	ts, _ := db.GetNextLamportTS()

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindGroupAdd),
		Payload:   mustJSON(t, payload),
	}

	if err := db.InsertEvent(event); err != nil {
		t.Fatalf("failed to insert group add event: %v", err)
	}
	if err := db.ProjectGroupAddEvent(event); err != nil {
		t.Fatalf("failed to project group add event: %v", err)
	}
}

func removeFromGroup(t *testing.T, db *database.DB, groupID string, itemID string) {
	payload := types.GroupRemovePayload{
		ContainerID: groupID,
		ItemID:      itemID,
	}

	eventID, _ := database.GenerateEventID(db)
	ts, _ := db.GetNextLamportTS()

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindGroupRemove),
		Payload:   mustJSON(t, payload),
	}

	if err := db.InsertEvent(event); err != nil {
		t.Fatalf("failed to insert group remove event: %v", err)
	}
	if err := db.ProjectGroupRemoveEvent(event); err != nil {
		t.Fatalf("failed to project group remove event: %v", err)
	}
}

func getQueueMembers(t *testing.T, db *database.DB, queueID string) []string {
	rows, err := db.Db.Query(`
		SELECT item_id FROM container_members
		WHERE container_id = ? AND removed = 0
		ORDER BY position ASC
	`, queueID)
	if err != nil {
		t.Fatalf("failed to query queue members: %v", err)
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var itemID string
		rows.Scan(&itemID)
		members = append(members, itemID)
	}
	return members
}

func getGroupMembers(t *testing.T, db *database.DB, groupID string) []string {
	rows, err := db.Db.Query(`
		SELECT item_id FROM container_members
		WHERE container_id = ? AND removed = 0
		ORDER BY item_id
	`, groupID)
	if err != nil {
		t.Fatalf("failed to query group members: %v", err)
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var itemID string
		rows.Scan(&itemID)
		members = append(members, itemID)
	}
	return members
}

func exportSchema(t *testing.T, db *database.DB) schema.SchemaExport {
	rows, err := db.Db.Query(`
		SELECT name, primitive, description
		FROM container_kinds
		WHERE deprecated = 0
		ORDER BY primitive, name
	`)
	if err != nil {
		t.Fatalf("failed to query container kinds: %v", err)
	}
	defer rows.Close()

	export := schema.SchemaExport{
		QueueKinds: []schema.ContainerKindExport{},
		StackKinds: []schema.ContainerKindExport{},
		GroupKinds: []schema.ContainerKindExport{},
	}

	for rows.Next() {
		var name string
		var primitive string
		var description sql.NullString

		rows.Scan(&name, &primitive, &description)

		kind := schema.ContainerKindExport{
			Name: name,
		}
		if description.Valid {
			kind.Description = &description.String
		}

		switch primitive {
		case "queue":
			export.QueueKinds = append(export.QueueKinds, kind)
		case "stack":
			export.StackKinds = append(export.StackKinds, kind)
		case "group":
			export.GroupKinds = append(export.GroupKinds, kind)
		}
	}

	return export
}

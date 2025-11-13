package database

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

func TestProjectProjectCreatedEvent(t *testing.T) {
	db := openTempDB(t)

	projectUID := string(types.NewProjectUID())
	payload := types.ProjectCreatedPayload{
		ProjectUID:  projectUID,
		Type:        "local",
		Name:        "test-project",
		Description: "Test description",
		CreatedBy:   "tester",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindProjectCreated),
		Payload:   payloadJSON,
	}

	if err := db.ProjectProjectCreatedEvent(event); err != nil {
		t.Fatalf("ProjectProjectCreatedEvent() error = %v", err)
	}

	// Verify project was created in database
	var count int
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM projects WHERE project_uid = ?`, projectUID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query projects: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 project, got %d", count)
	}
}

// TestProjectProjectAliasAddEvent removed - alias events are now filtered by transformers
// and never reach the projection layer

func TestProjectTaskCreatedEvent(t *testing.T) {
	db := openTempDB(t)

	// Create a project first
	projectUID := seedProject(t, db, "test")

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node ID: %v", err)
	}

	taskUID := string(types.NewTaskUID())
	payload := types.TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: 1,
		CreatedNode:    nodeID,
		Title:          "Test task",
		CreatedBy:      "tester",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   payloadJSON,
	}

	if err := db.ProjectTaskCreatedEvent(event); err != nil {
		t.Fatalf("ProjectTaskCreatedEvent() error = %v", err)
	}

	// Verify task was created
	var count int
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE task_uid = ?`, taskUID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query tasks: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 task, got %d", count)
	}
}

func TestProjectTaskNumberSetEvent(t *testing.T) {
	db := openTempDB(t)

	// Create a project and task
	projectUID := seedProject(t, db, "test")
	taskUID := seedTask(t, db, projectUID, "Test task", 1)

	// Update the number
	payload := types.TaskNumberSetPayload{
		TaskUID:    taskUID,
		ProjectUID: projectUID,
		Number:     42,
		Reason:     "manual",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindTaskNumberSet),
		Payload:   payloadJSON,
	}

	if err := db.ProjectTaskNumberSetEvent(event); err != nil {
		t.Fatalf("ProjectTaskNumberSetEvent() error = %v", err)
	}

	// Verify number was updated
	var number int64
	err = db.Db.QueryRow(`SELECT number FROM task_numbers WHERE task_uid = ?`, taskUID).Scan(&number)
	if err != nil {
		t.Fatalf("failed to query task number: %v", err)
	}

	if number != 42 {
		t.Errorf("expected number 42, got %d", number)
	}
}

func TestProjectTaskTitleSetEvent(t *testing.T) {
	db := openTempDB(t)

	// Create a project and task
	projectUID := seedProject(t, db, "test")
	taskUID := seedTask(t, db, projectUID, "Old title", 1)

	// Update the title
	payload := types.TaskTitleSetPayload{
		TaskUID: taskUID,
		Title:   "New title",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindTaskTitleSet),
		Payload:   payloadJSON,
	}

	if err := db.ProjectTaskTitleSetEvent(event); err != nil {
		t.Fatalf("ProjectTaskTitleSetEvent() error = %v", err)
	}

	// Verify title was updated
	var title string
	err = db.Db.QueryRow(`SELECT title FROM tasks WHERE task_uid = ?`, taskUID).Scan(&title)
	if err != nil {
		t.Fatalf("failed to query task title: %v", err)
	}

	if title != "New title" {
		t.Errorf("expected title 'New title', got %s", title)
	}
}

func TestProjectTaskDeleteEvent(t *testing.T) {
	db := openTempDB(t)

	// First create a project and task
	projectUID := seedProject(t, db, "deltest")
	taskUID := seedTask(t, db, projectUID, "Task to delete", 1)

	// Verify task exists in tasks table
	var countTasks int
	err := db.Db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE task_uid = ?`, taskUID).Scan(&countTasks)
	if err != nil {
		t.Fatalf("failed to query tasks: %v", err)
	}
	if countTasks != 1 {
		t.Fatalf("expected 1 task, got %d", countTasks)
	}

	// Verify task exists in task_numbers table
	var countNumbers int
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM task_numbers WHERE task_uid = ?`, taskUID).Scan(&countNumbers)
	if err != nil {
		t.Fatalf("failed to query task_numbers: %v", err)
	}
	if countNumbers != 1 {
		t.Fatalf("expected 1 task_number entry, got %d", countNumbers)
	}

	// Create and project delete event
	payload := types.TaskDeletePayload{
		TaskUUID: taskUID,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindTaskDelete),
		Payload:   payloadJSON,
	}

	if err := db.ProjectTaskDeleteEvent(event); err != nil {
		t.Fatalf("ProjectTaskDeleteEvent() error = %v", err)
	}

	// Verify task was deleted from tasks table
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE task_uid = ?`, taskUID).Scan(&countTasks)
	if err != nil {
		t.Fatalf("failed to query tasks after delete: %v", err)
	}
	if countTasks != 0 {
		t.Errorf("expected 0 tasks after delete, got %d", countTasks)
	}

	// Verify task was deleted from task_numbers table
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM task_numbers WHERE task_uid = ?`, taskUID).Scan(&countNumbers)
	if err != nil {
		t.Fatalf("failed to query task_numbers after delete: %v", err)
	}
	if countNumbers != 0 {
		t.Errorf("expected 0 task_number entries after delete, got %d", countNumbers)
	}
}

func TestProjectTaskDeleteEvent_Idempotency(t *testing.T) {
	db := openTempDB(t)

	// First create a project and task
	projectUID := seedProject(t, db, "idemtest")
	taskUID := seedTask(t, db, projectUID, "Task to delete", 1)

	// Create delete event
	payload := types.TaskDeletePayload{
		TaskUUID: taskUID,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindTaskDelete),
		Payload:   payloadJSON,
	}

	// Project delete event first time
	if err := db.ProjectTaskDeleteEvent(event); err != nil {
		t.Fatalf("first ProjectTaskDeleteEvent() error = %v", err)
	}

	// Project delete event second time (should be idempotent)
	if err := db.ProjectTaskDeleteEvent(event); err != nil {
		t.Fatalf("second ProjectTaskDeleteEvent() error = %v (should be idempotent)", err)
	}

	// Verify task is still deleted (only deleted once)
	var count int
	err = db.Db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE task_uid = ?`, taskUID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query tasks: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 tasks after double delete, got %d", count)
	}
}

func TestProjectContainerKindDefineEvent(t *testing.T) {
	db := openTempDB(t)

	payload := types.DefineContainerKindPayload{
		Name:        "sprint",
		Primitive:   types.PrimitiveQueue,
		Description: "Timeboxed work period",
		CreatedBy:   "tester",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerKindDefine),
		Payload:   payloadJSON,
	}

	if err := db.ProjectContainerKindDefineEvent(event); err != nil {
		t.Fatalf("ProjectContainerKindDefineEvent() error = %v", err)
	}

	// Verify kind was created
	var count int
	var primitive string
	var description string
	var deprecated int
	err = db.Db.QueryRow(`
		SELECT COUNT(*), primitive, description, deprecated
		FROM container_kinds
		WHERE name = ?
	`, payload.Name).Scan(&count, &primitive, &description, &deprecated)
	if err != nil {
		t.Fatalf("failed to query container_kinds: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 container kind, got %d", count)
	}
	if primitive != string(payload.Primitive) {
		t.Errorf("primitive = %q, want %q", primitive, payload.Primitive)
	}
	if description != payload.Description {
		t.Errorf("description = %q, want %q", description, payload.Description)
	}
	if deprecated != 0 {
		t.Errorf("deprecated = %d, want 0", deprecated)
	}
}

func TestProjectContainerKindDefineEvent_Idempotent(t *testing.T) {
	db := openTempDB(t)

	payload := types.DefineContainerKindPayload{
		Name:        "sprint",
		Primitive:   types.PrimitiveQueue,
		Description: "Initial description",
		CreatedBy:   "tester",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerKindDefine),
		Payload:   payloadJSON,
	}

	// Project first time
	if err := db.ProjectContainerKindDefineEvent(event); err != nil {
		t.Fatalf("first ProjectContainerKindDefineEvent() error = %v", err)
	}

	// Update description
	payload.Description = "Updated description"
	payloadJSON, err = json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal updated payload: %v", err)
	}
	event.Payload = payloadJSON

	// Project second time (should update description)
	if err := db.ProjectContainerKindDefineEvent(event); err != nil {
		t.Fatalf("second ProjectContainerKindDefineEvent() error = %v", err)
	}

	// Verify description was updated
	var description string
	err = db.Db.QueryRow(`
		SELECT description FROM container_kinds WHERE name = ?
	`, payload.Name).Scan(&description)
	if err != nil {
		t.Fatalf("failed to query container_kinds: %v", err)
	}

	if description != "Updated description" {
		t.Errorf("description = %q, want %q", description, "Updated description")
	}
}

func TestProjectContainerKindDeprecateEvent(t *testing.T) {
	db := openTempDB(t)

	// First define a kind
	definePayload := types.DefineContainerKindPayload{
		Name:        "old-sprint",
		Primitive:   types.PrimitiveQueue,
		Description: "Old sprint format",
		CreatedBy:   "tester",
	}

	definePayloadJSON, err := json.Marshal(definePayload)
	if err != nil {
		t.Fatalf("failed to marshal define payload: %v", err)
	}

	defineEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerKindDefine),
		Payload:   definePayloadJSON,
	}

	if err := db.ProjectContainerKindDefineEvent(defineEvent); err != nil {
		t.Fatalf("ProjectContainerKindDefineEvent() error = %v", err)
	}

	// Now deprecate it
	deprecatePayload := types.DeprecateContainerKindPayload{
		Name: "old-sprint",
	}

	deprecatePayloadJSON, err := json.Marshal(deprecatePayload)
	if err != nil {
		t.Fatalf("failed to marshal deprecate payload: %v", err)
	}

	deprecateEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerKindDeprecate),
		Payload:   deprecatePayloadJSON,
	}

	if err := db.ProjectContainerKindDeprecateEvent(deprecateEvent); err != nil {
		t.Fatalf("ProjectContainerKindDeprecateEvent() error = %v", err)
	}

	// Verify kind is deprecated
	var deprecated int
	err = db.Db.QueryRow(`
		SELECT deprecated FROM container_kinds WHERE name = ?
	`, deprecatePayload.Name).Scan(&deprecated)
	if err != nil {
		t.Fatalf("failed to query container_kinds: %v", err)
	}

	if deprecated != 1 {
		t.Errorf("deprecated = %d, want 1", deprecated)
	}
}

func TestProjectContainerCreateEvent(t *testing.T) {
	db := openTempDB(t)

	// First define a kind
	definePayload := types.DefineContainerKindPayload{
		Name:        "sprint",
		Primitive:   types.PrimitiveQueue,
		Description: "Sprint container",
		CreatedBy:   "tester",
	}
	definePayloadJSON, _ := json.Marshal(definePayload)
	defineEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerKindDefine),
		Payload:   definePayloadJSON,
	}
	db.ProjectContainerKindDefineEvent(defineEvent)

	// Now create a container
	payload := types.CreateContainerPayload{
		ID:        "q-1",
		Primitive: types.PrimitiveQueue,
		Kind:      "sprint",
		Name:      "Nov Sprint",
		Metadata: map[string]interface{}{
			"project": "lovable",
		},
		CreatedBy: "tester",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerCreate),
		Payload:   payloadJSON,
	}

	if err := db.ProjectContainerCreateEvent(event); err != nil {
		t.Fatalf("ProjectContainerCreateEvent() error = %v", err)
	}

	// Verify container was created
	var id string
	var primitive string
	var kind string
	var name string
	var removed int
	err = db.Db.QueryRow(`
		SELECT id, primitive, kind, name, removed
		FROM containers
		WHERE id = ?
	`, payload.ID).Scan(&id, &primitive, &kind, &name, &removed)
	if err != nil {
		t.Fatalf("failed to query containers: %v", err)
	}

	if id != payload.ID {
		t.Errorf("id = %q, want %q", id, payload.ID)
	}
	if primitive != string(payload.Primitive) {
		t.Errorf("primitive = %q, want %q", primitive, payload.Primitive)
	}
	if kind != payload.Kind {
		t.Errorf("kind = %q, want %q", kind, payload.Kind)
	}
	if name != payload.Name {
		t.Errorf("name = %q, want %q", name, payload.Name)
	}
	if removed != 0 {
		t.Errorf("removed = %d, want 0", removed)
	}
}

func TestProjectContainerRenameEvent(t *testing.T) {
	db := openTempDB(t)

	// Create a kind and container first
	definePayload := types.DefineContainerKindPayload{
		Name:        "sprint",
		Primitive:   types.PrimitiveQueue,
		Description: "Sprint container",
		CreatedBy:   "tester",
	}
	definePayloadJSON, _ := json.Marshal(definePayload)
	defineEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerKindDefine),
		Payload:   definePayloadJSON,
	}
	db.ProjectContainerKindDefineEvent(defineEvent)

	createPayload := types.CreateContainerPayload{
		ID:        "q-1",
		Primitive: types.PrimitiveQueue,
		Kind:      "sprint",
		Name:      "Old Name",
		CreatedBy: "tester",
	}
	createPayloadJSON, _ := json.Marshal(createPayload)
	createEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerCreate),
		Payload:   createPayloadJSON,
	}
	db.ProjectContainerCreateEvent(createEvent)

	// Now rename it
	renamePayload := types.RenameContainerPayload{
		ID:   "q-1",
		Name: "New Name",
	}

	renamePayloadJSON, err := json.Marshal(renamePayload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	renameEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerRename),
		Payload:   renamePayloadJSON,
	}

	if err := db.ProjectContainerRenameEvent(renameEvent); err != nil {
		t.Fatalf("ProjectContainerRenameEvent() error = %v", err)
	}

	// Verify name was updated
	var name string
	err = db.Db.QueryRow(`SELECT name FROM containers WHERE id = ?`, renamePayload.ID).Scan(&name)
	if err != nil {
		t.Fatalf("failed to query containers: %v", err)
	}

	if name != "New Name" {
		t.Errorf("name = %q, want %q", name, "New Name")
	}
}

func TestProjectContainerRemoveEvent(t *testing.T) {
	db := openTempDB(t)

	// Create a kind and container first
	definePayload := types.DefineContainerKindPayload{
		Name:        "sprint",
		Primitive:   types.PrimitiveQueue,
		Description: "Sprint container",
		CreatedBy:   "tester",
	}
	definePayloadJSON, _ := json.Marshal(definePayload)
	defineEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerKindDefine),
		Payload:   definePayloadJSON,
	}
	db.ProjectContainerKindDefineEvent(defineEvent)

	createPayload := types.CreateContainerPayload{
		ID:        "q-1",
		Primitive: types.PrimitiveQueue,
		Kind:      "sprint",
		Name:      "Sprint to Remove",
		CreatedBy: "tester",
	}
	createPayloadJSON, _ := json.Marshal(createPayload)
	createEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerCreate),
		Payload:   createPayloadJSON,
	}
	db.ProjectContainerCreateEvent(createEvent)

	// Now remove it
	removePayload := types.RemoveContainerPayload{
		ID: "q-1",
	}

	removePayloadJSON, err := json.Marshal(removePayload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	removeEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerRemove),
		Payload:   removePayloadJSON,
	}

	if err := db.ProjectContainerRemoveEvent(removeEvent); err != nil {
		t.Fatalf("ProjectContainerRemoveEvent() error = %v", err)
	}

	// Verify container is marked as removed
	var removed int
	err = db.Db.QueryRow(`SELECT removed FROM containers WHERE id = ?`, removePayload.ID).Scan(&removed)
	if err != nil {
		t.Fatalf("failed to query containers: %v", err)
	}

	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}
}

func TestProjectQueuePushEvent_PositionAssignment(t *testing.T) {
	db := openTempDB(t)

	// Create kind and container
	seedContainerKindAndInstance(t, db, "sprint", types.PrimitiveQueue, "q-1", "Nov Sprint")

	// Push three items
	for i, itemID := range []string{"tk-1", "tk-2", "tk-3"} {
		payload := types.QueuePushPayload{
			ContainerID: "q-1",
			ItemID:      itemID,
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

func TestProjectGroupAddEvent(t *testing.T) {
	db := openTempDB(t)

	// Create kind and container
	seedContainerKindAndInstance(t, db, "today", types.PrimitiveGroup, "g-1", "Today's Tasks")

	// Add items to group
	for i, itemID := range []string{"tk-1", "tk-2", "tk-3"} {
		payload := types.GroupAddPayload{
			ContainerID: "g-1",
			ItemID:      itemID,
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

// Helper functions for tests

func seedContainerKindAndInstance(t *testing.T, db *DB, kindName string, primitive types.ContainerPrimitive, containerID string, containerName string) {
	// Define kind
	definePayload := types.DefineContainerKindPayload{
		Name:        kindName,
		Primitive:   primitive,
		Description: "Test container",
		CreatedBy:   "tester",
	}
	definePayloadJSON, _ := json.Marshal(definePayload)
	defineEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerKindDefine),
		Payload:   definePayloadJSON,
	}
	db.ProjectContainerKindDefineEvent(defineEvent)

	// Create container
	createPayload := types.CreateContainerPayload{
		ID:        containerID,
		Primitive: primitive,
		Kind:      kindName,
		Name:      containerName,
		CreatedBy: "tester",
	}
	createPayloadJSON, _ := json.Marshal(createPayload)
	createEvent := types.Event{
		ID:        string(types.NewEventID()),
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindContainerCreate),
		Payload:   createPayloadJSON,
	}
	db.ProjectContainerCreateEvent(createEvent)
}

func seedQueueItems(t *testing.T, db *DB, containerID string, itemIDs []string) {
	for i, itemID := range itemIDs {
		payload := types.QueuePushPayload{
			ContainerID: containerID,
			ItemID:      itemID,
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
}

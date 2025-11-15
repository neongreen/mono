package database

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)
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
		Metadata: map[string]any{
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

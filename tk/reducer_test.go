package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/payloads"
	"github.com/neongreen/mono/tk/internal/types"
)

func TestReducer_TaskCreated(t *testing.T) {
	reducer := NewReducer()

	taskUID := string(NewTaskUID())
	projectUID := string(NewProjectUID())
	payload := TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: 1,
		CreatedNode:    string(NewNodeID()),
		Title:          "Test task",
		CreatedBy:      "alice",
	}
	payloadJSON, _ := json.Marshal(payload)

	now := time.Now()
	event := types.Event{
		ID:        "event_01",
		TS:        1,
		CreatedAt: now,
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.created",
		Payload:   payloadJSON,
	}

	err := reducer.Apply(event)
	if err != nil {
		t.Fatalf("Failed to apply task.created event: %v", err)
	}

	task, ok := reducer.GetTask(taskUID)
	if !ok {
		t.Fatal("Task not found")
	}

	if task.Title != "Test task" {
		t.Errorf("Expected title 'Test task', got %s", task.Title)
	}

	if task.CreatedBy != "alice" {
		t.Errorf("Expected creator alice, got %s", task.CreatedBy)
	}

	if !task.CreatedAt.Equal(now) {
		t.Errorf("Expected created_at to be %v, got %v", now, task.CreatedAt)
	}
}

func TestReducer_StatusSet(t *testing.T) {
	reducer := NewReducer()

	// Create task first
	taskUID := string(NewTaskUID())
	projectUID := string(NewProjectUID())
	createPayload := TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: 1,
		CreatedNode:    string(NewNodeID()),
		Title:          "Test task",
		CreatedBy:      "alice",
	}
	createPayloadJSON, _ := json.Marshal(createPayload)

	createEvent := types.Event{
		ID:        "event_01",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.created",
		Payload:   createPayloadJSON,
	}

	reducer.Apply(createEvent)

	// Set status
	statusPayload := payloads.TaskStatusSetPayload{
		TaskUUID: taskUID,
		Axis:     "generic",
		State:    "in_progress",
		Role:     "human",
	}
	statusPayloadJSON, _ := json.Marshal(statusPayload)

	statusEvent := types.Event{
		ID:        "event_02",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.status.set",
		Payload:   statusPayloadJSON,
	}

	err := reducer.Apply(statusEvent)
	if err != nil {
		t.Fatalf("Failed to apply task.status.set event: %v", err)
	}

	task, _ := reducer.GetTask(taskUID)

	axis, ok := task.Axes["generic"]
	if !ok {
		t.Fatal("Generic axis not found")
	}

	if axis.Effective != "in_progress" {
		t.Errorf("Expected effective status 'in_progress', got %s", axis.Effective)
	}

	if len(axis.Claims) != 1 {
		t.Errorf("Expected 1 claim, got %d", len(axis.Claims))
	}

	if axis.Claims[0].State != "in_progress" {
		t.Errorf("Expected claim state 'in_progress', got %s", axis.Claims[0].State)
	}

	if axis.Claims[0].Tentative {
		t.Error("Expected claim to not be tentative")
	}
}

func TestReducer_AuthorityResolution(t *testing.T) {
	reducer := NewReducer()

	// Create task
	taskUID := string(NewTaskUID())
	projectUID := string(NewProjectUID())
	createPayload := TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: 1,
		CreatedNode:    string(NewNodeID()),
		Title:          "Test task",
		CreatedBy:      "alice",
	}
	createPayloadJSON, _ := json.Marshal(createPayload)

	createEvent := types.Event{
		ID:        "event_01",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.created",
		Payload:   createPayloadJSON,
	}

	reducer.Apply(createEvent)

	// Agent sets status to done
	agentPayload := payloads.TaskStatusSetPayload{
		TaskUUID: taskUID,
		Axis:     "generic",
		State:    "done",
		Role:     "agent",
	}
	agentPayloadJSON, _ := json.Marshal(agentPayload)

	agentEvent := types.Event{
		ID:        "event_02",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "claude",
		Role:      "agent",
		Kind:      "task.status.set",
		Payload:   agentPayloadJSON,
	}

	reducer.Apply(agentEvent)

	// Human sets status to in_progress (same timestamp for concurrent claim)
	humanPayload := payloads.TaskStatusSetPayload{
		TaskUUID: taskUID,
		Axis:     "generic",
		State:    "in_progress",
		Role:     "human",
	}
	humanPayloadJSON, _ := json.Marshal(humanPayload)

	humanEvent := types.Event{
		ID:        "event_03",
		TS:        2, // Same timestamp as agent
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.status.set",
		Payload:   humanPayloadJSON,
	}

	reducer.Apply(humanEvent)

	task, _ := reducer.GetTask(taskUID)
	axis := task.Axes["generic"]

	// Human claim should win due to higher authority
	if axis.Effective != "in_progress" {
		t.Errorf("Expected effective status 'in_progress' (human wins), got %s", axis.Effective)
	}

	// Check tentative flags
	var humanClaim, agentClaim *types.Claim
	for i := range axis.Claims {
		if axis.Claims[i].Role == "human" {
			humanClaim = &axis.Claims[i]
		}
		if axis.Claims[i].Role == "agent" {
			agentClaim = &axis.Claims[i]
		}
	}

	if humanClaim == nil || agentClaim == nil {
		t.Fatal("Missing claims")
	}

	if humanClaim.Tentative {
		t.Error("Human claim should not be tentative")
	}

	if !agentClaim.Tentative {
		t.Error("Agent claim should be tentative")
	}
}

func TestReducer_NoteAdd(t *testing.T) {
	reducer := NewReducer()

	// Create task
	taskUID := string(NewTaskUID())
	projectUID := string(NewProjectUID())
	createPayload := TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: 1,
		CreatedNode:    string(NewNodeID()),
		Title:          "Test task",
		CreatedBy:      "alice",
	}
	createPayloadJSON, _ := json.Marshal(createPayload)

	createEvent := types.Event{
		ID:        "event_01",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.created",
		Payload:   createPayloadJSON,
	}

	reducer.Apply(createEvent)

	// Add note
	notePayload := payloads.TaskNoteAddPayload{
		TaskUUID: taskUID,
		Markdown: "This is a test note",
	}
	notePayloadJSON, _ := json.Marshal(notePayload)

	noteEvent := types.Event{
		ID:        "event_02",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.note.add",
		Payload:   notePayloadJSON,
	}

	err := reducer.Apply(noteEvent)
	if err != nil {
		t.Fatalf("Failed to apply task.note.add event: %v", err)
	}

	task, _ := reducer.GetTask(taskUID)

	if len(task.Notes) != 1 {
		t.Fatalf("Expected 1 note, got %d", len(task.Notes))
	}

	if task.Notes[0].Markdown != "This is a test note" {
		t.Errorf("Expected note 'This is a test note', got %s", task.Notes[0].Markdown)
	}

	if task.Notes[0].Actor != "alice" {
		t.Errorf("Expected note actor alice, got %s", task.Notes[0].Actor)
	}
}

func TestGetRoleAuthority(t *testing.T) {
	tests := []struct {
		role     string
		expected int
	}{
		{"human", 5},
		{"qa", 4},
		{"rel", 3},
		{"agent", 2},
		{"bot", 1},
		{"unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			authority := types.GetRoleAuthority(tt.role)
			if authority != tt.expected {
				t.Errorf("Expected authority %d for role %s, got %d", tt.expected, tt.role, authority)
			}
		})
	}
}

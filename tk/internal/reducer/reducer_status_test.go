package reducer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

func TestReducer_StatusSet(t *testing.T) {
	r := NewReducer()

	// Create task first
	taskUID := string(types.NewTaskUID())
	projectUID := string(types.NewProjectUID())
	createPayload := types.TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: 1,
		CreatedNode:    string(types.NewNodeID()),
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

	r.Apply(createEvent)

	// Set status
	statusPayload := types.TaskStatusSetPayload{
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

	err := r.Apply(statusEvent)
	if err != nil {
		t.Fatalf("Failed to apply task.status.set event: %v", err)
	}

	task, _ := r.GetTask(taskUID)

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
	r := NewReducer()

	// Create task
	taskUID := string(types.NewTaskUID())
	projectUID := string(types.NewProjectUID())
	createPayload := types.TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: 1,
		CreatedNode:    string(types.NewNodeID()),
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

	r.Apply(createEvent)

	// Agent sets status to done
	agentPayload := types.TaskStatusSetPayload{
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

	r.Apply(agentEvent)

	// Human sets status to in_progress (same timestamp for concurrent claim)
	humanPayload := types.TaskStatusSetPayload{
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

	r.Apply(humanEvent)

	task, _ := r.GetTask(taskUID)
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

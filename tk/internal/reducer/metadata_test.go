package reducer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

// marshalPayload is a local test helper (can't use testutil due to import cycle)
func marshalPayload(v any) json.RawMessage {
	data, _ := json.Marshal(v)
	return json.RawMessage(data)
}

// TestMetadata_SingleClaim tests that a single metadata claim sets the effective value
func TestMetadata_SingleClaim(t *testing.T) {
	reducer := NewReducer()

	// Create a task
	createEvent := types.Event{
		ID:        "ev-1-node1",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	if err := reducer.Apply(createEvent); err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Set metadata
	metaEvent := types.Event{
		ID:        "ev-2-node1",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", TaskID: "", Key: "priority", Value: json.RawMessage(`1`)}),
	}
	if err := reducer.Apply(metaEvent); err != nil {
		t.Fatalf("Failed to set metadata: %v", err)
	}

	// Verify
	task, ok := reducer.GetTask("tsk_123")
	if !ok {
		t.Fatal("Task not found")
	}

	metaStatus, ok := task.Metadata["priority"]
	if !ok {
		t.Fatal("Metadata 'priority' not found")
	}

	if string(metaStatus.Effective) != "1" {
		t.Errorf("Expected effective value 1, got %s", metaStatus.Effective)
	}

	if len(metaStatus.Claims) != 1 {
		t.Errorf("Expected 1 claim, got %d", len(metaStatus.Claims))
	}

	if metaStatus.Claims[0].Tentative {
		t.Error("Single claim should not be tentative")
	}
}

// TestMetadata_HumanOverridesAgent tests authority resolution
func TestMetadata_HumanOverridesAgent(t *testing.T) {
	reducer := NewReducer()

	// Create task
	createEvent := types.Event{
		ID:        "ev-1-node1",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	// Agent sets priority to 3
	agentEvent := types.Event{
		ID:        "ev-2-node1",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "claude",
		Role:      "agent",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`3`)}),
	}
	reducer.Apply(agentEvent)

	// Human sets priority to 1
	humanEvent := types.Event{
		ID:        "ev-3-node1",
		TS:        3,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`1`)}),
	}
	reducer.Apply(humanEvent)

	// Verify human wins
	task, _ := reducer.GetTask("tsk_123")
	metaStatus := task.Metadata["priority"]

	if string(metaStatus.Effective) != "1" {
		t.Errorf("Expected effective value 1 (human), got %s", metaStatus.Effective)
	}

	// Find agent claim
	var agentClaim *types.MetadataClaim
	for i := range metaStatus.Claims {
		if metaStatus.Claims[i].Role == "agent" {
			agentClaim = &metaStatus.Claims[i]
			break
		}
	}

	if agentClaim == nil {
		t.Fatal("Agent claim not found")
	}

	if !agentClaim.Tentative {
		t.Error("Agent claim should be tentative when human overrides")
	}
}

// TestMetadata_SameRole_LatestWins tests timestamp tie-breaking
func TestMetadata_SameRole_LatestWins(t *testing.T) {
	reducer := NewReducer()

	createEvent := types.Event{
		ID:        "ev-1-node1",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	// First human claim at TS=2
	event1 := types.Event{
		ID:        "ev-2-node1",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`1`)}),
	}
	reducer.Apply(event1)

	// Second human claim at TS=5
	event2 := types.Event{
		ID:        "ev-3-node1",
		TS:        5,
		CreatedAt: time.Now(),
		Actor:     "bob",
		Role:      "human",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`2`)}),
	}
	reducer.Apply(event2)

	// Verify latest timestamp wins
	task, _ := reducer.GetTask("tsk_123")
	metaStatus := task.Metadata["priority"]

	if string(metaStatus.Effective) != "2" {
		t.Errorf("Expected effective value 2 (latest), got %s", metaStatus.Effective)
	}

	// Earlier claim should be tentative
	if !metaStatus.Claims[0].Tentative {
		t.Error("Earlier claim should be tentative")
	}
	if metaStatus.Claims[1].Tentative {
		t.Error("Latest claim should not be tentative")
	}
}

// TestMetadata_MultipleKeys tests that different keys are independent
func TestMetadata_MultipleKeys(t *testing.T) {
	reducer := NewReducer()

	createEvent := types.Event{
		ID:        "ev-1-node1",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	// Set priority
	priorityEvent := types.Event{
		ID:        "ev-2-node1",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`1`)}),
	}
	reducer.Apply(priorityEvent)

	// Set labels
	labelsEvent := types.Event{
		ID:        "ev-3-node1",
		TS:        3,
		CreatedAt: time.Now(),
		Actor:     "bob",
		Role:      "agent",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "labels", Value: json.RawMessage(`["bug","urgent"]`)}),
	}
	reducer.Apply(labelsEvent)

	// Verify both keys exist independently
	task, _ := reducer.GetTask("tsk_123")

	if _, ok := task.Metadata["priority"]; !ok {
		t.Error("Priority metadata not found")
	}
	if _, ok := task.Metadata["labels"]; !ok {
		t.Error("Labels metadata not found")
	}

	if string(task.Metadata["priority"].Effective) != "1" {
		t.Errorf("Expected priority=1, got %s", task.Metadata["priority"].Effective)
	}
	if string(task.Metadata["labels"].Effective) != `["bug","urgent"]` {
		t.Errorf("Expected labels array, got %s", task.Metadata["labels"].Effective)
	}
}

// TestMetadata_ConcurrentClaims tests multiple concurrent claims
func TestMetadata_ConcurrentClaims(t *testing.T) {
	reducer := NewReducer()

	createEvent := types.Event{
		ID:        "ev-1-node1",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	// Agent claim
	agentEvent := types.Event{
		ID:        "ev-2-node1",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "claude",
		Role:      "agent",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`3`)}),
	}
	reducer.Apply(agentEvent)

	// QA claim
	qaEvent := types.Event{
		ID:        "ev-3-node1",
		TS:        3,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "qa",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`2`)}),
	}
	reducer.Apply(qaEvent)

	// Bot claim
	botEvent := types.Event{
		ID:        "ev-4-node1",
		TS:        4,
		CreatedAt: time.Now(),
		Actor:     "bot",
		Role:      "bot",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`4`)}),
	}
	reducer.Apply(botEvent)

	// Verify QA wins (highest authority among claims)
	task, _ := reducer.GetTask("tsk_123")
	metaStatus := task.Metadata["priority"]

	if string(metaStatus.Effective) != "2" {
		t.Errorf("Expected effective value 2 (QA), got %s", metaStatus.Effective)
	}

	// Check tentative flags
	for i := range metaStatus.Claims {
		claim := metaStatus.Claims[i]
		if claim.Role == "qa" && claim.Tentative {
			t.Error("QA claim should not be tentative")
		}
		if (claim.Role == "agent" || claim.Role == "bot") && !claim.Tentative {
			t.Errorf("Lower authority claim (%s) should be tentative", claim.Role)
		}
	}
}

// TestMetadata_EventsOutOfOrder tests deterministic resolution regardless of metadata event order
func TestMetadata_EventsOutOfOrder(t *testing.T) {
	// Create task first
	createEvent := types.Event{
		ID: "ev-1-node1", TS: 1, CreatedAt: time.Now(), Actor: "alice", Role: "human",
		Kind:    string(types.EventKindTaskCreated),
		Payload: marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}

	// Metadata events to test ordering
	metaEvents := []types.Event{
		{ID: "ev-2-node1", TS: 2, CreatedAt: time.Now(), Actor: "claude", Role: "agent", Kind: string(types.EventKindTaskMetaSet),
			Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`3`)})},
		{ID: "ev-3-node1", TS: 3, CreatedAt: time.Now(), Actor: "alice", Role: "human", Kind: string(types.EventKindTaskMetaSet),
			Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`1`)})},
	}

	// Test with metadata events in chronological order
	reducer1 := NewReducer()
	reducer1.Apply(createEvent)
	for _, e := range metaEvents {
		reducer1.Apply(e)
	}

	// Test with metadata events in reverse order
	reducer2 := NewReducer()
	reducer2.Apply(createEvent)
	for i := len(metaEvents) - 1; i >= 0; i-- {
		reducer2.Apply(metaEvents[i])
	}

	// Both should have same effective value (determined by authority+ts, not apply order)
	task1, _ := reducer1.GetTask("tsk_123")
	task2, _ := reducer2.GetTask("tsk_123")

	effective1 := string(task1.Metadata["priority"].Effective)
	effective2 := string(task2.Metadata["priority"].Effective)

	if effective1 != effective2 {
		t.Errorf("Out-of-order events produced different results: %s vs %s", effective1, effective2)
	}

	if effective1 != "1" {
		t.Errorf("Expected human claim (1) to win, got %s", effective1)
	}
}

// TestMetadata_NumberValue tests number values
func TestMetadata_NumberValue(t *testing.T) {
	reducer := NewReducer()

	createEvent := types.Event{
		ID: "ev-1-node1", TS: 1, Actor: "alice", Role: "human",
		Kind:    string(types.EventKindTaskCreated),
		Payload: marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	metaEvent := types.Event{
		ID:      "ev-2-node1",
		TS:      2,
		Actor:   "alice",
		Role:    "human",
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`42`)}),
	}
	reducer.Apply(metaEvent)

	task, _ := reducer.GetTask("tsk_123")
	if string(task.Metadata["priority"].Effective) != "42" {
		t.Errorf("Expected 42, got %s", task.Metadata["priority"].Effective)
	}
}

// TestMetadata_StringValue tests string values
func TestMetadata_StringValue(t *testing.T) {
	reducer := NewReducer()

	createEvent := types.Event{
		ID: "ev-1-node1", TS: 1, Actor: "alice", Role: "human",
		Kind:    string(types.EventKindTaskCreated),
		Payload: marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	metaEvent := types.Event{
		ID:      "ev-2-node1",
		TS:      2,
		Actor:   "alice",
		Role:    "human",
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "assignee", Value: json.RawMessage(`"alice"`)}),
	}
	reducer.Apply(metaEvent)

	task, _ := reducer.GetTask("tsk_123")
	if string(task.Metadata["assignee"].Effective) != `"alice"` {
		t.Errorf("Expected \"alice\", got %s", task.Metadata["assignee"].Effective)
	}
}

// TestMetadata_ArrayValue tests array values
func TestMetadata_ArrayValue(t *testing.T) {
	reducer := NewReducer()

	createEvent := types.Event{
		ID: "ev-1-node1", TS: 1, Actor: "alice", Role: "human",
		Kind:    string(types.EventKindTaskCreated),
		Payload: marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	metaEvent := types.Event{
		ID:      "ev-2-node1",
		TS:      2,
		Actor:   "alice",
		Role:    "human",
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "labels", Value: json.RawMessage(`["bug","urgent","p1"]`)}),
	}
	reducer.Apply(metaEvent)

	task, _ := reducer.GetTask("tsk_123")
	expected := `["bug","urgent","p1"]`
	if string(task.Metadata["labels"].Effective) != expected {
		t.Errorf("Expected %s, got %s", expected, task.Metadata["labels"].Effective)
	}
}

// TestMetadata_ObjectValue tests object values
func TestMetadata_ObjectValue(t *testing.T) {
	reducer := NewReducer()

	createEvent := types.Event{
		ID: "ev-1-node1", TS: 1, Actor: "alice", Role: "human",
		Kind:    string(types.EventKindTaskCreated),
		Payload: marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	metaEvent := types.Event{
		ID:      "ev-2-node1",
		TS:      2,
		Actor:   "alice",
		Role:    "human",
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "sla", Value: json.RawMessage(`{"days":7,"hours":168}`)}),
	}
	reducer.Apply(metaEvent)

	task, _ := reducer.GetTask("tsk_123")
	expected := `{"days":7,"hours":168}`
	if string(task.Metadata["sla"].Effective) != expected {
		t.Errorf("Expected %s, got %s", expected, task.Metadata["sla"].Effective)
	}
}

// TestMetadata_InvalidJSON tests error handling for invalid JSON
func TestMetadata_InvalidJSON(t *testing.T) {
	reducer := NewReducer()

	createEvent := types.Event{
		ID: "ev-1-node1", TS: 1, Actor: "alice", Role: "human",
		Kind:    string(types.EventKindTaskCreated),
		Payload: marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	// Try to set invalid JSON
	metaEvent := types.Event{
		ID:      "ev-2-node1",
		TS:      2,
		Actor:   "alice",
		Role:    "human",
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "bad", Value: json.RawMessage(`{invalid}`)}),
	}

	err := reducer.Apply(metaEvent)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

// TestMetadata_MissingRole tests that missing role defaults to human
func TestMetadata_MissingRole(t *testing.T) {
	reducer := NewReducer()

	createEvent := types.Event{
		ID: "ev-1-node1", TS: 1, Actor: "alice", Role: "human",
		Kind:    string(types.EventKindTaskCreated),
		Payload: marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	// Event with empty role
	metaEvent := types.Event{
		ID:      "ev-2-node1",
		TS:      2,
		Actor:   "alice",
		Role:    "", // Empty role
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`1`)}),
	}
	reducer.Apply(metaEvent)

	task, _ := reducer.GetTask("tsk_123")
	metaStatus := task.Metadata["priority"]

	if len(metaStatus.Claims) == 0 {
		t.Fatal("Expected at least one claim")
	}

	if metaStatus.Claims[0].Role != "human" {
		t.Errorf("Expected role to default to 'human', got %s", metaStatus.Claims[0].Role)
	}
}

// TestMetadata_ThreeWayConflict tests resolution with three different roles
func TestMetadata_ThreeWayConflict(t *testing.T) {
	reducer := NewReducer()

	createEvent := types.Event{
		ID: "ev-1-node1", TS: 1, Actor: "alice", Role: "human",
		Kind:    string(types.EventKindTaskCreated),
		Payload: marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	// Bot sets priority=4
	reducer.Apply(types.Event{
		ID: "ev-2-node1", TS: 2, Actor: "bot", Role: "bot",
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`4`)}),
	})

	// Agent sets priority=3
	reducer.Apply(types.Event{
		ID: "ev-3-node1", TS: 3, Actor: "agent", Role: "agent",
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`3`)}),
	})

	// Human sets priority=1
	reducer.Apply(types.Event{
		ID: "ev-4-node1", TS: 4, Actor: "human", Role: "human",
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`1`)}),
	})

	// Verify human wins
	task, _ := reducer.GetTask("tsk_123")
	metaStatus := task.Metadata["priority"]

	if string(metaStatus.Effective) != "1" {
		t.Errorf("Expected human value (1), got %s", metaStatus.Effective)
	}

	// Verify all claims present but lower ones tentative
	if len(metaStatus.Claims) != 3 {
		t.Errorf("Expected 3 claims, got %d", len(metaStatus.Claims))
	}

	for _, claim := range metaStatus.Claims {
		if claim.Role == "human" && claim.Tentative {
			t.Error("Human claim should not be tentative")
		}
		if claim.Role != "human" && !claim.Tentative {
			t.Errorf("Non-human claim (%s) should be tentative", claim.Role)
		}
	}
}

// TestMetadata_Idempotency tests reapplying same event
func TestMetadata_Idempotency(t *testing.T) {
	reducer := NewReducer()

	createEvent := types.Event{
		ID: "ev-1-node1", TS: 1, Actor: "alice", Role: "human",
		Kind:    string(types.EventKindTaskCreated),
		Payload: marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	metaEvent := types.Event{
		ID:      "ev-2-node1",
		TS:      2,
		Actor:   "alice",
		Role:    "human",
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`1`)}),
	}

	// Apply once
	reducer.Apply(metaEvent)

	// Apply again (simulating duplicate event)
	reducer.Apply(metaEvent)
	task, _ := reducer.GetTask("tsk_123")
	claims := len(task.Metadata["priority"].Claims)

	// Should have 2 claims (reducer doesn't dedupe by event ID, that's DB's job)
	if claims != 2 {
		t.Errorf("Expected 2 claims after reapplying, got %d", claims)
	}
}

// TestMetadata_AgentBeatsBot tests authority chain
func TestMetadata_AgentBeatsBot(t *testing.T) {
	reducer := NewReducer()

	createEvent := types.Event{
		ID: "ev-1-node1", TS: 1, Actor: "alice", Role: "human",
		Kind:    string(types.EventKindTaskCreated),
		Payload: marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	// Bot claim
	reducer.Apply(types.Event{
		ID: "ev-2-node1", TS: 2, Actor: "bot", Role: "bot",
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`4`)}),
	})

	// Agent claim
	reducer.Apply(types.Event{
		ID: "ev-3-node1", TS: 3, Actor: "agent", Role: "agent",
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`2`)}),
	})

	task, _ := reducer.GetTask("tsk_123")
	if string(task.Metadata["priority"].Effective) != "2" {
		t.Errorf("Expected agent value (2), got %s", task.Metadata["priority"].Effective)
	}
}

// TestMetadata_QABeatsAgent tests qa > agent authority
func TestMetadata_QABeatsAgent(t *testing.T) {
	reducer := NewReducer()

	createEvent := types.Event{
		ID: "ev-1-node1", TS: 1, Actor: "alice", Role: "human",
		Kind:    string(types.EventKindTaskCreated),
		Payload: marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	// Agent claim (later timestamp)
	reducer.Apply(types.Event{
		ID: "ev-2-node1", TS: 10, Actor: "agent", Role: "agent",
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`3`)}),
	})

	// QA claim (earlier timestamp but higher authority)
	reducer.Apply(types.Event{
		ID: "ev-3-node1", TS: 5, Actor: "qa", Role: "qa",
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "priority", Value: json.RawMessage(`1`)}),
	})

	task, _ := reducer.GetTask("tsk_123")
	if string(task.Metadata["priority"].Effective) != "1" {
		t.Errorf("Expected QA value (1) despite earlier timestamp, got %s", task.Metadata["priority"].Effective)
	}
}

// TestMetadata_EmptyArray tests empty array value
func TestMetadata_EmptyArray(t *testing.T) {
	reducer := NewReducer()

	createEvent := types.Event{
		ID: "ev-1-node1", TS: 1, Actor: "alice", Role: "human",
		Kind:    string(types.EventKindTaskCreated),
		Payload: marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	metaEvent := types.Event{
		ID:      "ev-2-node1",
		TS:      2,
		Actor:   "alice",
		Role:    "human",
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "labels", Value: json.RawMessage(`[]`)}),
	}
	reducer.Apply(metaEvent)

	task, _ := reducer.GetTask("tsk_123")
	if string(task.Metadata["labels"].Effective) != "[]" {
		t.Errorf("Expected empty array [], got %s", task.Metadata["labels"].Effective)
	}
}

// TestMetadata_NullValue tests null value
func TestMetadata_NullValue(t *testing.T) {
	reducer := NewReducer()

	createEvent := types.Event{
		ID: "ev-1-node1", TS: 1, Actor: "alice", Role: "human",
		Kind:    string(types.EventKindTaskCreated),
		Payload: marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	metaEvent := types.Event{
		ID:      "ev-2-node1",
		TS:      2,
		Actor:   "alice",
		Role:    "human",
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "notes", Value: json.RawMessage(`null`)}),
	}
	reducer.Apply(metaEvent)

	task, _ := reducer.GetTask("tsk_123")
	if string(task.Metadata["notes"].Effective) != "null" {
		t.Errorf("Expected null, got %s", task.Metadata["notes"].Effective)
	}
}

// TestMetadata_BoolValue tests boolean value
func TestMetadata_BoolValue(t *testing.T) {
	reducer := NewReducer()

	createEvent := types.Event{
		ID: "ev-1-node1", TS: 1, Actor: "alice", Role: "human",
		Kind:    string(types.EventKindTaskCreated),
		Payload: marshalPayload(types.TaskCreatedPayload{TaskUID: "tsk_123", ProjectUID: "prj_1", Title: "Test", CreatedBy: "alice", CreatedNode: "node1"}),
	}
	reducer.Apply(createEvent)

	metaEvent := types.Event{
		ID:      "ev-2-node1",
		TS:      2,
		Actor:   "alice",
		Role:    "human",
		Kind:    string(types.EventKindTaskMetaSet),
		Payload: marshalPayload(types.TaskMetaSetPayload{TaskUUID: "tsk_123", Key: "urgent", Value: json.RawMessage(`true`)}),
	}
	reducer.Apply(metaEvent)

	task, _ := reducer.GetTask("tsk_123")
	if string(task.Metadata["urgent"].Effective) != "true" {
		t.Errorf("Expected true, got %s", task.Metadata["urgent"].Effective)
	}
}

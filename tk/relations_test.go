package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/payloads"
	"github.com/neongreen/mono/tk/internal/relations"
	"github.com/neongreen/mono/tk/internal/types"
)

func TestRelationGraph_AddRemove(t *testing.T) {
	graph := relations.NewRelationsGraph()

	// Add a relation
	graph.AddRelation("task-a", "blocks", "task-b", "note1", "ev-1-node1", "node1", 1)

	// Verify it exists
	out := graph.GetOutgoingRelations("task-a", "blocks")
	if len(out) != 1 {
		t.Fatalf("Expected 1 outgoing relation, got %d", len(out))
	}
	if out[0].TaskUUID != "task-b" {
		t.Errorf("Expected task-b, got %s", out[0].TaskUUID)
	}

	in := graph.GetIncomingRelations("task-b", "blocks")
	if len(in) != 1 {
		t.Fatalf("Expected 1 incoming relation, got %d", len(in))
	}
	if in[0].TaskUUID != "task-a" {
		t.Errorf("Expected task-a, got %s", in[0].TaskUUID)
	}

	// Remove the relation
	graph.RemoveRelation("task-a", "blocks", "task-b", "ev-2-node1", "node1", 2)

	// Verify it's removed
	out = graph.GetOutgoingRelations("task-a", "blocks")
	if len(out) != 0 {
		t.Errorf("Expected 0 outgoing relations after remove, got %d", len(out))
	}
}

func TestRelationGraph_ORSetSemantics(t *testing.T) {
	graph := relations.NewRelationsGraph()

	// Add from two different nodes (two separate add events)
	graph.AddRelation("task-a", "blocks", "task-b", "", "ev-1-node1", "node1", 1)
	graph.AddRelation("task-a", "blocks", "task-b", "", "ev-2-node2", "node2", 2)

	// Verify it exists
	out := graph.GetOutgoingRelations("task-a", "blocks")
	if len(out) != 1 {
		t.Fatalf("Expected 1 outgoing relation, got %d", len(out))
	}

	// Remove - this observes both adds and tombstones them
	graph.RemoveRelation("task-a", "blocks", "task-b", "ev-3-node1", "node1", 3)

	// Should be removed (all observed adds are tombstoned)
	out = graph.GetOutgoingRelations("task-a", "blocks")
	if len(out) != 0 {
		t.Errorf("Expected 0 outgoing relations after remove (remove-wins), got %d", len(out))
	}

	// Add again after remove - this should resurrect the edge
	graph.AddRelation("task-a", "blocks", "task-b", "", "ev-4-node2", "node2", 4)

	// Should exist again (new add not observed by previous remove)
	out = graph.GetOutgoingRelations("task-a", "blocks")
	if len(out) != 1 {
		t.Errorf("Expected 1 outgoing relation after add-after-remove, got %d", len(out))
	}
}

func TestRelationGraph_CycleDetection(t *testing.T) {
	graph := relations.NewRelationsGraph()

	// Create a cycle: A -> B -> C -> A
	graph.AddRelation("task-a", "blocks", "task-b", "", "ev-1-node1", "node1", 1)
	graph.AddRelation("task-b", "blocks", "task-c", "", "ev-2-node1", "node1", 2)
	graph.AddRelation("task-c", "blocks", "task-a", "", "ev-3-node1", "node1", 3)

	cycles := graph.DetectCycles("blocks")
	if len(cycles) == 0 {
		t.Fatal("Expected to detect a cycle")
	}

	// Should detect cycle involving a, b, c
	foundCycle := false
	for _, cycle := range cycles {
		if len(cycle) == 3 {
			foundCycle = true
			break
		}
	}
	if !foundCycle {
		t.Error("Expected to find a 3-node cycle")
	}
}

func TestRelationGraph_ComputeBlocked(t *testing.T) {
	graph := relations.NewRelationsGraph()
	tasks := make(map[string]*Task)

	// Create tasks
	tasks["task-a"] = &Task{
		TaskUUID: "task-a",
		TaskID:   "tk-1",
		Title:    "Task A",
		Axes: map[string]types.AxisStatus{
			"generic": {Effective: "in_progress"},
		},
	}
	tasks["task-b"] = &Task{
		TaskUUID: "task-b",
		TaskID:   "tk-2",
		Title:    "Task B",
		Axes: map[string]types.AxisStatus{
			"generic": {Effective: "in_progress"},
		},
	}
	tasks["task-c"] = &Task{
		TaskUUID: "task-c",
		TaskID:   "tk-3",
		Title:    "Task C",
		Axes: map[string]types.AxisStatus{
			"generic": {Effective: "done"},
		},
	}

	// A blocks B (A is in progress, so B is blocked)
	graph.AddRelation("task-a", "blocks", "task-b", "", "ev-1-node1", "node1", 1)

	// C blocks B (C is done, so shouldn't block B)
	graph.AddRelation("task-c", "blocks", "task-b", "", "ev-2-node1", "node1", 2)

	// Compute blocked status
	ComputeBlocked(graph, tasks, "generic", []string{"done"})

	// Task A should not be blocked
	if tasks["task-a"].Blocked {
		t.Error("Task A should not be blocked")
	}

	// Task B should be blocked (by task A, not by task C which is done)
	if !tasks["task-b"].Blocked {
		t.Error("Task B should be blocked")
	}

	if len(tasks["task-b"].Blockers) != 1 {
		t.Errorf("Task B should have 1 blocker, got %d", len(tasks["task-b"].Blockers))
	}

	if len(tasks["task-b"].Blockers) > 0 && tasks["task-b"].Blockers[0].TaskID != "tk-1" {
		t.Errorf("Task B should be blocked by tk-1, got %s", tasks["task-b"].Blockers[0].TaskID)
	}

	// Task C should not be blocked
	if tasks["task-c"].Blocked {
		t.Error("Task C should not be blocked")
	}
}

func TestReducer_RelationEvents(t *testing.T) {
	reducer := NewReducer()

	// Create tasks first
	createTaskA := Event{
		ID:        "ev-1-node1",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.created",
		Payload:   marshalPayload(TaskCreatedPayload{TaskUID: "task-a", ProjectUID: string(NewProjectUID()), ProposedNumber: 1, CreatedNode: string(NewNodeID()), Title: "Task A", CreatedBy: "alice"}),
	}
	createTaskB := Event{
		ID:        "ev-2-node1",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.created",
		Payload:   marshalPayload(TaskCreatedPayload{TaskUID: "task-b", ProjectUID: string(NewProjectUID()), ProposedNumber: 1, CreatedNode: string(NewNodeID()), Title: "Task B", CreatedBy: "alice"}),
	}

	if err := reducer.Apply(createTaskA); err != nil {
		t.Fatalf("Failed to apply task.created: %v", err)
	}
	if err := reducer.Apply(createTaskB); err != nil {
		t.Fatalf("Failed to apply task.created: %v", err)
	}

	// Add a relation
	relationAddEvent := Event{
		ID:        "ev-3-node1",
		TS:        3,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "relation.add",
		Payload:   marshalPayload(payloads.RelationAddPayload{Src: "task-a", Type: "blocks", Dst: "task-b", Note: "test"}),
	}

	if err := reducer.Apply(relationAddEvent); err != nil {
		t.Fatalf("Failed to apply relation.add: %v", err)
	}

	// Verify relation exists
	out := reducer.relations.GetOutgoingRelations("task-a", "blocks")
	if len(out) != 1 {
		t.Fatalf("Expected 1 outgoing relation, got %d", len(out))
	}

	// Remove the relation
	relationRemoveEvent := Event{
		ID:        "ev-4-node1",
		TS:        4,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "relation.remove",
		Payload:   marshalPayload(payloads.RelationRemovePayload{Src: "task-a", Type: "blocks", Dst: "task-b"}),
	}

	if err := reducer.Apply(relationRemoveEvent); err != nil {
		t.Fatalf("Failed to apply relation.remove: %v", err)
	}

	// Verify relation is removed
	out = reducer.relations.GetOutgoingRelations("task-a", "blocks")
	if len(out) != 0 {
		t.Errorf("Expected 0 outgoing relations after remove, got %d", len(out))
	}
}

func TestBuildTaskRelations(t *testing.T) {
	graph := relations.NewRelationsGraph()

	// Add various relations
	graph.AddRelation("task-a", "blocks", "task-b", "", "ev-1-node1", "node1", 1)
	graph.AddRelation("task-a", "subtask", "task-c", "", "ev-2-node1", "node1", 2)
	graph.AddRelation("task-a", "related", "task-d", "", "ev-3-node1", "node1", 3)

	// Build relations for task-a
	relations := graph.BuildTaskRelations("task-a")

	if relations == nil {
		t.Fatal("Expected relations to be non-nil")
	}

	// Check blocks
	if len(relations.Blocks.Out) != 1 {
		t.Errorf("Expected 1 blocks out, got %d", len(relations.Blocks.Out))
	}
	if relations.Blocks.Out[0].TaskUUID != "task-b" {
		t.Errorf("Expected task-b, got %s", relations.Blocks.Out[0].TaskUUID)
	}

	// Check subtask
	if len(relations.Subtask.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(relations.Subtask.Children))
	}
	if relations.Subtask.Children[0] != "task-c" {
		t.Errorf("Expected task-c, got %s", relations.Subtask.Children[0])
	}

	// Check related
	if len(relations.Related.Out) != 1 {
		t.Errorf("Expected 1 related out, got %d", len(relations.Related.Out))
	}
	if relations.Related.Out[0].TaskUUID != "task-d" {
		t.Errorf("Expected task-d, got %s", relations.Related.Out[0].TaskUUID)
	}
}

// Helper to marshal payloads
func marshalPayload(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return json.RawMessage(data)
}

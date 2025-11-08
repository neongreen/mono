package cmd

import (
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/reducer"
	"github.com/neongreen/mono/tk/internal/relations"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
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
	tasks := make(map[string]*types.Task)

	// Create tasks
	tasks["task-a"] = &types.Task{
		TaskUUID: "task-a",
		TaskID:   "tk-1",
		Title:    "Task A",
		Axes: map[string]types.AxisStatus{
			"generic": {Effective: "in_progress"},
		},
	}
	tasks["task-b"] = &types.Task{
		TaskUUID: "task-b",
		TaskID:   "tk-2",
		Title:    "Task B",
		Axes: map[string]types.AxisStatus{
			"generic": {Effective: "in_progress"},
		},
	}
	tasks["task-c"] = &types.Task{
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
	utils.ComputeBlocked(graph, tasks, "generic", []string{"done"})

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
	reducer := reducer.NewReducer()

	// Create tasks first
	createTaskA := types.Event{
		ID:        "ev-1-node1",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.created",
		Payload:   marshalPayload(types.TaskCreatedPayload{TaskUID: "task-a", ProjectUID: string(types.NewProjectUID()), ProposedNumber: 1, CreatedNode: string(types.NewNodeID()), Title: "Task A", CreatedBy: "alice"}),
	}
	createTaskB := types.Event{
		ID:        "ev-2-node1",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.created",
		Payload:   marshalPayload(types.TaskCreatedPayload{TaskUID: "task-b", ProjectUID: string(types.NewProjectUID()), ProposedNumber: 1, CreatedNode: string(types.NewNodeID()), Title: "Task B", CreatedBy: "alice"}),
	}

	if err := reducer.Apply(createTaskA); err != nil {
		t.Fatalf("Failed to apply task.created: %v", err)
	}
	if err := reducer.Apply(createTaskB); err != nil {
		t.Fatalf("Failed to apply task.created: %v", err)
	}

	// Add a relation
	relationAddEvent := types.Event{
		ID:        "ev-3-node1",
		TS:        3,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "relation.add",
		Payload:   marshalPayload(types.RelationAddPayload{Src: "task-a", Type: "blocks", Dst: "task-b", Note: "test"}),
	}

	if err := reducer.Apply(relationAddEvent); err != nil {
		t.Fatalf("Failed to apply relation.add: %v", err)
	}

	// Verify relation exists
	out := reducer.Relations().GetOutgoingRelations("task-a", "blocks")
	if len(out) != 1 {
		t.Fatalf("Expected 1 outgoing relation, got %d", len(out))
	}

	// Remove the relation
	relationRemoveEvent := types.Event{
		ID:        "ev-4-node1",
		TS:        4,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "relation.remove",
		Payload:   marshalPayload(types.RelationRemovePayload{Src: "task-a", Type: "blocks", Dst: "task-b"}),
	}

	if err := reducer.Apply(relationRemoveEvent); err != nil {
		t.Fatalf("Failed to apply relation.remove: %v", err)
	}

	// Verify relation is removed
	out = reducer.Relations().GetOutgoingRelations("task-a", "blocks")
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

func TestRelationGraph_RemoveTaskRelations(t *testing.T) {
	graph := relations.NewRelationsGraph()

	// Create a task with multiple relations
	// task-a blocks task-b
	graph.AddRelation("task-a", "blocks", "task-b", "", "ev-1", "node1", 1)
	// task-c blocks task-a
	graph.AddRelation("task-c", "blocks", "task-a", "", "ev-2", "node1", 2)
	// task-a has subtask task-d
	graph.AddRelation("task-a", "subtask", "task-d", "", "ev-3", "node1", 3)
	// task-a related to task-e
	graph.AddRelation("task-a", "related", "task-e", "", "ev-4", "node1", 4)

	// Verify relations exist before delete
	outBlocks := graph.GetOutgoingRelations("task-a", "blocks")
	if len(outBlocks) != 1 {
		t.Fatalf("Expected 1 outgoing blocks relation, got %d", len(outBlocks))
	}

	inBlocks := graph.GetIncomingRelations("task-a", "blocks")
	if len(inBlocks) != 1 {
		t.Fatalf("Expected 1 incoming blocks relation, got %d", len(inBlocks))
	}

	outSubtask := graph.GetOutgoingRelations("task-a", "subtask")
	if len(outSubtask) != 1 {
		t.Fatalf("Expected 1 outgoing subtask relation, got %d", len(outSubtask))
	}

	outRelated := graph.GetOutgoingRelations("task-a", "related")
	if len(outRelated) != 1 {
		t.Fatalf("Expected 1 outgoing related relation, got %d", len(outRelated))
	}

	// Remove all relations for task-a
	graph.RemoveTaskRelations("task-a")

	// Verify all relations involving task-a are removed
	outBlocks = graph.GetOutgoingRelations("task-a", "blocks")
	if len(outBlocks) != 0 {
		t.Errorf("Expected 0 outgoing blocks relations after RemoveTaskRelations, got %d", len(outBlocks))
	}

	inBlocks = graph.GetIncomingRelations("task-a", "blocks")
	if len(inBlocks) != 0 {
		t.Errorf("Expected 0 incoming blocks relations after RemoveTaskRelations, got %d", len(inBlocks))
	}

	outSubtask = graph.GetOutgoingRelations("task-a", "subtask")
	if len(outSubtask) != 0 {
		t.Errorf("Expected 0 outgoing subtask relations after RemoveTaskRelations, got %d", len(outSubtask))
	}

	outRelated = graph.GetOutgoingRelations("task-a", "related")
	if len(outRelated) != 0 {
		t.Errorf("Expected 0 outgoing related relations after RemoveTaskRelations, got %d", len(outRelated))
	}

	// Verify relations from task-a to other tasks are also removed
	inBlocksB := graph.GetIncomingRelations("task-b", "blocks")
	if len(inBlocksB) != 0 {
		t.Errorf("Expected 0 incoming blocks relations for task-b after RemoveTaskRelations on task-a, got %d", len(inBlocksB))
	}

	inSubtaskD := graph.GetIncomingRelations("task-d", "subtask")
	if len(inSubtaskD) != 0 {
		t.Errorf("Expected 0 incoming subtask relations for task-d after RemoveTaskRelations on task-a, got %d", len(inSubtaskD))
	}
}

// TestRelationsIntegration tests a realistic workflow with multiple tasks and relations
func TestRelationsIntegration(t *testing.T) {
	reducer := reducer.NewReducer()

	// Create a project with dependencies:
	// - Task A (design) blocks Tasks B and C
	// - Task B (implement) has subtasks B1 and B2
	// - Task C (test)
	// - Task D (deploy) is blocked by C

	tasks := []struct {
		uuid  string
		id    string
		title string
	}{
		{"task-a", "proj-1", "Design API"},
		{"task-b", "proj-2", "Implement API"},
		{"task-b1", "proj-3", "Implement endpoint 1"},
		{"task-b2", "proj-4", "Implement endpoint 2"},
		{"task-c", "proj-5", "Write tests"},
		{"task-d", "proj-6", "Deploy to production"},
	}

	// Create all tasks
	ts := int64(1)
	projectUID := string(types.NewProjectUID())
	for _, task := range tasks {
		event := types.Event{
			ID:        "ev-" + task.uuid,
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     "alice",
			Role:      "human",
			Kind:      "task.created",
			Payload:   marshalPayload(types.TaskCreatedPayload{TaskUID: task.uuid, ProjectUID: projectUID, ProposedNumber: int64(ts), CreatedNode: string(types.NewNodeID()), Title: task.title, CreatedBy: "alice"}),
		}
		if err := reducer.Apply(event); err != nil {
			t.Fatalf("Failed to create task %s: %v", task.id, err)
		}
		ts++
	}

	// Set up relations
	relations := []struct {
		src     string
		relType string
		dst     string
	}{
		// A blocks B and C
		{"task-a", "blocks", "task-b"},
		{"task-a", "blocks", "task-c"},
		// B has subtasks B1 and B2
		{"task-b", "subtask", "task-b1"},
		{"task-b", "subtask", "task-b2"},
		// C blocks D
		{"task-c", "blocks", "task-d"},
	}

	for _, rel := range relations {
		event := types.Event{
			ID:        "ev-rel-" + rel.src + "-" + rel.dst,
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     "alice",
			Role:      "human",
			Kind:      "relation.add",
			Payload:   marshalPayload(types.RelationAddPayload{Src: rel.src, Type: rel.relType, Dst: rel.dst}),
		}
		if err := reducer.Apply(event); err != nil {
			t.Fatalf("Failed to add relation %s %s %s: %v", rel.src, rel.relType, rel.dst, err)
		}
		ts++
	}

	// Set some statuses
	statuses := []struct {
		task  string
		state string
	}{
		{"task-a", "in_progress"},
		{"task-b1", "done"},
		{"task-b2", "in_progress"},
	}

	for _, status := range statuses {
		event := types.Event{
			ID:        "ev-status-" + status.task,
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     "alice",
			Role:      "human",
			Kind:      "task.status.set",
			Payload:   marshalPayload(types.TaskStatusSetPayload{TaskUUID: status.task, Axis: "generic", State: status.state, Role: "human"}),
		}
		if err := reducer.Apply(event); err != nil {
			t.Fatalf("Failed to set status for %s: %v", status.task, err)
		}
		ts++
	}

	// Finalize relations with config
	cfg := &config.Config{
		Blocking: config.BlockingConfig{
			BlockingAxis: "generic",
			DoneStates:   []string{"done"},
		},
	}
	reducer.FinalizeRelations(cfg)

	// Verify task relations
	taskB, _ := reducer.GetTask("task-b")
	if taskB.Relations == nil {
		t.Fatal("Task B should have relations")
	}
	if len(taskB.Relations.Subtask.Children) != 2 {
		t.Errorf("Task B should have 2 children, got %d", len(taskB.Relations.Subtask.Children))
	}

	// Verify blocked status
	// Task B should be blocked by A (which is in_progress)
	if !taskB.Blocked {
		t.Error("Task B should be blocked")
	}
	if len(taskB.Blockers) != 1 {
		t.Errorf("Task B should have 1 blocker, got %d", len(taskB.Blockers))
	}

	// Task C should be blocked by A
	taskC, _ := reducer.GetTask("task-c")
	if !taskC.Blocked {
		t.Error("Task C should be blocked")
	}

	// Task D should be blocked by C (which is blocked by A - transitive)
	taskD, _ := reducer.GetTask("task-d")
	if !taskD.Blocked {
		t.Error("Task D should be blocked")
	}

	// Task A should not be blocked
	taskA, _ := reducer.GetTask("task-a")
	if taskA.Blocked {
		t.Error("Task A should not be blocked")
	}

	// Test transitive blockers for D
	transitiveBlockers := utils.GetTransitiveBlockers(reducer.Relations(), "task-d", reducer.Tasks(), "generic", []string{"done"}, 10)
	if len(transitiveBlockers) != 2 {
		t.Errorf("Task D should have 2 transitive blockers (C and A), got %d", len(transitiveBlockers))
	}

	// Now mark A as done
	event := types.Event{
		ID:        "ev-status-a-done",
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.status.set",
		Payload:   marshalPayload(types.TaskStatusSetPayload{TaskUUID: "task-a", Axis: "generic", State: "done", Role: "human"}),
	}
	if err := reducer.Apply(event); err != nil {
		t.Fatalf("Failed to set status for task-a: %v", err)
	}
	ts++

	// Recompute blocked status
	reducer.FinalizeRelations(cfg)

	// Task B and C should no longer be blocked
	taskB, _ = reducer.GetTask("task-b")
	if taskB.Blocked {
		t.Error("Task B should not be blocked after A is done")
	}

	taskC, _ = reducer.GetTask("task-c")
	if taskC.Blocked {
		t.Error("Task C should not be blocked after A is done")
	}

	// Task D should still be blocked (by C which is not done)
	taskD, _ = reducer.GetTask("task-d")
	if !taskD.Blocked {
		t.Error("Task D should still be blocked by C")
	}

	// Mark C as done
	event = types.Event{
		ID:        "ev-status-c-done",
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.status.set",
		Payload:   marshalPayload(types.TaskStatusSetPayload{TaskUUID: "task-c", Axis: "generic", State: "done", Role: "human"}),
	}
	if err := reducer.Apply(event); err != nil {
		t.Fatalf("Failed to set status for task-c: %v", err)
	}

	// Recompute blocked status
	reducer.FinalizeRelations(cfg)

	// Task D should now be unblocked
	taskD, _ = reducer.GetTask("task-d")
	if taskD.Blocked {
		t.Error("Task D should not be blocked after C is done")
	}
}

// TestCycleDetectionIntegration tests that cycles are properly detected
func TestCycleDetectionIntegration(t *testing.T) {
	graph := relations.NewRelationsGraph()

	// Create a complex graph with one cycle
	// A -> B -> C -> D
	//      ^         |
	//      |_________|
	graph.AddRelation("task-a", "blocks", "task-b", "", "ev-1", "node1", 1)
	graph.AddRelation("task-b", "blocks", "task-c", "", "ev-2", "node1", 2)
	graph.AddRelation("task-c", "blocks", "task-d", "", "ev-3", "node1", 3)
	graph.AddRelation("task-d", "blocks", "task-b", "", "ev-4", "node1", 4) // Creates cycle

	cycles := graph.DetectCycles("blocks")

	if len(cycles) == 0 {
		t.Fatal("Expected to detect at least one cycle")
	}

	// Should detect the B -> C -> D -> B cycle
	foundCycle := false
	for _, cycle := range cycles {
		// The cycle should contain at least B, C, and D
		if len(cycle) >= 3 {
			foundCycle = true
			break
		}
	}

	if !foundCycle {
		t.Error("Expected to find the B-C-D cycle")
	}
}

// TestRelationRemoval tests that removing relations works correctly
func TestRelationRemovalIntegration(t *testing.T) {
	reducer := reducer.NewReducer()

	// Create two tasks
	taskData := []struct {
		uuid  string
		id    string
		title string
	}{
		{"task-a", "test-1", "Task A"},
		{"task-b", "test-2", "Task B"},
	}

	for _, task := range taskData {
		event := types.Event{
			ID:        "ev-" + task.uuid,
			TS:        1,
			CreatedAt: time.Now(),
			Actor:     "alice",
			Role:      "human",
			Kind:      "task.created",
			Payload:   marshalPayload(types.TaskCreatedPayload{TaskUID: task.uuid, ProjectUID: string(types.NewProjectUID()), ProposedNumber: 1, CreatedNode: string(types.NewNodeID()), Title: task.title, CreatedBy: "alice"}),
		}
		reducer.Apply(event)
	}

	// Add a blocks relation
	addEvent := types.Event{
		ID:        "ev-add",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "relation.add",
		Payload:   marshalPayload(types.RelationAddPayload{Src: "task-a", Type: "blocks", Dst: "task-b"}),
	}
	reducer.Apply(addEvent)

	cfg := &config.Config{
		Blocking: config.BlockingConfig{
			BlockingAxis: "generic",
			DoneStates:   []string{"done"},
		},
	}
	reducer.FinalizeRelations(cfg)

	// Task B should be blocked
	taskB, _ := reducer.GetTask("task-b")
	if !taskB.Blocked {
		t.Error("Task B should be blocked initially")
	}

	// Remove the relation
	removeEvent := types.Event{
		ID:        "ev-remove",
		TS:        3,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "relation.remove",
		Payload:   marshalPayload(types.RelationRemovePayload{Src: "task-a", Type: "blocks", Dst: "task-b"}),
	}
	reducer.Apply(removeEvent)

	reducer.FinalizeRelations(cfg)

	// Task B should no longer be blocked
	taskB, _ = reducer.GetTask("task-b")
	if taskB.Blocked {
		t.Error("Task B should not be blocked after relation is removed")
	}
	if taskB.Relations != nil && len(taskB.Relations.Blocks.In) > 0 {
		t.Error("Task B should have no incoming blocks relations")
	}
}

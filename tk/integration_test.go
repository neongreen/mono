package main

import (
	"testing"
	"time"
)

// TestRelationsIntegration tests a realistic workflow with multiple tasks and relations
func TestRelationsIntegration(t *testing.T) {
	reducer := NewReducer()

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
	projectUID := string(NewProjectUID())
	for _, task := range tasks {
		event := Event{
			ID:        "ev-" + task.uuid,
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     "alice",
			Role:      "human",
			Kind:      "task.created",
			Payload:   marshalPayload(TaskCreatedPayload{TaskUID: task.uuid, ProjectUID: projectUID, ProposedNumber: int64(ts), CreatedNode: string(NewNodeID()), Title: task.title, CreatedBy: "alice"}),
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
		event := Event{
			ID:        "ev-rel-" + rel.src + "-" + rel.dst,
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     "alice",
			Role:      "human",
			Kind:      "relation.add",
			Payload:   marshalPayload(RelationAddPayload{Src: rel.src, Type: rel.relType, Dst: rel.dst}),
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
		event := Event{
			ID:        "ev-status-" + status.task,
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     "alice",
			Role:      "human",
			Kind:      "task.status.set",
			Payload:   marshalPayload(TaskStatusSetPayload{TaskUUID: status.task, Axis: "generic", State: status.state, Role: "human"}),
		}
		if err := reducer.Apply(event); err != nil {
			t.Fatalf("Failed to set status for %s: %v", status.task, err)
		}
		ts++
	}

	// Finalize relations with config
	config := &Config{
		Blocking: BlockingConfig{
			BlockingAxis: "generic",
			DoneStates:   []string{"done"},
		},
	}
	reducer.FinalizeRelations(config)

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
	transitiveBlockers := reducer.relations.GetTransitiveBlockers("task-d", reducer.tasks, "generic", []string{"done"}, 10)
	if len(transitiveBlockers) != 2 {
		t.Errorf("Task D should have 2 transitive blockers (C and A), got %d", len(transitiveBlockers))
	}

	// Now mark A as done
	event := Event{
		ID:        "ev-status-a-done",
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.status.set",
		Payload:   marshalPayload(TaskStatusSetPayload{TaskUUID: "task-a", Axis: "generic", State: "done", Role: "human"}),
	}
	if err := reducer.Apply(event); err != nil {
		t.Fatalf("Failed to set status for task-a: %v", err)
	}
	ts++

	// Recompute blocked status
	reducer.FinalizeRelations(config)

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
	event = Event{
		ID:        "ev-status-c-done",
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.status.set",
		Payload:   marshalPayload(TaskStatusSetPayload{TaskUUID: "task-c", Axis: "generic", State: "done", Role: "human"}),
	}
	if err := reducer.Apply(event); err != nil {
		t.Fatalf("Failed to set status for task-c: %v", err)
	}

	// Recompute blocked status
	reducer.FinalizeRelations(config)

	// Task D should now be unblocked
	taskD, _ = reducer.GetTask("task-d")
	if taskD.Blocked {
		t.Error("Task D should not be blocked after C is done")
	}
}

// TestCycleDetectionIntegration tests that cycles are properly detected
func TestCycleDetectionIntegration(t *testing.T) {
	graph := NewRelationsGraph()

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
	reducer := NewReducer()

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
		event := Event{
			ID:        "ev-" + task.uuid,
			TS:        1,
			CreatedAt: time.Now(),
			Actor:     "alice",
			Role:      "human",
			Kind:      "task.created",
			Payload:   marshalPayload(TaskCreatedPayload{TaskUID: task.uuid, ProjectUID: string(NewProjectUID()), ProposedNumber: 1, CreatedNode: string(NewNodeID()), Title: task.title, CreatedBy: "alice"}),
		}
		reducer.Apply(event)
	}

	// Add a blocks relation
	addEvent := Event{
		ID:        "ev-add",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "relation.add",
		Payload:   marshalPayload(RelationAddPayload{Src: "task-a", Type: "blocks", Dst: "task-b"}),
	}
	reducer.Apply(addEvent)

	config := &Config{
		Blocking: BlockingConfig{
			BlockingAxis: "generic",
			DoneStates:   []string{"done"},
		},
	}
	reducer.FinalizeRelations(config)

	// Task B should be blocked
	taskB, _ := reducer.GetTask("task-b")
	if !taskB.Blocked {
		t.Error("Task B should be blocked initially")
	}

	// Remove the relation
	removeEvent := Event{
		ID:        "ev-remove",
		TS:        3,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "relation.remove",
		Payload:   marshalPayload(RelationRemovePayload{Src: "task-a", Type: "blocks", Dst: "task-b"}),
	}
	reducer.Apply(removeEvent)

	reducer.FinalizeRelations(config)

	// Task B should no longer be blocked
	taskB, _ = reducer.GetTask("task-b")
	if taskB.Blocked {
		t.Error("Task B should not be blocked after relation is removed")
	}
	if taskB.Relations != nil && len(taskB.Relations.Blocks.In) > 0 {
		t.Error("Task B should have no incoming blocks relations")
	}
}

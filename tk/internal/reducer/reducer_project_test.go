package reducer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

func TestReducer_ProjectDelete_WithMultipleTasks(t *testing.T) {
	r := NewReducer()

	projectA := string(types.NewProjectUID())
	projectB := string(types.NewProjectUID())

	// Create task 1 in project B (created there)
	task1UID := string(types.NewTaskUID())
	create1Payload := types.TaskCreatedPayload{
		TaskUID:        task1UID,
		ProjectUID:     projectB,
		ProposedNumber: 1,
		CreatedNode:    string(types.NewNodeID()),
		Title:          "Task 1",
		CreatedBy:      "alice",
	}
	create1PayloadJSON, _ := json.Marshal(create1Payload)

	create1Event := types.Event{
		ID:        "event_01",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.created",
		Payload:   create1PayloadJSON,
	}

	r.Apply(create1Event)

	// Create task 2 in project A (will be relocated to B)
	task2UID := string(types.NewTaskUID())
	create2Payload := types.TaskCreatedPayload{
		TaskUID:        task2UID,
		ProjectUID:     projectA,
		ProposedNumber: 1,
		CreatedNode:    string(types.NewNodeID()),
		Title:          "Task 2",
		CreatedBy:      "alice",
	}
	create2PayloadJSON, _ := json.Marshal(create2Payload)

	create2Event := types.Event{
		ID:        "event_02",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.created",
		Payload:   create2PayloadJSON,
	}

	r.Apply(create2Event)

	// Create task 3 in project A (will stay in A)
	task3UID := string(types.NewTaskUID())
	create3Payload := types.TaskCreatedPayload{
		TaskUID:        task3UID,
		ProjectUID:     projectA,
		ProposedNumber: 2,
		CreatedNode:    string(types.NewNodeID()),
		Title:          "Task 3",
		CreatedBy:      "alice",
	}
	create3PayloadJSON, _ := json.Marshal(create3Payload)

	create3Event := types.Event{
		ID:        "event_03",
		TS:        3,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.created",
		Payload:   create3PayloadJSON,
	}

	r.Apply(create3Event)

	// Relocate task 2 from A to B
	relocatePayload := types.TaskRelocatePayload{
		TaskUID:        types.TaskUID(task2UID),
		FromProjectUID: types.ProjectUID(projectA),
		ToProjectUID:   types.ProjectUID(projectB),
		NumberPolicy: types.NumberPolicyPayload{
			Mode: "keep",
		},
	}
	relocatePayloadJSON, _ := json.Marshal(relocatePayload)

	relocateEvent := types.Event{
		ID:        "event_04",
		TS:        4,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.relocate",
		Payload:   relocatePayloadJSON,
	}

	r.Apply(relocateEvent)

	// Verify all tasks exist
	if _, ok := r.GetTask(task1UID); !ok {
		t.Fatal("Task 1 should exist")
	}
	if _, ok := r.GetTask(task2UID); !ok {
		t.Fatal("Task 2 should exist")
	}
	if _, ok := r.GetTask(task3UID); !ok {
		t.Fatal("Task 3 should exist")
	}

	// Delete project B - should delete task 1 (created in B) and task 2 (relocated to B)
	deleteBPayload := types.ProjectDeletePayload{ProjectUID: types.ProjectUID(projectB)}
	deleteBPayloadJSON, _ := json.Marshal(deleteBPayload)

	deleteBEvent := types.Event{
		ID:        "event_05",
		TS:        5,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "project.delete",
		Payload:   deleteBPayloadJSON,
	}

	err := r.Apply(deleteBEvent)
	if err != nil {
		t.Fatalf("Failed to delete project B: %v", err)
	}

	// Task 1 and 2 should be deleted (both belong to project B)
	if _, ok := r.GetTask(task1UID); ok {
		t.Error("Task 1 should be deleted (created in project B)")
	}
	if _, ok := r.GetTask(task2UID); ok {
		t.Error("Task 2 should be deleted (relocated to project B)")
	}

	// Task 3 should still exist (belongs to project A)
	if _, ok := r.GetTask(task3UID); !ok {
		t.Error("Task 3 should still exist (belongs to project A)")
	}
}

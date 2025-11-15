package reducer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

func TestReducer_TaskRelocate_UpdatesProjectTracking(t *testing.T) {
	r := NewReducer()

	// Create project A
	projectA := string(types.NewProjectUID())

	// Create project B
	projectB := string(types.NewProjectUID())

	// Create task in project A
	taskUID := string(types.NewTaskUID())
	createPayload := types.TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectA,
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

	err := r.Apply(createEvent)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Verify task exists and is in project A
	task, ok := r.GetTask(taskUID)
	if !ok {
		t.Fatal("Task should exist after creation")
	}
	if task.Title != "Test task" {
		t.Errorf("Expected task title 'Test task', got %s", task.Title)
	}

	// Relocate task from project A to project B
	relocatePayload := types.TaskRelocatePayload{
		TaskUID:        taskUID,
		FromProjectUID: projectA,
		ToProjectUID:   projectB,
		NumberPolicy: types.NumberPolicyPayload{
			Mode: "keep",
		},
	}
	relocatePayloadJSON, _ := json.Marshal(relocatePayload)

	relocateEvent := types.Event{
		ID:        "event_02",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.relocate",
		Payload:   relocatePayloadJSON,
	}

	err = r.Apply(relocateEvent)
	if err != nil {
		t.Fatalf("Failed to relocate task: %v", err)
	}

	// Verify task still exists
	_, ok = r.GetTask(taskUID)
	if !ok {
		t.Fatal("Task should exist after relocation")
	}

	// Delete project B (the task's new home)
	deleteProjectBPayload := types.ProjectDeletePayload{
		ProjectUID: projectB,
	}
	deleteProjectBPayloadJSON, _ := json.Marshal(deleteProjectBPayload)

	deleteProjectBEvent := types.Event{
		ID:        "event_03",
		TS:        3,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "project.delete",
		Payload:   deleteProjectBPayloadJSON,
	}

	err = r.Apply(deleteProjectBEvent)
	if err != nil {
		t.Fatalf("Failed to delete project B: %v", err)
	}

	// Task should be deleted because it belongs to project B now
	_, ok = r.GetTask(taskUID)
	if ok {
		t.Error("Task should be deleted when its current project (B) is deleted")
	}
}

func TestReducer_TaskRelocate_DeletesFromCorrectProject(t *testing.T) {
	r := NewReducer()

	// Create project A
	projectA := string(types.NewProjectUID())

	// Create project B
	projectB := string(types.NewProjectUID())

	// Create task in project A
	taskUID := string(types.NewTaskUID())
	createPayload := types.TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectA,
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

	err := r.Apply(createEvent)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Relocate task from project A to project B
	relocatePayload := types.TaskRelocatePayload{
		TaskUID:        taskUID,
		FromProjectUID: projectA,
		ToProjectUID:   projectB,
		NumberPolicy: types.NumberPolicyPayload{
			Mode: "keep",
		},
	}
	relocatePayloadJSON, _ := json.Marshal(relocatePayload)

	relocateEvent := types.Event{
		ID:        "event_02",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.relocate",
		Payload:   relocatePayloadJSON,
	}

	err = r.Apply(relocateEvent)
	if err != nil {
		t.Fatalf("Failed to relocate task: %v", err)
	}

	// Delete project A (the task's original home)
	deleteProjectAPayload := types.ProjectDeletePayload{
		ProjectUID: projectA,
	}
	deleteProjectAPayloadJSON, _ := json.Marshal(deleteProjectAPayload)

	deleteProjectAEvent := types.Event{
		ID:        "event_03",
		TS:        3,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "project.delete",
		Payload:   deleteProjectAPayloadJSON,
	}

	err = r.Apply(deleteProjectAEvent)
	if err != nil {
		t.Fatalf("Failed to delete project A: %v", err)
	}

	// Task should still exist because it was moved to project B
	task, ok := r.GetTask(taskUID)
	if !ok {
		t.Error("Task should still exist after deleting its original project (A), because it was relocated to project B")
	}
	if task.Title != "Test task" {
		t.Errorf("Expected task title 'Test task', got %s", task.Title)
	}
}

func TestReducer_TaskRelocate_MultipleRelocations(t *testing.T) {
	r := NewReducer()

	// Create three projects
	projectA := string(types.NewProjectUID())
	projectB := string(types.NewProjectUID())
	projectC := string(types.NewProjectUID())

	// Create task in project A
	taskUID := string(types.NewTaskUID())
	createPayload := types.TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectA,
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

	err := r.Apply(createEvent)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Relocate task from A to B
	relocate1Payload := types.TaskRelocatePayload{
		TaskUID:        taskUID,
		FromProjectUID: projectA,
		ToProjectUID:   projectB,
		NumberPolicy: types.NumberPolicyPayload{
			Mode: "keep",
		},
	}
	relocate1PayloadJSON, _ := json.Marshal(relocate1Payload)

	relocate1Event := types.Event{
		ID:        "event_02",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.relocate",
		Payload:   relocate1PayloadJSON,
	}

	err = r.Apply(relocate1Event)
	if err != nil {
		t.Fatalf("Failed to relocate task A->B: %v", err)
	}

	// Relocate task from B to C
	relocate2Payload := types.TaskRelocatePayload{
		TaskUID:        taskUID,
		FromProjectUID: projectB,
		ToProjectUID:   projectC,
		NumberPolicy: types.NumberPolicyPayload{
			Mode: "keep",
		},
	}
	relocate2PayloadJSON, _ := json.Marshal(relocate2Payload)

	relocate2Event := types.Event{
		ID:        "event_03",
		TS:        3,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.relocate",
		Payload:   relocate2PayloadJSON,
	}

	err = r.Apply(relocate2Event)
	if err != nil {
		t.Fatalf("Failed to relocate task B->C: %v", err)
	}

	// Delete project A (original home) - task should still exist
	deleteAPayload := types.ProjectDeletePayload{ProjectUID: projectA}
	deleteAPayloadJSON, _ := json.Marshal(deleteAPayload)
	deleteAEvent := types.Event{
		ID:        "event_04",
		TS:        4,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "project.delete",
		Payload:   deleteAPayloadJSON,
	}

	err = r.Apply(deleteAEvent)
	if err != nil {
		t.Fatalf("Failed to delete project A: %v", err)
	}

	if _, ok := r.GetTask(taskUID); !ok {
		t.Error("Task should still exist after deleting original project A (task is in C)")
	}

	// Delete project B (intermediate) - task should still exist
	deleteBPayload := types.ProjectDeletePayload{ProjectUID: projectB}
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

	err = r.Apply(deleteBEvent)
	if err != nil {
		t.Fatalf("Failed to delete project B: %v", err)
	}

	if _, ok := r.GetTask(taskUID); !ok {
		t.Error("Task should still exist after deleting intermediate project B (task is in C)")
	}

	// Delete project C (current home) - task should be deleted
	deleteCPayload := types.ProjectDeletePayload{ProjectUID: projectC}
	deleteCPayloadJSON, _ := json.Marshal(deleteCPayload)
	deleteCEvent := types.Event{
		ID:        "event_06",
		TS:        6,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "project.delete",
		Payload:   deleteCPayloadJSON,
	}

	err = r.Apply(deleteCEvent)
	if err != nil {
		t.Fatalf("Failed to delete project C: %v", err)
	}

	if _, ok := r.GetTask(taskUID); ok {
		t.Error("Task should be deleted after deleting current project C")
	}
}

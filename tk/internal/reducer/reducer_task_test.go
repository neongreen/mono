package reducer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

func TestReducer_TaskCreated(t *testing.T) {
	r := NewReducer()

	taskUID := string(types.NewTaskUID())
	projectUID := string(types.NewProjectUID())
	payload := types.TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: 1,
		CreatedNode:    string(types.NewNodeID()),
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

	err := r.Apply(event)
	if err != nil {
		t.Fatalf("Failed to apply task.created event: %v", err)
	}

	task, ok := r.GetTask(taskUID)
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

func TestReducer_TaskDelete(t *testing.T) {
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

	err := r.Apply(createEvent)
	if err != nil {
		t.Fatalf("Failed to apply task.created event: %v", err)
	}

	// Verify task exists
	_, ok := r.GetTask(taskUID)
	if !ok {
		t.Fatal("Task should exist before delete")
	}

	// Delete task
	deletePayload := types.TaskDeletePayload{
		TaskUUID: taskUID,
	}
	deletePayloadJSON, _ := json.Marshal(deletePayload)

	deleteEvent := types.Event{
		ID:        "event_02",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.delete",
		Payload:   deletePayloadJSON,
	}

	err = r.Apply(deleteEvent)
	if err != nil {
		t.Fatalf("Failed to apply task.delete event: %v", err)
	}

	// Verify task is soft deleted (still in map but marked deleted)
	task, ok := r.GetTaskIncludingDeleted(taskUID)
	if !ok {
		t.Error("Task should still exist in map after soft delete")
	}
	if !task.Deleted {
		t.Error("Task should be marked as deleted")
	}
	if task.DeletedAt.IsZero() {
		t.Error("Task should have DeletedAt timestamp")
	}

	// Verify task is not visible in GetAllTasks
	allTasks := r.GetAllTasks()
	for _, visibleTask := range allTasks {
		if visibleTask.TaskUUID == taskUID {
			t.Error("Deleted task should not appear in GetAllTasks()")
		}
	}
}

func TestReducer_TaskDelete_RemovesTaskIDMappings(t *testing.T) {
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

	// Verify task ID mapping exists
	if mappedUID, exists := r.taskByID[taskUID]; !exists || mappedUID != taskUID {
		t.Fatal("Task ID mapping should exist")
	}

	// Delete task
	deletePayload := types.TaskDeletePayload{
		TaskUUID: taskUID,
	}
	deletePayloadJSON, _ := json.Marshal(deletePayload)

	deleteEvent := types.Event{
		ID:        "event_02",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.delete",
		Payload:   deletePayloadJSON,
	}

	r.Apply(deleteEvent)

	// Verify task ID mapping still exists (soft delete doesn't remove mappings)
	if _, exists := r.taskByID[taskUID]; !exists {
		t.Error("Task ID mapping should still exist after soft delete")
	}

	// Verify task is marked deleted
	task, ok := r.GetTaskIncludingDeleted(taskUID)
	if !ok {
		t.Error("Task should still be retrievable after soft delete")
	}
	if !task.Deleted {
		t.Error("Task should be marked as deleted")
	}
}

func TestReducer_TaskDelete_Idempotency(t *testing.T) {
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

	// Delete task
	deletePayload := types.TaskDeletePayload{
		TaskUUID: taskUID,
	}
	deletePayloadJSON, _ := json.Marshal(deletePayload)

	deleteEvent := types.Event{
		ID:        "event_02",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.delete",
		Payload:   deletePayloadJSON,
	}

	// Delete first time
	err := r.Apply(deleteEvent)
	if err != nil {
		t.Fatalf("First delete failed: %v", err)
	}

	// Delete second time (should be idempotent, no error)
	err = r.Apply(deleteEvent)
	if err != nil {
		t.Errorf("Second delete should be idempotent but got error: %v", err)
	}

	// Verify task is still soft deleted
	task, ok := r.GetTask(taskUID)
	if !ok {
		t.Error("Task should still exist after soft delete")
	}
	if !task.Deleted {
		t.Error("Task should remain marked as deleted after idempotent delete")
	}
}

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

	// Create projects first
	projectNames := []string{"project-a", "project-b"}
	for i, projUID := range []string{projectA, projectB} {
		projPayload := types.ProjectCreatedPayload{
			ProjectUID:  types.ProjectUID(projUID),
			Type:        types.ProjectTypeLocal,
			Name:        projectNames[i],
			Description: "Test",
			CreatedBy:   "alice",
		}
		projJSON, _ := json.Marshal(projPayload)
		projEvent := types.Event{
			ID:        types.NewEventID().String(),
			TS:        int64(i),
			CreatedAt: time.Now(),
			Actor:     "alice",
			Role:      "human",
			Kind:      string(types.EventKindProjectCreated),
			Payload:   projJSON,
		}
		if err := r.Apply(projEvent); err != nil {
			t.Fatalf("Failed to create project: %v", err)
		}
	}

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

	if err := r.Apply(relocateEvent); err != nil {
		t.Fatalf("Failed to relocate task: %v", err)
	}

	// Verify task 2 was relocated to project B
	if projUID, ok := r.taskProjects[task2UID]; !ok {
		t.Fatal("Task 2 should have project mapping")
	} else if projUID != projectB {
		t.Errorf("Task 2 should belong to project B after relocate, got %s", projUID)
	}

	// Verify all tasks exist
	if _, ok := r.GetTask(task1UID); !ok {
		t.Fatal("Task 1 should exist")
	}
	task2Pre, ok := r.GetTask(task2UID)
	if !ok {
		t.Fatal("Task 2 should exist")
	}
	if task2Pre.ProjectUUID != projectB {
		t.Errorf("Task 2 ProjectUUID should be %s after relocate, got %s", projectB, task2Pre.ProjectUUID)
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

	// Task 1 and 2 should still exist but be hidden (project B deleted)
	task1, ok1 := r.GetTask(task1UID)
	if !ok1 {
		t.Error("Task 1 should still exist in map (soft delete)")
	}
	// Task is hidden because its project is deleted
	if !r.isProjectDeleted(projectB) {
		t.Error("Project B should be marked deleted")
	}
	if r.isTaskVisible(task1) {
		t.Error("Task 1 should not be visible (project deleted)")
	}

	task2, ok2 := r.GetTask(task2UID)
	if !ok2 {
		t.Error("Task 2 should still exist in map (soft delete)")
	}
	if r.isTaskVisible(task2) {
		t.Error("Task 2 should not be visible (relocated to deleted project)")
	}

	// Task 3 should still exist and be visible (belongs to project A which is not deleted)
	task3, ok3 := r.GetTask(task3UID)
	if !ok3 {
		t.Error("Task 3 should still exist (belongs to project A)")
	}
	if !r.isTaskVisible(task3) {
		t.Error("Task 3 should be visible (project A not deleted)")
	}

	// Verify GetAllTasks doesn't include tasks 1 and 2
	allTasks := r.GetAllTasks()
	foundTask1, foundTask2, foundTask3 := false, false, false
	for _, task := range allTasks {
		if task.TaskUUID == task1UID {
			foundTask1 = true
		}
		if task.TaskUUID == task2UID {
			foundTask2 = true
		}
		if task.TaskUUID == task3UID {
			foundTask3 = true
		}
	}
	if foundTask1 {
		t.Error("Task 1 should not be in GetAllTasks (project deleted)")
	}
	if foundTask2 {
		t.Error("Task 2 should not be in GetAllTasks (project deleted)")
	}
	if !foundTask3 {
		t.Error("Task 3 should be in GetAllTasks (project not deleted)")
	}
}

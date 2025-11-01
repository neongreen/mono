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

func TestReducer_NoteAdd(t *testing.T) {
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

	// Add note
	notePayload := types.TaskNoteAddPayload{
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

	err := r.Apply(noteEvent)
	if err != nil {
		t.Fatalf("Failed to apply task.note.add event: %v", err)
	}

	task, _ := r.GetTask(taskUID)

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

	// Verify task was removed
	_, ok = r.GetTask(taskUID)
	if ok {
		t.Error("Task should not exist after delete")
	}

	// Verify task is not in the tasks map
	if _, exists := r.tasks[taskUID]; exists {
		t.Error("Task should be removed from tasks map")
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

	// Verify task ID mapping was removed
	if _, exists := r.taskByID[taskUID]; exists {
		t.Error("Task ID mapping should be removed after delete")
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

	// Verify task is still deleted
	_, ok := r.GetTask(taskUID)
	if ok {
		t.Error("Task should remain deleted after idempotent delete")
	}
}

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
	task, ok = r.GetTask(taskUID)
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

	// BUG: Task should be deleted because it belongs to project B now,
	// but applyTaskRelocate doesn't update taskProjects, so the task
	// is still mapped to project A and won't be deleted
	_, ok = r.GetTask(taskUID)
	if ok {
		t.Error("Task should be deleted when its current project (B) is deleted, but it still exists because applyTaskRelocate doesn't update taskProjects")
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

	// BUG: Task should still exist because it was moved to project B,
	// but applyTaskRelocate doesn't update taskProjects, so the task
	// is still mapped to project A and will be incorrectly deleted
	task, ok := r.GetTask(taskUID)
	if !ok {
		t.Error("Task should still exist after deleting its original project (A), because it was relocated to project B, but applyTaskRelocate doesn't update taskProjects")
	} else if task.Title != "Test task" {
		t.Errorf("Expected task title 'Test task', got %s", task.Title)
	}
}

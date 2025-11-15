package reducer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

// ReducerTest provides a fluent API for writing reducer tests
type ReducerTest struct {
	t       *testing.T
	reducer *Reducer
	ts      int64
	tasks   map[string]*TaskHandle // taskUID -> handle
}

// NewReducerTest creates a new test context with a fresh reducer
func NewReducerTest(t *testing.T) *ReducerTest {
	t.Helper()
	return &ReducerTest{
		t:       t,
		reducer: NewReducer(),
		ts:      0,
		tasks:   make(map[string]*TaskHandle),
	}
}

// Reducer returns the underlying reducer for advanced usage
func (rt *ReducerTest) Reducer() *Reducer {
	return rt.reducer
}

// CreateProject creates a new project and returns its UID
func (rt *ReducerTest) CreateProject(name string) string {
	rt.t.Helper()
	return string(types.NewProjectUID())
}

// DeleteProject deletes a project and all its tasks
func (rt *ReducerTest) DeleteProject(projectUID string) {
	rt.t.Helper()
	rt.ts++

	payload := types.ProjectDeletePayload{
		ProjectUID: projectUID,
	}
	payloadJSON, _ := json.Marshal(payload)

	event := types.Event{
		ID:        rt.eventID(),
		TS:        rt.ts,
		CreatedAt: time.Now(),
		Actor:     "test-actor",
		Role:      "human",
		Kind:      "project.delete",
		Payload:   payloadJSON,
	}

	if err := rt.reducer.Apply(event); err != nil {
		rt.t.Fatalf("Failed to delete project %s: %v", projectUID, err)
	}
}

// CreateTask creates a new task with the given actor, title, and project
func (rt *ReducerTest) CreateTask(actor, title, projectUID string) *TaskHandle {
	rt.t.Helper()

	taskUID := string(types.NewTaskUID())
	rt.ts++

	payload := types.TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: 1,
		CreatedNode:    string(types.NewNodeID()),
		Title:          title,
		CreatedBy:      actor,
	}
	payloadJSON, _ := json.Marshal(payload)

	event := types.Event{
		ID:        rt.eventID(),
		TS:        rt.ts,
		CreatedAt: time.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      "task.created",
		Payload:   payloadJSON,
	}

	if err := rt.reducer.Apply(event); err != nil {
		rt.t.Fatalf("Failed to create task: %v", err)
	}

	handle := &TaskHandle{
		UID:        taskUID,
		ProjectUID: projectUID,
		Title:      title,
		test:       rt,
	}
	rt.tasks[taskUID] = handle

	return handle
}

// Task retrieves a task handle by UID (useful if you lost the handle)
func (rt *ReducerTest) Task(taskUID string) *TaskHandle {
	rt.t.Helper()

	if handle, ok := rt.tasks[taskUID]; ok {
		return handle
	}

	// Try to find it in the reducer
	if task, ok := rt.reducer.GetTask(taskUID); ok {
		handle := &TaskHandle{
			UID:        taskUID,
			ProjectUID: rt.reducer.taskProjects[taskUID],
			Title:      task.Title,
			test:       rt,
		}
		rt.tasks[taskUID] = handle
		return handle
	}

	rt.t.Fatalf("Task %s not found", taskUID)
	return nil
}

// AssertTaskCount verifies the number of tasks in the reducer
func (rt *ReducerTest) AssertTaskCount(expected int) {
	rt.t.Helper()

	actual := len(rt.reducer.tasks)
	if actual != expected {
		rt.t.Errorf("Expected %d tasks, got %d", expected, actual)
	}
}

// AssertNoTasks verifies there are no tasks
func (rt *ReducerTest) AssertNoTasks() {
	rt.t.Helper()
	rt.AssertTaskCount(0)
}

// eventID generates a unique event ID
func (rt *ReducerTest) eventID() string {
	return string(types.NewEventID())
}

// TaskHandle represents a task in the test with fluent operations
type TaskHandle struct {
	UID        string
	ProjectUID string
	Title      string
	test       *ReducerTest
}

// Relocate moves the task to a different project
func (th *TaskHandle) Relocate(toProjectUID string) *TaskHandle {
	th.test.t.Helper()
	th.test.ts++

	payload := types.TaskRelocatePayload{
		TaskUID:        th.UID,
		FromProjectUID: th.ProjectUID,
		ToProjectUID:   toProjectUID,
		NumberPolicy: types.NumberPolicyPayload{
			Mode: "keep",
		},
	}
	payloadJSON, _ := json.Marshal(payload)

	event := types.Event{
		ID:        th.test.eventID(),
		TS:        th.test.ts,
		CreatedAt: time.Now(),
		Actor:     "test-actor",
		Role:      "human",
		Kind:      "task.relocate",
		Payload:   payloadJSON,
	}

	if err := th.test.reducer.Apply(event); err != nil {
		th.test.t.Fatalf("Failed to relocate task %s: %v", th.UID, err)
	}

	th.ProjectUID = toProjectUID
	return th
}

// SetStatus sets the task status (defaults to human role)
func (th *TaskHandle) SetStatus(status string) *TaskHandle {
	th.test.t.Helper()
	return th.SetStatusAs(status, "human", "test-actor")
}

// SetStatusAs sets the task status with specific role and actor
func (th *TaskHandle) SetStatusAs(status, role, actor string) *TaskHandle {
	th.test.t.Helper()
	th.test.ts++

	payload := types.TaskStatusSetPayload{
		TaskUUID: th.UID,
		Axis:     "generic",
		State:    status,
		Role:     role,
	}
	payloadJSON, _ := json.Marshal(payload)

	event := types.Event{
		ID:        th.test.eventID(),
		TS:        th.test.ts,
		CreatedAt: time.Now(),
		Actor:     actor,
		Role:      role,
		Kind:      "task.status.set",
		Payload:   payloadJSON,
	}

	if err := th.test.reducer.Apply(event); err != nil {
		th.test.t.Fatalf("Failed to set status for task %s: %v", th.UID, err)
	}

	return th
}

// SetTitle updates the task title
func (th *TaskHandle) SetTitle(newTitle string) *TaskHandle {
	th.test.t.Helper()
	th.test.ts++

	payload := types.TaskTitleSetPayload{
		TaskUID: th.UID,
		Title:   newTitle,
	}
	payloadJSON, _ := json.Marshal(payload)

	event := types.Event{
		ID:        th.test.eventID(),
		TS:        th.test.ts,
		CreatedAt: time.Now(),
		Actor:     "test-actor",
		Role:      "human",
		Kind:      "task.title.set",
		Payload:   payloadJSON,
	}

	if err := th.test.reducer.Apply(event); err != nil {
		th.test.t.Fatalf("Failed to set title for task %s: %v", th.UID, err)
	}

	th.Title = newTitle
	return th
}

// Delete marks the task as deleted
func (th *TaskHandle) Delete() *TaskHandle {
	th.test.t.Helper()
	th.test.ts++

	payload := types.TaskDeletePayload{
		TaskUUID: th.UID,
	}
	payloadJSON, _ := json.Marshal(payload)

	event := types.Event{
		ID:        th.test.eventID(),
		TS:        th.test.ts,
		CreatedAt: time.Now(),
		Actor:     "test-actor",
		Role:      "human",
		Kind:      "task.delete",
		Payload:   payloadJSON,
	}

	if err := th.test.reducer.Apply(event); err != nil {
		th.test.t.Fatalf("Failed to delete task %s: %v", th.UID, err)
	}

	return th
}

// AssertExists verifies the task exists in the reducer
func (th *TaskHandle) AssertExists() *TaskHandle {
	th.test.t.Helper()

	_, ok := th.test.reducer.GetTask(th.UID)
	if !ok {
		th.test.t.Errorf("Task %s (%q) should exist but was not found", th.UID, th.Title)
	}

	return th
}

// AssertDeleted verifies the task does not exist in the reducer
func (th *TaskHandle) AssertDeleted() *TaskHandle {
	th.test.t.Helper()

	_, ok := th.test.reducer.GetTask(th.UID)
	if ok {
		th.test.t.Errorf("Task %s (%q) should be deleted but still exists", th.UID, th.Title)
	}

	return th
}

// AssertStatus verifies the task has the expected status
func (th *TaskHandle) AssertStatus(expectedStatus string) *TaskHandle {
	th.test.t.Helper()

	task, ok := th.test.reducer.GetTask(th.UID)
	if !ok {
		th.test.t.Fatalf("Task %s not found", th.UID)
		return th
	}

	axis, ok := task.Axes["generic"]
	if !ok {
		th.test.t.Fatalf("Task %s has no generic axis", th.UID)
		return th
	}

	if axis.Effective != expectedStatus {
		th.test.t.Errorf("Task %s (%q): expected status %q, got %q", th.UID, th.Title, expectedStatus, axis.Effective)
	}

	return th
}

// AssertTitle verifies the task has the expected title
func (th *TaskHandle) AssertTitle(expectedTitle string) *TaskHandle {
	th.test.t.Helper()

	task, ok := th.test.reducer.GetTask(th.UID)
	if !ok {
		th.test.t.Fatalf("Task %s not found", th.UID)
		return th
	}

	if task.Title != expectedTitle {
		th.test.t.Errorf("Task %s: expected title %q, got %q", th.UID, expectedTitle, task.Title)
	}

	return th
}

// AssertInProject verifies the task is in the expected project
func (th *TaskHandle) AssertInProject(expectedProjectUID string) *TaskHandle {
	th.test.t.Helper()

	actualProjectUID := th.test.reducer.taskProjects[th.UID]
	if actualProjectUID != expectedProjectUID {
		th.test.t.Errorf("Task %s (%q): expected project %s, got %s", th.UID, th.Title, expectedProjectUID, actualProjectUID)
	}

	return th
}

// AssertClaimCount verifies the number of claims for a given axis
func (th *TaskHandle) AssertClaimCount(axis string, expectedCount int) *TaskHandle {
	th.test.t.Helper()

	task, ok := th.test.reducer.GetTask(th.UID)
	if !ok {
		th.test.t.Fatalf("Task %s not found", th.UID)
		return th
	}

	axisData, ok := task.Axes[axis]
	if !ok {
		th.test.t.Fatalf("Task %s has no %s axis", th.UID, axis)
		return th
	}

	actualCount := len(axisData.Claims)
	if actualCount != expectedCount {
		th.test.t.Errorf("Task %s (%q) axis %s: expected %d claims, got %d", th.UID, th.Title, axis, expectedCount, actualCount)
	}

	return th
}

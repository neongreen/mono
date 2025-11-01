package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

// TestTypeValidation tests type validation
func TestTypeValidation(t *testing.T) {
	// Test ProjectUID
	t.Run("ProjectUID validation", func(t *testing.T) {
		validUID := types.NewProjectUID()
		if err := validUID.Validate(); err != nil {
			t.Errorf("valid ProjectUID failed validation: %v", err)
		}

		invalidUID := types.ProjectUID("invalid")
		if err := invalidUID.Validate(); err == nil {
			t.Error("invalid ProjectUID passed validation")
		}
	})

	// Test TaskUID
	t.Run("TaskUID validation", func(t *testing.T) {
		validUID := types.NewTaskUID()
		if err := validUID.Validate(); err != nil {
			t.Errorf("valid TaskUID failed validation: %v", err)
		}

		invalidUID := types.TaskUID("invalid")
		if err := invalidUID.Validate(); err == nil {
			t.Error("invalid TaskUID passed validation")
		}
	})

	// Test Alias
	t.Run("Alias validation", func(t *testing.T) {
		validAlias := types.Alias("tk")
		if err := validAlias.Validate(); err != nil {
			t.Errorf("valid alias failed validation: %v", err)
		}

		tooShort := types.Alias("x")
		if err := tooShort.Validate(); err == nil {
			t.Error("too short alias passed validation")
		}

		tooLong := types.Alias("this-is-way-too-long-for-an-alias")
		if err := tooLong.Validate(); err == nil {
			t.Error("too long alias passed validation")
		}
	})

	// Test TaskNumber
	t.Run("TaskNumber validation", func(t *testing.T) {
		validNum := types.TaskNumber(1)
		if err := validNum.Validate(); err != nil {
			t.Errorf("valid TaskNumber failed validation: %v", err)
		}

		invalidNum := types.TaskNumber(0)
		if err := invalidNum.Validate(); err == nil {
			t.Error("zero TaskNumber passed validation")
		}

		negativeNum := types.TaskNumber(-1)
		if err := negativeNum.Validate(); err == nil {
			t.Error("negative TaskNumber passed validation")
		}
	})
}

// TestEvents tests event handling in reducer
func TestEvents(t *testing.T) {
	r := NewReducer()

	// Test project.created event
	t.Run("project.created", func(t *testing.T) {
		payload := ProjectCreatedPayload{
			ProjectUID:  string(types.NewProjectUID()),
			Type:        "local",
			Name:        "Test Project",
			Description: "A test project",
			CreatedBy:   "testuser",
		}
		payloadJSON, _ := json.Marshal(payload)

		event := types.Event{
			ID:        string(types.NewEventID()),
			TS:        1,
			CreatedAt: time.Now(),
			Actor:     "testuser",
			Role:      "human",
			Kind:      string(types.EventKindProjectCreated),
			Payload:   payloadJSON,
		}

		if err := r.Apply(event); err != nil {
			t.Errorf("failed to apply project.created event: %v", err)
		}
	})

	// Test task.created event
	t.Run("task.created", func(t *testing.T) {
		taskUID := types.NewTaskUID()
		projectUID := types.NewProjectUID()

		payload := TaskCreatedPayload{
			TaskUID:        string(taskUID),
			ProjectUID:     string(projectUID),
			ProposedNumber: 1,
			CreatedNode:    string(types.NewNodeID()),
			Title:          "Test Task",
			CreatedBy:      "testuser",
		}
		payloadJSON, _ := json.Marshal(payload)

		event := types.Event{
			ID:        string(types.NewEventID()),
			TS:        2,
			CreatedAt: time.Now(),
			Actor:     "testuser",
			Role:      "human",
			Kind:      string(types.EventKindTaskCreated),
			Payload:   payloadJSON,
		}

		if err := r.Apply(event); err != nil {
			t.Errorf("failed to apply task.created event: %v", err)
		}

		// Verify task was created
		if _, exists := r.tasks[string(taskUID)]; !exists {
			t.Error("task was not created in reducer")
		}
	})

	// Test task.title.set event
	t.Run("task.title.set", func(t *testing.T) {
		taskUID := types.NewTaskUID()

		// First create the task
		createPayload := TaskCreatedPayload{
			TaskUID:        string(taskUID),
			ProjectUID:     string(types.NewProjectUID()),
			ProposedNumber: 1,
			CreatedNode:    string(types.NewNodeID()),
			Title:          "Original Title",
			CreatedBy:      "testuser",
		}
		createPayloadJSON, _ := json.Marshal(createPayload)

		createEvent := types.Event{
			ID:        string(types.NewEventID()),
			TS:        3,
			CreatedAt: time.Now(),
			Actor:     "testuser",
			Role:      "human",
			Kind:      string(types.EventKindTaskCreated),
			Payload:   createPayloadJSON,
		}

		if err := r.Apply(createEvent); err != nil {
			t.Fatalf("failed to create task: %v", err)
		}

		// Now change the title
		titlePayload := TaskTitleSetPayload{
			TaskUID: string(taskUID),
			Title:   "Updated Title",
		}
		titlePayloadJSON, _ := json.Marshal(titlePayload)

		titleEvent := types.Event{
			ID:        string(types.NewEventID()),
			TS:        4,
			CreatedAt: time.Now(),
			Actor:     "testuser",
			Role:      "human",
			Kind:      string(types.EventKindTaskTitleSet),
			Payload:   titlePayloadJSON,
		}

		if err := r.Apply(titleEvent); err != nil {
			t.Errorf("failed to apply task.title.set event: %v", err)
		}

		// Verify title was updated
		task := r.tasks[string(taskUID)]
		if task.Title != "Updated Title" {
			t.Errorf("expected title 'Updated Title', got '%s'", task.Title)
		}
	})
}

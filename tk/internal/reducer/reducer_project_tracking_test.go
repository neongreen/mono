package reducer

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

func TestReducer_ProjectTracking_Creation(t *testing.T) {
	r := NewReducer()

	projectUID := types.NewProjectUID()

	// Create a project
	payload := types.ProjectCreatedPayload{
		ProjectUID:  projectUID,
		Type:        "local",
		Name:        "test-project",
		Description: "A test project",
		CreatedBy:   "alice",
	}
	payloadJSON, _ := json.Marshal(payload)

	event := types.Event{
		ID:        "event_01",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "project.created",
		Payload:   payloadJSON,
	}

	if err := r.Apply(event); err != nil {
		t.Fatalf("Failed to apply project.created event: %v", err)
	}

	// Verify project exists
	project, ok := r.GetProject(string(projectUID))
	if !ok {
		t.Fatal("Project should exist after creation")
	}

	if project.ProjectUID != string(projectUID) {
		t.Errorf("ProjectUID = %s, want %s", project.ProjectUID, projectUID)
	}
	if project.Name != "test-project" {
		t.Errorf("Name = %s, want test-project", project.Name)
	}
	if project.Type != "local" {
		t.Errorf("Type = %s, want local", project.Type)
	}
	if project.IsSynthetic {
		t.Error("Project should not be synthetic")
	}
}

func TestReducer_ProjectTracking_GetAllProjects(t *testing.T) {
	r := NewReducer()

	// Create multiple projects
	projectUIDs := []types.ProjectUID{
		types.NewProjectUID(),
		types.NewProjectUID(),
		types.NewProjectUID(),
	}

	for i, uid := range projectUIDs {
		payload := types.ProjectCreatedPayload{
			ProjectUID:  uid,
			Type:        "local",
			Name:        fmt.Sprintf("project-%d", i),
			Description: "Test project",
			CreatedBy:   "alice",
		}
		payloadJSON, _ := json.Marshal(payload)

		event := types.Event{
			ID:        fmt.Sprintf("event_%d", i),
			TS:        int64(i + 1),
			CreatedAt: time.Now(),
			Actor:     "alice",
			Role:      "human",
			Kind:      "project.created",
			Payload:   payloadJSON,
		}

		if err := r.Apply(event); err != nil {
			t.Fatalf("Failed to apply project.created event: %v", err)
		}
	}

	// Get all projects
	projects := r.GetAllProjects()
	if len(projects) != 3 {
		t.Errorf("GetAllProjects() returned %d projects, want 3", len(projects))
	}
}

func TestReducer_ProjectTracking_SyntheticProject(t *testing.T) {
	r := NewReducer()

	projectUID := "corrupt-project-uid"
	taskUID := string(types.NewTaskUID())

	// Create a task with a non-existent project
	// This should create a synthetic project
	taskPayload := types.TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: 1,
		CreatedNode:    string(types.NewNodeID()),
		Title:          "Test task",
		CreatedBy:      "alice",
	}
	taskPayloadJSON, _ := json.Marshal(taskPayload)

	taskEvent := types.Event{
		ID:        "event_01",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.created",
		Payload:   taskPayloadJSON,
	}

	if err := r.Apply(taskEvent); err != nil {
		t.Fatalf("Failed to apply task.created event: %v", err)
	}

	// Verify synthetic project was created
	project, ok := r.GetProject(projectUID)
	if !ok {
		t.Fatal("Synthetic project should have been created")
	}

	if !project.IsSynthetic {
		t.Error("Project should be marked as synthetic")
	}
	if project.Name != projectUID {
		t.Errorf("Synthetic project name = %s, want %s", project.Name, projectUID)
	}
	if project.Type != "local" {
		t.Errorf("Synthetic project type = %s, want local", project.Type)
	}
	if project.CreatedBy != "system" {
		t.Errorf("Synthetic project created_by = %s, want system", project.CreatedBy)
	}
}

func TestReducer_ProjectTracking_DeleteProject(t *testing.T) {
	r := NewReducer()

	projectUID := types.NewProjectUID()

	// Create a project
	createPayload := types.ProjectCreatedPayload{
		ProjectUID:  projectUID,
		Type:        "local",
		Name:        "test-project",
		Description: "A test project",
		CreatedBy:   "alice",
	}
	createPayloadJSON, _ := json.Marshal(createPayload)

	createEvent := types.Event{
		ID:        "event_01",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "project.created",
		Payload:   createPayloadJSON,
	}

	if err := r.Apply(createEvent); err != nil {
		t.Fatalf("Failed to apply project.created event: %v", err)
	}

	// Verify project exists
	if _, ok := r.GetProject(string(projectUID)); !ok {
		t.Fatal("Project should exist after creation")
	}

	// Delete the project
	deletePayload := types.ProjectDeletePayload{ProjectUID: projectUID}
	deletePayloadJSON, _ := json.Marshal(deletePayload)

	deleteEvent := types.Event{
		ID:        "event_02",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "project.delete",
		Payload:   deletePayloadJSON,
	}

	if err := r.Apply(deleteEvent); err != nil {
		t.Fatalf("Failed to apply project.delete event: %v", err)
	}

	// Verify project no longer exists
	if _, ok := r.GetProject(string(projectUID)); ok {
		t.Error("Project should not exist after deletion")
	}
}

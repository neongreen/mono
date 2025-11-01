package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

func TestProjectProjectCreatedEvent(t *testing.T) {
	db := openTempDB(t)

	projectUID := string(NewProjectUID())
	payload := ProjectCreatedPayload{
		ProjectUID:  projectUID,
		Type:        "local",
		Name:        "test-project",
		Description: "Test description",
		CreatedBy:   "tester",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event := types.Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(EventKindProjectCreated),
		Payload:   payloadJSON,
	}

	if err := db.ProjectProjectCreatedEvent(event); err != nil {
		t.Fatalf("ProjectProjectCreatedEvent() error = %v", err)
	}

	// Verify project was created in database
	var count int
	err = db.db.QueryRow(`SELECT COUNT(*) FROM projects WHERE project_uid = ?`, projectUID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query projects: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 project, got %d", count)
	}
}

func TestProjectProjectAliasAddEvent(t *testing.T) {
	db := openTempDB(t)

	// First create a project
	projectUID := seedProject(t, db, "oldname")

	// Add a new alias
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node ID: %v", err)
	}

	payload := ProjectAliasAddPayload{
		ProjectUID: projectUID,
		Alias:      "newalias",
		Node:       nodeID,
		AddedBy:    "tester",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event := types.Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(EventKindProjectAliasAdd),
		Payload:   payloadJSON,
	}

	if err := db.ProjectProjectAliasAddEvent(event); err != nil {
		t.Fatalf("ProjectProjectAliasAddEvent() error = %v", err)
	}

	// Verify alias was added
	var count int
	err = db.db.QueryRow(`SELECT COUNT(*) FROM project_aliases WHERE project_uid = ? AND alias = ?`, projectUID, "newalias").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query project_aliases: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 alias, got %d", count)
	}
}

func TestProjectTaskCreatedEvent(t *testing.T) {
	db := openTempDB(t)

	// Create a project first
	projectUID := seedProject(t, db, "test")

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node ID: %v", err)
	}

	taskUID := string(NewTaskUID())
	payload := TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: 1,
		CreatedNode:    nodeID,
		Title:          "Test task",
		CreatedBy:      "tester",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event := types.Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(EventKindTaskCreated),
		Payload:   payloadJSON,
	}

	if err := db.ProjectTaskCreatedEvent(event); err != nil {
		t.Fatalf("ProjectTaskCreatedEvent() error = %v", err)
	}

	// Verify task was created
	var count int
	err = db.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE task_uid = ?`, taskUID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query tasks: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 task, got %d", count)
	}
}

func TestProjectTaskNumberSetEvent(t *testing.T) {
	db := openTempDB(t)

	// Create a project and task
	projectUID := seedProject(t, db, "test")
	taskUID := seedTask(t, db, projectUID, "Test task", 1)

	// Update the number
	payload := TaskNumberSetPayload{
		TaskUID:    taskUID,
		ProjectUID: projectUID,
		Number:     42,
		Reason:     "manual",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event := types.Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(EventKindTaskNumberSet),
		Payload:   payloadJSON,
	}

	if err := db.ProjectTaskNumberSetEvent(event); err != nil {
		t.Fatalf("ProjectTaskNumberSetEvent() error = %v", err)
	}

	// Verify number was updated
	var number int64
	err = db.db.QueryRow(`SELECT number FROM task_numbers WHERE task_uid = ?`, taskUID).Scan(&number)
	if err != nil {
		t.Fatalf("failed to query task number: %v", err)
	}

	if number != 42 {
		t.Errorf("expected number 42, got %d", number)
	}
}

func TestProjectTaskTitleSetEvent(t *testing.T) {
	db := openTempDB(t)

	// Create a project and task
	projectUID := seedProject(t, db, "test")
	taskUID := seedTask(t, db, projectUID, "Old title", 1)

	// Update the title
	payload := TaskTitleSetPayload{
		TaskUID: taskUID,
		Title:   "New title",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	event := types.Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(EventKindTaskTitleSet),
		Payload:   payloadJSON,
	}

	if err := db.ProjectTaskTitleSetEvent(event); err != nil {
		t.Fatalf("ProjectTaskTitleSetEvent() error = %v", err)
	}

	// Verify title was updated
	var title string
	err = db.db.QueryRow(`SELECT title FROM tasks WHERE task_uid = ?`, taskUID).Scan(&title)
	if err != nil {
		t.Fatalf("failed to query task title: %v", err)
	}

	if title != "New title" {
		t.Errorf("expected title 'New title', got %s", title)
	}
}

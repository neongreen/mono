package main

import (
	"encoding/json"
	"fmt"
)

// V4 Reducer Functions
// Handles v4 events (project.created, task.created, task.number.set, task.relocate, etc.)

// V4Reducer extends Reducer with v4-specific state
type V4Reducer struct {
	projects       map[string]*V4Project         // Key: project_uid
	projectAliases map[string]map[string]string  // Key: node -> alias -> project_uid
	taskNumbers    map[string]map[int64][]string // Key: project_uid -> number -> []task_uid (for collisions)
	taskProjects   map[string]string             // Key: task_uid -> project_uid
}

// V4Project represents project state
type V4Project struct {
	ProjectUID  string
	Type        string
	Name        string
	Description string
	CreatedBy   string
	CreatedAt   int64
	Aliases     map[string][]string // Key: node -> []alias
}

// NewV4Reducer creates a new v4 reducer
func NewV4Reducer() *V4Reducer {
	return &V4Reducer{
		projects:       make(map[string]*V4Project),
		projectAliases: make(map[string]map[string]string),
		taskNumbers:    make(map[string]map[int64][]string),
		taskProjects:   make(map[string]string),
	}
}

// ApplyV4Event applies a v4-specific event
// Returns (handled=true, error) if the event was a v4 event that was handled
// Returns (handled=false, nil) if the event was not a v4 event
func (r *Reducer) ApplyV4Event(e Event) (bool, error) {
	switch EventKind(e.Kind) {
	case EventKindProjectCreated:
		return true, r.applyProjectCreated(e)
	case EventKindProjectAliasAdd:
		return true, r.applyProjectAliasAdd(e)
	case EventKindProjectAliasRemove:
		return true, r.applyProjectAliasRemove(e)
	case EventKindTaskCreated:
		// Check if this is a v4 task.created event by checking for project_uid field
		var testPayload struct {
			ProjectUID string `json:"project_uid"`
		}
		if err := json.Unmarshal(e.Payload, &testPayload); err == nil && testPayload.ProjectUID != "" {
			// This is a v4 event
			return true, r.applyTaskCreatedV4(e)
		}
		// This is a v1/v2 event, let the legacy handler process it
		return false, nil
	case EventKindTaskNumberSet:
		return true, r.applyTaskNumberSet(e)
	case EventKindTaskRelocate:
		return true, r.applyTaskRelocate(e)
	case EventKindTaskTitleSet:
		return true, r.applyTaskTitleSet(e)
	default:
		// Not a v4 event, skip
		return false, nil
	}
}

func (r *Reducer) applyProjectCreated(e Event) error {
	var payload ProjectCreatedPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project.created payload: %w", err)
	}

	// Project state is managed by DB projections, not in-memory reducer
	// This is just for completeness
	return nil
}

func (r *Reducer) applyProjectAliasAdd(e Event) error {
	var payload ProjectAliasAddPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project.alias.add payload: %w", err)
	}

	// Alias state is managed by DB projections
	return nil
}

func (r *Reducer) applyProjectAliasRemove(e Event) error {
	var payload ProjectAliasRemovePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project.alias.remove payload: %w", err)
	}

	// Alias state is managed by DB projections
	return nil
}

func (r *Reducer) applyTaskCreatedV4(e Event) error {
	var payload TaskCreatedV4Payload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.created payload: %w", err)
	}

	taskUID := payload.TaskUID

	// Guard against duplicate task.created for same UID
	if _, exists := r.tasks[taskUID]; exists {
		return nil
	}

	// In v4, task display ID is derived from project alias + number
	// For now, we'll use the task_uid as the task_id until we compute the display ID
	r.tasks[taskUID] = &Task{
		TaskUUID:  taskUID,
		TaskID:    taskUID, // Placeholder, will be replaced by display ID
		Aliases:   []string{},
		Title:     payload.Title,
		Axes:      make(map[string]AxisStatus),
		Notes:     []Note{},
		CreatedBy: payload.CreatedBy,
		CreatedAt: e.CreatedAt,
	}

	// Register task by UID
	r.taskByID[taskUID] = taskUID

	return nil
}

func (r *Reducer) applyTaskNumberSet(e Event) error {
	var payload TaskNumberSetPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.number.set payload: %w", err)
	}

	// Number assignments are managed by DB projections (task_numbers table)
	// The reducer doesn't need to track this in memory
	return nil
}

func (r *Reducer) applyTaskRelocate(e Event) error {
	var payload TaskRelocatePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.relocate payload: %w", err)
	}

	// Task relocations are managed by DB projections
	// The task itself doesn't change in the in-memory reducer
	return nil
}

func (r *Reducer) applyTaskTitleSet(e Event) error {
	var payload TaskTitleSetPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.title.set payload: %w", err)
	}

	taskUID := payload.TaskUID

	// Find the task
	task, exists := r.tasks[taskUID]
	if !exists {
		return fmt.Errorf("task %s not found", taskUID)
	}

	// Update the title
	task.Title = payload.Title

	return nil
}

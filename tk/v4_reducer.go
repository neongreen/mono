package main

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/types"
)

// Reducer Functions
// Handles events (project.created, task.created, task.number.set, task.relocate, etc.)

// ApplyProjectEvent applies an event-specific handler
// Returns (handled=true, error) if the event was handled
// Returns (handled=false, nil) if the event was not handled
func (r *Reducer) ApplyProjectEvent(e Event) (bool, error) {
	switch EventKind(e.Kind) {
	case EventKindProjectCreated:
		return true, r.applyProjectCreated(e)
	case EventKindProjectAliasAdd:
		return true, r.applyProjectAliasAdd(e)
	case EventKindProjectAliasRemove:
		return true, r.applyProjectAliasRemove(e)
	case EventKindTaskCreated:
		return true, r.applyTaskCreated(e)
	case EventKindTaskNumberSet:
		return true, r.applyTaskNumberSet(e)
	case EventKindTaskRelocate:
		return true, r.applyTaskRelocate(e)
	case EventKindTaskTitleSet:
		return true, r.applyTaskTitleSet(e)
	default:
		// Not a handled event, skip
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

func (r *Reducer) applyTaskCreated(e Event) error {
	var payload TaskCreatedPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.created payload: %w", err)
	}

	taskUID := payload.TaskUID

	// Guard against duplicate task.created for same UID
	if _, exists := r.tasks[taskUID]; exists {
		return nil
	}

	// Task display ID is derived from project alias + number
	// For now, we'll use the task_uid as the task_id until we compute the display ID
	r.tasks[taskUID] = &Task{
		TaskUUID:  taskUID,
		TaskID:    taskUID, // Placeholder, will be replaced by display ID
		Aliases:   []string{},
		Title:     payload.Title,
		Axes:      make(map[string]types.AxisStatus),
		Notes:     []types.Note{},
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

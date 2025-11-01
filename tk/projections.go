package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/types"
)

// Projection Functions
// These functions project events from the events table into projection tables

// ProjectProjectCreatedEvent projects a project.created event into the projects table (idempotent)
func (d *DB) ProjectProjectCreatedEvent(e types.Event) error {
	if e.Kind != string(types.EventKindProjectCreated) {
		return fmt.Errorf("expected project.created event, got %s", e.Kind)
	}

	var payload types.ProjectCreatedPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project.created payload: %w", err)
	}

	// Project into projects table (idempotent)
	_, err := d.db.Exec(`
		INSERT OR IGNORE INTO projects (project_uid, type, name, description, created_at, created_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, payload.ProjectUID, payload.Type, payload.Name, payload.Description, e.CreatedAt.Unix(), payload.CreatedBy)

	return err
}

// ProjectProjectAliasAddEvent projects a project.alias.add event into the project_aliases table (idempotent)
func (d *DB) ProjectProjectAliasAddEvent(e types.Event) error {
	if e.Kind != string(types.EventKindProjectAliasAdd) {
		return fmt.Errorf("expected project.alias.add event, got %s", e.Kind)
	}

	var payload types.ProjectAliasAddPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project.alias.add payload: %w", err)
	}

	// Project into project_aliases table (idempotent)
	_, err := d.db.Exec(`
		INSERT OR REPLACE INTO project_aliases (project_uid, alias, node, added_by)
		VALUES (?, ?, ?, ?)
	`, payload.ProjectUID, payload.Alias, payload.Node, payload.AddedBy)

	return err
}

// ProjectProjectAliasRemoveEvent projects a project.alias.remove event by removing from project_aliases table (idempotent)
func (d *DB) ProjectProjectAliasRemoveEvent(e types.Event) error {
	if e.Kind != string(types.EventKindProjectAliasRemove) {
		return fmt.Errorf("expected project.alias.remove event, got %s", e.Kind)
	}

	var payload types.ProjectAliasRemovePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project.alias.remove payload: %w", err)
	}

	// Remove from project_aliases table (idempotent)
	_, err := d.db.Exec(`
		DELETE FROM project_aliases 
		WHERE project_uid = ? AND alias = ? AND node = ?
	`, payload.ProjectUID, payload.Alias, payload.Node)

	return err
}

// ProjectProjectDeleteEvent projects a project.delete event by removing from projects, project_aliases, and all tasks in the project (idempotent)
func (d *DB) ProjectProjectDeleteEvent(e types.Event) error {
	if e.Kind != string(types.EventKindProjectDelete) {
		return fmt.Errorf("expected project.delete event, got %s", e.Kind)
	}

	var payload types.ProjectDeletePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project.delete payload: %w", err)
	}

	// Use a transaction to ensure atomicity
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete all tasks in this project from task_numbers table
	_, err = tx.Exec(`DELETE FROM task_numbers WHERE project_uid = ?`, payload.ProjectUID)
	if err != nil {
		return fmt.Errorf("failed to delete task_numbers for project: %w", err)
	}

	// Delete all tasks in this project from tasks table
	_, err = tx.Exec(`DELETE FROM tasks WHERE project_uid = ?`, payload.ProjectUID)
	if err != nil {
		return fmt.Errorf("failed to delete tasks for project: %w", err)
	}

	// Delete from project_aliases table
	_, err = tx.Exec(`DELETE FROM project_aliases WHERE project_uid = ?`, payload.ProjectUID)
	if err != nil {
		return fmt.Errorf("failed to delete from project_aliases: %w", err)
	}

	// Delete from projects table (idempotent)
	_, err = tx.Exec(`DELETE FROM projects WHERE project_uid = ?`, payload.ProjectUID)
	if err != nil {
		return fmt.Errorf("failed to delete from projects: %w", err)
	}

	return tx.Commit()
}

// ProjectTaskCreatedEvent projects a task.created event into the tasks table (idempotent)
func (d *DB) ProjectTaskCreatedEvent(e types.Event) error {
	if e.Kind != string(types.EventKindTaskCreated) {
		return fmt.Errorf("expected task.created event, got %s", e.Kind)
	}

	var payload types.TaskCreatedPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.created payload: %w", err)
	}

	// Project into tasks table (idempotent)
	_, err := d.db.Exec(`
		INSERT OR IGNORE INTO tasks (task_uid, project_uid, created_node, title, created_at, created_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, payload.TaskUID, payload.ProjectUID, payload.CreatedNode, payload.Title, e.CreatedAt.Unix(), payload.CreatedBy)

	return err
}

// ProjectTaskNumberSetEvent projects a task.number.set event into the task_numbers table (idempotent)
func (d *DB) ProjectTaskNumberSetEvent(e types.Event) error {
	if e.Kind != string(types.EventKindTaskNumberSet) {
		return fmt.Errorf("expected task.number.set event, got %s", e.Kind)
	}

	var payload types.TaskNumberSetPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.number.set payload: %w", err)
	}

	// First, remove any existing number assignment for this task (to handle renumbering)
	_, err := d.db.Exec(`
		DELETE FROM task_numbers WHERE task_uid = ?
	`, payload.TaskUID)
	if err != nil {
		return err
	}

	// Then insert the new number (idempotent)
	_, err = d.db.Exec(`
		INSERT INTO task_numbers (project_uid, number, task_uid)
		VALUES (?, ?, ?)
	`, payload.ProjectUID, payload.Number, payload.TaskUID)

	return err
}

// ProjectTaskRelocateEvent projects a task.relocate event by updating project and number (idempotent)
func (d *DB) ProjectTaskRelocateEvent(e types.Event) error {
	if e.Kind != string(types.EventKindTaskRelocate) {
		return fmt.Errorf("expected task.relocate event, got %s", e.Kind)
	}

	var payload types.TaskRelocatePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.relocate payload: %w", err)
	}

	// Update project in tasks table
	_, err := d.db.Exec(`
		UPDATE tasks SET project_uid = ? WHERE task_uid = ?
	`, payload.ToProjectUID, payload.TaskUID)
	if err != nil {
		return err
	}

	// Update number in task_numbers table
	// For "keep" mode, retrieve old number before deleting
	var oldNumber int64
	if payload.NumberPolicy.Mode == "keep" {
		err = d.db.QueryRow(`
			SELECT number FROM task_numbers WHERE task_uid = ?
		`, payload.TaskUID).Scan(&oldNumber)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to get old number: %w", err)
		}
	}

	// Remove old number assignment
	_, err = d.db.Exec(`
		DELETE FROM task_numbers WHERE task_uid = ?
	`, payload.TaskUID)
	if err != nil {
		return err
	}

	// Then add new number based on policy
	var number int64
	switch payload.NumberPolicy.Mode {
	case "force":
		number = payload.NumberPolicy.Number
	case "keep":
		// Keep the old number - retrieve it from the old assignment
		// If provided in the policy, use that; otherwise use the old number we retrieved
		if payload.NumberPolicy.Number > 0 {
			number = payload.NumberPolicy.Number
		} else {
			number = oldNumber
		}
	case "auto":
		// Auto-assign next available number
		var maxNumber int64
		err = d.db.QueryRow(`
			SELECT COALESCE(MAX(number), 0) FROM task_numbers 
			WHERE project_uid = ?
		`, payload.ToProjectUID).Scan(&maxNumber)
		if err != nil {
			return err
		}
		number = maxNumber + 1
	default:
		return fmt.Errorf("unknown number policy mode: %s", payload.NumberPolicy.Mode)
	}

	_, err = d.db.Exec(`
		INSERT INTO task_numbers (project_uid, number, task_uid)
		VALUES (?, ?, ?)
	`, payload.ToProjectUID, number, payload.TaskUID)

	return err
}

// ProjectTaskTitleSetEvent projects a task.title.set event by updating the title (idempotent)
func (d *DB) ProjectTaskTitleSetEvent(e types.Event) error {
	if e.Kind != string(types.EventKindTaskTitleSet) {
		return fmt.Errorf("expected task.title.set event, got %s", e.Kind)
	}

	var payload types.TaskTitleSetPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.title.set payload: %w", err)
	}

	// Update title in tasks table (idempotent)
	_, err := d.db.Exec(`
		UPDATE tasks SET title = ? WHERE task_uid = ?
	`, payload.Title, payload.TaskUID)

	return err
}

// ProjectTaskDeleteEvent projects a task.delete event by removing the task from tables (idempotent)
func (d *DB) ProjectTaskDeleteEvent(e types.Event) error {
	if e.Kind != string(types.EventKindTaskDelete) {
		return fmt.Errorf("expected task.delete event, got %s", e.Kind)
	}

	var payload types.TaskDeletePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.delete payload: %w", err)
	}

	// Delete from task_numbers table
	_, err := d.db.Exec(`DELETE FROM task_numbers WHERE task_uid = ?`, payload.TaskUUID)
	if err != nil {
		return fmt.Errorf("failed to delete from task_numbers: %w", err)
	}

	// Delete from tasks table (idempotent)
	_, err = d.db.Exec(`DELETE FROM tasks WHERE task_uid = ?`, payload.TaskUUID)
	if err != nil {
		return fmt.Errorf("failed to delete from tasks: %w", err)
	}

	return nil
}

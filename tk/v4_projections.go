package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// V4 Projection Functions
// These functions project v4 events from the events table into projection tables

// ProjectProjectCreatedEvent projects a project.created event into the projects table (idempotent)
func (d *DB) ProjectProjectCreatedEvent(e Event) error {
	if e.Kind != string(EventKindProjectCreated) {
		return fmt.Errorf("expected project.created event, got %s", e.Kind)
	}

	var payload ProjectCreatedPayload
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
func (d *DB) ProjectProjectAliasAddEvent(e Event) error {
	if e.Kind != string(EventKindProjectAliasAdd) {
		return fmt.Errorf("expected project.alias.add event, got %s", e.Kind)
	}

	var payload ProjectAliasAddPayload
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
func (d *DB) ProjectProjectAliasRemoveEvent(e Event) error {
	if e.Kind != string(EventKindProjectAliasRemove) {
		return fmt.Errorf("expected project.alias.remove event, got %s", e.Kind)
	}

	var payload ProjectAliasRemovePayload
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

// ProjectTaskCreatedV4Event projects a task.created (v4) event into the tasks table (idempotent)
func (d *DB) ProjectTaskCreatedV4Event(e Event) error {
	if e.Kind != string(EventKindTaskCreated) {
		return fmt.Errorf("expected task.created event, got %s", e.Kind)
	}

	var payload TaskCreatedV4Payload
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
func (d *DB) ProjectTaskNumberSetEvent(e Event) error {
	if e.Kind != string(EventKindTaskNumberSet) {
		return fmt.Errorf("expected task.number.set event, got %s", e.Kind)
	}

	var payload TaskNumberSetPayload
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
func (d *DB) ProjectTaskRelocateEvent(e Event) error {
	if e.Kind != string(EventKindTaskRelocate) {
		return fmt.Errorf("expected task.relocate event, got %s", e.Kind)
	}

	var payload TaskRelocatePayload
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
func (d *DB) ProjectTaskTitleSetEvent(e Event) error {
	if e.Kind != string(EventKindTaskTitleSet) {
		return fmt.Errorf("expected task.title.set event, got %s", e.Kind)
	}

	var payload TaskTitleSetPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.title.set payload: %w", err)
	}

	// Update title in tasks table (idempotent)
	_, err := d.db.Exec(`
		UPDATE tasks SET title = ? WHERE task_uid = ?
	`, payload.Title, payload.TaskUID)

	return err
}

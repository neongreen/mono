package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/neongreen/mono/tk/internal/types"
)

// Projection Functions
// These functions project events from the events table into projection tables

// ProjectEvent projects an event into its respective table based on event kind
func (d *DB) ProjectEvent(event types.Event) error {
	switch event.Kind {
	case string(types.EventKindProjectCreated):
		return d.ProjectProjectCreatedEvent(event)
	case string(types.EventKindProjectDelete):
		return d.ProjectProjectDeleteEvent(event)
	case string(types.EventKindProjectNameSet):
		return d.ProjectProjectNameSetEvent(event)
	case string(types.EventKindTaskCreated):
		return d.ProjectTaskCreatedEvent(event)
	case string(types.EventKindTaskNumberSet):
		return d.ProjectTaskNumberSetEvent(event)
	case string(types.EventKindTaskRelocate):
		return d.ProjectTaskRelocateEvent(event)
	case string(types.EventKindTaskTitleSet):
		return d.ProjectTaskTitleSetEvent(event)
	case string(types.EventKindTaskDelete):
		return d.ProjectTaskDeleteEvent(event)

	// Container schema events
	case string(types.EventKindContainerKindDefine):
		return d.ProjectContainerKindDefineEvent(event)
	case string(types.EventKindContainerKindDeprecate):
		return d.ProjectContainerKindDeprecateEvent(event)

	// Container instance events
	case string(types.EventKindContainerCreate):
		return d.ProjectContainerCreateEvent(event)
	case string(types.EventKindContainerRename):
		return d.ProjectContainerRenameEvent(event)
	case string(types.EventKindContainerMetadataUpdate):
		return d.ProjectContainerMetadataUpdateEvent(event)
	case string(types.EventKindContainerRemove):
		return d.ProjectContainerRemoveEvent(event)
	}
	return nil
}

// ProjectProjectCreatedEvent projects a project.created event into the projects table (idempotent)
func (d *DB) ProjectProjectCreatedEvent(e types.Event) error {
	if e.Kind != string(types.EventKindProjectCreated) {
		return fmt.Errorf("expected project.created event, got %s", e.Kind)
	}

	var payload types.ProjectCreatedPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project.created payload: %w", err)
	}

	// Check if project already exists (might be synthetic)
	var exists bool
	var isSynthetic int
	err := d.Db.QueryRow(`SELECT 1, COALESCE(is_synthetic, 0) FROM projects WHERE project_uid = ?`,
		payload.ProjectUID).Scan(&exists, &isSynthetic)

	if errors.Is(err, sql.ErrNoRows) {
		// Project doesn't exist - insert it
		_, err = d.Db.Exec(`
			INSERT INTO projects (project_uid, type, name, description, created_at, created_by, is_synthetic)
			VALUES (?, ?, ?, ?, ?, ?, 0)
		`, payload.ProjectUID, payload.Type, payload.Name, payload.Description, e.CreatedAt.Unix(), payload.CreatedBy)
		return err
	} else if err != nil {
		return fmt.Errorf("failed to check if project exists: %w", err)
	}

	// Project exists - update it if it was synthetic
	if isSynthetic == 1 {
		_, err = d.Db.Exec(`
			UPDATE projects
			SET type = ?, name = ?, description = ?, created_at = ?, created_by = ?, is_synthetic = 0
			WHERE project_uid = ?
		`, payload.Type, payload.Name, payload.Description, e.CreatedAt.Unix(), payload.CreatedBy, payload.ProjectUID)
		return err
	}

	// Real project already exists - do nothing (idempotent)
	return nil
}

// ProjectProjectDeleteEvent projects a project.delete event by removing from projects and all tasks in the project (idempotent)
func (d *DB) ProjectProjectDeleteEvent(e types.Event) error {
	if e.Kind != string(types.EventKindProjectDelete) {
		return fmt.Errorf("expected project.delete event, got %s", e.Kind)
	}

	var payload types.ProjectDeletePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project.delete payload: %w", err)
	}

	// Use a transaction to ensure atomicity
	tx, err := d.Db.Begin()
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

	// Delete from projects table (idempotent)
	_, err = tx.Exec(`DELETE FROM projects WHERE project_uid = ?`, payload.ProjectUID)
	if err != nil {
		return fmt.Errorf("failed to delete from projects: %w", err)
	}

	return tx.Commit()
}

// ProjectProjectNameSetEvent projects a project.name.set event by updating the project name (idempotent)
func (d *DB) ProjectProjectNameSetEvent(e types.Event) error {
	if e.Kind != string(types.EventKindProjectNameSet) {
		return fmt.Errorf("expected project.name.set event, got %s", e.Kind)
	}

	var payload types.ProjectNameSetPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project.name.set payload: %w", err)
	}

	// Update project name in projects table (idempotent)
	_, err := d.Db.Exec(`
		UPDATE projects
		SET name = ?
		WHERE project_uid = ?
	`, payload.Name, payload.ProjectUID)

	return err
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

	// Check if project exists
	// If not, create a synthetic project (v5 feature for handling corrupt data)
	var exists bool
	err := d.Db.QueryRow(`SELECT 1 FROM projects WHERE project_uid = ?`, payload.ProjectUID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		// Project doesn't exist - create synthetic project
		// Use the literal corrupt value as both uid and name
		_, err = d.Db.Exec(`
			INSERT OR IGNORE INTO projects (project_uid, name, type, is_synthetic, description, created_at, created_by)
			VALUES (?, ?, 'local', 1, 'Synthetic project created by projection layer', ?, 'system')
		`, payload.ProjectUID, payload.ProjectUID, e.CreatedAt.Unix())
		if err != nil {
			return fmt.Errorf("failed to create synthetic project: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check if project exists: %w", err)
	}

	// Project into tasks table (idempotent)
	_, err = d.Db.Exec(`
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

	// Check if project exists
	// If not, create a synthetic project (v5 feature for handling corrupt data)
	var exists bool
	err := d.Db.QueryRow(`SELECT 1 FROM projects WHERE project_uid = ?`, payload.ProjectUID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		// Project doesn't exist - create synthetic project
		// Use the literal corrupt value as both uid and name
		_, err = d.Db.Exec(`
			INSERT OR IGNORE INTO projects (project_uid, name, type, is_synthetic, description, created_at, created_by)
			VALUES (?, ?, 'local', 1, 'Synthetic project created by projection layer', ?, 'system')
		`, payload.ProjectUID, payload.ProjectUID, e.CreatedAt.Unix())
		if err != nil {
			return fmt.Errorf("failed to create synthetic project: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check if project exists: %w", err)
	}

	// First, remove any existing number assignment for this task (to handle renumbering)
	_, err = d.Db.Exec(`
		DELETE FROM task_numbers WHERE task_uid = ?
	`, payload.TaskUID)
	if err != nil {
		return err
	}

	// Then insert the new number (idempotent)
	_, err = d.Db.Exec(`
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

	// Check if destination project exists
	// If not, create a synthetic project (v5 feature for handling corrupt data)
	var exists bool
	err := d.Db.QueryRow(`SELECT 1 FROM projects WHERE project_uid = ?`, payload.ToProjectUID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		// Project doesn't exist - create synthetic project
		// Use the literal corrupt value as both uid and name
		_, err = d.Db.Exec(`
			INSERT OR IGNORE INTO projects (project_uid, name, type, is_synthetic, description, created_at, created_by)
			VALUES (?, ?, 'local', 1, 'Synthetic project created by projection layer', ?, 'system')
		`, payload.ToProjectUID, payload.ToProjectUID, e.CreatedAt.Unix())
		if err != nil {
			return fmt.Errorf("failed to create synthetic project: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check if project exists: %w", err)
	}

	// Update project in tasks table
	_, err = d.Db.Exec(`
		UPDATE tasks SET project_uid = ? WHERE task_uid = ?
	`, payload.ToProjectUID, payload.TaskUID)
	if err != nil {
		return err
	}

	// Update number in task_numbers table
	// For "keep" mode, retrieve old number before deleting
	var oldNumber int64
	if payload.NumberPolicy.Mode == "keep" {
		err = d.Db.QueryRow(`
			SELECT number FROM task_numbers WHERE task_uid = ?
		`, payload.TaskUID).Scan(&oldNumber)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to get old number: %w", err)
		}
	}

	// Remove old number assignment
	_, err = d.Db.Exec(`
		DELETE FROM task_numbers WHERE task_uid = ?
	`, payload.TaskUID)
	if err != nil {
		return err
	}

	// Then add new number based on policy
	// NOTE: "auto" mode should have been resolved to "force" at event creation time.
	// However, for backward compatibility with old events, we still handle "auto" mode here.
	// New events (after our fix) will always use "force" mode.
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
		// Legacy support: handle old "auto" mode events
		// This is non-deterministic, but needed for backward compatibility
		// All new events will use "force" mode instead
		var maxNumber int64
		err = d.Db.QueryRow(`
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

	_, err = d.Db.Exec(`
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
	_, err := d.Db.Exec(`
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
	_, err := d.Db.Exec(`DELETE FROM task_numbers WHERE task_uid = ?`, payload.TaskUUID)
	if err != nil {
		return fmt.Errorf("failed to delete from task_numbers: %w", err)
	}

	// Delete from tasks table (idempotent)
	_, err = d.Db.Exec(`DELETE FROM tasks WHERE task_uid = ?`, payload.TaskUUID)
	if err != nil {
		return fmt.Errorf("failed to delete from tasks: %w", err)
	}

	return nil
}

// Container projection functions (v6)

// ProjectContainerKindDefineEvent projects a container.kind.define event into the container_kinds table (idempotent)
func (d *DB) ProjectContainerKindDefineEvent(e types.Event) error {
	if e.Kind != string(types.EventKindContainerKindDefine) {
		return fmt.Errorf("expected container.kind.define event, got %s", e.Kind)
	}

	var payload types.DefineContainerKindPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal container.kind.define payload: %w", err)
	}

	// Check if kind already exists
	var exists bool
	err := d.Db.QueryRow(`SELECT 1 FROM container_kinds WHERE name = ?`, payload.Name).Scan(&exists)

	if errors.Is(err, sql.ErrNoRows) {
		// Kind doesn't exist - insert it
		_, err = d.Db.Exec(`
			INSERT INTO container_kinds (name, primitive, description, deprecated, created_at, created_by)
			VALUES (?, ?, ?, 0, ?, ?)
		`, payload.Name, payload.Primitive, payload.Description, e.CreatedAt.Unix(), payload.CreatedBy)
		return err
	} else if err != nil {
		return fmt.Errorf("failed to check if container kind exists: %w", err)
	}

	// Kind exists - update description only (idempotent, allows hint updates)
	_, err = d.Db.Exec(`
		UPDATE container_kinds
		SET description = ?
		WHERE name = ?
	`, payload.Description, payload.Name)

	return err
}

// ProjectContainerKindDeprecateEvent projects a container.kind.deprecate event by marking the kind as deprecated (idempotent)
func (d *DB) ProjectContainerKindDeprecateEvent(e types.Event) error {
	if e.Kind != string(types.EventKindContainerKindDeprecate) {
		return fmt.Errorf("expected container.kind.deprecate event, got %s", e.Kind)
	}

	var payload types.DeprecateContainerKindPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal container.kind.deprecate payload: %w", err)
	}

	// Mark as deprecated (idempotent)
	_, err := d.Db.Exec(`
		UPDATE container_kinds
		SET deprecated = 1
		WHERE name = ?
	`, payload.Name)

	return err
}

// ProjectContainerCreateEvent projects a container.create event into the containers table (idempotent)
func (d *DB) ProjectContainerCreateEvent(e types.Event) error {
	if e.Kind != string(types.EventKindContainerCreate) {
		return fmt.Errorf("expected container.create event, got %s", e.Kind)
	}

	var payload types.CreateContainerPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal container.create payload: %w", err)
	}

	// Serialize metadata to JSON
	var metadataJSON []byte
	if payload.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(payload.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	// Insert container (idempotent with INSERT OR IGNORE)
	_, err := d.Db.Exec(`
		INSERT OR IGNORE INTO containers (id, primitive, kind, name, metadata, removed)
		VALUES (?, ?, ?, ?, ?, 0)
	`, payload.ID, payload.Primitive, payload.Kind, payload.Name, metadataJSON)

	return err
}

// ProjectContainerRenameEvent projects a container.rename event by updating the container name (idempotent)
func (d *DB) ProjectContainerRenameEvent(e types.Event) error {
	if e.Kind != string(types.EventKindContainerRename) {
		return fmt.Errorf("expected container.rename event, got %s", e.Kind)
	}

	var payload types.RenameContainerPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal container.rename payload: %w", err)
	}

	// Update name (idempotent)
	_, err := d.Db.Exec(`
		UPDATE containers
		SET name = ?
		WHERE id = ?
	`, payload.Name, payload.ID)

	return err
}

// ProjectContainerMetadataUpdateEvent projects a container.metadata.update event by updating metadata (idempotent)
func (d *DB) ProjectContainerMetadataUpdateEvent(e types.Event) error {
	if e.Kind != string(types.EventKindContainerMetadataUpdate) {
		return fmt.Errorf("expected container.metadata.update event, got %s", e.Kind)
	}

	var payload types.UpdateContainerMetadataPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal container.metadata.update payload: %w", err)
	}

	// Serialize metadata to JSON
	var metadataJSON []byte
	if payload.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(payload.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	// Update metadata (idempotent, overwrites entire metadata blob)
	_, err := d.Db.Exec(`
		UPDATE containers
		SET metadata = ?
		WHERE id = ?
	`, metadataJSON, payload.ID)

	return err
}

// ProjectContainerRemoveEvent projects a container.remove event by soft-deleting the container (idempotent)
func (d *DB) ProjectContainerRemoveEvent(e types.Event) error {
	if e.Kind != string(types.EventKindContainerRemove) {
		return fmt.Errorf("expected container.remove event, got %s", e.Kind)
	}

	var payload types.RemoveContainerPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal container.remove payload: %w", err)
	}

	// Soft delete by setting removed=1 (idempotent)
	// Also soft delete all members
	tx, err := d.Db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE containers
		SET removed = 1
		WHERE id = ?
	`, payload.ID)
	if err != nil {
		return fmt.Errorf("failed to mark container as removed: %w", err)
	}

	_, err = tx.Exec(`
		UPDATE container_members
		SET removed = 1
		WHERE container_id = ?
	`, payload.ID)
	if err != nil {
		return fmt.Errorf("failed to mark container members as removed: %w", err)
	}

	return tx.Commit()
}

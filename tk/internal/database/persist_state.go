package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/reducer"
)

// PersistReducerState persists the complete reducer state to the database.
// This replaces event-based projections with state-based persistence.
// Call this after applying all events to the reducer during ingestion.
func (d *DB) PersistReducerState(r *reducer.Reducer) error {
	// Start a transaction for atomic updates
	tx, err := d.Db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Persist task numbers
	if err := d.persistTaskNumbers(tx, r); err != nil {
		return fmt.Errorf("failed to persist task numbers: %w", err)
	}

	// Persist projects
	if err := d.persistProjects(tx, r); err != nil {
		return fmt.Errorf("failed to persist projects: %w", err)
	}

	// Persist tasks
	if err := d.persistTasks(tx, r); err != nil {
		return fmt.Errorf("failed to persist tasks: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// persistTaskNumbers persists task number assignments from reducer state
func (d *DB) persistTaskNumbers(tx *sql.Tx, r *reducer.Reducer) error {
	// Clear existing task numbers (we're doing a full rebuild)
	if _, err := tx.Exec(`DELETE FROM task_numbers`); err != nil {
		return fmt.Errorf("failed to clear task_numbers: %w", err)
	}

	// Insert all task numbers from reducer
	for _, taskNum := range r.TaskNumbers() {
		_, err := tx.Exec(`
			INSERT INTO task_numbers (project_uid, number, task_uid)
			VALUES (?, ?, ?)
		`, taskNum.ProjectUID, taskNum.Number, taskNum.TaskUID)
		if err != nil {
			return fmt.Errorf("failed to insert task number for task %s: %w", taskNum.TaskUID, err)
		}
	}

	return nil
}

// persistProjects persists projects from reducer state
func (d *DB) persistProjects(tx *sql.Tx, r *reducer.Reducer) error {
	// Note: We do upserts here since projects table may have synthetic projects
	// that were created during event replay

	for projectUID, project := range r.Projects() {
		// Convert deleted flag to integer for SQLite
		deletedInt := 0
		if project.Deleted {
			deletedInt = 1
		}

		// Format deleted_at as RFC3339 string (or empty if zero)
		var deletedAtStr string
		if !project.DeletedAt.IsZero() {
			deletedAtStr = project.DeletedAt.Format(time.RFC3339)
		}

		// Convert time.Time to Unix timestamp for INTEGER column
		createdAtUnix := project.CreatedAt.Unix()

		_, err := tx.Exec(`
			INSERT INTO projects (project_uid, type, name, description, created_at, created_by, deleted, deleted_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(project_uid) DO UPDATE SET
				type = excluded.type,
				name = excluded.name,
				description = excluded.description,
				created_at = excluded.created_at,
				created_by = excluded.created_by,
				deleted = excluded.deleted,
				deleted_at = excluded.deleted_at
		`, projectUID, project.Type, project.Name, project.Description, createdAtUnix, project.CreatedBy, deletedInt, deletedAtStr)
		if err != nil {
			return fmt.Errorf("failed to upsert project %s: %w", projectUID, err)
		}
	}

	return nil
}

// persistTasks persists tasks from reducer state
func (d *DB) persistTasks(tx *sql.Tx, r *reducer.Reducer) error {
	// Clear existing tasks (we're doing a full rebuild)
	if _, err := tx.Exec(`DELETE FROM tasks`); err != nil {
		return fmt.Errorf("failed to clear tasks: %w", err)
	}

	// Insert all tasks from reducer (including deleted ones)
	for taskUID, task := range r.Tasks() {
		// Convert deleted flag to integer for SQLite
		deletedInt := 0
		if task.Deleted {
			deletedInt = 1
		}

		// Format deleted_at as RFC3339 string (or empty if zero)
		var deletedAtStr string
		if !task.DeletedAt.IsZero() {
			deletedAtStr = task.DeletedAt.Format(time.RFC3339)
		}

		// Default item_kind to "task" if not specified
		itemKind := task.ItemKind
		if itemKind == "" {
			itemKind = "task"
		}

		// Convert time.Time to Unix timestamp for INTEGER column
		createdAtUnix := task.CreatedAt.Unix()

		_, err := tx.Exec(`
			INSERT INTO tasks (task_uid, project_uid, created_node, title, created_at, created_by, item_kind, deleted, deleted_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, taskUID, task.ProjectUUID, task.CreatedNode, task.Title, createdAtUnix, task.CreatedBy, itemKind, deletedInt, deletedAtStr)
		if err != nil {
			return fmt.Errorf("failed to insert task %s: %w", taskUID, err)
		}
	}

	return nil
}

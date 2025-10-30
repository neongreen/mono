package main

import (
	"database/sql"
	"fmt"
	"strconv"
)

// CreateProjectTables creates the projection tables for projects and tasks
func (d *DB) CreateProjectTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS projects (
		project_uid TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		created_by TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS project_aliases (
		project_uid TEXT NOT NULL,
		alias TEXT NOT NULL,
		node TEXT NOT NULL,
		added_by TEXT NOT NULL,
		PRIMARY KEY (alias, node)
	);

	CREATE TABLE IF NOT EXISTS tasks (
		task_uid TEXT PRIMARY KEY,
		project_uid TEXT NOT NULL,
		created_node TEXT NOT NULL,
		title TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		created_by TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS task_numbers (
		project_uid TEXT NOT NULL,
		number INTEGER NOT NULL,
		task_uid TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_task_numbers_task_uid ON task_numbers(task_uid);
	CREATE INDEX IF NOT EXISTS idx_task_numbers_project_number ON task_numbers(project_uid, number);
	`

	if _, err := d.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create project tables: %w", err)
	}

	return nil
}

// GetDBVersion returns the database version (always 4)
func (d *DB) GetDBVersion() (int, error) {
	var versionStr string
	err := d.db.QueryRow("SELECT value FROM metadata WHERE key = 'db_version'").Scan(&versionStr)
	if err == sql.ErrNoRows {

		return 4, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get DB version: %w", err)
	}

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return 0, fmt.Errorf("invalid DB version format: %w", err)
	}

	return version, nil
}

// SetDBVersion sets the database version in metadata
func (d *DB) SetDBVersion(version int) error {
	_, err := d.db.Exec(`
		INSERT OR REPLACE INTO metadata (key, value)
		VALUES ('db_version', ?)
	`, strconv.Itoa(version))
	if err != nil {
		return fmt.Errorf("failed to set DB version: %w", err)
	}
	return nil
}

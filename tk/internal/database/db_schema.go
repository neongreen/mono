package database

import (
	"database/sql"
	"errors"
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

	CREATE TABLE IF NOT EXISTS export_state (
		remote_name TEXT NOT NULL,
		space TEXT NOT NULL,
		last_exported_event_id TEXT NOT NULL,
		segment_seq INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		PRIMARY KEY (remote_name, space)
	);

	CREATE TABLE IF NOT EXISTS attachments (
		hash TEXT PRIMARY KEY,
		content BLOB NOT NULL,
		mime_type TEXT,
		size INTEGER NOT NULL,
		created_at TEXT NOT NULL,
		created_by TEXT NOT NULL
	);
	`

	if _, err := d.Db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create project tables: %w", err)
	}

	return nil
}

// CreateItemKindTables creates the projection tables for item kinds
func (d *DB) CreateItemKindTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS item_kinds (
		name TEXT PRIMARY KEY,
		description TEXT,
		llm_hint TEXT,
		builtin INTEGER NOT NULL DEFAULT 0,
		deprecated INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL,
		created_by TEXT NOT NULL
	);
	`

	if _, err := d.Db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create item kind tables: %w", err)
	}

	return nil
}

// AddItemKindToTasks adds the item_kind column to the tasks table
// This is used during migration from v6 to v7
func (d *DB) AddItemKindToTasks() error {
	// Check if column already exists
	var columnExists bool
	row := d.Db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM pragma_table_info('tasks')
		WHERE name = 'item_kind'
	`)
	if err := row.Scan(&columnExists); err != nil {
		return fmt.Errorf("failed to check if item_kind column exists: %w", err)
	}

	if columnExists {
		return nil // Already migrated
	}

	// Add item_kind column with default 'task'
	_, err := d.Db.Exec(`
		ALTER TABLE tasks ADD COLUMN item_kind TEXT NOT NULL DEFAULT 'task'
	`)
	if err != nil {
		return fmt.Errorf("failed to add item_kind column: %w", err)
	}

	// Add foreign key constraint (note: SQLite doesn't enforce existing FKs on ALTER)
	// The constraint will be checked for new rows
	// Full enforcement requires recreating the table, which we'll do in migration

	return nil
}

// CreateContainerTables creates the projection tables for containers
func (d *DB) CreateContainerTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS container_kinds (
		name TEXT PRIMARY KEY,
		primitive TEXT NOT NULL,
		description TEXT,
		deprecated INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL,
		created_by TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS containers (
		id TEXT PRIMARY KEY,
		primitive TEXT NOT NULL,
		kind TEXT NOT NULL,
		name TEXT NOT NULL,
		metadata TEXT,
		removed INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY (kind) REFERENCES container_kinds(name)
	);

	CREATE TABLE IF NOT EXISTS container_members (
		container_id TEXT NOT NULL,
		item_id TEXT NOT NULL,
		position INTEGER,
		removed INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY (container_id, item_id),
		FOREIGN KEY (container_id) REFERENCES containers(id)
	);

	CREATE INDEX IF NOT EXISTS idx_container_members_container_pos
		ON container_members(container_id, position);
	CREATE INDEX IF NOT EXISTS idx_containers_kind
		ON containers(kind);
	CREATE INDEX IF NOT EXISTS idx_containers_removed
		ON containers(removed);
	CREATE INDEX IF NOT EXISTS idx_container_members_removed
		ON container_members(removed);
	`

	if _, err := d.Db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create container tables: %w", err)
	}

	return nil
}

// GetDBVersion returns the database version (always 4)
func (d *DB) GetDBVersion() (int, error) {
	var versionStr string
	err := d.Db.QueryRow("SELECT value FROM metadata WHERE key = 'db_version'").Scan(&versionStr)
	if errors.Is(err, sql.ErrNoRows) {
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
	_, err := d.Db.Exec(`
		INSERT OR REPLACE INTO metadata (key, value)
		VALUES ('db_version', ?)
	`, strconv.Itoa(version))
	if err != nil {
		return fmt.Errorf("failed to set DB version: %w", err)
	}
	return nil
}

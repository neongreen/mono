package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// V4 Migration Functions
// Based on tk/specs/v4-migration.md

const (
	v4MigrationLock = "migrate-v4.lock"
	v4BackupSuffix  = ".v3.bak"
	v4SegmentSubdir = "v4"
	v4SpecVersion   = 4
	v3SpecVersion   = 3 // or could be 2 or 1
)

// GetDBVersion reads the version_major from metadata table
func (d *DB) GetDBVersion() (int, error) {
	var version int
	err := d.db.QueryRow("SELECT value FROM metadata WHERE key = 'version_major'").Scan(&version)
	if err == sql.ErrNoRows {
		// No version found, assume v1/v2
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to read version: %w", err)
	}
	return version, nil
}

// SetDBVersion sets the version_major in metadata table
func (d *DB) SetDBVersion(version int) error {
	_, err := d.db.Exec(`
		INSERT OR REPLACE INTO metadata (key, value) 
		VALUES ('version_major', ?)
	`, fmt.Sprintf("%d", version))
	return err
}

// NeedsMigrationToV4 checks if the database needs v4 migration
func (d *DB) NeedsMigrationToV4() (bool, error) {
	version, err := d.GetDBVersion()
	if err != nil {
		return false, err
	}
	return version < v4SpecVersion, nil
}

// CreateV4Tables adds the v4 schema tables
func (d *DB) CreateV4Tables() error {
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
	CREATE INDEX IF NOT EXISTS idx_project_aliases_project ON project_aliases(project_uid);
	
	CREATE TABLE IF NOT EXISTS tasks (
		task_uid TEXT PRIMARY KEY,
		project_uid TEXT NOT NULL,
		created_node TEXT NOT NULL,
		title TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		created_by TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_uid);
	
	CREATE TABLE IF NOT EXISTS task_numbers (
		project_uid TEXT NOT NULL,
		number INTEGER NOT NULL,
		task_uid TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_task_numbers_project_number ON task_numbers(project_uid, number);
	CREATE INDEX IF NOT EXISTS idx_task_numbers_task ON task_numbers(task_uid);
	`

	if _, err := d.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create v4 schema: %w", err)
	}

	return nil
}

// MigrateToV4 performs automatic migration from v1/v2 to v4
func (d *DB) MigrateToV4(dbPath string) error {
	// Step 1: Create lock file
	lockPath := filepath.Join(filepath.Dir(dbPath), v4MigrationLock)
	if _, err := os.Stat(lockPath); err == nil {
		return fmt.Errorf("migration already in progress (lock file exists)")
	}

	lockFile, err := os.Create(lockPath)
	if err != nil {
		return fmt.Errorf("failed to create lock file: %w", err)
	}
	lockFile.Close()
	defer os.Remove(lockPath)

	// Step 2: Create backup
	backupPath := dbPath + v4BackupSuffix
	if err := copyFile(dbPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Step 3: Add v4 tables
	if err := d.CreateV4Tables(); err != nil {
		return fmt.Errorf("failed to create v4 tables: %w", err)
	}

	// Step 4: Backfill v4 events from legacy data
	if err := d.backfillV4Events(); err != nil {
		return fmt.Errorf("failed to backfill events: %w", err)
	}

	// Step 5: Set version and config
	if err := d.SetDBVersion(v4SpecVersion); err != nil {
		return fmt.Errorf("failed to set version: %w", err)
	}

	_, err = d.db.Exec(`
		INSERT OR REPLACE INTO metadata (key, value) 
		VALUES ('remote_subdir', ?)
	`, v4SegmentSubdir)
	if err != nil {
		return fmt.Errorf("failed to set remote_subdir: %w", err)
	}

	return nil
}

// backfillV4Events creates v4 events from legacy v1/v2 data
func (d *DB) backfillV4Events() error {
	// Get current node ID
	nodeID, err := d.GetOrCreateNodeID()
	if err != nil {
		return fmt.Errorf("failed to get node ID: %w", err)
	}

	// Get current username
	actor, err := getCurrentUser()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	// Migrate prefixes to projects
	if err := d.migratePrefixesToProjects(nodeID, actor); err != nil {
		return fmt.Errorf("failed to migrate prefixes: %w", err)
	}

	// Migrate tasks
	if err := d.migrateTasksToV4(nodeID, actor); err != nil {
		return fmt.Errorf("failed to migrate tasks: %w", err)
	}

	return nil
}

// migratePrefixesToProjects converts prefix.created events to project.created + project.alias.add
func (d *DB) migratePrefixesToProjects(nodeID string, actor string) error {
	// Query all prefixes
	rows, err := d.db.Query(`
		SELECT prefix, node, description, created_at, created_by 
		FROM prefixes 
		WHERE removed = 0
		ORDER BY created_at
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	projectMap := make(map[string]string) // prefix -> project_uid

	for rows.Next() {
		var prefix, node, description, createdBy string
		var createdAt int64

		if err := rows.Scan(&prefix, &node, &description, &createdAt, &createdBy); err != nil {
			return err
		}

		// Check if we already created a project for this prefix
		projectUID, exists := projectMap[prefix]
		if !exists {
			// Create new project
			projectUID = string(NewProjectUID())
			projectMap[prefix] = projectUID

			// Emit project.created event
			if err := d.emitProjectCreatedEvent(projectUID, "local", prefix, description, createdBy, createdAt); err != nil {
				return err
			}
		}

		// Emit project.alias.add event for this node
		if err := d.emitProjectAliasAddEvent(projectUID, prefix, node, createdBy, createdAt); err != nil {
			return err
		}
	}

	return rows.Err()
}

// migrateTasksToV4 converts legacy task records to v4 task.created + task.number.set events
func (d *DB) migrateTasksToV4(nodeID string, actor string) error {
	// We need to reconstruct tasks from events
	// Query all task.created events
	rows, err := d.db.Query(`
		SELECT id, ts, created_at, actor, payload 
		FROM events 
		WHERE kind = 'task.created'
		ORDER BY ts, id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Track prefix -> project_uid mapping from project_aliases
	prefixToProject := make(map[string]string)
	aliasRows, err := d.db.Query(`SELECT project_uid, alias FROM project_aliases WHERE node = ?`, nodeID)
	if err != nil {
		return err
	}
	for aliasRows.Next() {
		var projectUID, alias string
		if err := aliasRows.Scan(&projectUID, &alias); err != nil {
			aliasRows.Close()
			return err
		}
		prefixToProject[alias] = projectUID
	}
	aliasRows.Close()

	for rows.Next() {
		var eventID string
		var ts, createdAt int64
		var eventActor string
		var payload []byte

		if err := rows.Scan(&eventID, &ts, &createdAt, &eventActor, &payload); err != nil {
			return err
		}

		// Parse legacy task.created payload
		var legacyPayload TaskCreatedPayload
		if err := json.Unmarshal(payload, &legacyPayload); err != nil {
			return fmt.Errorf("failed to parse task.created payload: %w", err)
		}

		// Extract prefix and number from task_id (format: prefix-number-node)
		prefix, number, _, err := parseTaskID(legacyPayload.TaskID)
		if err != nil {
			return fmt.Errorf("failed to parse task ID %s: %w", legacyPayload.TaskID, err)
		}

		// Find corresponding project
		projectUID, ok := prefixToProject[prefix]
		if !ok {
			return fmt.Errorf("no project found for prefix %s", prefix)
		}

		// The task_uuid from v1/v2 becomes the task_uid in v4
		taskUID := legacyPayload.TaskUUID
		if taskUID == "" {
			// Generate new task UID if not present
			taskUID = string(NewTaskUID())
		}

		// Emit task.created (v4) event
		if err := d.emitTaskCreatedV4Event(taskUID, projectUID, number, nodeID, legacyPayload.Title, legacyPayload.CreatedBy, createdAt); err != nil {
			return err
		}

		// Emit task.number.set event
		if err := d.emitTaskNumberSetEvent(taskUID, projectUID, number, "migration", createdAt); err != nil {
			return err
		}
	}

	return rows.Err()
}

// Helper functions to emit v4 events during migration

func (d *DB) emitProjectCreatedEvent(projectUID, typ, name, description, createdBy string, createdAt int64) error {
	// Create and insert the event
	// For now, we'll add to events table and project the data into projects table
	_, err := d.db.Exec(`
		INSERT INTO projects (project_uid, type, name, description, created_at, created_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, projectUID, typ, name, description, createdAt, createdBy)
	return err
}

func (d *DB) emitProjectAliasAddEvent(projectUID, alias, node, addedBy string, createdAt int64) error {
	_, err := d.db.Exec(`
		INSERT OR REPLACE INTO project_aliases (project_uid, alias, node, added_by)
		VALUES (?, ?, ?, ?)
	`, projectUID, alias, node, addedBy)
	return err
}

func (d *DB) emitTaskCreatedV4Event(taskUID, projectUID string, proposedNumber int64, createdNode, title, createdBy string, createdAt int64) error {
	_, err := d.db.Exec(`
		INSERT OR REPLACE INTO tasks (task_uid, project_uid, created_node, title, created_at, created_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, taskUID, projectUID, createdNode, title, createdAt, createdBy)
	return err
}

func (d *DB) emitTaskNumberSetEvent(taskUID, projectUID string, number int64, reason string, createdAt int64) error {
	_, err := d.db.Exec(`
		INSERT INTO task_numbers (project_uid, number, task_uid)
		VALUES (?, ?, ?)
	`, projectUID, number, taskUID)
	return err
}

// RollbackV4 restores the v3 backup
func RollbackV4(dbPath string) error {
	backupPath := dbPath + v4BackupSuffix

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup found at %s", backupPath)
	}

	// Close any open connections to the database
	// (caller should handle this)

	// Restore backup
	if err := os.Remove(dbPath); err != nil {
		return fmt.Errorf("failed to remove current database: %w", err)
	}

	if err := copyFile(backupPath, dbPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// parseTaskID extracts prefix, number, and node from a task ID
// Format: prefix-number-node
func parseTaskID(taskID string) (prefix string, number int64, node string, err error) {
	parts := strings.Split(taskID, "-")
	if len(parts) != 3 {
		return "", 0, "", fmt.Errorf("invalid task ID format: expected prefix-number-node")
	}

	prefix = parts[0]
	number, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid number in task ID: %w", err)
	}
	node = parts[2]

	return prefix, number, node, nil
}

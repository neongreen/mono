package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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

type v4MigrationContext struct {
	db              *DB
	nodeID          string
	actor           string
	prefixToProject map[string]string
	legacyTaskUIDs  map[string]string
	legacyTaskIDs   map[string]string
	taskProjects    map[string]string
}

type migrationEventHandler func(*v4MigrationContext, Event) error

var v4MigrationHandlers = [...]migrationEventHandler{
	eventKindProjectCreatedIndex:     migrationSkip,
	eventKindProjectAliasAddIndex:    migrationSkip,
	eventKindProjectAliasRemoveIndex: migrationSkip,
	eventKindTaskCreatedIndex:        migrateLegacyTaskCreated,
	eventKindTaskNumberSetIndex:      migrationSkip,
	eventKindTaskRelocateIndex:       migrationSkip,
	eventKindTaskStatusSetIndex:      migrateTaskStatusSet,
	eventKindTaskNoteAddIndex:        migrateTaskNoteAdd,
	eventKindTaskTitleSetIndex:       migrateTaskTitleSet,
	eventKindRelationAddIndex:        migrateRelationAdd,
	eventKindRelationRemoveIndex:     migrateRelationRemove,
	eventKindRelationNoteIndex:       migrateRelationNote,
	eventKindPrefixCreatedIndex:      migrationSkip,
	eventKindPrefixRemovedIndex:      migrationSkip,
	eventKindTaskReprefixLegacyIndex: migrateTaskReprefix,
	eventKindTaskAliasAddedIndex:     migrateTaskAliasAdded,
}

var (
	_ [int(eventKindCount) - len(v4MigrationHandlers)]struct{}
	_ [len(v4MigrationHandlers) - int(eventKindCount)]struct{}
)

func migrationHandlerForKind(kind EventKind) (migrationEventHandler, bool) {
	idx, ok := eventKindIndexOf(kind)
	if !ok {
		return nil, false
	}
	return v4MigrationHandlers[idx], true
}

func migrationUnsupported(kind EventKind) migrationEventHandler {
	return func(_ *v4MigrationContext, _ Event) error {
		return fmt.Errorf("v4 migration: handler for %s not implemented", kind)
	}
}

func migrationSkip(_ *v4MigrationContext, _ Event) error {
	return nil
}

func (ctx *v4MigrationContext) registerTask(taskUID string, legacyUUID string, ids ...string) {
	if ctx.legacyTaskUIDs == nil {
		ctx.legacyTaskUIDs = make(map[string]string)
	}
	if ctx.legacyTaskIDs == nil {
		ctx.legacyTaskIDs = make(map[string]string)
	}

	if taskUID != "" {
		ctx.legacyTaskUIDs[taskUID] = taskUID
	}

	if legacyUUID != "" {
		ctx.legacyTaskUIDs[legacyUUID] = taskUID
	}

	for _, id := range ids {
		if id == "" {
			continue
		}
		ctx.legacyTaskIDs[id] = taskUID
	}
}

func (ctx *v4MigrationContext) resolveTaskUID(legacyUUID string, taskID string) (string, error) {
	if legacyUUID != "" {
		if uid, ok := ctx.legacyTaskUIDs[legacyUUID]; ok {
			return uid, nil
		}
		if strings.HasPrefix(legacyUUID, "tsk_") {
			// Already a v4 task UID
			return legacyUUID, nil
		}
		// Check if legacyUUID is actually a task ID (e.g., "tk-30-wiWhKW" in the uuid field)
		// This happens when old events stored full task IDs in the task_uuid field
		if uid, ok := ctx.legacyTaskIDs[legacyUUID]; ok {
			return uid, nil
		}
	}

	if taskID != "" {
		if uid, ok := ctx.legacyTaskIDs[taskID]; ok {
			return uid, nil
		}
	}

	// Task not found in cache, try looking it up in the events table
	// This can happen if a task.status.set or other event comes before task.created
	// in the event ordering during migration
	if taskID != "" {
		var payload []byte
		err := ctx.db.db.QueryRow(`
			SELECT payload
			FROM events
			WHERE kind = 'task.created'
			AND json_extract(payload, '$.task_id') = ?
			LIMIT 1
		`, taskID).Scan(&payload)

		if err == nil {
			var taskCreated TaskCreatedPayload
			if err := json.Unmarshal(payload, &taskCreated); err == nil {
				// Found the task, register it and return the UUID
				taskUID := taskCreated.TaskUUID
				if taskUID == "" {
					taskUID = string(NewTaskUID())
				}
				// Register with the original legacy UUID (even if empty) for consistency
				ctx.registerTask(taskUID, taskCreated.TaskUUID, taskCreated.TaskID)
				return taskUID, nil
			}
		}
	}

	if legacyUUID != "" {
		var payload []byte
		err := ctx.db.db.QueryRow(`
			SELECT payload
			FROM events
			WHERE kind = 'task.created'
			AND json_extract(payload, '$.task_uuid') = ?
			LIMIT 1
		`, legacyUUID).Scan(&payload)

		if err == nil {
			var taskCreated TaskCreatedPayload
			if err := json.Unmarshal(payload, &taskCreated); err == nil {
				// Found the task, register it and return the UUID
				taskUID := taskCreated.TaskUUID
				if taskUID == "" {
					taskUID = string(NewTaskUID())
				}
				// Register with the original legacy UUID (even if empty) for consistency
				ctx.registerTask(taskUID, taskCreated.TaskUUID, taskCreated.TaskID)
				return taskUID, nil
			}
		}

		// Also try searching by task_id in case legacyUUID is actually a task ID
		// This can happen when the TaskUUID field in old events contained task IDs
		err = ctx.db.db.QueryRow(`
			SELECT payload
			FROM events
			WHERE kind = 'task.created'
			AND json_extract(payload, '$.task_id') = ?
			LIMIT 1
		`, legacyUUID).Scan(&payload)

		if err == nil {
			var taskCreated TaskCreatedPayload
			if err := json.Unmarshal(payload, &taskCreated); err == nil {
				// Found the task, register it and return the UUID
				taskUID := taskCreated.TaskUUID
				if taskUID == "" {
					taskUID = string(NewTaskUID())
				}
				// Register with the original legacy UUID (even if empty) for consistency
				ctx.registerTask(taskUID, taskCreated.TaskUUID, taskCreated.TaskID)
				return taskUID, nil
			}
		}
	}

	return "", fmt.Errorf("v4 migration: unknown task reference (uuid=%q id=%q)", legacyUUID, taskID)
}

func (ctx *v4MigrationContext) projectUIDForPrefix(prefix string) (string, error) {
	if projectUID, ok := ctx.prefixToProject[prefix]; ok {
		return projectUID, nil
	}

	var projectUID string
	err := ctx.db.db.QueryRow(`
		SELECT project_uid 
		FROM project_aliases 
		WHERE alias = ? 
		LIMIT 1
	`, prefix).Scan(&projectUID)
	if err == sql.ErrNoRows {
		// Prefix not found - create a project on-demand for this prefix
		// This can happen if:
		// 1. A task exists with a prefix that was removed (removed = 1)
		// 2. A task exists with a prefix that was never created
		projectUID = string(NewProjectUID())
		ctx.prefixToProject[prefix] = projectUID

		// Find the earliest task creation time for this prefix to use as a deterministic timestamp
		var earliestCreatedAtNano int64
		err := ctx.db.db.QueryRow(`
			SELECT MIN(created_at)
			FROM events
			WHERE kind = 'task.created'
			AND json_extract(payload, '$.task_id') LIKE ? || '-%'
		`, prefix).Scan(&earliestCreatedAtNano)

		createdAt := time.Now()
		if err == nil && earliestCreatedAtNano > 0 {
			// Use the earliest task creation time (created_at is stored in nanoseconds)
			createdAt = time.Unix(0, earliestCreatedAtNano)
		}

		// Create and insert project.created event
		payload := ProjectCreatedPayload{
			ProjectUID:  projectUID,
			Type:        "local",
			Name:        prefix,
			Description: "", // No description available for missing prefix
			CreatedBy:   ctx.actor,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return "", fmt.Errorf("v4 migration: failed to marshal project.created payload: %w", err)
		}

		event := Event{
			ID:        string(NewEventID()),
			TS:        0,
			CreatedAt: createdAt,
			Actor:     ctx.actor,
			Role:      "human",
			Kind:      string(EventKindProjectCreated),
			Payload:   payloadJSON,
		}

		if err := ctx.db.InsertEvent(event); err != nil {
			return "", fmt.Errorf("v4 migration: failed to insert project.created event: %w", err)
		}

		if err := ctx.db.ProjectProjectCreatedEvent(event); err != nil {
			return "", fmt.Errorf("v4 migration: failed to project project.created event: %w", err)
		}

		// Create and insert project.alias.add event (use same timestamp)
		aliasPayload := ProjectAliasAddPayload{
			ProjectUID: projectUID,
			Alias:      prefix,
			Node:       ctx.nodeID,
			AddedBy:    ctx.actor,
		}
		aliasPayloadJSON, err := json.Marshal(aliasPayload)
		if err != nil {
			return "", fmt.Errorf("v4 migration: failed to marshal project.alias.add payload: %w", err)
		}

		aliasEvent := Event{
			ID:        string(NewEventID()),
			TS:        0,
			CreatedAt: createdAt, // Use same timestamp as project.created
			Actor:     ctx.actor,
			Role:      "human",
			Kind:      string(EventKindProjectAliasAdd),
			Payload:   aliasPayloadJSON,
		}

		if err := ctx.db.InsertEvent(aliasEvent); err != nil {
			return "", fmt.Errorf("v4 migration: failed to insert project.alias.add event: %w", err)
		}

		if err := ctx.db.ProjectProjectAliasAddEvent(aliasEvent); err != nil {
			return "", fmt.Errorf("v4 migration: failed to project project.alias.add event: %w", err)
		}

		return projectUID, nil
	}
	if err != nil {
		return "", fmt.Errorf("v4 migration: failed to resolve prefix %s: %w", prefix, err)
	}

	ctx.prefixToProject[prefix] = projectUID
	return projectUID, nil
}

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
	prefixToProject, err := d.migratePrefixesToProjects(nodeID, actor)
	if err != nil {
		return fmt.Errorf("failed to migrate prefixes: %w", err)
	}

	// Migrate tasks
	if err := d.migrateTasksToV4(nodeID, actor, prefixToProject); err != nil {
		return fmt.Errorf("failed to migrate tasks: %w", err)
	}

	// Delete old v3 events (events with ts > 0)
	// The migrated v4 events all have ts=0, so any event with ts > 0 is a legacy event
	// that should be removed to avoid conflicts when the reducer runs
	result, err := d.db.Exec(`DELETE FROM events WHERE ts > 0`)
	if err != nil {
		return fmt.Errorf("failed to delete legacy v3 events: %w", err)
	}
	rowsDeleted, _ := result.RowsAffected()
	if rowsDeleted > 0 {
		fmt.Printf("Deleted %d legacy v3 events\n", rowsDeleted)
	}

	return nil
}

// migratePrefixesToProjects converts prefix.created events to project.created + project.alias.add
func (d *DB) migratePrefixesToProjects(nodeID string, actor string) (map[string]string, error) {
	// Query all prefixes
	rows, err := d.db.Query(`
		SELECT prefix, node, description, created_at, created_by 
		FROM prefixes 
		WHERE removed = 0
		ORDER BY created_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prefixToProject := make(map[string]string) // prefix -> project_uid

	for rows.Next() {
		var prefix, node, description, createdBy string
		var createdAt int64

		if err := rows.Scan(&prefix, &node, &description, &createdAt, &createdBy); err != nil {
			return nil, err
		}

		// Check if we already created a project for this prefix
		projectUID, exists := prefixToProject[prefix]
		if !exists {
			// Create new project
			projectUID = string(NewProjectUID())
			prefixToProject[prefix] = projectUID

			// Create and insert project.created event
			payload := ProjectCreatedPayload{
				ProjectUID:  projectUID,
				Type:        "local",
				Name:        prefix,
				Description: description,
				CreatedBy:   createdBy,
			}
			payloadJSON, err := json.Marshal(payload)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal project.created payload: %w", err)
			}

			event := Event{
				ID:        string(NewEventID()),
				TS:        0, // Will be set during ingest/replay
				CreatedAt: time.Unix(createdAt, 0),
				Actor:     createdBy,
				Role:      "human",
				Kind:      string(EventKindProjectCreated),
				Payload:   payloadJSON,
			}

			// Insert event
			if err := d.InsertEvent(event); err != nil {
				return nil, fmt.Errorf("failed to insert project.created event: %w", err)
			}

			// Project immediately
			if err := d.ProjectProjectCreatedEvent(event); err != nil {
				return nil, fmt.Errorf("failed to project project.created event: %w", err)
			}
		}

		// Create and insert project.alias.add event for this node
		aliasPayload := ProjectAliasAddPayload{
			ProjectUID: projectUID,
			Alias:      prefix,
			Node:       node,
			AddedBy:    createdBy,
		}
		aliasPayloadJSON, err := json.Marshal(aliasPayload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal project.alias.add payload: %w", err)
		}

		aliasEvent := Event{
			ID:        string(NewEventID()),
			TS:        0,
			CreatedAt: time.Unix(createdAt, 0),
			Actor:     createdBy,
			Role:      "human",
			Kind:      string(EventKindProjectAliasAdd),
			Payload:   aliasPayloadJSON,
		}

		// Insert event
		if err := d.InsertEvent(aliasEvent); err != nil {
			return nil, fmt.Errorf("failed to insert project.alias.add event: %w", err)
		}

		// Project immediately
		if err := d.ProjectProjectAliasAddEvent(aliasEvent); err != nil {
			return nil, fmt.Errorf("failed to project project.alias.add event: %w", err)
		}
	}

	return prefixToProject, rows.Err()
}

// migrateTasksToV4 converts legacy task records into v4 events
func (d *DB) migrateTasksToV4(nodeID string, actor string, prefixToProject map[string]string) error {
	ctx := &v4MigrationContext{
		db:              d,
		nodeID:          nodeID,
		actor:           actor,
		prefixToProject: prefixToProject,
		legacyTaskUIDs:  make(map[string]string),
		legacyTaskIDs:   make(map[string]string),
		taskProjects:    make(map[string]string),
	}

	// First pass: migrate all task.created events to establish task UID mappings
	rows, err := d.db.Query(`
		SELECT id, ts, created_at, actor, role, kind, payload
		FROM events
		WHERE ts > 0 AND kind = 'task.created'
		ORDER BY ts, id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var event Event
		var createdAtNano int64

		if err := rows.Scan(&event.ID, &event.TS, &createdAtNano, &event.Actor, &event.Role, &event.Kind, &event.Payload); err != nil {
			return err
		}

		event.CreatedAt = time.Unix(0, createdAtNano)

		handler, ok := migrationHandlerForKind(EventKind(event.Kind))
		if !ok {
			return fmt.Errorf("v4 migration: unknown event kind %s", event.Kind)
		}

		if handler == nil {
			return fmt.Errorf("v4 migration: nil handler for %s", event.Kind)
		}

		if err := handler(ctx, event); err != nil {
			return fmt.Errorf("v4 migration: handler for %s failed: %w", event.Kind, err)
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()

	// Second pass: migrate all other events
	rows, err = d.db.Query(`
		SELECT id, ts, created_at, actor, role, kind, payload
		FROM events
		WHERE ts > 0 AND kind != 'task.created'
		ORDER BY ts, id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var event Event
		var createdAtNano int64

		if err := rows.Scan(&event.ID, &event.TS, &createdAtNano, &event.Actor, &event.Role, &event.Kind, &event.Payload); err != nil {
			return err
		}

		event.CreatedAt = time.Unix(0, createdAtNano)

		handler, ok := migrationHandlerForKind(EventKind(event.Kind))
		if !ok {
			return fmt.Errorf("v4 migration: unknown event kind %s", event.Kind)
		}

		if handler == nil {
			return fmt.Errorf("v4 migration: nil handler for %s", event.Kind)
		}

		if err := handler(ctx, event); err != nil {
			return fmt.Errorf("v4 migration: handler for %s failed: %w", event.Kind, err)
		}
	}

	return rows.Err()
}

func migrateLegacyTaskCreated(ctx *v4MigrationContext, event Event) error {
	var legacyPayload TaskCreatedPayload
	if err := json.Unmarshal(event.Payload, &legacyPayload); err != nil {
		return fmt.Errorf("failed to parse task.created payload: %w", err)
	}

	prefix, number, node, err := parseTaskID(legacyPayload.TaskID)
	if err != nil {
		return fmt.Errorf("failed to parse task ID %s: %w", legacyPayload.TaskID, err)
	}

	projectUID, err := ctx.projectUIDForPrefix(prefix)
	if err != nil {
		return err
	}

	taskUID := ""
	if legacyPayload.TaskUUID != "" {
		if mapped, ok := ctx.legacyTaskUIDs[legacyPayload.TaskUUID]; ok {
			taskUID = mapped
		}
	}
	if taskUID == "" {
		taskUID = string(NewTaskUID())
	}

	ctx.registerTask(taskUID, legacyPayload.TaskUUID, legacyPayload.TaskID)
	ctx.taskProjects[taskUID] = projectUID

	taskPayload := TaskCreatedV4Payload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: number,
		CreatedNode:    node,
		Title:          legacyPayload.Title,
		CreatedBy:      legacyPayload.CreatedBy,
	}

	taskPayloadJSON, err := json.Marshal(taskPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal task.created payload: %w", err)
	}

	role := event.Role
	if role == "" {
		role = "human"
	}

	taskEvent := Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: event.CreatedAt,
		Actor:     legacyPayload.CreatedBy,
		Role:      role,
		Kind:      string(EventKindTaskCreated),
		Payload:   taskPayloadJSON,
	}

	if err := ctx.db.InsertEvent(taskEvent); err != nil {
		return fmt.Errorf("failed to insert task.created event: %w", err)
	}

	if err := ctx.db.ProjectTaskCreatedV4Event(taskEvent); err != nil {
		return fmt.Errorf("failed to project task.created event: %w", err)
	}

	numberPayload := TaskNumberSetPayload{
		TaskUID:    taskUID,
		ProjectUID: projectUID,
		Number:     number,
		Reason:     "migration",
	}

	numberPayloadJSON, err := json.Marshal(numberPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal task.number.set payload: %w", err)
	}

	numberEvent := Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: event.CreatedAt,
		Actor:     legacyPayload.CreatedBy,
		Role:      role,
		Kind:      string(EventKindTaskNumberSet),
		Payload:   numberPayloadJSON,
	}

	if err := ctx.db.InsertEvent(numberEvent); err != nil {
		return fmt.Errorf("failed to insert task.number.set event: %w", err)
	}

	if err := ctx.db.ProjectTaskNumberSetEvent(numberEvent); err != nil {
		return fmt.Errorf("failed to project task.number.set event: %w", err)
	}

	return nil
}

func migrateTaskStatusSet(ctx *v4MigrationContext, event Event) error {
	var payload TaskStatusSetPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse task.status.set payload: %w", err)
	}

	taskUID, err := ctx.resolveTaskUID(payload.TaskUUID, payload.TaskID)
	if err != nil {
		return err
	}

	ctx.registerTask(taskUID, "", payload.TaskID)

	role := payload.Role
	if role == "" {
		role = event.Role
	}
	if role == "" {
		role = "human"
	}

	newPayload := TaskStatusSetPayload{
		TaskUUID: taskUID,
		TaskID:   payload.TaskID,
		Axis:     payload.Axis,
		State:    payload.State,
		Role:     role,
	}

	payloadJSON, err := json.Marshal(newPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal task.status.set payload: %w", err)
	}

	newEvent := Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      role,
		Kind:      string(EventKindTaskStatusSet),
		Payload:   payloadJSON,
	}

	if err := ctx.db.InsertEvent(newEvent); err != nil {
		return fmt.Errorf("failed to insert migrated task.status.set event: %w", err)
	}

	return nil
}

func migrateTaskNoteAdd(ctx *v4MigrationContext, event Event) error {
	var payload TaskNoteAddPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse task.note.add payload: %w", err)
	}

	taskUID, err := ctx.resolveTaskUID(payload.TaskUUID, payload.TaskID)
	if err != nil {
		return err
	}

	ctx.registerTask(taskUID, "", payload.TaskID)

	newPayload := TaskNoteAddPayload{
		TaskUUID: taskUID,
		TaskID:   payload.TaskID,
		Markdown: payload.Markdown,
	}

	payloadJSON, err := json.Marshal(newPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal task.note.add payload: %w", err)
	}

	newEvent := Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindTaskNoteAdd),
		Payload:   payloadJSON,
	}

	if err := ctx.db.InsertEvent(newEvent); err != nil {
		return fmt.Errorf("failed to insert migrated task.note.add event: %w", err)
	}

	return nil
}

func migrateTaskTitleSet(ctx *v4MigrationContext, event Event) error {
	var payload TaskTitleSetPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse task.title.set payload: %w", err)
	}

	taskUID, err := ctx.resolveTaskUID(payload.TaskUID, "")
	if err != nil {
		return err
	}

	newPayload := TaskTitleSetPayload{
		TaskUID: taskUID,
		Title:   payload.Title,
	}

	payloadJSON, err := json.Marshal(newPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal task.title.set payload: %w", err)
	}

	newEvent := Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindTaskTitleSet),
		Payload:   payloadJSON,
	}

	if err := ctx.db.InsertEvent(newEvent); err != nil {
		return fmt.Errorf("failed to insert migrated task.title.set event: %w", err)
	}

	if err := ctx.db.ProjectTaskTitleSetEvent(newEvent); err != nil {
		return fmt.Errorf("failed to project migrated task.title.set event: %w", err)
	}

	return nil
}

func migrateTaskAliasAdded(ctx *v4MigrationContext, event Event) error {
	var payload TaskAliasAddedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse task.alias.added payload: %w", err)
	}

	taskUID, err := ctx.resolveTaskUID(payload.TaskUUID, "")
	if err != nil {
		return err
	}

	ctx.registerTask(taskUID, "", payload.AliasID)

	newPayload := TaskAliasAddedPayload{
		TaskUUID: taskUID,
		AliasID:  payload.AliasID,
	}

	payloadJSON, err := json.Marshal(newPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal task.alias.added payload: %w", err)
	}

	newEvent := Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindTaskAliasAdded),
		Payload:   payloadJSON,
	}

	if err := ctx.db.InsertEvent(newEvent); err != nil {
		return fmt.Errorf("failed to insert migrated task.alias.added event: %w", err)
	}

	return nil
}

func migrateRelationAdd(ctx *v4MigrationContext, event Event) error {
	var payload RelationAddPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse relation.add payload: %w", err)
	}

	srcUID, err := ctx.resolveTaskUID(payload.Src, "")
	if err != nil {
		return fmt.Errorf("failed to resolve relation.add src %s: %w", payload.Src, err)
	}
	dstUID, err := ctx.resolveTaskUID(payload.Dst, "")
	if err != nil {
		return fmt.Errorf("failed to resolve relation.add dst %s: %w", payload.Dst, err)
	}

	newPayload := RelationAddPayload{
		Src:  srcUID,
		Type: payload.Type,
		Dst:  dstUID,
		Note: payload.Note,
	}

	payloadJSON, err := json.Marshal(newPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal relation.add payload: %w", err)
	}

	newEvent := Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindRelationAdd),
		Payload:   payloadJSON,
	}

	if err := ctx.db.InsertEvent(newEvent); err != nil {
		return fmt.Errorf("failed to insert migrated relation.add event: %w", err)
	}

	return nil
}

func migrateRelationRemove(ctx *v4MigrationContext, event Event) error {
	var payload RelationRemovePayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse relation.remove payload: %w", err)
	}

	srcUID, err := ctx.resolveTaskUID(payload.Src, "")
	if err != nil {
		return fmt.Errorf("failed to resolve relation.remove src %s: %w", payload.Src, err)
	}
	dstUID, err := ctx.resolveTaskUID(payload.Dst, "")
	if err != nil {
		return fmt.Errorf("failed to resolve relation.remove dst %s: %w", payload.Dst, err)
	}

	newPayload := RelationRemovePayload{
		Src:  srcUID,
		Type: payload.Type,
		Dst:  dstUID,
	}

	payloadJSON, err := json.Marshal(newPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal relation.remove payload: %w", err)
	}

	newEvent := Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindRelationRemove),
		Payload:   payloadJSON,
	}

	if err := ctx.db.InsertEvent(newEvent); err != nil {
		return fmt.Errorf("failed to insert migrated relation.remove event: %w", err)
	}

	return nil
}

func migrateRelationNote(ctx *v4MigrationContext, event Event) error {
	var payload RelationNotePayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse relation.note payload: %w", err)
	}

	srcUID, err := ctx.resolveTaskUID(payload.Src, "")
	if err != nil {
		return fmt.Errorf("failed to resolve relation.note src %s: %w", payload.Src, err)
	}
	dstUID, err := ctx.resolveTaskUID(payload.Dst, "")
	if err != nil {
		return fmt.Errorf("failed to resolve relation.note dst %s: %w", payload.Dst, err)
	}

	newPayload := RelationNotePayload{
		Src:      srcUID,
		Type:     payload.Type,
		Dst:      dstUID,
		Markdown: payload.Markdown,
	}

	payloadJSON, err := json.Marshal(newPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal relation.note payload: %w", err)
	}

	newEvent := Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindRelationNote),
		Payload:   payloadJSON,
	}

	if err := ctx.db.InsertEvent(newEvent); err != nil {
		return fmt.Errorf("failed to insert migrated relation.note event: %w", err)
	}

	return nil
}

func migrateTaskReprefix(ctx *v4MigrationContext, event Event) error {
	var payload TaskReprefixPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("failed to parse task.reprefix payload: %w", err)
	}

	taskUID, err := ctx.resolveTaskUID(payload.TaskUUID, "")
	if err != nil {
		return err
	}

	fromProject := ctx.taskProjects[taskUID]
	if fromProject == "" {
		fromProject, err = ctx.projectUIDForPrefix(payload.OldPrefix)
		if err != nil {
			return err
		}
	}

	toProject, err := ctx.projectUIDForPrefix(payload.NewPrefix)
	if err != nil {
		return err
	}

	newPayload := TaskRelocatePayload{
		TaskUID:        taskUID,
		FromProjectUID: fromProject,
		ToProjectUID:   toProject,
		NumberPolicy: NumberPolicyPayload{
			Mode:   "force",
			Number: payload.NewNumber,
		},
	}

	payloadJSON, err := json.Marshal(newPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal task.relocate payload: %w", err)
	}

	newEvent := Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindTaskRelocate),
		Payload:   payloadJSON,
	}

	if err := ctx.db.InsertEvent(newEvent); err != nil {
		return fmt.Errorf("failed to insert migrated task.relocate event: %w", err)
	}

	if err := ctx.db.ProjectTaskRelocateEvent(newEvent); err != nil {
		return fmt.Errorf("failed to project migrated task.relocate event: %w", err)
	}

	ctx.taskProjects[taskUID] = toProject

	newTaskID := fmt.Sprintf("%s-%d-%s", payload.NewPrefix, payload.NewNumber, payload.OldNode)
	ctx.registerTask(taskUID, "", newTaskID)

	return nil
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

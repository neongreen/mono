package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	dbFileName = "tk.db"
)

// DB wraps a SQLite database for tk events
type DB struct {
	db            *sql.DB
	reducerCache  *Reducer // Cached reducer built from all events
	reducerConfig *Config  // Config used to build cached reducer
}

// OpenDB opens or creates a tk database at the given path
func OpenDB(path string) (*DB, error) {
	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Disable foreign keys (as per spec)
	if _, err := db.Exec("PRAGMA foreign_keys=OFF"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to disable foreign keys: %w", err)
	}

	return &DB{db: db}, nil
}

// InitDB creates the events table and metadata table
func (d *DB) InitDB() error {
	schema := `
	CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		ts INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		actor TEXT NOT NULL,
		role TEXT NOT NULL,
		kind TEXT NOT NULL,
		payload JSON NOT NULL,
		ctx JSON,
		repo_uuid TEXT,
		branch TEXT,
		commit_sha TEXT,
		jj_op_id TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_events_ts ON events(ts, id);
	CREATE INDEX IF NOT EXISTS idx_events_kind ON events(kind);
	
	CREATE TABLE IF NOT EXISTS metadata (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS task_counter (
		last_id INTEGER NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS event_counter (
		last_id INTEGER NOT NULL
	);
	
	CREATE TABLE IF NOT EXISTS event_id_map (
		rowid INTEGER PRIMARY KEY,
		event_id TEXT UNIQUE NOT NULL
	);
	
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
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Initialize task counter if it doesn't exist (legacy support)
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM task_counter").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check task counter: %w", err)
	}
	if count == 0 {
		if _, err := d.db.Exec("INSERT INTO task_counter (last_id) VALUES (0)"); err != nil {
			return fmt.Errorf("failed to initialize task counter: %w", err)
		}
	}

	// Initialize event counter if it doesn't exist
	err = d.db.QueryRow("SELECT COUNT(*) FROM event_counter").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check event counter: %w", err)
	}
	if count == 0 {
		if _, err := d.db.Exec("INSERT INTO event_counter (last_id) VALUES (0)"); err != nil {
			return fmt.Errorf("failed to initialize event counter: %w", err)
		}
	}

	return nil
}

// InsertEvent adds an event to the database
func (d *DB) InsertEvent(e Event) error {
	query := `
		INSERT INTO events (id, ts, created_at, actor, role, kind, payload, ctx, repo_uuid, branch, commit_sha, jj_op_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(query, e.ID, e.TS, e.CreatedAt.UnixNano(), e.Actor, e.Role, e.Kind, e.Payload, e.Ctx, e.RepoUUID, e.Branch, e.Commit, e.JJOpID)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	// Invalidate reducer cache when a new event is added
	d.reducerCache = nil
	d.reducerConfig = nil

	return nil
}

// GetEvents retrieves all events in chronological order
func (d *DB) GetEvents() ([]Event, error) {
	query := `SELECT id, ts, created_at, actor, role, kind, payload, ctx, repo_uuid, branch, commit_sha, jj_op_id
	          FROM events ORDER BY created_at, id`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var ctx, repoUUID, branch, commit, jjOpID sql.NullString
		var createdAtNano int64

		err := rows.Scan(&e.ID, &e.TS, &createdAtNano, &e.Actor, &e.Role, &e.Kind, &e.Payload, &ctx, &repoUUID, &branch, &commit, &jjOpID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		e.CreatedAt = time.Unix(0, createdAtNano)
		if ctx.Valid {
			e.Ctx = json.RawMessage(ctx.String)
		}
		if repoUUID.Valid {
			e.RepoUUID = repoUUID.String
		}
		if branch.Valid {
			e.Branch = branch.String
		}
		if commit.Valid {
			e.Commit = commit.String
		}
		if jjOpID.Valid {
			e.JJOpID = jjOpID.String
		}

		events = append(events, e)
	}

	return events, rows.Err()
}

// GetEventsByTaskID retrieves events for a specific task
func (d *DB) GetEventsByTaskID(taskID string) ([]Event, error) {
	query := `
		SELECT id, ts, created_at, actor, role, kind, payload, ctx, repo_uuid, branch, commit_sha, jj_op_id
		FROM events
		WHERE json_extract(payload, '$.task_id') = ?
		ORDER BY created_at, id
	`

	rows, err := d.db.Query(query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events for task %s: %w", taskID, err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var ctx, repoUUID, branch, commit, jjOpID sql.NullString
		var createdAtNano int64

		err := rows.Scan(&e.ID, &e.TS, &createdAtNano, &e.Actor, &e.Role, &e.Kind, &e.Payload, &ctx, &repoUUID, &branch, &commit, &jjOpID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		e.CreatedAt = time.Unix(0, createdAtNano)
		if ctx.Valid {
			e.Ctx = json.RawMessage(ctx.String)
		}
		if repoUUID.Valid {
			e.RepoUUID = repoUUID.String
		}
		if branch.Valid {
			e.Branch = branch.String
		}
		if commit.Valid {
			e.Commit = commit.String
		}
		if jjOpID.Valid {
			e.JJOpID = jjOpID.String
		}

		events = append(events, e)
	}

	return events, rows.Err()
}

// GetEventsByTaskUUID retrieves events for a specific task UUID
func (d *DB) GetEventsByTaskUUID(taskUUID string) ([]Event, error) {
	query := `
		SELECT id, ts, created_at, actor, role, kind, payload, ctx, repo_uuid, branch, commit_sha, jj_op_id
		FROM events
		WHERE json_extract(payload, '$.task_uuid') = ?
		   OR json_extract(payload, '$.task_id') = ?
		ORDER BY created_at, id
	`

	rows, err := d.db.Query(query, taskUUID, taskUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events for task UUID %s: %w", taskUUID, err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var ctx, repoUUID, branch, commit, jjOpID sql.NullString
		var createdAtNano int64

		err := rows.Scan(&e.ID, &e.TS, &createdAtNano, &e.Actor, &e.Role, &e.Kind, &e.Payload, &ctx, &repoUUID, &branch, &commit, &jjOpID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		e.CreatedAt = time.Unix(0, createdAtNano)
		if ctx.Valid {
			e.Ctx = json.RawMessage(ctx.String)
		}
		if repoUUID.Valid {
			e.RepoUUID = repoUUID.String
		}
		if branch.Valid {
			e.Branch = branch.String
		}
		if commit.Valid {
			e.Commit = commit.String
		}
		if jjOpID.Valid {
			e.JJOpID = jjOpID.String
		}

		events = append(events, e)
	}

	return events, rows.Err()
}

// GetDBVersion reads the version_major from metadata table (defaults to 4)
func (d *DB) GetDBVersion() (int, error) {
	var versionStr string
	err := d.db.QueryRow("SELECT value FROM metadata WHERE key = 'version_major'").Scan(&versionStr)
	if err == sql.ErrNoRows {
		return 4, nil // Default to v4 if not set
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get version: %w", err)
	}
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return 0, fmt.Errorf("invalid version in metadata: %w", err)
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

// ResolveTaskIDToUUID resolves a task reference to its UUID (legacy helper).
func (d *DB) ResolveTaskIDToUUID(taskID string) (string, error) {
	return ResolveTaskReference(d, taskID)
}

// Close closes the database
func (d *DB) Close() error {
	return d.db.Close()
}

// GetDBPath returns the database path in ~/.tk/ directory
func GetDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	tkDir := filepath.Join(home, ".tk")
	return filepath.Join(tkDir, dbFileName), nil
}

// DBExists checks if a database exists at the given path
func DBExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetOrCreateNodeID gets the node ID or creates one if it doesn't exist
func (d *DB) GetOrCreateNodeID() (string, error) {
	// Try to get existing node ID
	var nodeID string
	err := d.db.QueryRow("SELECT value FROM metadata WHERE key = 'node_id'").Scan(&nodeID)
	if err == nil {
		return nodeID, nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to query node ID: %w", err)
	}

	// Generate new node ID (6 random alphanumeric characters, mixed case)
	nodeID, err = generateNodeID(6)
	if err != nil {
		return "", fmt.Errorf("failed to generate node ID: %w", err)
	}

	// Store it
	_, err = d.db.Exec("INSERT INTO metadata (key, value) VALUES ('node_id', ?)", nodeID)
	if err != nil {
		return "", fmt.Errorf("failed to store node ID: %w", err)
	}

	return nodeID, nil
}

// RegenerateNodeID generates a new node ID and updates the metadata
func (d *DB) RegenerateNodeID() (string, error) {
	// Generate new node ID
	newNodeID, err := generateNodeID(6)
	if err != nil {
		return "", fmt.Errorf("failed to generate node ID: %w", err)
	}

	// Update the metadata
	_, err = d.db.Exec("UPDATE metadata SET value = ? WHERE key = 'node_id'", newNodeID)
	if err != nil {
		return "", fmt.Errorf("failed to update node ID: %w", err)
	}

	return newNodeID, nil
}

// GetNextLamportTS gets the next Lamport timestamp and increments the counter
func (d *DB) GetNextLamportTS() (int64, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current counter value
	var counter int64
	err = tx.QueryRow("SELECT value FROM metadata WHERE key = 'lamport_counter'").Scan(&counter)
	if err == sql.ErrNoRows {
		counter = 0
	} else if err != nil {
		return 0, fmt.Errorf("failed to query lamport counter: %w", err)
	}

	// Increment counter
	nextTS := counter + 1

	// Update or insert
	if counter == 0 {
		_, err = tx.Exec("INSERT INTO metadata (key, value) VALUES ('lamport_counter', ?)", nextTS)
	} else {
		_, err = tx.Exec("UPDATE metadata SET value = ? WHERE key = 'lamport_counter'", nextTS)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to update lamport counter: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nextTS, nil
}

// BumpLamport updates the lamport counter if the given value is higher
func (d *DB) BumpLamport(newValue int64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current counter value
	var counter int64
	err = tx.QueryRow("SELECT value FROM metadata WHERE key = 'lamport_counter'").Scan(&counter)
	if err == sql.ErrNoRows {
		counter = 0
	} else if err != nil {
		return fmt.Errorf("failed to query lamport counter: %w", err)
	}

	// Only update if new value is higher
	if newValue > counter {
		if counter == 0 {
			_, err = tx.Exec("INSERT INTO metadata (key, value) VALUES ('lamport_counter', ?)", newValue)
		} else {
			_, err = tx.Exec("UPDATE metadata SET value = ? WHERE key = 'lamport_counter'", newValue)
		}
		if err != nil {
			return fmt.Errorf("failed to update lamport counter: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetNextTaskNumber gets the next task number and increments the counter
func (d *DB) GetNextTaskNumber() (int64, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var lastID int64
	err = tx.QueryRow("SELECT last_id FROM task_counter").Scan(&lastID)
	if err != nil {
		return 0, fmt.Errorf("failed to get last task ID: %w", err)
	}

	nextID := lastID + 1
	_, err = tx.Exec("UPDATE task_counter SET last_id = ?", nextID)
	if err != nil {
		return 0, fmt.Errorf("failed to update task counter: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nextID, nil
}

// GetNextEventNumber gets the next event number and increments the counter
func (d *DB) GetNextEventNumber() (int64, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var lastID int64
	err = tx.QueryRow("SELECT last_id FROM event_counter").Scan(&lastID)
	if err != nil {
		return 0, fmt.Errorf("failed to get last event ID: %w", err)
	}

	nextID := lastID + 1
	_, err = tx.Exec("UPDATE event_counter SET last_id = ?", nextID)
	if err != nil {
		return 0, fmt.Errorf("failed to update event counter: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nextID, nil
}

// generateNodeID generates a random alphanumeric node ID with mixed case
func generateNodeID(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}

// ResolveTaskID resolves a short task ID to a full task ID
// Accepts formats: "1", "tk-1", "foo-2", "tk-1-abc123"
// Returns an error if the ID is ambiguous or doesn't exist
func (d *DB) ResolveTaskID(shortID string) (string, error) {
	// Check if it's already a full ID (contains at least 2 hyphens)
	hyphenCount := strings.Count(shortID, "-")
	if hyphenCount >= 2 {
		// Verify it exists
		var count int
		err := d.db.QueryRow(`
			SELECT COUNT(*)
			FROM events
			WHERE kind = 'task.created' AND json_extract(payload, '$.task_id') = ?
		`, shortID).Scan(&count)
		if err != nil {
			return "", fmt.Errorf("failed to query task ID: %w", err)
		}
		if count > 0 {
			return shortID, nil
		}
		return "", fmt.Errorf("task not found: %s", shortID)
	}

	// Try to match as a number or prefix-number format
	if _, err := strconv.Atoi(shortID); err == nil {
		// Just a number - use SQL to find matches efficiently
		query := `
			SELECT DISTINCT json_extract(payload, '$.task_id') as task_id
			FROM events
			WHERE kind = 'task.created'
			  AND json_extract(payload, '$.task_id') LIKE '%-' || ? || '-%'
		`

		rows, err := d.db.Query(query, shortID)
		if err != nil {
			return "", fmt.Errorf("failed to query task IDs: %w", err)
		}
		defer rows.Close()

		var matches []string
		for rows.Next() {
			var taskID string
			if err := rows.Scan(&taskID); err != nil {
				return "", fmt.Errorf("failed to scan task ID: %w", err)
			}
			matches = append(matches, taskID)
		}

		if len(matches) == 0 {
			return "", fmt.Errorf("task not found: %s", shortID)
		}

		// Group by prefix-number to detect ambiguity
		prefixNumberMap := make(map[string][]string)
		for _, taskID := range matches {
			parts := strings.Split(taskID, "-")
			if len(parts) >= 2 {
				prefixNumber := strings.Join(parts[:2], "-")
				prefixNumberMap[prefixNumber] = append(prefixNumberMap[prefixNumber], taskID)
			}
		}

		if len(prefixNumberMap) > 1 {
			prefixes := make([]string, 0, len(prefixNumberMap))
			for pn := range prefixNumberMap {
				prefixes = append(prefixes, pn)
			}
			return "", fmt.Errorf("ambiguous task ID %s (matches %v) — use <prefix>-%s instead", shortID, prefixes, shortID)
		}

		// Only one prefix-number combination
		if len(matches) == 1 {
			return matches[0], nil
		}

		// Multiple matches with same prefix-number but different nodes - this is ambiguous
		// Extract the prefix-number for error message
		var prefixNumber string
		for pn := range prefixNumberMap {
			prefixNumber = pn
			break
		}
		return "", fmt.Errorf("ambiguous task ID %s (multiple nodes created %s) — use full ID like %s", shortID, prefixNumber, matches[0])
	}

	// Format is prefix-number - use SQL to find matches
	query := `
		SELECT DISTINCT json_extract(payload, '$.task_id') as task_id
		FROM events
		WHERE kind = 'task.created'
		  AND json_extract(payload, '$.task_id') LIKE ? || '-%'
	`

	rows, err := d.db.Query(query, shortID)
	if err != nil {
		return "", fmt.Errorf("failed to query task IDs: %w", err)
	}
	defer rows.Close()

	var matches []string
	for rows.Next() {
		var taskID string
		if err := rows.Scan(&taskID); err != nil {
			return "", fmt.Errorf("failed to scan task ID: %w", err)
		}
		// Verify it actually starts with prefix-number-
		parts := strings.Split(taskID, "-")
		if len(parts) >= 2 {
			shortForm := strings.Join(parts[:2], "-")
			if shortForm == shortID {
				matches = append(matches, taskID)
			}
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("task not found: %s", shortID)
	}

	if len(matches) > 1 {
		return "", fmt.Errorf("ambiguous task ID %s (multiple nodes created %s) — use full ID like %s", shortID, shortID, matches[0])
	}

	return matches[0], nil
}

// GetAllTaskIDs returns all task IDs in the database
func (d *DB) GetAllTaskIDs() ([]string, error) {
	query := `SELECT DISTINCT task_uid FROM tasks ORDER BY created_at`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query task UIDs: %w", err)
	}
	defer rows.Close()

	var taskUIDs []string
	for rows.Next() {
		var taskUID string
		if err := rows.Scan(&taskUID); err != nil {
			return nil, fmt.Errorf("failed to scan task UID: %w", err)
		}
		taskUIDs = append(taskUIDs, taskUID)
	}
	return taskUIDs, rows.Err()
}

// GetTaskIDsByPrefixes returns task IDs filtered by prefix list
func (d *DB) GetTaskIDsByPrefixes(prefixes []string) ([]string, error) {
	if len(prefixes) == 0 {
		return d.GetAllTaskIDs()
	}

	// Get task UIDs by project aliases
	// First, get node ID
	nodeID, err := d.GetOrCreateNodeID()
	if err != nil {
		return nil, fmt.Errorf("failed to get node ID: %w", err)
	}

	// Build query to find tasks by project alias
	var conditions []string
	var args []interface{}
	for _, alias := range prefixes {
		conditions = append(conditions, "project_aliases.alias = ?")
		args = append(args, alias)
	}
	args = append(args, nodeID)

	query := `
		SELECT DISTINCT tasks.task_uid
		FROM tasks
		JOIN project_aliases ON tasks.project_uid = project_aliases.project_uid
		WHERE (` + strings.Join(conditions, " OR ") + `)
		  AND project_aliases.node = ?
		ORDER BY tasks.created_at
	`

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query task UIDs by prefix: %w", err)
	}
	defer rows.Close()

	var taskUIDs []string
	for rows.Next() {
		var taskUID string
		if err := rows.Scan(&taskUID); err != nil {
			return nil, fmt.Errorf("failed to scan task UID: %w", err)
		}
		taskUIDs = append(taskUIDs, taskUID)
	}
	return taskUIDs, rows.Err()
}

// FormatTaskID formats a task ID for display, hiding the suffix unless needed for disambiguation
func FormatTaskID(fullID string, allTaskIDs []string) string {
	// Extract parts: <prefix>-<number>-<suffix>
	parts := strings.Split(fullID, "-")
	if len(parts) < 3 {
		return fullID // Malformed ID, return as-is
	}

	shortForm := strings.Join(parts[:2], "-") // <prefix>-<number>

	// Check if any other task has the same short form but different suffix
	needsSuffix := false
	for _, otherID := range allTaskIDs {
		if otherID == fullID {
			continue
		}
		otherParts := strings.Split(otherID, "-")
		if len(otherParts) >= 2 {
			otherShortForm := strings.Join(otherParts[:2], "-")
			if otherShortForm == shortForm {
				needsSuffix = true
				break
			}
		}
	}

	if needsSuffix {
		return fullID
	}
	return shortForm
}

// Prefix represents a task prefix definition
type Prefix struct {
	Prefix      string
	Node        string
	Description string
	CreatedAt   time.Time
	CreatedBy   string
	Removed     bool
}


// GetCachedReducerWithConfig returns a cached reducer or builds a new one if needed.
// The cache is invalidated when new events are inserted.
// This significantly improves performance for operations that need to query task state.
//
// Note: The cache uses pointer identity for config comparison. This is safe because:
// - Each command typically loads config once and reuses the same instance
// - The DB instance is scoped to a single command execution
// - Cache is invalidated on any event insertion
// If config pointer doesn't match, we rebuild the reducer (safe but may miss some cache hits).
func (d *DB) GetCachedReducerWithConfig(config *Config) (*Reducer, error) {
	// Check if we have a valid cached reducer with the same config pointer
	// Using pointer comparison is conservative - we rebuild if in doubt
	if d.reducerCache != nil && d.reducerConfig == config {
		return d.reducerCache, nil
	}

	// Build a new reducer from all events
	events, err := d.GetEvents()
	if err != nil {
		return nil, err
	}

	reducer, err := BuildFromEventsWithConfig(events, config)
	if err != nil {
		return nil, err
	}

	// Cache the reducer
	d.reducerCache = reducer
	d.reducerConfig = config

	return reducer, nil
}

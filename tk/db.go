package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sort"
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
	db *sql.DB
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
	
	CREATE TABLE IF NOT EXISTS prefixes (
		prefix TEXT NOT NULL,
		node TEXT NOT NULL,
		description TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		created_by TEXT NOT NULL,
		PRIMARY KEY (prefix, node)
	);
	
	CREATE TABLE IF NOT EXISTS prefix_counters (
		prefix TEXT NOT NULL,
		node TEXT NOT NULL,
		last_id INTEGER NOT NULL,
		PRIMARY KEY (prefix, node)
	);
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

	// Create default "tk" prefix if no prefixes exist
	nodeID, err := d.GetOrCreateNodeID()
	if err != nil {
		return fmt.Errorf("failed to get node ID: %w", err)
	}

	var prefixCount int
	err = d.db.QueryRow("SELECT COUNT(*) FROM prefixes WHERE node = ?", nodeID).Scan(&prefixCount)
	if err != nil {
		return fmt.Errorf("failed to check prefixes: %w", err)
	}

	if prefixCount == 0 {
		// Migrate legacy counter to "tk" prefix
		var legacyCounter int64
		err = d.db.QueryRow("SELECT last_id FROM task_counter").Scan(&legacyCounter)
		if err != nil {
			return fmt.Errorf("failed to get legacy counter: %w", err)
		}

		// Check if there are existing tasks that would indicate a legacy DB
		var taskCount int
		err = d.db.QueryRow("SELECT COUNT(*) FROM events WHERE kind = 'task.created'").Scan(&taskCount)
		if err != nil {
			return fmt.Errorf("failed to check task count: %w", err)
		}

		description := "Default task prefix"
		if taskCount > 0 && legacyCounter > 0 {
			// This is a legacy DB with existing tasks, adjust the description
			description = "Imported from legacy (default)"
		}

		// Create default "tk" prefix
		_, err = d.db.Exec(
			"INSERT OR IGNORE INTO prefixes (prefix, node, description, created_at, created_by) VALUES (?, ?, ?, ?, ?)",
			"tk", nodeID, description, time.Now().UnixNano(), "system",
		)
		if err != nil {
			return fmt.Errorf("failed to create default prefix: %w", err)
		}

		// Initialize prefix counter with legacy value
		_, err = d.db.Exec(
			"INSERT OR IGNORE INTO prefix_counters (prefix, node, last_id) VALUES (?, ?, ?)",
			"tk", nodeID, legacyCounter,
		)
		if err != nil {
			return fmt.Errorf("failed to initialize prefix counter: %w", err)
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
	nodeID = generateNodeID(6)

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
	newNodeID := generateNodeID(6)

	// Update the metadata
	_, err := d.db.Exec("UPDATE metadata SET value = ? WHERE key = 'node_id'", newNodeID)
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
func generateNodeID(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
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
			return "", fmt.Errorf("ambiguous task ID %s (multiple prefixes match: %v), please specify prefix", shortID, prefixes)
		}

		// Only one prefix-number combination, so return any match (they differ only in node suffix)
		if len(matches) == 1 {
			return matches[0], nil
		}

		// Multiple matches with same prefix-number but different nodes - return first (or could error)
		return matches[0], nil
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
		return "", fmt.Errorf("ambiguous task ID %s, matches: %v", shortID, matches)
	}

	return matches[0], nil
}

// GetAllTaskIDs returns all task IDs in the database
func (d *DB) GetAllTaskIDs() ([]string, error) {
	query := `
		SELECT DISTINCT json_extract(payload, '$.task_id') as task_id
		FROM events
		WHERE kind = 'task.created'
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query task IDs: %w", err)
	}
	defer rows.Close()

	var taskIDs []string
	for rows.Next() {
		var taskID string
		if err := rows.Scan(&taskID); err != nil {
			return nil, fmt.Errorf("failed to scan task ID: %w", err)
		}
		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs, rows.Err()
}

// GetTaskIDsByPrefixes returns task IDs filtered by prefix list
func (d *DB) GetTaskIDsByPrefixes(prefixes []string) ([]string, error) {
	if len(prefixes) == 0 {
		return d.GetAllTaskIDs()
	}

	// Build SQL query with OR conditions for each prefix
	var conditions []string
	var args []interface{}
	for _, prefix := range prefixes {
		conditions = append(conditions, "json_extract(payload, '$.task_id') LIKE ?")
		args = append(args, prefix+"-%")
	}

	query := `
		SELECT DISTINCT json_extract(payload, '$.task_id') as task_id
		FROM events
		WHERE kind = 'task.created' AND (` + strings.Join(conditions, " OR ") + `)
	`

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query task IDs: %w", err)
	}
	defer rows.Close()

	var taskIDs []string
	for rows.Next() {
		var taskID string
		if err := rows.Scan(&taskID); err != nil {
			return nil, fmt.Errorf("failed to scan task ID: %w", err)
		}
		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs, rows.Err()
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
}

// CreatePrefix creates a new prefix and emits a prefix.created event
func (d *DB) CreatePrefix(prefix, description, createdBy string) error {
	// Normalize to lowercase first
	prefix = strings.ToLower(prefix)

	// Validate prefix format
	if err := ValidatePrefixName(prefix); err != nil {
		return err
	}

	nodeID, err := d.GetOrCreateNodeID()
	if err != nil {
		return fmt.Errorf("failed to get node ID: %w", err)
	}

	// Check if prefix already exists for this node
	var count int
	err = d.db.QueryRow("SELECT COUNT(*) FROM prefixes WHERE prefix = ? AND node = ?", prefix, nodeID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check prefix existence: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("prefix %q already exists for this node", prefix)
	}

	// Generate event ID
	eventID, err := GenerateEventID(d)
	if err != nil {
		return fmt.Errorf("failed to generate event ID: %w", err)
	}

	// Get next Lamport timestamp
	lamportTS, err := d.GetNextLamportTS()
	if err != nil {
		return fmt.Errorf("failed to get lamport timestamp: %w", err)
	}

	// Create prefix.created event
	payload := PrefixCreatedPayload{
		Prefix:      prefix,
		Description: description,
		CreatedBy:   createdBy,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	now := time.Now()
	event := Event{
		ID:        eventID,
		TS:        lamportTS,
		CreatedAt: now,
		Actor:     createdBy,
		Role:      "human",
		Kind:      "prefix.created",
		Payload:   payloadJSON,
	}

	// Insert event first (event-sourcing principle)
	if err := d.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert prefix.created event: %w", err)
	}

	// Project into prefixes table (idempotent)
	_, err = d.db.Exec(
		"INSERT OR IGNORE INTO prefixes (prefix, node, description, created_at, created_by) VALUES (?, ?, ?, ?, ?)",
		prefix, nodeID, description, now.UnixNano(), createdBy,
	)
	if err != nil {
		return fmt.Errorf("failed to create prefix: %w", err)
	}

	// Initialize counter for this prefix (idempotent)
	_, err = d.db.Exec(
		"INSERT OR IGNORE INTO prefix_counters (prefix, node, last_id) VALUES (?, ?, ?)",
		prefix, nodeID, 0,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize prefix counter: %w", err)
	}

	return nil
}

// GetPrefixes returns all prefixes for this node
func (d *DB) GetPrefixes() ([]Prefix, error) {
	nodeID, err := d.GetOrCreateNodeID()
	if err != nil {
		return nil, fmt.Errorf("failed to get node ID: %w", err)
	}

	query := `
		SELECT prefix, node, description, created_at, created_by
		FROM prefixes
		WHERE node = ?
		ORDER BY created_at
	`

	rows, err := d.db.Query(query, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query prefixes: %w", err)
	}
	defer rows.Close()

	var prefixes []Prefix
	for rows.Next() {
		var p Prefix
		var createdAtNano int64
		err := rows.Scan(&p.Prefix, &p.Node, &p.Description, &createdAtNano, &p.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to scan prefix: %w", err)
		}
		p.CreatedAt = time.Unix(0, createdAtNano)
		prefixes = append(prefixes, p)
	}

	return prefixes, rows.Err()
}

// GetAllPrefixes returns all prefixes from all nodes (event-backed)
func (d *DB) GetAllPrefixes() ([]Prefix, error) {
	// Get prefixes from the prefixes table (from prefix.created events)
	query := `
		SELECT prefix, node, description, created_at, created_by
		FROM prefixes
		ORDER BY created_at
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query prefixes: %w", err)
	}
	defer rows.Close()

	prefixMap := make(map[string]Prefix) // key: prefix-node
	for rows.Next() {
		var p Prefix
		var createdAtNano int64
		err := rows.Scan(&p.Prefix, &p.Node, &p.Description, &createdAtNano, &p.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to scan prefix: %w", err)
		}
		p.CreatedAt = time.Unix(0, createdAtNano)
		key := p.Prefix + "-" + p.Node
		prefixMap[key] = p
	}

	// Also derive prefixes from task.created events (for prefixes that don't have metadata)
	taskQuery := `
		SELECT DISTINCT substr(json_extract(payload, '$.task_id'), 1, instr(json_extract(payload, '$.task_id'), '-') - 1) as prefix,
		       json_extract(payload, '$.created_by') as created_by
		FROM events
		WHERE kind = 'task.created'
	`

	taskRows, err := d.db.Query(taskQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query task prefixes: %w", err)
	}
	defer taskRows.Close()

	for taskRows.Next() {
		var prefix, createdBy string
		if err := taskRows.Scan(&prefix, &createdBy); err != nil {
			return nil, fmt.Errorf("failed to scan task prefix: %w", err)
		}

		// Extract node from task ID to build the key
		// We need to query for a task with this prefix to get the node
		var taskID string
		err = d.db.QueryRow(`
			SELECT json_extract(payload, '$.task_id')
			FROM events
			WHERE kind = 'task.created' 
			  AND json_extract(payload, '$.task_id') LIKE ?
			LIMIT 1
		`, prefix+"-%").Scan(&taskID)
		if err != nil {
			continue
		}

		// Extract node from task ID (format: prefix-number-node)
		parts := strings.Split(taskID, "-")
		if len(parts) < 3 {
			continue
		}
		node := parts[2]
		key := prefix + "-" + node

		// Only add if not already in map (from prefix.created events)
		if _, exists := prefixMap[key]; !exists {
			prefixMap[key] = Prefix{
				Prefix:      prefix,
				Node:        node,
				Description: "(discovered from tasks, no metadata)",
				CreatedBy:   createdBy,
				CreatedAt:   time.Time{}, // Unknown creation time
			}
		}
	}

	// Convert map to slice
	var prefixes []Prefix
	for _, p := range prefixMap {
		prefixes = append(prefixes, p)
	}

	// Sort by creation time
	sort.Slice(prefixes, func(i, j int) bool {
		// Put items with unknown time at the end
		if prefixes[i].CreatedAt.IsZero() && !prefixes[j].CreatedAt.IsZero() {
			return false
		}
		if !prefixes[i].CreatedAt.IsZero() && prefixes[j].CreatedAt.IsZero() {
			return true
		}
		return prefixes[i].CreatedAt.Before(prefixes[j].CreatedAt)
	})

	return prefixes, nil
}

// PrefixExists checks if a prefix exists for this node
func (d *DB) PrefixExists(prefix string) (bool, error) {
	nodeID, err := d.GetOrCreateNodeID()
	if err != nil {
		return false, fmt.Errorf("failed to get node ID: %w", err)
	}

	var count int
	err = d.db.QueryRow("SELECT COUNT(*) FROM prefixes WHERE prefix = ? AND node = ?", prefix, nodeID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check prefix existence: %w", err)
	}

	return count > 0, nil
}

// GetNextTaskNumberForPrefix gets the next task number for a specific prefix and increments the counter
func (d *DB) GetNextTaskNumberForPrefix(prefix string) (int64, error) {
	nodeID, err := d.GetOrCreateNodeID()
	if err != nil {
		return 0, fmt.Errorf("failed to get node ID: %w", err)
	}

	// Check if prefix exists
	exists, err := d.PrefixExists(prefix)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, fmt.Errorf("prefix %q does not exist for this node", prefix)
	}

	tx, err := d.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var lastID int64
	err = tx.QueryRow("SELECT last_id FROM prefix_counters WHERE prefix = ? AND node = ?", prefix, nodeID).Scan(&lastID)
	if err != nil {
		return 0, fmt.Errorf("failed to get last task ID for prefix %q: %w", prefix, err)
	}

	nextID := lastID + 1
	_, err = tx.Exec("UPDATE prefix_counters SET last_id = ? WHERE prefix = ? AND node = ?", nextID, prefix, nodeID)
	if err != nil {
		return 0, fmt.Errorf("failed to update task counter for prefix %q: %w", prefix, err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nextID, nil
}

// ValidatePrefixName validates a prefix name according to the spec
func ValidatePrefixName(prefix string) error {
	if len(prefix) < 2 || len(prefix) > 20 {
		return fmt.Errorf("prefix must be 2-20 characters long")
	}

	// Must start with lowercase letter
	if prefix[0] < 'a' || prefix[0] > 'z' {
		return fmt.Errorf("prefix must start with a lowercase letter (a-z)")
	}

	// Must contain only lowercase letters, digits, and underscores (no hyphens)
	for i, c := range prefix {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return fmt.Errorf("prefix must contain only lowercase letters, digits, and underscores (char %d: %c)", i, c)
		}
	}

	// Check for reserved prefixes
	reserved := []string{"ev", "event", "task", "node", "remote", "sync"}
	for _, r := range reserved {
		if prefix == r {
			return fmt.Errorf("prefix %q is reserved", prefix)
		}
	}

	return nil
}

// ProjectPrefixCreatedEvent projects a prefix.created event into the prefixes table (idempotent)
func (d *DB) ProjectPrefixCreatedEvent(e Event) error {
	if e.Kind != "prefix.created" {
		return fmt.Errorf("expected prefix.created event, got %s", e.Kind)
	}

	var payload PrefixCreatedPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal prefix.created payload: %w", err)
	}

	// Extract node from event (we need to determine which node created this)
	// The event actor should match the creating node
	nodeID, err := d.GetOrCreateNodeID()
	if err != nil {
		return fmt.Errorf("failed to get node ID: %w", err)
	}

	// For events from other nodes, we'd need to extract the node from the event ID
	// Format: ev-<number>-<node>
	parts := strings.Split(e.ID, "-")
	if len(parts) >= 3 {
		nodeID = parts[2]
	}

	// Project into prefixes table (idempotent)
	_, err = d.db.Exec(
		"INSERT OR IGNORE INTO prefixes (prefix, node, description, created_at, created_by) VALUES (?, ?, ?, ?, ?)",
		payload.Prefix, nodeID, payload.Description, e.CreatedAt.UnixNano(), payload.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to project prefix.created event: %w", err)
	}

	// Initialize counter if it doesn't exist (idempotent)
	_, err = d.db.Exec(
		"INSERT OR IGNORE INTO prefix_counters (prefix, node, last_id) VALUES (?, ?, ?)",
		payload.Prefix, nodeID, 0,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize prefix counter: %w", err)
	}

	return nil
}

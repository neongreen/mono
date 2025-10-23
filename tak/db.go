package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	dbFileName = "tak.db"
)

// DB wraps a SQLite database for tak events
type DB struct {
	db *sql.DB
}

// OpenDB opens or creates a tak database at the given path
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
	`

	if _, err := d.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Initialize counter if it doesn't exist
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM task_counter").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check counter: %w", err)
	}
	if count == 0 {
		if _, err := d.db.Exec("INSERT INTO task_counter (last_id) VALUES (0)"); err != nil {
			return fmt.Errorf("failed to initialize counter: %w", err)
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

// GetDBPath returns the database path in ~/.tak/ directory
func GetDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	takDir := filepath.Join(home, ".tak")
	return filepath.Join(takDir, dbFileName), nil
}

// DBExists checks if a database exists at the given path
func DBExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetOrCreateInstallationSuffix gets the installation suffix or creates one if it doesn't exist
func (d *DB) GetOrCreateInstallationSuffix() (string, error) {
	// Try to get existing suffix
	var suffix string
	err := d.db.QueryRow("SELECT value FROM metadata WHERE key = 'installation_suffix'").Scan(&suffix)
	if err == nil {
		return suffix, nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to query installation suffix: %w", err)
	}

	// Generate new suffix (6 random alphanumeric characters)
	suffix = generateRandomSuffix(6)

	// Store it
	_, err = d.db.Exec("INSERT INTO metadata (key, value) VALUES ('installation_suffix', ?)", suffix)
	if err != nil {
		return "", fmt.Errorf("failed to store installation suffix: %w", err)
	}

	return suffix, nil
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

// generateRandomSuffix generates a random alphanumeric suffix
func generateRandomSuffix(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
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

// ResolveTaskID resolves a short task ID (e.g., "tak-2") to a full task ID (e.g., "tak-2-abc123")
// If the input is already a full ID, it validates that it exists
// Returns an error if the ID is ambiguous or doesn't exist
func (d *DB) ResolveTaskID(shortID string) (string, error) {
	// Get all task IDs from the database
	query := `
		SELECT DISTINCT json_extract(payload, '$.task_id') as task_id
		FROM events
		WHERE kind = 'task.created'
	`

	rows, err := d.db.Query(query)
	if err != nil {
		return "", fmt.Errorf("failed to query task IDs: %w", err)
	}
	defer rows.Close()

	var taskIDs []string
	for rows.Next() {
		var taskID string
		if err := rows.Scan(&taskID); err != nil {
			return "", fmt.Errorf("failed to scan task ID: %w", err)
		}
		taskIDs = append(taskIDs, taskID)
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("failed to iterate task IDs: %w", err)
	}

	// Check if shortID is already a full ID
	for _, fullID := range taskIDs {
		if fullID == shortID {
			return fullID, nil
		}
	}

	// Try to match as a short ID (without suffix)
	var matches []string
	for _, fullID := range taskIDs {
		// Extract the numeric part (e.g., "tak-2" from "tak-2-abc123")
		// Format is tak-<number>-<suffix>
		parts := strings.Split(fullID, "-")
		if len(parts) >= 2 {
			shortForm := strings.Join(parts[:2], "-") // tak-<number>
			if shortForm == shortID {
				matches = append(matches, fullID)
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

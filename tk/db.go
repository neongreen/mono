package main

import (
	"database/sql"

	"fmt"

	"os"
	"path/filepath"

	"github.com/neongreen/mono/tk/internal/reducer"
	"github.com/neongreen/mono/tk/internal/sync"
	_ "modernc.org/sqlite"
)

const (
	dbFileName = "tk.db"
)

// DB wraps a SQLite database for tk events
type DB struct {
	db            *sql.DB
	reducerCache  *reducer.Reducer // Cached reducer built from all events
	reducerConfig *sync.Config     // Config used to build cached reducer
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

	// Always create project tables (projects, project_aliases, tasks, task_numbers)
	if err := d.CreateProjectTables(); err != nil {
		return fmt.Errorf("failed to create project tables: %w", err)
	}

	return nil
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

// Query executes a query that returns rows
func (d *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

// QueryRow executes a query that is expected to return at most one row
func (d *DB) QueryRow(query string, args ...any) *sql.Row {
	return d.db.QueryRow(query, args...)
}

// Begin starts a transaction
func (d *DB) Begin() (*sql.Tx, error) {
	return d.db.Begin()
}

// Exec executes a query without returning any rows
func (d *DB) Exec(query string, args ...any) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

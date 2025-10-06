package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

type Run struct {
	ID           int64
	RepoPath     string
	RunType      string
	StartTime    time.Time
	EndTime      *time.Time
	CommitCount  int
	FileCount    int
	Status       string
}

type Commit struct {
	ID          int64
	RunID       int64
	Hash        string
	Author      string
	AuthorEmail string
	Date        time.Time
	Message     string
}

type File struct {
	ID       int64
	CommitID int64
	Path     string
	Size     int64
	Mode     string
}

// Open opens or creates the SQLite database in ~/.ingest/ingest.db
func Open() (*Database, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	ingestDir := filepath.Join(homeDir, ".ingest")
	if err := os.MkdirAll(ingestDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .ingest directory: %w", err)
	}

	dbPath := filepath.Join(ingestDir, "ingest.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{db: db}
	if err := database.createTables(); err != nil {
		db.Close()
		return nil, err
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// createTables creates the necessary tables if they don't exist
func (d *Database) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		repo_path TEXT NOT NULL,
		run_type TEXT NOT NULL DEFAULT 'git',
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		commit_count INTEGER DEFAULT 0,
		file_count INTEGER DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'in_progress'
	);

	CREATE TABLE IF NOT EXISTS commits (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		hash TEXT NOT NULL,
		author TEXT NOT NULL,
		author_email TEXT NOT NULL,
		date DATETIME NOT NULL,
		message TEXT NOT NULL,
		FOREIGN KEY (run_id) REFERENCES runs(id)
	);

	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		commit_id INTEGER NOT NULL,
		path TEXT NOT NULL,
		size INTEGER NOT NULL,
		mode TEXT NOT NULL,
		FOREIGN KEY (commit_id) REFERENCES commits(id)
	);

	CREATE TABLE IF NOT EXISTS fs_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		path TEXT NOT NULL,
		is_dir INTEGER NOT NULL,
		size INTEGER NOT NULL,
		mode TEXT NOT NULL,
		mod_time DATETIME NOT NULL,
		content BLOB,
		FOREIGN KEY (run_id) REFERENCES runs(id)
	);

	CREATE TABLE IF NOT EXISTS cmd_runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		command TEXT NOT NULL,
		exit_code INTEGER NOT NULL,
		stdout TEXT,
		stderr TEXT,
		duration_ms INTEGER NOT NULL,
		FOREIGN KEY (run_id) REFERENCES runs(id)
	);

	CREATE INDEX IF NOT EXISTS idx_commits_run_id ON commits(run_id);
	CREATE INDEX IF NOT EXISTS idx_files_commit_id ON files(commit_id);
	CREATE INDEX IF NOT EXISTS idx_fs_entries_run_id ON fs_entries(run_id);
	CREATE INDEX IF NOT EXISTS idx_cmd_runs_run_id ON cmd_runs(run_id);
	`

	_, err := d.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// CreateRun creates a new ingestion run
func (d *Database) CreateRun(repoPath string, runType string) (int64, error) {
	result, err := d.db.Exec(
		"INSERT INTO runs (repo_path, run_type, start_time, status) VALUES (?, ?, ?, ?)",
		repoPath,
		runType,
		time.Now(),
		"in_progress",
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create run: %w", err)
	}

	return result.LastInsertId()
}

// FinishRun marks a run as completed
func (d *Database) FinishRun(runID int64, status string) error {
	_, err := d.db.Exec(
		"UPDATE runs SET end_time = ?, status = ? WHERE id = ?",
		time.Now(),
		status,
		runID,
	)
	if err != nil {
		return fmt.Errorf("failed to finish run: %w", err)
	}

	return nil
}

// UpdateRunCounts updates commit and file counts for a run
func (d *Database) UpdateRunCounts(runID int64) error {
	_, err := d.db.Exec(`
		UPDATE runs 
		SET commit_count = (SELECT COUNT(*) FROM commits WHERE run_id = ?),
		    file_count = (SELECT COUNT(*) FROM files WHERE commit_id IN (SELECT id FROM commits WHERE run_id = ?))
		WHERE id = ?
	`, runID, runID, runID)
	if err != nil {
		return fmt.Errorf("failed to update run counts: %w", err)
	}

	return nil
}

// CreateCommit creates a new commit record
func (d *Database) CreateCommit(runID int64, hash, author, authorEmail string, date time.Time, message string) (int64, error) {
	result, err := d.db.Exec(
		"INSERT INTO commits (run_id, hash, author, author_email, date, message) VALUES (?, ?, ?, ?, ?, ?)",
		runID,
		hash,
		author,
		authorEmail,
		date,
		message,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create commit: %w", err)
	}

	return result.LastInsertId()
}

// CreateFile creates a new file record
func (d *Database) CreateFile(commitID int64, path string, size int64, mode string) error {
	_, err := d.db.Exec(
		"INSERT INTO files (commit_id, path, size, mode) VALUES (?, ?, ?, ?)",
		commitID,
		path,
		size,
		mode,
	)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

// GetAllRuns retrieves all ingestion runs
func (d *Database) GetAllRuns() ([]Run, error) {
	rows, err := d.db.Query(`
		SELECT id, repo_path, run_type, start_time, end_time, commit_count, file_count, status
		FROM runs
		ORDER BY start_time DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query runs: %w", err)
	}
	defer rows.Close()

	var runs []Run
	for rows.Next() {
		var run Run
		var endTime sql.NullTime
		err := rows.Scan(
			&run.ID,
			&run.RepoPath,
			&run.RunType,
			&run.StartTime,
			&endTime,
			&run.CommitCount,
			&run.FileCount,
			&run.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan run: %w", err)
		}

		if endTime.Valid {
			run.EndTime = &endTime.Time
		}

		runs = append(runs, run)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating runs: %w", err)
	}

	return runs, nil
}

// CreateFSEntry creates a new filesystem entry record
func (d *Database) CreateFSEntry(runID int64, path string, isDir bool, size int64, mode string, modTime time.Time, content []byte) error {
	_, err := d.db.Exec(
		"INSERT INTO fs_entries (run_id, path, is_dir, size, mode, mod_time, content) VALUES (?, ?, ?, ?, ?, ?, ?)",
		runID,
		path,
		isDir,
		size,
		mode,
		modTime,
		content,
	)
	if err != nil {
		return fmt.Errorf("failed to create fs entry: %w", err)
	}

	return nil
}

// CreateCmdRun creates a new command run record
func (d *Database) CreateCmdRun(runID int64, command string, exitCode int, stdout, stderr string, durationMs int64) error {
	_, err := d.db.Exec(
		"INSERT INTO cmd_runs (run_id, command, exit_code, stdout, stderr, duration_ms) VALUES (?, ?, ?, ?, ?, ?)",
		runID,
		command,
		exitCode,
		stdout,
		stderr,
		durationMs,
	)
	if err != nil {
		return fmt.Errorf("failed to create cmd run: %w", err)
	}

	return nil
}

// UpdateRunFileCount updates the file count for filesystem and command runs
func (d *Database) UpdateRunFileCount(runID int64) error {
	_, err := d.db.Exec(`
		UPDATE runs 
		SET file_count = (SELECT COUNT(*) FROM fs_entries WHERE run_id = ?)
		WHERE id = ?
	`, runID, runID)
	if err != nil {
		return fmt.Errorf("failed to update run file count: %w", err)
	}

	return nil
}

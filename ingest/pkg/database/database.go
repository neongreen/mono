package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

type Run struct {
	ID        int64
	RepoPath  string
	RunType   string
	StartTime time.Time
	EndTime   *time.Time
	ItemCount int
	Status    string
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
		item_count INTEGER DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'in_progress'
	);

	CREATE TABLE IF NOT EXISTS blobs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sha256 TEXT NOT NULL UNIQUE,
		content BLOB NOT NULL,
		size INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS commits (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		hash TEXT NOT NULL,
		author TEXT NOT NULL,
		author_email TEXT NOT NULL,
		committer TEXT NOT NULL,
		committer_email TEXT NOT NULL,
		date DATETIME NOT NULL,
		message TEXT NOT NULL,
		FOREIGN KEY (run_id) REFERENCES runs(id)
	);

	CREATE TABLE IF NOT EXISTS commit_parents (
		commit_id INTEGER NOT NULL,
		parent_hash TEXT NOT NULL,
		FOREIGN KEY (commit_id) REFERENCES commits(id),
		PRIMARY KEY (commit_id, parent_hash)
	);

	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		commit_id INTEGER NOT NULL,
		path TEXT NOT NULL,
		size INTEGER NOT NULL,
		mode TEXT NOT NULL,
		blob_id INTEGER,
		FOREIGN KEY (commit_id) REFERENCES commits(id),
		FOREIGN KEY (blob_id) REFERENCES blobs(id)
	);

	CREATE TABLE IF NOT EXISTS git_refs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		ref_type TEXT NOT NULL,
		name TEXT NOT NULL,
		target_hash TEXT NOT NULL,
		FOREIGN KEY (run_id) REFERENCES runs(id)
	);

	CREATE TABLE IF NOT EXISTS git_remotes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		url TEXT NOT NULL,
		FOREIGN KEY (run_id) REFERENCES runs(id)
	);

	CREATE TABLE IF NOT EXISTS fs_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		path TEXT NOT NULL,
		is_dir INTEGER NOT NULL,
		size INTEGER NOT NULL,
		mode TEXT NOT NULL,
		mod_time DATETIME NOT NULL,
		blob_id INTEGER,
		FOREIGN KEY (run_id) REFERENCES runs(id),
		FOREIGN KEY (blob_id) REFERENCES blobs(id)
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

	CREATE TABLE IF NOT EXISTS github_issues (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		number INTEGER NOT NULL,
		title TEXT NOT NULL,
		body TEXT,
		state TEXT NOT NULL,
		author TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		closed_at DATETIME,
		labels TEXT,
		assignees TEXT,
		milestone TEXT,
		FOREIGN KEY (run_id) REFERENCES runs(id)
	);

	CREATE TABLE IF NOT EXISTS github_prs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		number INTEGER NOT NULL,
		title TEXT NOT NULL,
		body TEXT,
		state TEXT NOT NULL,
		author TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		closed_at DATETIME,
		merged_at DATETIME,
		merged INTEGER NOT NULL DEFAULT 0,
		draft INTEGER NOT NULL DEFAULT 0,
		base_branch TEXT NOT NULL,
		head_branch TEXT NOT NULL,
		labels TEXT,
		assignees TEXT,
		reviewers TEXT,
		milestone TEXT,
		FOREIGN KEY (run_id) REFERENCES runs(id)
	);

	CREATE TABLE IF NOT EXISTS github_comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		item_type TEXT NOT NULL,
		item_number INTEGER NOT NULL,
		comment_id INTEGER NOT NULL,
		author TEXT NOT NULL,
		body TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (run_id) REFERENCES runs(id)
	);

	CREATE INDEX IF NOT EXISTS idx_blobs_sha256 ON blobs(sha256);
	CREATE INDEX IF NOT EXISTS idx_commits_run_id ON commits(run_id);
	CREATE INDEX IF NOT EXISTS idx_commits_hash ON commits(hash);
	CREATE INDEX IF NOT EXISTS idx_commit_parents_commit_id ON commit_parents(commit_id);
	CREATE INDEX IF NOT EXISTS idx_files_commit_id ON files(commit_id);
	CREATE INDEX IF NOT EXISTS idx_files_blob_id ON files(blob_id);
	CREATE INDEX IF NOT EXISTS idx_git_refs_run_id ON git_refs(run_id);
	CREATE INDEX IF NOT EXISTS idx_git_remotes_run_id ON git_remotes(run_id);
	CREATE INDEX IF NOT EXISTS idx_fs_entries_run_id ON fs_entries(run_id);
	CREATE INDEX IF NOT EXISTS idx_fs_entries_blob_id ON fs_entries(blob_id);
	CREATE INDEX IF NOT EXISTS idx_cmd_runs_run_id ON cmd_runs(run_id);
	CREATE INDEX IF NOT EXISTS idx_github_issues_run_id ON github_issues(run_id);
	CREATE INDEX IF NOT EXISTS idx_github_issues_number ON github_issues(run_id, number);
	CREATE INDEX IF NOT EXISTS idx_github_prs_run_id ON github_prs(run_id);
	CREATE INDEX IF NOT EXISTS idx_github_prs_number ON github_prs(run_id, number);
	CREATE INDEX IF NOT EXISTS idx_github_comments_run_id ON github_comments(run_id);
	CREATE INDEX IF NOT EXISTS idx_github_comments_item ON github_comments(run_id, item_type, item_number);
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

// UpdateRunItemCount updates item count for a run based on run type
func (d *Database) UpdateRunItemCount(runID int64) error {
	// Get run type first
	var runType string
	err := d.db.QueryRow("SELECT run_type FROM runs WHERE id = ?", runID).Scan(&runType)
	if err != nil {
		return fmt.Errorf("failed to get run type: %w", err)
	}

	var query string
	var args []interface{}
	switch runType {
	case "git":
		query = "UPDATE runs SET item_count = (SELECT COUNT(*) FROM commits WHERE run_id = ?) WHERE id = ?"
		args = []interface{}{runID, runID}
	case "fs":
		query = "UPDATE runs SET item_count = (SELECT COUNT(*) FROM fs_entries WHERE run_id = ?) WHERE id = ?"
		args = []interface{}{runID, runID}
	case "cmd":
		lineCount, err := d.totalCommandOutputLines(runID)
		if err != nil {
			return fmt.Errorf("failed to calculate command output lines: %w", err)
		}
		_, err = d.db.Exec("UPDATE runs SET item_count = ? WHERE id = ?", lineCount, runID)
		if err != nil {
			return fmt.Errorf("failed to update run item count: %w", err)
		}
		return nil
	case "github":
		query = "UPDATE runs SET item_count = (SELECT COUNT(*) FROM github_issues WHERE run_id = ?) + (SELECT COUNT(*) FROM github_prs WHERE run_id = ?) WHERE id = ?"
		args = []interface{}{runID, runID, runID}
	default:
		query = "UPDATE runs SET item_count = 0 WHERE id = ?"
		args = []interface{}{runID}
	}

	_, err = d.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update run item count: %w", err)
	}

	return nil
}

func (d *Database) totalCommandOutputLines(runID int64) (int, error) {
	rows, err := d.db.Query("SELECT stdout, stderr FROM cmd_runs WHERE run_id = ?", runID)
	if err != nil {
		return 0, fmt.Errorf("failed to select command outputs: %w", err)
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		var stdout, stderr sql.NullString
		if err := rows.Scan(&stdout, &stderr); err != nil {
			return 0, fmt.Errorf("failed to scan command outputs: %w", err)
		}
		if stdout.Valid {
			total += countOutputLines(stdout.String)
		}
		if stderr.Valid {
			total += countOutputLines(stderr.String)
		}
	}

	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("error iterating command outputs: %w", err)
	}

	return total, nil
}

func countOutputLines(output string) int {
	if output == "" {
		return 0
	}

	count := 1
	for _, r := range output {
		if r == '\n' {
			count++
		}
	}
	if strings.HasSuffix(output, "\n") {
		count--
	}
	return count
}

// UpdateRunCounts is deprecated, use UpdateRunItemCount instead
func (d *Database) UpdateRunCounts(runID int64) error {
	return d.UpdateRunItemCount(runID)
}

// CreateCommit creates a new commit record
func (d *Database) CreateCommit(runID int64, hash, author, authorEmail, committer, committerEmail string, date time.Time, message string, parentHashes []string) (int64, error) {
	result, err := d.db.Exec(
		"INSERT INTO commits (run_id, hash, author, author_email, committer, committer_email, date, message) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		runID,
		hash,
		author,
		authorEmail,
		committer,
		committerEmail,
		date,
		message,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create commit: %w", err)
	}

	commitID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get commit ID: %w", err)
	}

	// Insert parent relationships
	for _, parentHash := range parentHashes {
		_, err := d.db.Exec(
			"INSERT INTO commit_parents (commit_id, parent_hash) VALUES (?, ?)",
			commitID,
			parentHash,
		)
		if err != nil {
			return 0, fmt.Errorf("failed to create commit parent: %w", err)
		}
	}

	return commitID, nil
}

// GetOrCreateBlob stores a blob if it doesn't exist and returns its ID
func (d *Database) GetOrCreateBlob(content []byte, sha256Hash string) (int64, error) {
	// Check if blob already exists
	var blobID int64
	err := d.db.QueryRow("SELECT id FROM blobs WHERE sha256 = ?", sha256Hash).Scan(&blobID)
	if err == nil {
		// Blob exists, return its ID
		return blobID, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to check for existing blob: %w", err)
	}

	// Blob doesn't exist, create it
	result, err := d.db.Exec(
		"INSERT INTO blobs (sha256, content, size) VALUES (?, ?, ?)",
		sha256Hash,
		content,
		len(content),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create blob: %w", err)
	}

	return result.LastInsertId()
}

// CreateFile creates a new file record
func (d *Database) CreateFile(commitID int64, path string, size int64, mode string, blobID *int64) error {
	_, err := d.db.Exec(
		"INSERT INTO files (commit_id, path, size, mode, blob_id) VALUES (?, ?, ?, ?, ?)",
		commitID,
		path,
		size,
		mode,
		blobID,
	)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

// GetAllRuns retrieves all ingestion runs
func (d *Database) GetAllRuns() ([]Run, error) {
	rows, err := d.db.Query(`
		SELECT id, repo_path, run_type, start_time, end_time, item_count, status
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
			&run.ItemCount,
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
func (d *Database) CreateFSEntry(runID int64, path string, isDir bool, size int64, mode string, modTime time.Time, blobID *int64) error {
	_, err := d.db.Exec(
		"INSERT INTO fs_entries (run_id, path, is_dir, size, mode, mod_time, blob_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		runID,
		path,
		isDir,
		size,
		mode,
		modTime,
		blobID,
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

// UpdateRunFileCount is deprecated, use UpdateRunItemCount instead
func (d *Database) UpdateRunFileCount(runID int64) error {
	return d.UpdateRunItemCount(runID)
}

// CreateGitRef creates a new git reference record (branch or tag)
func (d *Database) CreateGitRef(runID int64, refType, name, targetHash string) error {
	_, err := d.db.Exec(
		"INSERT INTO git_refs (run_id, ref_type, name, target_hash) VALUES (?, ?, ?, ?)",
		runID,
		refType,
		name,
		targetHash,
	)
	if err != nil {
		return fmt.Errorf("failed to create git ref: %w", err)
	}

	return nil
}

// CreateGitRemote creates a new git remote record
func (d *Database) CreateGitRemote(runID int64, name, url string) error {
	_, err := d.db.Exec(
		"INSERT INTO git_remotes (run_id, name, url) VALUES (?, ?, ?)",
		runID,
		name,
		url,
	)
	if err != nil {
		return fmt.Errorf("failed to create git remote: %w", err)
	}

	return nil
}

// Query executes a SQL query and returns results as JSON-serializable data
func (d *Database) Query(query string) ([]map[string]interface{}, error) {
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		// Create a slice of interface{} to hold each column's value
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Convert []byte to string for better JSON output
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}

		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// CreateGitHubIssue creates a new GitHub issue record
func (d *Database) CreateGitHubIssue(runID int64, number int, title, body, state, author string, createdAt, updatedAt time.Time, closedAt *time.Time, labels, assignees, milestone string) error {
	_, err := d.db.Exec(
		"INSERT INTO github_issues (run_id, number, title, body, state, author, created_at, updated_at, closed_at, labels, assignees, milestone) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		runID,
		number,
		title,
		body,
		state,
		author,
		createdAt,
		updatedAt,
		closedAt,
		labels,
		assignees,
		milestone,
	)
	if err != nil {
		return fmt.Errorf("failed to create github issue: %w", err)
	}

	return nil
}

// CreateGitHubPR creates a new GitHub pull request record
func (d *Database) CreateGitHubPR(runID int64, number int, title, body, state, author string, createdAt, updatedAt time.Time, closedAt, mergedAt *time.Time, merged, draft bool, baseBranch, headBranch, labels, assignees, reviewers, milestone string) error {
	_, err := d.db.Exec(
		"INSERT INTO github_prs (run_id, number, title, body, state, author, created_at, updated_at, closed_at, merged_at, merged, draft, base_branch, head_branch, labels, assignees, reviewers, milestone) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		runID,
		number,
		title,
		body,
		state,
		author,
		createdAt,
		updatedAt,
		closedAt,
		mergedAt,
		merged,
		draft,
		baseBranch,
		headBranch,
		labels,
		assignees,
		reviewers,
		milestone,
	)
	if err != nil {
		return fmt.Errorf("failed to create github pull request: %w", err)
	}

	return nil
}

// CreateGitHubComment creates a new GitHub comment record
func (d *Database) CreateGitHubComment(runID int64, itemType string, itemNumber int, commentID int64, author, body string, createdAt, updatedAt time.Time) error {
	_, err := d.db.Exec(
		"INSERT INTO github_comments (run_id, item_type, item_number, comment_id, author, body, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		runID,
		itemType,
		itemNumber,
		commentID,
		author,
		body,
		createdAt,
		updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create github comment: %w", err)
	}

	return nil
}

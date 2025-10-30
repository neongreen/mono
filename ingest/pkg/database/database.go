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

// GitHubIssueRecord represents a row in github_issues.
type GitHubIssueRecord struct {
	RunID             int64
	Number            int
	Title             string
	Body              string
	State             string
	Author            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ClosedAt          *time.Time
	Labels            string
	Assignees         string
	Milestone         string
	NodeID            string
	IssueID           int64
	HTMLURL           string
	APIURL            string
	CommentsURL       string
	EventsURL         string
	StateReason       string
	Locked            bool
	ActiveLockReason  string
	Draft             bool
	ClosedBy          string
	CommentCount      int
	ReactionsTotal    int
	ParticipantsCount int
}

// GitHubCommentReaction represents a reaction on a GitHub comment.
type GitHubCommentReaction struct {
	RunID      int64
	ItemType   string
	ItemNumber int
	CommentID  int64
	Reactor    string
	Content    string
}

// LinearIssue represents a Linear issue record stored in the database.
type LinearIssue struct {
	IssueID     string
	Identifier  string
	Title       string
	Description *string
	Priority    *int
	Status      *string
	Assignee    *string
	Team        *string
	URL         *string
	RawData     *string
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
	node_id TEXT,
	issue_id INTEGER,
	html_url TEXT,
	api_url TEXT,
	comments_url TEXT,
	events_url TEXT,
	state_reason TEXT,
	locked INTEGER NOT NULL DEFAULT 0,
	active_lock_reason TEXT,
	draft INTEGER NOT NULL DEFAULT 0,
	closed_by TEXT,
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

	CREATE TABLE IF NOT EXISTS github_comment_reactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		item_type TEXT NOT NULL,
		item_number INTEGER NOT NULL,
		comment_id INTEGER NOT NULL,
		reactor TEXT NOT NULL,
		content TEXT NOT NULL,
		FOREIGN KEY (run_id) REFERENCES runs(id)
	);

	CREATE TABLE IF NOT EXISTS linear_issues (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id INTEGER NOT NULL,
		issue_id TEXT NOT NULL,
		identifier TEXT NOT NULL,
		title TEXT NOT NULL,
		description TEXT,
		priority INTEGER,
		status TEXT,
		assignee TEXT,
		team TEXT,
		url TEXT,
		raw_data TEXT,
		FOREIGN KEY (run_id) REFERENCES runs(id),
		UNIQUE(run_id, issue_id)
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
	CREATE INDEX IF NOT EXISTS idx_github_comment_reactions_run_id ON github_comment_reactions(run_id);
	CREATE INDEX IF NOT EXISTS idx_github_comment_reactions_comment ON github_comment_reactions(comment_id);
	CREATE INDEX IF NOT EXISTS idx_linear_issues_run_id ON linear_issues(run_id);
	CREATE INDEX IF NOT EXISTS idx_linear_issues_identifier ON linear_issues(run_id, identifier);
	`

	_, err := d.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	if err := d.ensureSchemaUpgrades(); err != nil {
		return err
	}

	return nil
}

func (d *Database) ensureSchemaUpgrades() error {
	columns := []struct {
		table string
		name  string
		typ   string
	}{
		{"github_issues", "node_id", "TEXT"},
		{"github_issues", "issue_id", "INTEGER"},
		{"github_issues", "html_url", "TEXT"},
		{"github_issues", "api_url", "TEXT"},
		{"github_issues", "comments_url", "TEXT"},
		{"github_issues", "events_url", "TEXT"},
		{"github_issues", "state_reason", "TEXT"},
		{"github_issues", "locked", "INTEGER NOT NULL DEFAULT 0"},
		{"github_issues", "active_lock_reason", "TEXT"},
		{"github_issues", "draft", "INTEGER NOT NULL DEFAULT 0"},
		{"github_issues", "closed_by", "TEXT"},
		{"github_issues", "comment_count", "INTEGER"},
		{"github_issues", "reaction_total", "INTEGER"},
		{"github_issues", "participants_count", "INTEGER"},
	}

	for _, col := range columns {
		if err := d.ensureColumn(col.table, col.name, col.typ); err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) ensureColumn(table, column, columnType string) error {
	rows, err := d.db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return fmt.Errorf("failed to inspect %s: %w", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid      int
			name     string
			typ      string
			notnull  int
			defaultV any
			pk       int
		)
		if err := rows.Scan(&cid, &name, &typ, &notnull, &defaultV, &pk); err != nil {
			return fmt.Errorf("failed to scan table info for %s: %w", table, err)
		}
		if name == column {
			return nil
		}
	}

	if _, err := d.db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, columnType)); err != nil {
		return fmt.Errorf("failed to add column %s.%s: %w", table, column, err)
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


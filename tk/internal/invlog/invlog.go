package invlog

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// InvocationLog represents a single invocation of the tk command
type InvocationLog struct {
	Timestamp  time.Time `json:"timestamp"`
	Command    string    `json:"command"`
	Args       []string  `json:"args"`
	PID        int       `json:"pid"`
	PPID       int       `json:"ppid"`
	User       string    `json:"user"`
	Success    bool      `json:"success"`
	ExitCode   int       `json:"exit_code"`
	Stdout     string    `json:"stdout,omitempty"`
	Stderr     string    `json:"stderr,omitempty"`
	DurationMs int64     `json:"duration_ms"`
}

// GetLogPath returns the path to the invocation log file (legacy JSONL)
func GetLogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	tkDir := filepath.Join(home, ".tk")
	// Ensure the directory exists with private permissions
	if err := os.MkdirAll(tkDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create tk directory: %w", err)
	}

	return filepath.Join(tkDir, "log.jsonl"), nil
}

// GetDBPath returns the path to the invocation log SQLite database
func GetDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	tkDir := filepath.Join(home, ".tk")
	// Ensure the directory exists with private permissions
	if err := os.MkdirAll(tkDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create tk directory: %w", err)
	}

	return filepath.Join(tkDir, "invlog.db"), nil
}

// WriteLog appends an invocation log entry to the log file
func WriteLog(log InvocationLog) error {
	logPath, err := GetLogPath()
	if err != nil {
		return fmt.Errorf("failed to get log path: %w", err)
	}

	// Open file for appending, create if doesn't exist with private permissions
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer f.Close()

	// Acquire exclusive lock to prevent concurrent write interleaving
	// Uses platform-specific implementation (flock on Unix, no-op on Windows)
	if err := lockFile(f); err != nil {
		// If locking fails, continue without it
		// This ensures logging doesn't break the command
	} else {
		defer unlockFile(f)
	}

	// Marshal to JSON
	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Append newline for JSONL format
	data = append(data, '\n')

	// Write to file
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	return nil
}

// openDB opens or creates the invocation log SQLite database
func openDB() (*sql.DB, error) {
	dbPath, err := GetDBPath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open invlog database: %w", err)
	}

	// Create schema if needed
	schema := `
		CREATE TABLE IF NOT EXISTS invocations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp INTEGER NOT NULL,
			command TEXT NOT NULL,
			args TEXT NOT NULL,
			pid INTEGER,
			ppid INTEGER,
			user TEXT,
			success INTEGER NOT NULL,
			exit_code INTEGER NOT NULL,
			stdout BLOB,
			stderr BLOB,
			duration_ms INTEGER NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_invocations_timestamp ON invocations(timestamp);
		CREATE INDEX IF NOT EXISTS idx_invocations_success ON invocations(success);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	// Check if database is empty - if so, migrate from JSONL
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM invocations`).Scan(&count); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to check if database is empty: %w", err)
	}

	if count == 0 {
		// Try to migrate from JSONL
		if err := migrateFromJSONL(db); err != nil {
			// Migration failure is not fatal - just log it and continue
			// The database is still usable, just without historical data
		}
	}

	return db, nil
}

// migrateFromJSONL migrates existing JSONL log entries to SQLite database
func migrateFromJSONL(db *sql.DB) error {
	jsonlPath, err := GetLogPath()
	if err != nil {
		return err
	}

	// Check if JSONL file exists
	if _, err := os.Stat(jsonlPath); os.IsNotExist(err) {
		// No JSONL file to migrate
		return nil
	}

	// Read JSONL file
	data, err := os.ReadFile(jsonlPath)
	if err != nil {
		return fmt.Errorf("failed to read JSONL file: %w", err)
	}

	// Parse JSONL (one JSON object per line)
	lines := bytes.Split(data, []byte("\n"))
	migrated := 0

	for i, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		var entry InvocationLog
		if err := json.Unmarshal(line, &entry); err != nil {
			// Skip malformed entries
			continue
		}

		// Compress stdout and stderr
		stdoutCompressed, err := compress([]byte(entry.Stdout))
		if err != nil {
			return fmt.Errorf("failed to compress stdout for entry %d: %w", i, err)
		}

		stderrCompressed, err := compress([]byte(entry.Stderr))
		if err != nil {
			return fmt.Errorf("failed to compress stderr for entry %d: %w", i, err)
		}

		// Marshal args to JSON
		argsJSON, err := json.Marshal(entry.Args)
		if err != nil {
			return fmt.Errorf("failed to marshal args for entry %d: %w", i, err)
		}

		// Insert into database
		_, err = db.Exec(`
			INSERT INTO invocations (timestamp, command, args, pid, ppid, user, success, exit_code, stdout, stderr, duration_ms)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, entry.Timestamp.UnixNano(), entry.Command, string(argsJSON), entry.PID, entry.PPID, entry.User,
			boolToInt(entry.Success), entry.ExitCode, stdoutCompressed, stderrCompressed, entry.DurationMs)

		if err != nil {
			return fmt.Errorf("failed to insert entry %d: %w", i, err)
		}

		migrated++
	}

	// Migration successful - delete JSONL file
	if err := os.Remove(jsonlPath); err != nil {
		// Don't fail if we can't delete - just continue
		return nil
	}

	return nil
}

// compress compresses data using gzip
func compress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// WriteLogDB writes an invocation log entry to the SQLite database
func WriteLogDB(log InvocationLog) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Compress stdout and stderr
	stdoutCompressed, err := compress([]byte(log.Stdout))
	if err != nil {
		return fmt.Errorf("failed to compress stdout: %w", err)
	}

	stderrCompressed, err := compress([]byte(log.Stderr))
	if err != nil {
		return fmt.Errorf("failed to compress stderr: %w", err)
	}

	// Marshal args to JSON
	argsJSON, err := json.Marshal(log.Args)
	if err != nil {
		return fmt.Errorf("failed to marshal args: %w", err)
	}

	// Insert into database
	_, err = db.Exec(`
		INSERT INTO invocations (timestamp, command, args, pid, ppid, user, success, exit_code, stdout, stderr, duration_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, log.Timestamp.UnixNano(), log.Command, string(argsJSON), log.PID, log.PPID, log.User,
		boolToInt(log.Success), log.ExitCode, stdoutCompressed, stderrCompressed, log.DurationMs)

	if err != nil {
		return fmt.Errorf("failed to insert log entry: %w", err)
	}

	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

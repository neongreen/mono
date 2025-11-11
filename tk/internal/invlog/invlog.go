package invlog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/neongreen/mono/tk/lib/rotatinglog"
)

const (
	logDirName     = "invlogs"
	maxLogFileSize = 10 * 1024 * 1024 // 10MB
)

// InvocationLog represents a single invocation of the tk command
type InvocationLog struct {
	Timestamp  int64    `json:"timestamp"` // Unix nanoseconds
	Command    string   `json:"command"`
	Args       []string `json:"args"`
	PID        int      `json:"pid"`
	PPID       int      `json:"ppid"`
	User       string   `json:"user"`
	Success    bool     `json:"success"`
	ExitCode   int      `json:"exit_code"`
	Stdout     string   `json:"stdout,omitempty"`
	Stderr     string   `json:"stderr,omitempty"`
	DurationMs int64    `json:"duration_ms"`
}

// GetLogDir returns the path to the invocation log directory
func GetLogDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	tkDir := filepath.Join(home, ".tk")
	logDir := filepath.Join(tkDir, logDirName)

	return logDir, nil
}

// WriteLog appends an invocation log entry to the rotating log
func WriteLog(log InvocationLog) error {
	// Migrate legacy logs on first write (if they exist)
	// This is called once per invocation, but MigrateLegacyLogs is idempotent
	// (it does nothing if no legacy files exist)
	_ = MigrateLegacyLogs()

	logDir, err := GetLogDir()
	if err != nil {
		return err
	}

	// Create writer
	writer, err := rotatinglog.NewWriter(logDir, maxLogFileSize)
	if err != nil {
		return fmt.Errorf("failed to create log writer: %w", err)
	}
	defer writer.Close()

	// Convert timestamp from time.Time to Unix nanoseconds if needed
	// (Already done in main.go, but keeping the field as int64 for simplicity)

	// Marshal to JSON
	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Append to log
	if err := writer.Append(data); err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	return nil
}

// Query executes a SQL query against the invocation logs
// Use ? placeholders for parameters to avoid SQL injection
func Query(sqlQuery string, args ...interface{}) ([]rotatinglog.QueryResult, error) {
	logDir, err := GetLogDir()
	if err != nil {
		return nil, err
	}

	return rotatinglog.Query(logDir, sqlQuery, args...)
}

// Search searches for a pattern in the invocation logs
func Search(pattern string) ([]rotatinglog.QueryResult, error) {
	logDir, err := GetLogDir()
	if err != nil {
		return nil, err
	}

	return rotatinglog.Search(logDir, pattern)
}

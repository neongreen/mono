package invlog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
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

// GetLogPath returns the path to the invocation log file
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

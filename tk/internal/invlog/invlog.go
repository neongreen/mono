package invlog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
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
	// Ensure the directory exists
	if err := os.MkdirAll(tkDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create tk directory: %w", err)
	}

	return filepath.Join(tkDir, "invocations.jsonl"), nil
}

// WriteLog appends an invocation log entry to the log file
func WriteLog(log InvocationLog) error {
	logPath, err := GetLogPath()
	if err != nil {
		return fmt.Errorf("failed to get log path: %w", err)
	}

	// Open file for appending, create if doesn't exist
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer f.Close()

	// Acquire exclusive lock to prevent concurrent write interleaving
	// Note: syscall.Flock is Unix-specific. On Windows, this will fail gracefully
	// and logging will still work, just without concurrent write protection.
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		// If locking fails, continue without it (e.g., on unsupported platforms)
		// This ensures logging doesn't break the command on Windows
	} else {
		defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
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

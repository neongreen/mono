package invlog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteLog(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for testing
	origHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if origHome == "" {
			os.Unsetenv("HOME")
		} else {
			os.Setenv("HOME", origHome)
		}
	})
	os.Setenv("HOME", tmpDir)

	// Create .tk directory in temp home
	tkDir := filepath.Join(tmpDir, ".tk")
	if err := os.MkdirAll(tkDir, 0755); err != nil {
		t.Fatalf("failed to create .tk dir: %v", err)
	}

	// Write a test log entry
	testLog := InvocationLog{
		Timestamp:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Command:    "tk",
		Args:       []string{"new", "test task"},
		PID:        12345,
		PPID:       12344,
		User:       "testuser",
		Success:    true,
		ExitCode:   0,
		Stdout:     "Created task tk-1: test task\n",
		Stderr:     "",
		DurationMs: 150,
	}

	err := WriteLog(testLog)
	if err != nil {
		t.Fatalf("WriteLog failed: %v", err)
	}

	// Read back the log file
	expectedPath := filepath.Join(tkDir, "invocations.jsonl")
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Parse the JSON line (trim trailing newline)
	var readLog InvocationLog
	dataStr := string(data)
	if len(dataStr) > 0 && dataStr[len(dataStr)-1] == '\n' {
		dataStr = dataStr[:len(dataStr)-1]
	}
	if err := json.Unmarshal([]byte(dataStr), &readLog); err != nil {
		t.Fatalf("failed to unmarshal log entry: %v", err)
	}

	// Verify the log entry
	if readLog.Command != testLog.Command {
		t.Errorf("Command mismatch: got %v, want %v", readLog.Command, testLog.Command)
	}
	if len(readLog.Args) != len(testLog.Args) {
		t.Errorf("Args length mismatch: got %v, want %v", len(readLog.Args), len(testLog.Args))
	}
	if readLog.PID != testLog.PID {
		t.Errorf("PID mismatch: got %v, want %v", readLog.PID, testLog.PID)
	}
	if readLog.Success != testLog.Success {
		t.Errorf("Success mismatch: got %v, want %v", readLog.Success, testLog.Success)
	}
}

func TestWriteLogAppend(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Override the home directory for testing
	origHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if origHome == "" {
			os.Unsetenv("HOME")
		} else {
			os.Setenv("HOME", origHome)
		}
	})
	os.Setenv("HOME", tmpDir)

	// Create .tk directory in temp home
	tkDir := filepath.Join(tmpDir, ".tk")
	if err := os.MkdirAll(tkDir, 0755); err != nil {
		t.Fatalf("failed to create .tk dir: %v", err)
	}

	// Write two log entries
	log1 := InvocationLog{
		Timestamp:  time.Now(),
		Command:    "tk",
		Args:       []string{"new", "task 1"},
		PID:        100,
		Success:    true,
		ExitCode:   0,
		DurationMs: 100,
	}

	log2 := InvocationLog{
		Timestamp:  time.Now(),
		Command:    "tk",
		Args:       []string{"ls"},
		PID:        101,
		Success:    true,
		ExitCode:   0,
		DurationMs: 50,
	}

	if err := WriteLog(log1); err != nil {
		t.Fatalf("WriteLog(log1) failed: %v", err)
	}
	if err := WriteLog(log2); err != nil {
		t.Fatalf("WriteLog(log2) failed: %v", err)
	}

	// Read back the log file
	expectedPath := filepath.Join(tkDir, "invocations.jsonl")
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Count the number of lines
	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}

	if lines != 2 {
		t.Errorf("Expected 2 log entries, got %d", lines)
	}
}

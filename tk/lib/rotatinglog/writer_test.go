package rotatinglog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNewWriter(t *testing.T) {
	dir := t.TempDir()
	w, err := NewWriter(dir, 1024)
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}
	if w == nil {
		t.Fatal("expected non-nil writer")
	}

	// Verify directory was created
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Fatalf("directory was not created")
	}
}

func TestAppend(t *testing.T) {
	dir := t.TempDir()
	w, err := NewWriter(dir, 1024*1024) // 1MB threshold
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}

	// Append a JSONL entry
	data := []byte(`{"field":"value","number":123}`)
	if err := w.Append(data); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// Verify file was created
	logPath := filepath.Join(dir, "current.jsonl")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	// Should have data + newline
	expected := `{"field":"value","number":123}` + "\n"
	if string(content) != expected {
		t.Errorf("expected %q, got %q", expected, string(content))
	}
}

func TestRotation(t *testing.T) {
	dir := t.TempDir()
	// Set very small threshold to trigger rotation
	w, err := NewWriter(dir, 100) // 100 bytes
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}

	// Write data that exceeds threshold
	largeData := make([]byte, 150)
	for i := range largeData {
		largeData[i] = 'a'
	}
	entry := map[string]any{
		"data": string(largeData),
	}
	entryJSON, _ := json.Marshal(entry)

	// First append - should succeed
	if err := w.Append(entryJSON); err != nil {
		t.Fatalf("first append failed: %v", err)
	}

	// Second append - should trigger rotation
	if err := w.Append(entryJSON); err != nil {
		t.Fatalf("second append failed: %v", err)
	}

	// Should have compressed file now
	matches, err := filepath.Glob(filepath.Join(dir, "*.jsonl.zst"))
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}

	if len(matches) == 0 {
		t.Errorf("expected compressed file after rotation, found none")
	}
}

func TestQuery(t *testing.T) {
	// Skip test if DuckDB CLI is not available
	if err := checkDuckDBAvailable(); err != nil {
		t.Skip("DuckDB CLI not available, skipping test:", err)
	}

	dir := t.TempDir()
	w, err := NewWriter(dir, 1024*1024)
	if err != nil {
		t.Fatalf("NewWriter failed: %v", err)
	}

	// Write some test data
	entries := []map[string]any{
		{"timestamp": 100, "command": "tk", "status": "success"},
		{"timestamp": 200, "command": "tk", "status": "error"},
		{"timestamp": 300, "command": "tk", "status": "success"},
	}

	for _, entry := range entries {
		data, _ := json.Marshal(entry)
		if err := w.Append(data); err != nil {
			t.Fatalf("append failed: %v", err)
		}
	}

	// Query the data
	results, err := Query(dir, `SELECT * FROM logs WHERE status = 'error'`)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	if len(results) > 0 {
		if results[0]["status"] != "error" {
			t.Errorf("expected status=error, got %v", results[0]["status"])
		}
	}
}

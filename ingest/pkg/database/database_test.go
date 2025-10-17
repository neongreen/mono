package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDatabaseOperations(t *testing.T) {
	// Create a temporary directory for test database
	tempDir := t.TempDir()

	// Mock the home directory for testing
	originalHomeDir := os.Getenv("HOME")
	testHome := tempDir
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", originalHomeDir)

	// Open database
	db, err := Open()
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Verify .ingest directory was created
	ingestDir := filepath.Join(testHome, ".ingest")
	if _, err := os.Stat(ingestDir); os.IsNotExist(err) {
		t.Fatalf(".ingest directory was not created")
	}

	// Test creating a run
	runID, err := db.CreateRun("/test/repo", "git")
	if err != nil {
		t.Fatalf("Failed to create run: %v", err)
	}

	if runID != 1 {
		t.Errorf("Expected run ID 1, got %d", runID)
	}

	// Test creating a commit
	commitDate := time.Now()
	commitID, err := db.CreateCommit(runID, "abc123", "Test Author", "test@example.com", "Test Committer", "committer@example.com", commitDate, "Test commit", []string{})
	if err != nil {
		t.Fatalf("Failed to create commit: %v", err)
	}

	if commitID != 1 {
		t.Errorf("Expected commit ID 1, got %d", commitID)
	}

	// Test creating a file
	err = db.CreateFile(commitID, "/test/file.go", 1234, "100644", nil)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Test updating run counts
	err = db.UpdateRunItemCount(runID)
	if err != nil {
		t.Fatalf("Failed to update run counts: %v", err)
	}

	// Test finishing run
	err = db.FinishRun(runID, "completed")
	if err != nil {
		t.Fatalf("Failed to finish run: %v", err)
	}

	// Test getting all runs
	runs, err := db.GetAllRuns()
	if err != nil {
		t.Fatalf("Failed to get all runs: %v", err)
	}

	if len(runs) != 1 {
		t.Errorf("Expected 1 run, got %d", len(runs))
	}

	run := runs[0]
	if run.RepoPath != "/test/repo" {
		t.Errorf("Expected repo path '/test/repo', got '%s'", run.RepoPath)
	}

	if run.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", run.Status)
	}

	if run.RunType != "git" {
		t.Errorf("Expected run type 'git', got '%s'", run.RunType)
	}

	if run.ItemCount != 1 {
		t.Errorf("Expected 1 item, got %d", run.ItemCount)
	}

	if run.EndTime == nil {
		t.Error("Expected end time to be set")
	}
}

func TestMultipleRuns(t *testing.T) {
	// Create a temporary directory for test database
	tempDir := t.TempDir()

	// Mock the home directory for testing
	originalHomeDir := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHomeDir)

	// Open database
	db, err := Open()
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create multiple runs
	run1ID, err := db.CreateRun("/test/repo1", "git")
	if err != nil {
		t.Fatalf("Failed to create run 1: %v", err)
	}

	run2ID, err := db.CreateRun("/test/repo2", "fs")
	if err != nil {
		t.Fatalf("Failed to create run 2: %v", err)
	}

	// Finish runs
	db.FinishRun(run1ID, "completed")
	db.FinishRun(run2ID, "completed")

	// Get all runs
	runs, err := db.GetAllRuns()
	if err != nil {
		t.Fatalf("Failed to get all runs: %v", err)
	}

	if len(runs) != 2 {
		t.Errorf("Expected 2 runs, got %d", len(runs))
	}

	// Verify runs are ordered by start time descending
	if runs[0].ID < runs[1].ID {
		t.Error("Runs should be ordered by start time descending")
	}
}

func TestQuery(t *testing.T) {
	// Create a temporary directory for test database
	tempDir := t.TempDir()

	// Mock the home directory for testing
	originalHomeDir := os.Getenv("HOME")
	testHome := tempDir
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", originalHomeDir)

	// Open database
	db, err := Open()
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create test data
	runID, err := db.CreateRun("/test/repo", "git")
	if err != nil {
		t.Fatalf("Failed to create run: %v", err)
	}

	commitDate := time.Now()
	commitID, err := db.CreateCommit(runID, "abc123", "Test Author", "test@example.com", "Test Committer", "committer@example.com", commitDate, "Test commit", []string{})
	if err != nil {
		t.Fatalf("Failed to create commit: %v", err)
	}

	err = db.CreateFile(commitID, "/test/file.go", 1234, "100644", nil)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Test simple SELECT query
	results, err := db.Query("SELECT id, repo_path, run_type FROM runs")
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0]["repo_path"] != "/test/repo" {
		t.Errorf("Expected repo_path '/test/repo', got '%v'", results[0]["repo_path"])
	}

	if results[0]["run_type"] != "git" {
		t.Errorf("Expected run_type 'git', got '%v'", results[0]["run_type"])
	}

	// Test query with WHERE clause
	results, err = db.Query("SELECT hash, author FROM commits WHERE hash = 'abc123'")
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0]["hash"] != "abc123" {
		t.Errorf("Expected hash 'abc123', got '%v'", results[0]["hash"])
	}

	if results[0]["author"] != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%v'", results[0]["author"])
	}

	// Test query with COUNT
	results, err = db.Query("SELECT COUNT(*) as count FROM files")
	if err != nil {
		t.Fatalf("Failed to execute query: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	count, ok := results[0]["count"].(int64)
	if !ok {
		t.Errorf("Expected count to be int64, got %T", results[0]["count"])
	}

	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	// Test invalid query
	_, err = db.Query("SELECT * FROM nonexistent_table")
	if err == nil {
		t.Error("Expected error for invalid query, got nil")
	}
}

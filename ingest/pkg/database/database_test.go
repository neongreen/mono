package database

import (
	"os"
	"path/filepath"
	"strings"
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

func TestUpdateRunItemCountForCommandRun(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	db, err := Open()
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	cmdRunID, err := db.CreateRun("echo multiline", "cmd")
	if err != nil {
		t.Fatalf("Failed to create cmd run: %v", err)
	}

	stdout := "line one\nline two\n"
	stderr := "warning only\n"
	if err := db.CreateCmdRun(cmdRunID, "echo multiline", 0, stdout, stderr, 25); err != nil {
		t.Fatalf("Failed to create cmd run record: %v", err)
	}

	if err := db.UpdateRunItemCount(cmdRunID); err != nil {
		t.Fatalf("UpdateRunItemCount failed: %v", err)
	}

	var itemCount int
	if err := db.db.QueryRow("SELECT item_count FROM runs WHERE id = ?", cmdRunID).Scan(&itemCount); err != nil {
		t.Fatalf("Failed to fetch item count: %v", err)
	}
	if itemCount != 3 {
		t.Fatalf("expected item count 3 (stdout lines 2 + stderr lines 1), got %d", itemCount)
	}

	emptyRunID, err := db.CreateRun("true", "cmd")
	if err != nil {
		t.Fatalf("Failed to create empty cmd run: %v", err)
	}

	if err := db.CreateCmdRun(emptyRunID, "true", 0, "", "", 1); err != nil {
		t.Fatalf("Failed to insert empty cmd run: %v", err)
	}

	if err := db.UpdateRunItemCount(emptyRunID); err != nil {
		t.Fatalf("UpdateRunItemCount failed for empty run: %v", err)
	}

	if err := db.db.QueryRow("SELECT item_count FROM runs WHERE id = ?", emptyRunID).Scan(&itemCount); err != nil {
		t.Fatalf("Failed to fetch empty run item count: %v", err)
	}
	if itemCount != 0 {
		t.Fatalf("expected item count 0 for empty command output, got %d", itemCount)
	}
}

func TestGitHubOperations(t *testing.T) {
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

	// Test creating a GitHub run
	runID, err := db.CreateRun("owner/repo", "github")
	if err != nil {
		t.Fatalf("Failed to create run: %v", err)
	}

	if runID != 1 {
		t.Errorf("Expected run ID 1, got %d", runID)
	}

	// Test creating a GitHub issue
	now := time.Now()
	closedAt := now.Add(24 * time.Hour)
	err = db.CreateGitHubIssue(
		runID,
		1,
		"Test Issue",
		"Issue body",
		"closed",
		"testuser",
		now,
		now.Add(1*time.Hour),
		&closedAt,
		"bug,enhancement",
		"assignee1,assignee2",
		"v1.0",
	)
	if err != nil {
		t.Fatalf("Failed to create GitHub issue: %v", err)
	}

	// Test creating a GitHub PR
	mergedAt := now.Add(48 * time.Hour)
	err = db.CreateGitHubPR(
		runID,
		2,
		"Test PR",
		"PR body",
		"closed",
		"prauthor",
		now,
		now.Add(2*time.Hour),
		&closedAt,
		&mergedAt,
		true,
		false,
		"main",
		"feature-branch",
		"enhancement",
		"reviewer1",
		"reviewer2",
		"v1.0",
	)
	if err != nil {
		t.Fatalf("Failed to create GitHub PR: %v", err)
	}

	// Test creating a GitHub comment for issue
	err = db.CreateGitHubComment(
		runID,
		"issue",
		1,
		12345,
		"commenter1",
		"This is a comment",
		now,
		now.Add(30*time.Minute),
	)
	if err != nil {
		t.Fatalf("Failed to create GitHub issue comment: %v", err)
	}

	// Test creating a GitHub comment for PR
	err = db.CreateGitHubComment(
		runID,
		"pr",
		2,
		67890,
		"commenter2",
		"This is a PR comment",
		now,
		now.Add(45*time.Minute),
	)
	if err != nil {
		t.Fatalf("Failed to create GitHub PR comment: %v", err)
	}

	// Test UpdateRunItemCount for github type
	err = db.UpdateRunItemCount(runID)
	if err != nil {
		t.Fatalf("Failed to update run item count: %v", err)
	}

	// Verify the item count (1 issue + 1 PR = 2)
	runs, err := db.GetAllRuns()
	if err != nil {
		t.Fatalf("Failed to get runs: %v", err)
	}

	if len(runs) != 1 {
		t.Errorf("Expected 1 run, got %d", len(runs))
	}

	if runs[0].ItemCount != 2 {
		t.Errorf("Expected item count 2, got %d", runs[0].ItemCount)
	}

	// Test querying GitHub issues
	results, err := db.Query("SELECT number, title, state, labels FROM github_issues")
	if err != nil {
		t.Fatalf("Failed to query GitHub issues: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(results))
	}

	if results[0]["title"] != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got '%v'", results[0]["title"])
	}

	labels := results[0]["labels"].(string)
	if !strings.Contains(labels, "bug") || !strings.Contains(labels, "enhancement") {
		t.Errorf("Expected labels to contain 'bug' and 'enhancement', got '%s'", labels)
	}

	// Test querying GitHub PRs
	results, err = db.Query("SELECT number, title, merged, base_branch, head_branch FROM github_prs")
	if err != nil {
		t.Fatalf("Failed to query GitHub PRs: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 PR, got %d", len(results))
	}

	if results[0]["title"] != "Test PR" {
		t.Errorf("Expected title 'Test PR', got '%v'", results[0]["title"])
	}

	if results[0]["base_branch"] != "main" {
		t.Errorf("Expected base_branch 'main', got '%v'", results[0]["base_branch"])
	}

	// Test querying GitHub comments
	results, err = db.Query("SELECT item_type, item_number, author, body FROM github_comments ORDER BY item_type")
	if err != nil {
		t.Fatalf("Failed to query GitHub comments: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(results))
	}

	if results[0]["item_type"] != "issue" {
		t.Errorf("Expected item_type 'issue', got '%v'", results[0]["item_type"])
	}

	if results[1]["item_type"] != "pr" {
		t.Errorf("Expected item_type 'pr', got '%v'", results[1]["item_type"])
	}
}

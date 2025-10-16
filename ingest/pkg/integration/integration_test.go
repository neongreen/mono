package integration

import (
	"ingest/pkg/database"
	"ingest/pkg/git"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitIngestion(t *testing.T) {
	// Create a temporary directory for test repository
	tempDir := t.TempDir()
	testRepo := filepath.Join(tempDir, "testrepo")

	// Initialize a git repository
	if err := os.Mkdir(testRepo, 0755); err != nil {
		t.Fatalf("Failed to create test repo directory: %v", err)
	}

	// Initialize git
	cmd := exec.Command("git", "init")
	cmd.Dir = testRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init git: %v", err)
	}

	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = testRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git email: %v", err)
	}

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = testRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to configure git name: %v", err)
	}

	// Create a test file and commit it
	testFile := filepath.Join(testRepo, "test.txt")
	if err := os.WriteFile(testFile, []byte("Hello, World!"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = testRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = testRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Create a second commit
	testFile2 := filepath.Join(testRepo, "test2.txt")
	if err := os.WriteFile(testFile2, []byte("Second file"), 0644); err != nil {
		t.Fatalf("Failed to create second test file: %v", err)
	}

	cmd = exec.Command("git", "add", "test2.txt")
	cmd.Dir = testRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add second file: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "Second commit")
	cmd.Dir = testRepo
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create second commit: %v", err)
	}

	// Mock the home directory for testing
	originalHomeDir := os.Getenv("HOME")
	testHome := filepath.Join(tempDir, "home")
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", originalHomeDir)

	// Open database
	db, err := database.Open()
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create a run
	runID, err := db.CreateRun(testRepo, "git")
	if err != nil {
		t.Fatalf("Failed to create run: %v", err)
	}

	// Walk the repository
	commits, err := git.WalkRepository(testRepo, nil)
	if err != nil {
		t.Fatalf("Failed to walk repository: %v", err)
	}

	// Should have 2 commits
	if len(commits) != 2 {
		t.Errorf("Expected 2 commits, got %d", len(commits))
	}

	// Store commits
	for _, commit := range commits {
		commitID, err := db.CreateCommit(
			runID,
			commit.Hash,
			commit.Author,
			commit.AuthorEmail,
			commit.Committer,
			commit.CommitterEmail,
			commit.Date,
			commit.Message,
			commit.ParentHashes,
		)
		if err != nil {
			t.Fatalf("Failed to create commit: %v", err)
		}

		for _, file := range commit.Files {
			var blobID *int64
			if len(file.Content) > 0 {
				id, err := db.GetOrCreateBlob(file.Content, file.SHA256Hash)
				if err != nil {
					t.Fatalf("Failed to create blob: %v", err)
				}
				blobID = &id
			}

			err := db.CreateFile(commitID, file.Path, file.Size, file.Mode, blobID)
			if err != nil {
				t.Fatalf("Failed to create file: %v", err)
			}
		}
	}

	// Update run counts
	if err := db.UpdateRunItemCount(runID); err != nil {
		t.Fatalf("Failed to update run counts: %v", err)
	}

	// Finish run
	if err := db.FinishRun(runID, "completed"); err != nil {
		t.Fatalf("Failed to finish run: %v", err)
	}

	// Verify the data
	runs, err := db.GetAllRuns()
	if err != nil {
		t.Fatalf("Failed to get runs: %v", err)
	}

	if len(runs) != 1 {
		t.Errorf("Expected 1 run, got %d", len(runs))
	}

	run := runs[0]
	if run.RunType != "git" {
		t.Errorf("Expected run type 'git', got '%s'", run.RunType)
	}

	if run.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", run.Status)
	}

	if run.ItemCount != 2 {
		t.Errorf("Expected 2 items (commits), got %d", run.ItemCount)
	}

	if run.EndTime == nil {
		t.Error("Expected end time to be set")
	}
}

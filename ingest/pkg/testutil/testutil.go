package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"ingest/pkg/database"
)

// GitCommit describes files to write for a commit and the commit message.
type GitCommit struct {
	Message string
	Files   map[string]string
}

// WithTempHome sets HOME to a temporary directory owned by the test and returns the directory.
func WithTempHome(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	home := filepath.Join(root, "home")
	t.Setenv("HOME", home)
	return home
}

// NewTempGitRepo creates a temporary git repository populated with the provided commits.
func NewTempGitRepo(t *testing.T, commits []GitCommit) string {
	t.Helper()

	repoDir := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoDir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", strings.Join(args, " "), err, output)
		}
	}

	runGit("init")
	runGit("config", "user.email", "test@example.com")
	runGit("config", "user.name", "Test User")

	for _, commit := range commits {
		for path, content := range commit.Files {
			fullPath := filepath.Join(repoDir, path)
			if strings.HasSuffix(path, "/") {
				if err := os.MkdirAll(fullPath, 0o755); err != nil {
					t.Fatalf("mkdir %s: %v", path, err)
				}
				continue
			}
			if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
				t.Fatalf("mkdir parent for %s: %v", path, err)
			}
			if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
				t.Fatalf("write %s: %v", path, err)
			}
		}

		runGit("add", "-A")
		runGit("commit", "-m", commit.Message)
	}

	return repoDir
}

// WriteFiles materialises the provided relative path -> content mapping under root.
// Paths ending with "/" create directories.
func WriteFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()

	for rel, content := range files {
		fullPath := filepath.Join(root, rel)
		if strings.HasSuffix(rel, "/") {
			if err := os.MkdirAll(fullPath, 0o755); err != nil {
				t.Fatalf("mkdir %s: %v", rel, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("mkdir parent for %s: %v", rel, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
}

// OpenDatabase opens the ingest database using database.Open and registers cleanup.
func OpenDatabase(t *testing.T) *database.Database {
	t.Helper()

	db, err := database.Open()
	if err != nil {
		t.Fatalf("database.Open(): %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("db.Close(): %v", err)
		}
	})

	return db
}

// AssertSingleRun verifies that a single completed run exists with expected characteristics.
func AssertSingleRun(t *testing.T, db *database.Database, runType string, expectedItems int) database.Run {
	t.Helper()

	runs, err := db.GetAllRuns()
	if err != nil {
		t.Fatalf("GetAllRuns(): %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	run := runs[0]
	if run.RunType != runType {
		t.Fatalf("expected run type %q, got %q", runType, run.RunType)
	}
	if run.Status != "completed" {
		t.Fatalf("expected status completed, got %s", run.Status)
	}
	if run.EndTime == nil {
		t.Fatal("expected run end time to be set")
	}
	if run.ItemCount != expectedItems {
		t.Fatalf("expected item count %d, got %d", expectedItems, run.ItemCount)
	}

	return run
}

// CountRows executes a query returning a single integer column and returns it as int.
func CountRows(t *testing.T, db *database.Database, query string) int {
	t.Helper()

	results, err := db.Query(query)
	if err != nil {
		t.Fatalf("db.Query(%q): %v", query, err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 row, got %d", len(results))
	}

	switch value := results[0][resultsColumn(results[0])].(type) {
	case int64:
		return int(value)
	case int:
		return value
	default:
		t.Fatalf("expected integer result, got %#v", results[0])
		return 0
	}
}

func resultsColumn(row map[string]any) string {
	for name := range row {
		return name
	}
	return ""
}

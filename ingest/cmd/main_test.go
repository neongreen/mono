package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"ingest/pkg/database"
)

func runCLI(t *testing.T, args ...string) (string, string) {
	t.Helper()

	cmd := newRootCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute %v: %v\nstderr: %s", args, err, stderr.String())
	}

	return stdout.String(), stderr.String()
}

func createGitRepo(t *testing.T, repoPath string) {
	t.Helper()

	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}

	for _, cfg := range [][2]string{
		{"user.email", "test@example.com"},
		{"user.name", "Test User"},
	} {
		cmd = exec.Command("git", "config", cfg[0], cfg[1])
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			t.Fatalf("git config %s: %v", cfg[0], err)
		}
	}

	if err := os.WriteFile(filepath.Join(repoPath, "a.txt"), []byte("first"), 0o644); err != nil {
		t.Fatalf("write a.txt: %v", err)
	}

	cmd = exec.Command("git", "add", "a.txt")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("git add: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "first")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("git commit first: %v", err)
	}

	if err := os.WriteFile(filepath.Join(repoPath, "b.txt"), []byte("second"), 0o644); err != nil {
		t.Fatalf("write b.txt: %v", err)
	}

	cmd = exec.Command("git", "add", "b.txt")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("git add second: %v", err)
	}

	cmd = exec.Command("git", "commit", "-m", "second")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("git commit second: %v", err)
	}
}

func TestCLI_GitIngestion(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	t.Setenv("HOME", homeDir)

	repoPath := filepath.Join(tempDir, "repo")
	createGitRepo(t, repoPath)

	runCLI(t, "git", repoPath)

	db, err := database.Open()
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	runs, err := db.GetAllRuns()
	if err != nil {
		t.Fatalf("get runs: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].RunType != "git" {
		t.Fatalf("expected run type git, got %s", runs[0].RunType)
	}
	if runs[0].ItemCount != 2 {
		t.Fatalf("expected 2 commits, got %d", runs[0].ItemCount)
	}

	results, err := db.Query("SELECT COUNT(*) AS count FROM commits")
	if err != nil {
		t.Fatalf("query commits: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result row, got %d", len(results))
	}
	count, ok := results[0]["count"].(int64)
	if !ok {
		t.Fatalf("commit count not int64: %#v", results[0]["count"])
	}
	if count != 2 {
		t.Fatalf("expected 2 commits stored, got %d", count)
	}
}

func TestCLI_FSRespectGitignoreFlag(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	t.Setenv("HOME", homeDir)

	workdir := filepath.Join(tempDir, "workspace")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}

	if err := os.WriteFile(filepath.Join(workdir, ".gitignore"), []byte("ignored.txt\n"), 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workdir, "included.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write included.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workdir, "ignored.txt"), []byte("skip"), 0o644); err != nil {
		t.Fatalf("write ignored.txt: %v", err)
	}

	runCLI(t, "fs", workdir)

	db, err := database.Open()
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	runs, err := db.GetAllRuns()
	if err != nil {
		t.Fatalf("get runs: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	results, err := db.Query("SELECT path FROM fs_entries ORDER BY path")
	if err != nil {
		t.Fatalf("query fs_entries: %v", err)
	}

	var paths []string
	for _, row := range results {
		path, ok := row["path"].(string)
		if !ok {
			t.Fatalf("path column not string: %#v", row["path"])
		}
		paths = append(paths, path)
	}

	for _, path := range paths {
		if path == "ignored.txt" {
			t.Fatalf("ignored file was ingested despite respect-gitignore being true")
		}
	}

	if want := len(paths); runs[0].ItemCount != want {
		t.Fatalf("expected run item count %d, got %d", want, runs[0].ItemCount)
	}
}

func TestCLI_FSCanDisableGitignoreFiltering(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	t.Setenv("HOME", homeDir)

	workdir := filepath.Join(tempDir, "workspace")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}

	if err := os.WriteFile(filepath.Join(workdir, ".gitignore"), []byte("ignored.txt\n"), 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workdir, "ignored.txt"), []byte("skip"), 0o644); err != nil {
		t.Fatalf("write ignored.txt: %v", err)
	}

	runCLI(t, "fs", "--respect-gitignore=false", workdir)

	db, err := database.Open()
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	results, err := db.Query("SELECT path FROM fs_entries WHERE path = 'ignored.txt'")
	if err != nil {
		t.Fatalf("query fs_entries: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected ignored.txt to be ingested when respect-gitignore is false")
	}
}

func TestCLI_CommandIngestionCountsOutputLines(t *testing.T) {
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	t.Setenv("HOME", homeDir)

	runCLI(t, "cmd", "printf 'line1\\nline2\\n'")

	db, err := database.Open()
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	runs, err := db.GetAllRuns()
	if err != nil {
		t.Fatalf("get runs: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].ItemCount != 2 {
		t.Fatalf("expected item count 2 (lines), got %d", runs[0].ItemCount)
	}

	results, err := db.Query("SELECT stdout FROM cmd_runs")
	if err != nil {
		t.Fatalf("query cmd_runs: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 cmd run row, got %d", len(results))
	}
	stdout, ok := results[0]["stdout"].(string)
	if !ok {
		t.Fatalf("stdout column not string: %#v", results[0]["stdout"])
	}
	if stdout != "line1\nline2\n" {
		t.Fatalf("unexpected stdout stored: %q", stdout)
	}
}

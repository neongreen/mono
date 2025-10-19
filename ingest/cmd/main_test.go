package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ingest/pkg/testutil"
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

func runCLIExpectError(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	cmd := newRootCmd()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestCLI_GitIngestion(t *testing.T) {
	testutil.WithTempHome(t)

	repoPath := testutil.NewTempGitRepo(t, []testutil.GitCommit{
		{Message: "first", Files: map[string]string{"a.txt": "first"}},
		{Message: "second", Files: map[string]string{"b.txt": "second"}},
	})

	runCLI(t, "git", repoPath)

	db := testutil.OpenDatabase(t)

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
	testutil.WithTempHome(t)

	workdir := t.TempDir()
	testutil.WriteFiles(t, workdir, map[string]string{
		".gitignore":   "ignored.txt\n",
		"included.txt": "hello",
		"ignored.txt":  "skip",
	})

	runCLI(t, "fs", workdir)

	db := testutil.OpenDatabase(t)

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
	testutil.WithTempHome(t)

	workdir := t.TempDir()
	testutil.WriteFiles(t, workdir, map[string]string{
		".gitignore":  "ignored.txt\n",
		"ignored.txt": "skip",
	})

	runCLI(t, "fs", "--respect-gitignore=false", workdir)

	db := testutil.OpenDatabase(t)

	results, err := db.Query("SELECT path FROM fs_entries WHERE path = 'ignored.txt'")
	if err != nil {
		t.Fatalf("query fs_entries: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected ignored.txt to be ingested when respect-gitignore is false")
	}
}

func TestCLI_CommandIngestionCountsOutputLines(t *testing.T) {
	testutil.WithTempHome(t)

	runCLI(t, "cmd", "printf 'line1\\nline2\\n'")

	db := testutil.OpenDatabase(t)

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

func TestCLI_ConfigValidateSuccess(t *testing.T) {
	testutil.WithTempHome(t)

	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "ingest.config.toml")
	config := `
[[job]]
type = "command"
command = "printf 'hi'"
`
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	stdout, _ := runCLI(t, "config", "validate", "--config", configPath)

	if !strings.Contains(stdout, "is valid") {
		t.Fatalf("expected success message, got %q", stdout)
	}
	if !strings.Contains(stdout, "job 1") {
		t.Fatalf("expected job listing in output, got %q", stdout)
	}
}

func TestCLI_ConfigValidateFailure(t *testing.T) {
	testutil.WithTempHome(t)

	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "bad-config.toml")
	config := `
[[job]]
type = "github"
owner = "octo"
`
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, _, err := runCLIExpectError(t, "config", "validate", "--config", configPath)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "owner") {
		t.Fatalf("expected error about missing repo, got %v", err)
	}
}

func TestCLI_ConfigValidateWarnsForMissingMCPToken(t *testing.T) {
	testutil.WithTempHome(t)

	tmp := t.TempDir()
	configPath := filepath.Join(tmp, "warn-config.toml")
	config := `
[[job]]
type = "github_mcp"
owner = "octo"
repo = "mono"

  [job.mcp]
  endpoint = "https://example.com/mcp"
`
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("INGEST_GITHUB_MCP_TOKEN", "")
	t.Setenv("INGEST_MCP_TOKEN", "")
	t.Setenv("MISE_GITHUB_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")

	stdout, _ := runCLI(t, "config", "validate", "--config", configPath)

	if !strings.Contains(stdout, "Warnings:") {
		t.Fatalf("expected warnings section, got %q", stdout)
	}
	if !strings.Contains(stdout, "no MCP token resolved") {
		t.Fatalf("expected token warning, got %q", stdout)
	}
	if !strings.Contains(stdout, "MISE_GITHUB_TOKEN") {
		t.Fatalf("expected hint mentioning MISE_GITHUB_TOKEN, got %q", stdout)
	}
}

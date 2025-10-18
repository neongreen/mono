package runconfig

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	config := `
parallelism = 3

[[job]]
name = "fs"
type = "fs"
path = "./fixtures"
respect_gitignore = false

[[job]]
type = "github"
owner = "octo"
repo = "example"
`

	cfg, err := Parse([]byte(config))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cfg.Parallelism != 3 {
		t.Fatalf("expected parallelism 3, got %d", cfg.Parallelism)
	}

	if len(cfg.Jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(cfg.Jobs))
	}

	fsJob := cfg.Jobs[0]
	if fsJob.Type != "fs" {
		t.Fatalf("expected fs job, got %s", fsJob.Type)
	}
	if fsJob.RespectGitignore == nil || *fsJob.RespectGitignore != false {
		t.Fatalf("expected respect_gitignore=false, got %+v", fsJob.RespectGitignore)
	}

	ghJob := cfg.Jobs[1]
	if ghJob.Owner != "octo" || ghJob.Repo != "example" {
		t.Fatalf("unexpected github job: owner=%s repo=%s", ghJob.Owner, ghJob.Repo)
	}
}

func TestExecuteContinuesOnError(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	cfg := Config{
		Parallelism: 2,
		Jobs: []JobConfig{
			{
				Name:    "ok",
				Type:    "command",
				Command: "printf 'hello'",
			},
			{
				Name: "missing",
				Type: "git",
				Path: filepath.Join(tempDir, "does-not-exist"),
			},
		},
	}

	results, err := Execute(ctx, io.Discard, cfg)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Err != nil {
		t.Fatalf("expected first job to succeed, got error: %v", results[0].Err)
	}
	if results[1].Err == nil {
		t.Fatalf("expected second job to fail")
	}

	if err == nil {
		t.Fatalf("expected aggregated error but got nil")
	}

	if !strings.Contains(err.Error(), "missing") {
		t.Fatalf("expected error message to reference failing job, got %v", err)
	}

	// Ensure database writes ended up under the temp HOME directory.
	if _, statErr := os.Stat(filepath.Join(tempDir, ".ingest")); statErr != nil {
		t.Fatalf("expected .ingest directory in temp home: %v", statErr)
	}
}

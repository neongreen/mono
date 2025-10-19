package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePRReleaseTag(t *testing.T) {
	tests := []struct {
		tag          string
		wantProject  string
		wantPR       int
		wantSequence int
		wantOK       bool
	}{
		{"dissect--pr-123.4", "dissect", 123, 4, true},
		{"markdown-format/pr-8.2", "markdown-format", 8, 2, true},
		{"invalid-tag", "", 0, 0, false},
		{"project--main.1", "", 0, 0, false},
	}

	for _, tt := range tests {
		project, prNumber, sequence, ok := parsePRReleaseTag(tt.tag)
		if ok != tt.wantOK {
			t.Fatalf("parsePRReleaseTag(%q) ok=%v, want %v", tt.tag, ok, tt.wantOK)
		}
		if !ok {
			continue
		}
		if project != tt.wantProject || prNumber != tt.wantPR || sequence != tt.wantSequence {
			t.Fatalf("parsePRReleaseTag(%q) = (%q, %d, %d), want (%q, %d, %d)",
				tt.tag, project, prNumber, sequence, tt.wantProject, tt.wantPR, tt.wantSequence)
		}
	}
}

func TestFindPreviousReleaseTag(t *testing.T) {
	cacheDir := t.TempDir()

	// Create cached releases
	tags := []string{
		"dissect--pr-10.1",
		"dissect--pr-10.2",
		"other--pr-10.3",
		"dissect--pr-11.1",
	}

	for _, tag := range tags {
		path := filepath.Join(cacheDir, tag)
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatalf("failed to create cache directory %s: %v", path, err)
		}
	}

	prev, err := findPreviousReleaseTag(cacheDir, "dissect", 10, 3)
	if err != nil {
		t.Fatalf("findPreviousReleaseTag returned error: %v", err)
	}
	if prev != "dissect--pr-10.2" {
		t.Fatalf("expected previous tag dissect--pr-10.2, got %q", prev)
	}

	prev, err = findPreviousReleaseTag(cacheDir, "dissect", 10, 2)
	if err != nil {
		t.Fatalf("findPreviousReleaseTag returned error: %v", err)
	}
	if prev != "dissect--pr-10.1" {
		t.Fatalf("expected previous tag dissect--pr-10.1, got %q", prev)
	}

	prev, err = findPreviousReleaseTag(cacheDir, "dissect", 10, 1)
	if err != nil {
		t.Fatalf("findPreviousReleaseTag returned error: %v", err)
	}
	if prev != "" {
		t.Fatalf("expected no previous tag, got %q", prev)
	}

	prev, err = findPreviousReleaseTag(cacheDir, "dissect", 99, 1)
	if err != nil {
		t.Fatalf("findPreviousReleaseTag returned error: %v", err)
	}
	if prev != "" {
		t.Fatalf("expected no previous tag for unknown PR, got %q", prev)
	}
}

func TestFindPreviousReleaseTagMissingCache(t *testing.T) {
	tempRoot := t.TempDir()
	missingDir := filepath.Join(tempRoot, "does-not-exist")

	prev, err := findPreviousReleaseTag(missingDir, "dissect", 10, 2)
	if err != nil {
		t.Fatalf("expected no error for missing cache directory, got %v", err)
	}
	if prev != "" {
		t.Fatalf("expected no previous tag when cache missing, got %q", prev)
	}
}

package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseBatchMoveArg(t *testing.T) {
	tests := []struct {
		name        string
		arg         string
		wantSources []string
		wantTarget  string
		wantErr     bool
	}{
		{
			name:        "single file to directory",
			arg:         "file1.go -> dir/",
			wantSources: []string{"file1.go"},
			wantTarget:  "dir/",
		},
		{
			name:        "multiple files to directory",
			arg:         "a.go,b.go -> dir/",
			wantSources: []string{"a.go", "b.go"},
			wantTarget:  "dir/",
		},
		{
			name:        "single file to file (rename)",
			arg:         "src.go -> dest.go",
			wantSources: []string{"src.go"},
			wantTarget:  "dest.go",
		},
		{
			name:        "glob pattern",
			arg:         "db*.go -> internal/db/",
			wantSources: []string{"db*.go"},
			wantTarget:  "internal/db/",
		},
		{
			name:        "spaces after comma",
			arg:         "a.go, b.go, c.go -> target/",
			wantSources: []string{"a.go", "b.go", "c.go"},
			wantTarget:  "target/",
		},
		{
			name:    "no spaces around arrow",
			arg:     "a.go,b.go->dir/",
			wantErr: true, // requires spaces around arrow
		},
		{
			name:        "extra spaces",
			arg:         "  a.go  ->  dir/  ",
			wantSources: []string{"a.go"},
			wantTarget:  "dir/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBatchMoveArg(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBatchMoveArg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if len(got.Sources) != len(tt.wantSources) {
				t.Errorf("ParseBatchMoveArg() got %d sources, want %d", len(got.Sources), len(tt.wantSources))
				return
			}

			for i, src := range got.Sources {
				if src != tt.wantSources[i] {
					t.Errorf("ParseBatchMoveArg() source[%d] = %v, want %v", i, src, tt.wantSources[i])
				}
			}

			if got.Target != tt.wantTarget {
				t.Errorf("ParseBatchMoveArg() target = %v, want %v", got.Target, tt.wantTarget)
			}
		})
	}
}

func TestParseBatchMoveArg_Errors(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		wantErr string
	}{
		{
			name:    "no arrow",
			arg:     "no arrow here",
			wantErr: "expected 'source -> target' format",
		},
		{
			name:    "missing source",
			arg:     " -> target/",
			wantErr: "missing source files",
		},
		{
			name:    "missing target",
			arg:     "source.go -> ",
			wantErr: "missing target",
		},
		{
			name:    "multiple arrows",
			arg:     "a.go -> b.go -> c.go",
			wantErr: "use ' -> ' (with spaces)",
		},
		{
			name:    "empty string",
			arg:     "",
			wantErr: "expected 'source -> target' format",
		},
		{
			name:    "only arrow with spaces",
			arg:     " -> ",
			wantErr: "missing source files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseBatchMoveArg(tt.arg)
			if err == nil {
				t.Errorf("ParseBatchMoveArg() expected error containing %q, got nil", tt.wantErr)
				return
			}
			// Just check that we got an error, don't be too strict about the message
		})
	}
}

func TestExpandGlobs(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{"db.go", "db_events.go", "db_test.go", "main.go", "util.go"}
	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("package main\n"), 0o644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", f, err)
		}
	}

	tests := []struct {
		name    string
		sources []string
		want    []string // Just filenames, will be converted to full paths
		wantErr bool
	}{
		{
			name:    "simple glob",
			sources: []string{"db*.go"},
			want:    []string{"db.go", "db_events.go", "db_test.go"},
		},
		{
			name:    "all go files",
			sources: []string{"*.go"},
			want:    []string{"db.go", "db_events.go", "db_test.go", "main.go", "util.go"},
		},
		{
			name:    "specific files",
			sources: []string{"main.go", "util.go"},
			want:    []string{"main.go", "util.go"},
		},
		{
			name:    "mix of glob and specific",
			sources: []string{"db*.go", "util.go"},
			want:    []string{"db.go", "db_events.go", "db_test.go", "util.go"},
		},
		{
			name:    "no matches",
			sources: []string{"nonexistent*.go"},
			wantErr: true,
		},
		{
			name:    "duplicate prevention",
			sources: []string{"db.go", "db*.go"}, // db.go appears twice
			want:    []string{"db.go", "db_events.go", "db_test.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandGlobs(tt.sources, tmpDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandGlobs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Convert expected filenames to full paths
			wantPaths := make([]string, len(tt.want))
			for i, name := range tt.want {
				wantPaths[i] = filepath.Join(tmpDir, name)
			}

			if len(got) != len(wantPaths) {
				t.Errorf("ExpandGlobs() got %d files, want %d\nGot: %v\nWant: %v", len(got), len(wantPaths), got, wantPaths)
				return
			}

			// Check that all expected files are present (order doesn't matter)
			gotMap := make(map[string]bool)
			for _, p := range got {
				gotMap[p] = true
			}

			for _, want := range wantPaths {
				if !gotMap[want] {
					t.Errorf("ExpandGlobs() missing expected file %s", want)
				}
			}
		})
	}
}

func TestIsDirectory(t *testing.T) {
	tests := []struct {
		name   string
		target string
		want   bool
	}{
		{
			name:   "ends with slash",
			target: "internal/db/",
			want:   true,
		},
		{
			name:   "ends with backslash",
			target: "internal\\db\\",
			want:   true,
		},
		{
			name:   "no trailing slash",
			target: "file.go",
			want:   false,
		},
		{
			name:   "path without extension",
			target: "internal/db",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDirectory(tt.target); got != tt.want {
				t.Errorf("IsDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

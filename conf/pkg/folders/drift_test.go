package folders

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectDriftWithExcludes(t *testing.T) {
	// Create temporary directories for source and conf
	sourceDir, err := os.MkdirTemp("", "source")
	if err != nil {
		t.Fatalf("failed to create source temp dir: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	confDir, err := os.MkdirTemp("", "conf")
	if err != nil {
		t.Fatalf("failed to create conf temp dir: %v", err)
	}
	defer os.RemoveAll(confDir)

	// Create test files in source
	testFiles := []struct {
		path    string
		content string
	}{
		{"file1.txt", "content1"},
		{"file2.tmp", "temp file"},
		{".DS_Store", "ds store"},
		{"subdir/file3.txt", "content3"},
		{"subdir/file4.tmp", "temp file 4"},
	}

	for _, tf := range testFiles {
		fullPath := filepath.Join(sourceDir, tf.path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(tf.content), 0o644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	t.Run("no_excludes_detects_all_files", func(t *testing.T) {
		drifts, err := DetectDriftWithExcludes(sourceDir, confDir, nil)
		if err != nil {
			t.Fatalf("DetectDriftWithExcludes failed: %v", err)
		}

		// Should detect 5 files (4 files + 1 directory)
		if len(drifts) < 4 {
			t.Errorf("expected at least 4 drifts, got %d", len(drifts))
		}
	})

	t.Run("exclude_tmp_files", func(t *testing.T) {
		drifts, err := DetectDriftWithExcludes(sourceDir, confDir, []string{"*.tmp"})
		if err != nil {
			t.Fatalf("DetectDriftWithExcludes failed: %v", err)
		}

		// Check that no .tmp files are in the drift list
		for _, drift := range drifts {
			if filepath.Ext(drift.RelPath) == ".tmp" {
				t.Errorf("expected .tmp files to be excluded, but found: %s", drift.RelPath)
			}
		}
	})

	t.Run("exclude_DS_Store", func(t *testing.T) {
		drifts, err := DetectDriftWithExcludes(sourceDir, confDir, []string{".DS_Store"})
		if err != nil {
			t.Fatalf("DetectDriftWithExcludes failed: %v", err)
		}

		// Check that .DS_Store is not in the drift list
		for _, drift := range drifts {
			if filepath.Base(drift.RelPath) == ".DS_Store" {
				t.Errorf("expected .DS_Store to be excluded, but found: %s", drift.RelPath)
			}
		}
	})

	t.Run("multiple_exclude_patterns", func(t *testing.T) {
		drifts, err := DetectDriftWithExcludes(sourceDir, confDir, []string{"*.tmp", ".DS_Store"})
		if err != nil {
			t.Fatalf("DetectDriftWithExcludes failed: %v", err)
		}

		// Check that neither .tmp files nor .DS_Store are in the drift list
		for _, drift := range drifts {
			base := filepath.Base(drift.RelPath)
			if filepath.Ext(drift.RelPath) == ".tmp" {
				t.Errorf("expected .tmp files to be excluded, but found: %s", drift.RelPath)
			}
			if base == ".DS_Store" {
				t.Errorf("expected .DS_Store to be excluded, but found: %s", drift.RelPath)
			}
		}
	})
}

func TestShouldExclude(t *testing.T) {
	tests := []struct {
		name            string
		relPath         string
		baseName        string
		excludePatterns []string
		expected        bool
	}{
		{
			name:            "no patterns",
			relPath:         "file.txt",
			baseName:        "file.txt",
			excludePatterns: nil,
			expected:        false,
		},
		{
			name:            "match wildcard pattern",
			relPath:         "file.tmp",
			baseName:        "file.tmp",
			excludePatterns: []string{"*.tmp"},
			expected:        true,
		},
		{
			name:            "no match wildcard pattern",
			relPath:         "file.txt",
			baseName:        "file.txt",
			excludePatterns: []string{"*.tmp"},
			expected:        false,
		},
		{
			name:            "match exact pattern",
			relPath:         ".DS_Store",
			baseName:        ".DS_Store",
			excludePatterns: []string{".DS_Store"},
			expected:        true,
		},
		{
			name:            "match one of multiple patterns",
			relPath:         "file.tmp",
			baseName:        "file.tmp",
			excludePatterns: []string{".DS_Store", "*.tmp", "*.bak"},
			expected:        true,
		},
		{
			name:            "match none of multiple patterns",
			relPath:         "file.txt",
			baseName:        "file.txt",
			excludePatterns: []string{".DS_Store", "*.tmp", "*.bak"},
			expected:        false,
		},
		{
			name:            "match subdirectory file by base name",
			relPath:         "subdir/file.tmp",
			baseName:        "file.tmp",
			excludePatterns: []string{"*.tmp"},
			expected:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldExclude(tt.relPath, tt.baseName, tt.excludePatterns)
			if result != tt.expected {
				t.Errorf("shouldExclude(%q, %q, %v) = %v, want %v",
					tt.relPath, tt.baseName, tt.excludePatterns, result, tt.expected)
			}
		})
	}
}

func TestDetectDrift(t *testing.T) {
	// Create temporary directories for source and conf
	sourceDir, err := os.MkdirTemp("", "source")
	if err != nil {
		t.Fatalf("failed to create source temp dir: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	confDir, err := os.MkdirTemp("", "conf")
	if err != nil {
		t.Fatalf("failed to create conf temp dir: %v", err)
	}
	defer os.RemoveAll(confDir)

	t.Run("empty_directories", func(t *testing.T) {
		drifts, err := DetectDrift(sourceDir, confDir)
		if err != nil {
			t.Fatalf("DetectDrift failed: %v", err)
		}
		if len(drifts) != 0 {
			t.Errorf("expected 0 drifts for empty directories, got %d", len(drifts))
		}
	})

	t.Run("added_file", func(t *testing.T) {
		// Create file in source only
		if err := os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("new"), 0o644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		drifts, err := DetectDrift(sourceDir, confDir)
		if err != nil {
			t.Fatalf("DetectDrift failed: %v", err)
		}

		// Should detect 1 added file
		found := false
		for _, drift := range drifts {
			if drift.RelPath == "new.txt" && drift.Status == StatusAdded {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find added file 'new.txt'")
		}

		// Clean up
		os.Remove(filepath.Join(sourceDir, "new.txt"))
	})

	t.Run("deleted_file", func(t *testing.T) {
		// Create file in conf only
		if err := os.WriteFile(filepath.Join(confDir, "deleted.txt"), []byte("deleted"), 0o644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		drifts, err := DetectDrift(sourceDir, confDir)
		if err != nil {
			t.Fatalf("DetectDrift failed: %v", err)
		}

		// Should detect 1 deleted file
		found := false
		for _, drift := range drifts {
			if drift.RelPath == "deleted.txt" && drift.Status == StatusDeleted {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find deleted file 'deleted.txt'")
		}

		// Clean up
		os.Remove(filepath.Join(confDir, "deleted.txt"))
	})

	t.Run("modified_file", func(t *testing.T) {
		// Create file in both with different content
		if err := os.WriteFile(filepath.Join(sourceDir, "modified.txt"), []byte("source content"), 0o644); err != nil {
			t.Fatalf("failed to write source file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(confDir, "modified.txt"), []byte("conf content"), 0o644); err != nil {
			t.Fatalf("failed to write conf file: %v", err)
		}

		drifts, err := DetectDrift(sourceDir, confDir)
		if err != nil {
			t.Fatalf("DetectDrift failed: %v", err)
		}

		// Should detect 1 modified file
		found := false
		for _, drift := range drifts {
			if drift.RelPath == "modified.txt" && drift.Status == StatusModified {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find modified file 'modified.txt'")
		}

		// Clean up
		os.Remove(filepath.Join(sourceDir, "modified.txt"))
		os.Remove(filepath.Join(confDir, "modified.txt"))
	})

	t.Run("in_sync_file", func(t *testing.T) {
		// Create file in both with same content
		content := []byte("same content")
		if err := os.WriteFile(filepath.Join(sourceDir, "same.txt"), content, 0o644); err != nil {
			t.Fatalf("failed to write source file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(confDir, "same.txt"), content, 0o644); err != nil {
			t.Fatalf("failed to write conf file: %v", err)
		}

		drifts, err := DetectDrift(sourceDir, confDir)
		if err != nil {
			t.Fatalf("DetectDrift failed: %v", err)
		}

		// Should not include in-sync file in drifts
		for _, drift := range drifts {
			if drift.RelPath == "same.txt" {
				t.Errorf("expected in-sync file 'same.txt' to not be in drifts")
			}
		}

		// Clean up
		os.Remove(filepath.Join(sourceDir, "same.txt"))
		os.Remove(filepath.Join(confDir, "same.txt"))
	})
}

func TestFormatDriftSummary(t *testing.T) {
	tests := []struct {
		name     string
		drifts   []FileDrift
		contains []string
	}{
		{
			name:     "empty drifts",
			drifts:   nil,
			contains: []string{"No drift detected"},
		},
		{
			name: "modified files only",
			drifts: []FileDrift{
				{RelPath: "file1.txt", Status: StatusModified},
				{RelPath: "file2.txt", Status: StatusModified},
			},
			contains: []string{"2 files with drift", "2 modified"},
		},
		{
			name: "added files only",
			drifts: []FileDrift{
				{RelPath: "file1.txt", Status: StatusAdded},
			},
			contains: []string{"1 files with drift", "1 added"},
		},
		{
			name: "deleted files only",
			drifts: []FileDrift{
				{RelPath: "file1.txt", Status: StatusDeleted},
				{RelPath: "file2.txt", Status: StatusDeleted},
				{RelPath: "file3.txt", Status: StatusDeleted},
			},
			contains: []string{"3 files with drift", "3 deleted"},
		},
		{
			name: "mixed statuses",
			drifts: []FileDrift{
				{RelPath: "file1.txt", Status: StatusModified},
				{RelPath: "file2.txt", Status: StatusAdded},
				{RelPath: "file3.txt", Status: StatusDeleted},
			},
			contains: []string{"3 files with drift", "1 modified", "1 added", "1 deleted"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDriftSummary(tt.drifts)
			for _, substr := range tt.contains {
				if !containsString(result, substr) {
					t.Errorf("FormatDriftSummary(%v) = %q, want to contain %q", tt.drifts, result, substr)
				}
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

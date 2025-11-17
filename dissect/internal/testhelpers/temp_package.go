package testhelpers

import (
	"os"
	"path/filepath"
	"testing"
)

// CreateTempPackage creates a temporary Go module with the specified files.
// Returns the path to the temporary directory.
// The caller is responsible for cleaning up using t.Cleanup or defer os.RemoveAll.
func CreateTempPackage(t *testing.T, files map[string]string) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create go.mod
	gomod := "module test\n\ngo 1.21\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(gomod), 0o644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create all files
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o644); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to write %s: %v", name, err)
		}
	}

	return tmpDir
}

// CreateFileInDir creates a file with the given content in the specified directory.
func CreateFileInDir(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write %s: %v", name, err)
	}
}

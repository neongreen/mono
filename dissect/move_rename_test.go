package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neongreen/mono/dissect/pkg/refactor"
)

// Stub tests removed - they only verified file creation, not actual functionality.
// Real rename tests are in dissect/pkg/gopls/rename_test.go

func TestFileMoveSamePackageToSubdirectory(t *testing.T) {
	// This is the existing test from move_file_test.go
	// Kept here to ensure file move still works
	tmpDir := t.TempDir()

	// Create a Go module
	modFile := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(modFile, []byte("module testmod\n\ngo 1.24\n"), 0o644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create main.go with a reference to AdminCmd
	mainFile := filepath.Join(tmpDir, "main.go")
	mainContent := `package main

func main() {
	_ = AdminCmd
}
`
	if err := os.WriteFile(mainFile, []byte(mainContent), 0o644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	// Create admin_cmd.go with AdminCmd variable
	adminCmdFile := filepath.Join(tmpDir, "admin_cmd.go")
	adminCmdContent := `package main

var AdminCmd = "admin command"
`
	if err := os.WriteFile(adminCmdFile, []byte(adminCmdContent), 0o644); err != nil {
		t.Fatalf("Failed to create admin_cmd.go: %v", err)
	}

	// Create target directory
	cmdDir := filepath.Join(tmpDir, "cmd")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatalf("Failed to create cmd directory: %v", err)
	}

	// Move admin_cmd.go to cmd/admin.go
	targetFile := filepath.Join(cmdDir, "admin.go")
	err := refactor.MoveFileWithImportUpdates(adminCmdFile, targetFile, tmpDir, "goimports")
	if err != nil {
		t.Fatalf("MoveFileWithImportUpdates failed: %v", err)
	}

	// Verify the move succeeded
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		t.Errorf("Target file was not created: %s", targetFile)
	}

	if _, err := os.Stat(adminCmdFile); !os.IsNotExist(err) {
		t.Errorf("Source file still exists: %s", adminCmdFile)
	}

	// Verify package was updated in the moved file
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	if !strings.Contains(string(targetContent), "package cmd") {
		t.Errorf("Package declaration was not updated in target file")
	}

	// Verify main.go was updated to import and qualify the reference
	updatedMainContent, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read updated main.go: %v", err)
	}

	mainStr := string(updatedMainContent)

	// Should have import
	if !strings.Contains(mainStr, "\"testmod/cmd\"") {
		t.Errorf("Import was not added to main.go")
	}

	// Should have qualified reference
	if !strings.Contains(mainStr, "cmd.AdminCmd") {
		t.Errorf("Reference was not qualified in main.go")
	}
}

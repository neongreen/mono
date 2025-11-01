package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMoveFile tests basic file moving/renaming functionality
func TestMoveFile(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dissect_move_file_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a source file
	sourceFile := filepath.Join(tmpDir, "source.go")
	sourceContent := `package main

import "fmt"

func Foo() {
	fmt.Println("Hello from Foo")
}

func Bar() {
	fmt.Println("Hello from Bar")
}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Initialize go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Move the file
	targetFile := filepath.Join(tmpDir, "target.go")
	args := []string{"move", sourceFile, targetFile}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Run the move command
	runMove(moveCmd, args[1:])

	// Verify source file no longer exists
	if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
		t.Errorf("Source file still exists after move")
	}

	// Verify target file exists
	if _, err := os.Stat(targetFile); err != nil {
		t.Fatalf("Target file does not exist: %v", err)
	}

	// Verify target file has the correct content
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	if string(targetContent) != sourceContent {
		t.Errorf("Target file content doesn't match source.\nExpected:\n%s\nGot:\n%s", sourceContent, string(targetContent))
	}
}

// TestMoveFileToSubdirectory tests moving a file to a subdirectory
func TestMoveFileToSubdirectory(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dissect_move_file_subdir_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a source file
	sourceFile := filepath.Join(tmpDir, "admin_cmd.go")
	sourceContent := `package tk

func AdminCommand() {
	// admin command logic
}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Initialize go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Move the file to a subdirectory (like the user's example: tk/admin_cmd.go -> tk/cmd/admin.go)
	targetFile := filepath.Join(tmpDir, "cmd", "admin.go")
	args := []string{"move", sourceFile, targetFile}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Run the move command
	runMove(moveCmd, args[1:])

	// Verify source file no longer exists
	if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
		t.Errorf("Source file still exists after move")
	}

	// Verify target file exists
	if _, err := os.Stat(targetFile); err != nil {
		t.Fatalf("Target file does not exist: %v", err)
	}

	// Verify target file has the correct content
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	if string(targetContent) != sourceContent {
		t.Errorf("Target file content doesn't match source.\nExpected:\n%s\nGot:\n%s", sourceContent, string(targetContent))
	}
}

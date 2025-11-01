package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// TestMoveFileToSubdirectory tests moving a file to a subdirectory (package should change)
func TestMoveFileToSubdirectory(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dissect_move_file_subdir_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a source file
	sourceFile := filepath.Join(tmpDir, "admin_cmd.go")
	sourceContent := `package main

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

	// Move the file to a subdirectory
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

	// Verify target file has package changed to match directory
	expectedContent := `package cmd

func AdminCommand() {
	// admin command logic
}
`
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	if string(targetContent) != expectedContent {
		t.Errorf("Target file content doesn't match expected.\nExpected:\n%s\nGot:\n%s", expectedContent, string(targetContent))
	}
}

// TestMoveFileWithImportUpdates tests that imports are updated when moving files between packages
func TestMoveFileWithImportUpdates(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dissect_move_file_imports_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a source file in a helper subdirectory
	helperDir := filepath.Join(tmpDir, "helper")
	err = os.MkdirAll(helperDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create helper dir: %v", err)
	}

	sourceFile := filepath.Join(helperDir, "util.go")
	sourceContent := `package helper

func Helper() string {
	return "helper"
}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create a main file that imports from helper
	mainFile := filepath.Join(tmpDir, "main.go")
	mainContent := `package main

import "test/helper"

func main() {
	helper.Helper()
}
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	// Initialize go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Move the file from helper/ to utils/
	targetFile := filepath.Join(tmpDir, "utils", "helper.go")
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

	// Verify target file exists and has correct package
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	expectedTargetContent := `package utils

func Helper() string {
	return "helper"
}
`
	if string(targetContent) != expectedTargetContent {
		t.Errorf("Target file package not updated correctly.\nExpected:\n%s\nGot:\n%s", expectedTargetContent, string(targetContent))
	}

	// Verify main file has import updated from test/helper to test/utils
	mainContentBytes, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read main file: %v", err)
	}

	expectedMainContent := `package main

import "test/utils"

func main() {
	utils.Helper()
}
`
	if string(mainContentBytes) != expectedMainContent {
		t.Errorf("Main file imports not updated correctly.\nExpected:\n%s\nGot:\n%s", expectedMainContent, string(mainContentBytes))
	}
}

// TestMoveFileSamePackageToSubdirectory tests moving a file within the same package to a subdirectory
// This should break references from sibling files that use the moved definitions
func TestMoveFileSamePackageToSubdirectory(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dissect_move_same_package_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create admin_cmd.go that defines adminCmd variable in package main
	adminFile := filepath.Join(tmpDir, "admin_cmd.go")
	adminContent := `package main

var adminCmd = "admin"

func AdminHelper() string {
	return "admin helper"
}
`
	err = os.WriteFile(adminFile, []byte(adminContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write admin file: %v", err)
	}

	// Create main.go that uses adminCmd from admin_cmd.go (same package, no import)
	mainFile := filepath.Join(tmpDir, "main.go")
	mainContent := `package main

import "fmt"

func main() {
	fmt.Println(adminCmd)
	fmt.Println(AdminHelper())
}
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	// Initialize go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Move admin_cmd.go to cmd/admin.go (creates new package)
	targetFile := filepath.Join(tmpDir, "cmd", "admin.go")
	args := []string{"move", adminFile, targetFile}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Run the move command
	runMove(moveCmd, args[1:])

	// Verify source file no longer exists
	if _, err := os.Stat(adminFile); !os.IsNotExist(err) {
		t.Errorf("Source file still exists after move")
	}

	// Verify target file exists
	if _, err := os.Stat(targetFile); err != nil {
		t.Fatalf("Target file does not exist: %v", err)
	}

	// Verify target file has package cmd
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	if !strings.Contains(string(targetContent), "package cmd") {
		t.Errorf("Target file should have 'package cmd', got:\n%s", string(targetContent))
	}

	// The critical test: verify project does NOT build (adminCmd is now undefined in main.go)
	// This is the bug - dissect move leaves the code in a broken state
	buildCmd := exec.Command("go", "build", ".")
	buildCmd.Dir = tmpDir
	buildOutput, buildErr := buildCmd.CombinedOutput()

	// We EXPECT this to fail because main.go still references adminCmd without a package qualifier
	if buildErr == nil {
		t.Errorf("Expected build to fail after moving file to subdirectory, but it succeeded")
	}

	// Verify we get the expected error message about undefined symbols
	if !strings.Contains(string(buildOutput), "undefined: adminCmd") &&
		!strings.Contains(string(buildOutput), "undefined: AdminHelper") {
		t.Errorf("Expected 'undefined' error for adminCmd or AdminHelper, got: %s", string(buildOutput))
	}
}

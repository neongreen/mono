package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMoveFileMultipleImporters verifies that when a file is moved,
// all files that import it are updated correctly
func TestMoveFileMultipleImporters(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_move_file_multi_import_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create source file in helper/util.go
	helperDir := filepath.Join(tmpDir, "helper")
	err = os.MkdirAll(helperDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create helper dir: %v", err)
	}

	sourceFile := filepath.Join(helperDir, "util.go")
	sourceContent := `package helper

func Util() string {
	return "utility function"
}

func Helper() int {
	return 42
}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create importer 1 at root
	importer1 := filepath.Join(tmpDir, "main.go")
	importer1Content := `package main

import "test/helper"

func main() {
	s := helper.Util()
	n := helper.Helper()
	println(s, n)
}
`
	err = os.WriteFile(importer1, []byte(importer1Content), 0644)
	if err != nil {
		t.Fatalf("Failed to write importer1: %v", err)
	}

	// Create importer 2 in subdirectory
	handlersDir := filepath.Join(tmpDir, "handlers")
	err = os.MkdirAll(handlersDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create handlers dir: %v", err)
	}

	importer2 := filepath.Join(handlersDir, "handler.go")
	importer2Content := `package handlers

import "test/helper"

func GetUtil() string {
	return helper.Util()
}
`
	err = os.WriteFile(importer2, []byte(importer2Content), 0644)
	if err != nil {
		t.Fatalf("Failed to write importer2: %v", err)
	}

	// Create importer 3 in sibling package
	libDir := filepath.Join(tmpDir, "lib")
	err = os.MkdirAll(libDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create lib dir: %v", err)
	}

	importer3 := filepath.Join(libDir, "library.go")
	importer3Content := `package lib

import "test/helper"

var DefaultHelper = helper.Helper()
`
	err = os.WriteFile(importer3, []byte(importer3Content), 0644)
	if err != nil {
		t.Fatalf("Failed to write importer3: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Move helper/util.go to support/util.go
	targetFile := filepath.Join(tmpDir, "support", "util.go")
	args := []string{"move", sourceFile, targetFile}
	runMove(moveCmd, args[1:])

	// Verify source file no longer exists
	if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
		t.Errorf("Source file still exists after move")
	}

	// Verify target file exists
	if _, err := os.Stat(targetFile); err != nil {
		t.Fatalf("Target file does not exist: %v", err)
	}

	// Verify target file has correct package
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	if !strings.Contains(string(targetContent), "package support") {
		t.Errorf("Target file should have 'package support', got:\n%s", string(targetContent))
	}

	// Verify importer 1 updated
	importer1Updated, err := os.ReadFile(importer1)
	if err != nil {
		t.Fatalf("Failed to read importer1: %v", err)
	}
	importer1Str := string(importer1Updated)
	if !strings.Contains(importer1Str, `import "test/support"`) {
		t.Errorf("Importer1 should import test/support, got:\n%s", importer1Str)
	}
	if !strings.Contains(importer1Str, "support.Util()") {
		t.Errorf("Importer1 should use support.Util(), got:\n%s", importer1Str)
	}
	if !strings.Contains(importer1Str, "support.Helper()") {
		t.Errorf("Importer1 should use support.Helper(), got:\n%s", importer1Str)
	}

	// Verify importer 2 updated
	importer2Updated, err := os.ReadFile(importer2)
	if err != nil {
		t.Fatalf("Failed to read importer2: %v", err)
	}
	importer2Str := string(importer2Updated)
	if !strings.Contains(importer2Str, `import "test/support"`) {
		t.Errorf("Importer2 should import test/support, got:\n%s", importer2Str)
	}
	if !strings.Contains(importer2Str, "support.Util()") {
		t.Errorf("Importer2 should use support.Util(), got:\n%s", importer2Str)
	}

	// Verify importer 3 updated
	importer3Updated, err := os.ReadFile(importer3)
	if err != nil {
		t.Fatalf("Failed to read importer3: %v", err)
	}
	importer3Str := string(importer3Updated)
	if !strings.Contains(importer3Str, `import "test/support"`) {
		t.Errorf("Importer3 should import test/support, got:\n%s", importer3Str)
	}
	if !strings.Contains(importer3Str, "support.Helper()") {
		t.Errorf("Importer3 should use support.Helper(), got:\n%s", importer3Str)
	}

	// Verify code builds
	buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

// TestMoveFileRenameInPlace verifies renaming a file in the same directory
func TestMoveFileRenameInPlace(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_move_file_rename_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create source file
	sourceFile := filepath.Join(tmpDir, "admin_commands.go")
	sourceContent := `package main

import "fmt"

func AdminCommand() {
	fmt.Println("admin command")
}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create main file
	mainFile := filepath.Join(tmpDir, "main.go")
	mainContent := `package main

func main() {
	AdminCommand()
}
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Rename in same directory
	targetFile := filepath.Join(tmpDir, "admin.go")
	args := []string{"move", sourceFile, targetFile}
	runMove(moveCmd, args[1:])

	// Verify source file no longer exists
	if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
		t.Errorf("Source file still exists after move")
	}

	// Verify target file exists
	if _, err := os.Stat(targetFile); err != nil {
		t.Fatalf("Target file does not exist: %v", err)
	}

	// Verify target file has same package (main, unchanged)
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	if string(targetContent) != sourceContent {
		t.Errorf("Target file content should be unchanged, got:\n%s", string(targetContent))
	}

	// Verify main file unchanged (no imports to update)
	mainUpdated, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read main file: %v", err)
	}
	if string(mainUpdated) != mainContent {
		t.Errorf("Main file should be unchanged, got:\n%s", string(mainUpdated))
	}

	// Verify code builds
	buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

// TestMoveFileCreateTargetDir verifies that target directories are created automatically
func TestMoveFileCreateTargetDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_move_file_create_dir_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create source file in helpers/
	helpersDir := filepath.Join(tmpDir, "helpers")
	err = os.MkdirAll(helpersDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create helpers dir: %v", err)
	}

	sourceFile := filepath.Join(helpersDir, "util.go")
	sourceContent := `package helpers

func Utility() string {
	return "utility"
}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create importer
	mainFile := filepath.Join(tmpDir, "main.go")
	mainContent := `package main

import "test/helpers"

func main() {
	s := helpers.Utility()
	println(s)
}
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Move to deeply nested path that doesn't exist
	targetFile := filepath.Join(tmpDir, "deeply", "nested", "new", "path", "util.go")
	args := []string{"move", sourceFile, targetFile}
	runMove(moveCmd, args[1:])

	// Verify source file no longer exists
	if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
		t.Errorf("Source file still exists after move")
	}

	// Verify all intermediate directories were created
	dirs := []string{
		filepath.Join(tmpDir, "deeply"),
		filepath.Join(tmpDir, "deeply", "nested"),
		filepath.Join(tmpDir, "deeply", "nested", "new"),
		filepath.Join(tmpDir, "deeply", "nested", "new", "path"),
	}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory should exist: %s", dir)
		}
	}

	// Verify target file exists
	if _, err := os.Stat(targetFile); err != nil {
		t.Fatalf("Target file does not exist: %v", err)
	}

	// Verify target file has correct package (path, from directory name)
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	if !strings.Contains(string(targetContent), "package path") {
		t.Errorf("Target file should have 'package path', got:\n%s", string(targetContent))
	}

	// Verify main file updated
	mainUpdated, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read main file: %v", err)
	}
	mainStr := string(mainUpdated)
	if !strings.Contains(mainStr, `import "test/deeply/nested/new/path"`) {
		t.Errorf("Main should import deeply nested path, got:\n%s", mainStr)
	}
	if !strings.Contains(mainStr, "path.Utility()") {
		t.Errorf("Main should use path.Utility(), got:\n%s", mainStr)
	}

	// Verify code builds
	buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

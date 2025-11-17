package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMoveFileMixedDeclarations verifies that files with mixed declaration types are handled correctly
func TestMoveFileMixedDeclarations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_move_file_mixed_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create source file with mixed declarations
	modelsDir := filepath.Join(tmpDir, "models")
	err = os.MkdirAll(modelsDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create models dir: %v", err)
	}

	sourceFile := filepath.Join(modelsDir, "user.go")
	sourceContent := `package models

// User represents a user in the system
type User struct {
	ID   int
	Name string
}

// Repository defines user repository operations
type Repository interface {
	GetUser(id int) (*User, error)
}

const MaxUsers = 100

var DefaultUser = User{ID: 1, Name: "Default"}

// NewUser creates a new user
func NewUser(name string) *User {
	return &User{Name: name}
}

// Validate checks if the user is valid
func (u *User) Validate() error {
	if u.Name == "" {
		return nil
	}
	return nil
}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create importer using all declaration types
	importerFile := filepath.Join(tmpDir, "main.go")
	importerContent := `package main

import "test/models"

func main() {
	// Use type
	user := models.User{ID: 1, Name: "Alice"}
	
	// Use function
	newUser := models.NewUser("Bob")
	
	// Use method
	_ = user.Validate()
	_ = newUser.Validate()
	
	// Use const
	max := models.MaxUsers
	
	// Use var
	def := models.DefaultUser
	
	println(user.Name, max, def.Name)
}
`
	err = os.WriteFile(importerFile, []byte(importerContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write importer: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Move models/user.go to entities/user.go
	targetFile := filepath.Join(tmpDir, "entities", "user.go")
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

	// Verify target file has correct package and all declarations
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	targetStr := string(targetContent)
	if !strings.Contains(targetStr, "package entities") {
		t.Errorf("Target file should have 'package entities', got:\n%s", targetStr)
	}
	// Check all declarations are present
	expectedDecls := []string{"type User struct", "type Repository interface", "const MaxUsers", "var DefaultUser", "func NewUser", "func (u *User) Validate"}
	for _, decl := range expectedDecls {
		if !strings.Contains(targetStr, decl) {
			t.Errorf("Target file should contain '%s', got:\n%s", decl, targetStr)
		}
	}

	// Verify importer updated to use new package for all declarations
	importerUpdated, err := os.ReadFile(importerFile)
	if err != nil {
		t.Fatalf("Failed to read importer: %v", err)
	}
	importerStr := string(importerUpdated)
	if !strings.Contains(importerStr, `import "test/entities"`) {
		t.Errorf("Importer should import test/entities, got:\n%s", importerStr)
	}
	// Check all references use new package qualifier
	expectedRefs := []string{"entities.User{", "entities.NewUser", "entities.MaxUsers", "entities.DefaultUser"}
	for _, ref := range expectedRefs {
		if !strings.Contains(importerStr, ref) {
			t.Errorf("Importer should use '%s', got:\n%s", ref, importerStr)
		}
	}

	// Verify code builds
	buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

// TestMoveFilePreservesComments verifies that all comments are preserved when moving files
func TestMoveFilePreservesComments(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_move_file_comments_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create source file with various comment styles
	utilsDir := filepath.Join(tmpDir, "utils")
	err = os.MkdirAll(utilsDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create utils dir: %v", err)
	}

	sourceFile := filepath.Join(utilsDir, "helper.go")
	sourceContent := `// Package utils provides utility functions for the application.
// This is a package-level doc comment.
package utils

// Helper is a utility function that helps with things.
// It returns a helpful string.
func Helper() string {
	// This is an inline comment
	result := "helpful" // trailing comment
	return result
}

// Config holds configuration values
type Config struct {
	// Value is the config value
	Value string
}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create importer
	mainFile := filepath.Join(tmpDir, "main.go")
	mainContent := `package main

import "test/utils"

func main() {
	result := utils.Helper()
	cfg := utils.Config{Value: "test"}
	println(result, cfg.Value)
}
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Move utils/helper.go to support/helper.go
	targetFile := filepath.Join(tmpDir, "support", "helper.go")
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

	// Verify target file has correct package and all comments preserved
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	targetStr := string(targetContent)

	// Check package declaration updated
	if !strings.Contains(targetStr, "package support") {
		t.Errorf("Target file should have 'package support', got:\n%s", targetStr)
	}

	// Check all comments are preserved (note: package doc comments are not auto-updated)
	expectedComments := []string{
		"// This is a package-level doc comment",
		"// Helper is a utility function",
		"// It returns a helpful string",
		"// This is an inline comment",
		"// trailing comment",
		"// Config holds configuration values",
		"// Value is the config value",
	}
	for _, comment := range expectedComments {
		if !strings.Contains(targetStr, comment) {
			t.Errorf("Target file should contain comment '%s', got:\n%s", comment, targetStr)
		}
	}

	// Verify main file updated
	mainUpdated, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read main file: %v", err)
	}
	mainStr := string(mainUpdated)
	if !strings.Contains(mainStr, `import "test/support"`) {
		t.Errorf("Main should import test/support, got:\n%s", mainStr)
	}
	if !strings.Contains(mainStr, "support.Helper()") {
		t.Errorf("Main should use support.Helper(), got:\n%s", mainStr)
	}
	if !strings.Contains(mainStr, "support.Config{") {
		t.Errorf("Main should use support.Config{}, got:\n%s", mainStr)
	}

	// Verify code builds
	buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

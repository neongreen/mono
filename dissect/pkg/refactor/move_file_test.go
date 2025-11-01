package refactor

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMoveFileWithMultipleImporters tests that imports are updated in multiple files
func TestMoveFileWithMultipleImporters(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_move_multi_import_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file in models/
	modelsDir := filepath.Join(tmpDir, "models")
	err = os.MkdirAll(modelsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create models dir: %v", err)
	}

	sourceFile := filepath.Join(modelsDir, "user.go")
	sourceContent := `package models

type User struct {
	ID   int
	Name string
}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create first file that imports models
	file1 := filepath.Join(tmpDir, "handler1.go")
	file1Content := `package main

import "test/models"

func GetUser() *models.User {
	return &models.User{ID: 1, Name: "Alice"}
}
`
	err = os.WriteFile(file1, []byte(file1Content), 0644)
	if err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}

	// Create second file that imports models
	file2 := filepath.Join(tmpDir, "handler2.go")
	file2Content := `package main

import "test/models"

func UpdateUser(u *models.User) {
	models.ProcessUser(u)
}
`
	err = os.WriteFile(file2, []byte(file2Content), 0644)
	if err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	// Initialize go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Move the file from models/ to entities/
	targetFile := filepath.Join(tmpDir, "entities", "user.go")
	err = MoveFileWithImportUpdates(sourceFile, targetFile, tmpDir)
	if err != nil {
		t.Fatalf("MoveFileWithImportUpdates failed: %v", err)
	}

	// Verify target file has correct package
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	expectedTargetContent := `package entities

type User struct {
	ID   int
	Name string
}
`
	if string(targetContent) != expectedTargetContent {
		t.Errorf("Target file package not updated correctly.\nExpected:\n%s\nGot:\n%s", expectedTargetContent, string(targetContent))
	}

	// Verify both importing files were updated
	file1Updated, err := os.ReadFile(file1)
	if err != nil {
		t.Fatalf("Failed to read file1: %v", err)
	}

	expectedFile1 := `package main

import "test/entities"

func GetUser() *entities.User {
	return &entities.User{ID: 1, Name: "Alice"}
}
`
	if string(file1Updated) != expectedFile1 {
		t.Errorf("File1 not updated correctly.\nExpected:\n%s\nGot:\n%s", expectedFile1, string(file1Updated))
	}

	file2Updated, err := os.ReadFile(file2)
	if err != nil {
		t.Fatalf("Failed to read file2: %v", err)
	}

	expectedFile2 := `package main

import "test/entities"

func UpdateUser(u *entities.User) {
	entities.ProcessUser(u)
}
`
	if string(file2Updated) != expectedFile2 {
		t.Errorf("File2 not updated correctly.\nExpected:\n%s\nGot:\n%s", expectedFile2, string(file2Updated))
	}
}

// TestMoveFileWithImportAlias tests that files with import aliases are handled correctly
func TestMoveFileWithImportAlias(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_move_alias_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file
	helpDir := filepath.Join(tmpDir, "help")
	err = os.MkdirAll(helpDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create help dir: %v", err)
	}

	sourceFile := filepath.Join(helpDir, "util.go")
	sourceContent := `package help

func DoSomething() string {
	return "done"
}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create file with import alias
	mainFile := filepath.Join(tmpDir, "main.go")
	mainContent := `package main

import helplib "test/help"

func main() {
	helplib.DoSomething()
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

	// Move the file from help/ to support/
	targetFile := filepath.Join(tmpDir, "support", "util.go")
	err = MoveFileWithImportUpdates(sourceFile, targetFile, tmpDir)
	if err != nil {
		t.Fatalf("MoveFileWithImportUpdates failed: %v", err)
	}

	// Verify target file
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	expectedTargetContent := `package support

func DoSomething() string {
	return "done"
}
`
	if string(targetContent) != expectedTargetContent {
		t.Errorf("Target file package not updated correctly.\nExpected:\n%s\nGot:\n%s", expectedTargetContent, string(targetContent))
	}

	// Verify main file: import path should update but alias should be preserved
	mainUpdated, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read main file: %v", err)
	}

	expectedMain := `package main

import helplib "test/support"

func main() {
	helplib.DoSomething()
}
`
	if string(mainUpdated) != expectedMain {
		t.Errorf("Main file not updated correctly.\nExpected:\n%s\nGot:\n%s", expectedMain, string(mainUpdated))
	}
}

// TestMoveFileWithMultipleImportsInFile tests files that import multiple packages
func TestMoveFileWithMultipleImportsInFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_move_multi_imports_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file in config/
	configDir := filepath.Join(tmpDir, "config")
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	sourceFile := filepath.Join(configDir, "settings.go")
	sourceContent := `package config

type Settings struct {
	Debug bool
}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create a utils package that won't be moved
	utilsDir := filepath.Join(tmpDir, "utils")
	err = os.MkdirAll(utilsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create utils dir: %v", err)
	}

	utilsFile := filepath.Join(utilsDir, "helper.go")
	utilsContent := `package utils

func Helper() string {
	return "help"
}
`
	err = os.WriteFile(utilsFile, []byte(utilsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write utils file: %v", err)
	}

	// Create main file that imports both config and utils
	mainFile := filepath.Join(tmpDir, "main.go")
	mainContent := `package main

import (
	"test/config"
	"test/utils"
)

func main() {
	s := &config.Settings{Debug: true}
	utils.Helper()
	processSettings(s)
}

func processSettings(s *config.Settings) {
	// Process settings
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

	// Move config/ to options/
	targetFile := filepath.Join(tmpDir, "options", "settings.go")
	err = MoveFileWithImportUpdates(sourceFile, targetFile, tmpDir)
	if err != nil {
		t.Fatalf("MoveFileWithImportUpdates failed: %v", err)
	}

	// Verify target file
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	expectedTargetContent := `package options

type Settings struct {
	Debug bool
}
`
	if string(targetContent) != expectedTargetContent {
		t.Errorf("Target file package not updated correctly.\nExpected:\n%s\nGot:\n%s", expectedTargetContent, string(targetContent))
	}

	// Verify main file: only config import should be updated, utils should remain
	mainUpdated, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read main file: %v", err)
	}

	expectedMain := `package main

import (
	"test/options"
	"test/utils"
)

func main() {
	s := &options.Settings{Debug: true}
	utils.Helper()
	processSettings(s)
}

func processSettings(s *options.Settings) {
	// Process settings
}
`
	if string(mainUpdated) != expectedMain {
		t.Errorf("Main file not updated correctly.\nExpected:\n%s\nGot:\n%s", expectedMain, string(mainUpdated))
	}
}

// TestMoveFileDeeplyNested tests moving files between deeply nested directories
func TestMoveFileDeeplyNested(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_move_nested_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file in a/b/c/
	sourceDir := filepath.Join(tmpDir, "a", "b", "c")
	err = os.MkdirAll(sourceDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	sourceFile := filepath.Join(sourceDir, "deep.go")
	sourceContent := `package c

func DeepFunc() {}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create importer
	mainFile := filepath.Join(tmpDir, "main.go")
	mainContent := `package main

import "test/a/b/c"

func main() {
	c.DeepFunc()
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

	// Move from a/b/c/ to x/y/z/
	targetFile := filepath.Join(tmpDir, "x", "y", "z", "deep.go")
	err = MoveFileWithImportUpdates(sourceFile, targetFile, tmpDir)
	if err != nil {
		t.Fatalf("MoveFileWithImportUpdates failed: %v", err)
	}

	// Verify target file
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	expectedTargetContent := `package z

func DeepFunc() {}
`
	if string(targetContent) != expectedTargetContent {
		t.Errorf("Target file package not updated correctly.\nExpected:\n%s\nGot:\n%s", expectedTargetContent, string(targetContent))
	}

	// Verify main file
	mainUpdated, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read main file: %v", err)
	}

	expectedMain := `package main

import "test/x/y/z"

func main() {
	z.DeepFunc()
}
`
	if string(mainUpdated) != expectedMain {
		t.Errorf("Main file not updated correctly.\nExpected:\n%s\nGot:\n%s", expectedMain, string(mainUpdated))
	}
}

// TestMoveFileNoImporters tests moving a file that has no importers
func TestMoveFileNoImporters(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_move_no_importers_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create standalone file
	sourceFile := filepath.Join(tmpDir, "standalone.go")
	sourceContent := `package main

func main() {
	println("Hello")
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

	// Move to subdirectory
	targetFile := filepath.Join(tmpDir, "cmd", "app.go")
	err = MoveFileWithImportUpdates(sourceFile, targetFile, tmpDir)
	if err != nil {
		t.Fatalf("MoveFileWithImportUpdates failed: %v", err)
	}

	// Verify target file
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	expectedTargetContent := `package cmd

func main() {
	println("Hello")
}
`
	if string(targetContent) != expectedTargetContent {
		t.Errorf("Target file package not updated correctly.\nExpected:\n%s\nGot:\n%s", expectedTargetContent, string(targetContent))
	}
}

// TestMoveFileSameDirectoryRename tests renaming a file in the same directory
func TestMoveFileSameDirectoryRename(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_move_rename_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file
	sourceFile := filepath.Join(tmpDir, "old_name.go")
	sourceContent := `package main

import "fmt"

func Foo() {
	fmt.Println("foo")
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

	// Rename in same directory
	targetFile := filepath.Join(tmpDir, "new_name.go")
	err = MoveFileWithImportUpdates(sourceFile, targetFile, tmpDir)
	if err != nil {
		t.Fatalf("MoveFileWithImportUpdates failed: %v", err)
	}

	// Verify target file: package should NOT change
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	if string(targetContent) != sourceContent {
		t.Errorf("Target file should be identical to source.\nExpected:\n%s\nGot:\n%s", sourceContent, string(targetContent))
	}

	// Verify source was removed
	if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
		t.Errorf("Source file should have been deleted")
	}
}

// TestMoveFileComplexSelectors tests that complex selector expressions are updated
func TestMoveFileComplexSelectors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_move_complex_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file
	dataDir := filepath.Join(tmpDir, "data")
	err = os.MkdirAll(dataDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create data dir: %v", err)
	}

	sourceFile := filepath.Join(dataDir, "types.go")
	sourceContent := `package data

type Config struct {
	Value string
}

var DefaultConfig = Config{Value: "default"}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create file with complex usage patterns
	mainFile := filepath.Join(tmpDir, "main.go")
	mainContent := `package main

import "test/data"

func main() {
	cfg := data.Config{Value: "test"}
	cfg2 := data.DefaultConfig
	_ = data.Config{Value: "nested"}
	merge(data.Config{}, data.DefaultConfig)
}

func merge(c1, c2 data.Config) {}
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

	// Move from data/ to settings/
	targetFile := filepath.Join(tmpDir, "settings", "types.go")
	err = MoveFileWithImportUpdates(sourceFile, targetFile, tmpDir)
	if err != nil {
		t.Fatalf("MoveFileWithImportUpdates failed: %v", err)
	}

	// Read actual output and verify using substring checks
	mainUpdated, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read main file: %v", err)
	}

	mainStr := string(mainUpdated)

	// Check that import was updated to settings
	if !contains(mainStr, `"test/settings"`) {
		t.Errorf("Import not updated to test/settings. Got:\n%s", mainStr)
	}

	// Check that old package qualifiers are gone
	if contains(mainStr, "data.Config") || contains(mainStr, "data.DefaultConfig") {
		t.Errorf("Still contains old package qualifiers (data.). Got:\n%s", mainStr)
	}

	// Check that new package qualifiers are present
	if !contains(mainStr, "settings.Config") || !contains(mainStr, "settings.DefaultConfig") {
		t.Errorf("Missing new package qualifiers (settings.). Got:\n%s", mainStr)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestGetPackageNameFromPath tests the package name extraction function
func TestGetPackageNameFromPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "test/helper", "helper"},
		{"deeply nested", "github.com/user/repo/pkg/utils", "utils"},
		{"single component", "main", "main"},
		{"with dots", "test.com/pkg", "pkg"},
		{"trailing slash", "test/utils/", "utils"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPackageNameFromPath(tt.input)
			if result != tt.expected {
				t.Errorf("getPackageNameFromPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestMoveFileUpdatesTestFileImports tests that test files get import statements added
// when a file is moved to a new package. This is a regression test for a bug where
// test files would get type qualifiers added (e.g., sync.Type) but not the import statement.
func TestMoveFileUpdatesTestFileImports(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_test_imports_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a package with a types.go file
	pkgDir := filepath.Join(tmpDir, "pkg")
	err = os.MkdirAll(pkgDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create pkg dir: %v", err)
	}

	// Create types.go with a simple type
	typesFile := filepath.Join(pkgDir, "types.go")
	typesContent := `package pkg

type Foo struct {
	Value string
}

type Bar interface {
	DoSomething()
}
`
	err = os.WriteFile(typesFile, []byte(typesContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write types.go: %v", err)
	}

	// Create types_test.go that uses Foo and Bar without importing
	// (no import needed since it's in the same package)
	testFile := filepath.Join(pkgDir, "types_test.go")
	testContent := `package pkg

import "testing"

func TestFoo(t *testing.T) {
	f := Foo{Value: "test"}
	if f.Value != "test" {
		t.Error("unexpected value")
	}
}

func TestBar(t *testing.T) {
	var _ Bar
}
`
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write types_test.go: %v", err)
	}

	// Initialize go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module testmodule\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Move types.go to internal/types/types.go
	targetFile := filepath.Join(pkgDir, "internal", "types", "types.go")
	err = MoveFileWithImportUpdates(typesFile, targetFile, tmpDir)
	if err != nil {
		t.Fatalf("Failed to move file: %v", err)
	}

	// Read the test file to verify it was updated
	testFileContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	testFileStr := string(testFileContent)

	// Verify import was added
	if !contains(testFileStr, `"testmodule/pkg/internal/types"`) {
		t.Errorf("Test file missing import for internal/types package.\nGot:\n%s", testFileStr)
	}

	// Verify type references were qualified
	if !contains(testFileStr, "types.Foo") {
		t.Errorf("Test file missing qualified reference types.Foo.\nGot:\n%s", testFileStr)
	}

	if !contains(testFileStr, "types.Bar") {
		t.Errorf("Test file missing qualified reference types.Bar.\nGot:\n%s", testFileStr)
	}

	// Verify the moved file has correct package declaration
	movedContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read moved file: %v", err)
	}

	if !contains(string(movedContent), "package types") {
		t.Errorf("Moved file should have 'package types', got:\n%s", string(movedContent))
	}
}

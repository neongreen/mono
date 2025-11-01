package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMoveFileDeeplyNested verifies moving files between deeply nested directories
func TestMoveFileDeeplyNested(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_move_file_nested_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create deeply nested source file: pkg/api/handlers/v1/admin.go
	sourceDir := filepath.Join(tmpDir, "pkg", "api", "handlers", "v1")
	err = os.MkdirAll(sourceDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	sourceFile := filepath.Join(sourceDir, "admin.go")
	sourceContent := `package v1

func AdminFunc() string {
	return "admin function"
}

func AdminStatus() int {
	return 200
}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create importer at root
	importerFile := filepath.Join(tmpDir, "main.go")
	importerContent := `package main

import "test/pkg/api/handlers/v1"

func main() {
	result := v1.AdminFunc()
	status := v1.AdminStatus()
	println(result, status)
}
`
	err = os.WriteFile(importerFile, []byte(importerContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write importer: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Move to internal/services/admin/handler.go
	targetFile := filepath.Join(tmpDir, "internal", "services", "admin", "handler.go")
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

	// Verify target file has correct package (admin, from directory name)
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	if !strings.Contains(string(targetContent), "package admin") {
		t.Errorf("Target file should have 'package admin', got:\n%s", string(targetContent))
	}

	// Verify importer updated
	importerUpdated, err := os.ReadFile(importerFile)
	if err != nil {
		t.Fatalf("Failed to read importer: %v", err)
	}
	importerStr := string(importerUpdated)
	if !strings.Contains(importerStr, `import "test/internal/services/admin"`) {
		t.Errorf("Importer should import test/internal/services/admin, got:\n%s", importerStr)
	}
	if !strings.Contains(importerStr, "admin.AdminFunc()") {
		t.Errorf("Importer should use admin.AdminFunc(), got:\n%s", importerStr)
	}
	if !strings.Contains(importerStr, "admin.AdminStatus()") {
		t.Errorf("Importer should use admin.AdminStatus(), got:\n%s", importerStr)
	}

	// Verify code builds
	buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

// TestMoveFileComplexImports verifies that complex import patterns are handled correctly
func TestMoveFileComplexImports(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_move_file_imports_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create source file: data/store.go
	dataDir := filepath.Join(tmpDir, "data")
	err = os.MkdirAll(dataDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create data dir: %v", err)
	}

	sourceFile := filepath.Join(dataDir, "store.go")
	sourceContent := `package data

type Store struct {
	Name string
}

func NewStore(name string) *Store {
	return &Store{Name: name}
}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create importer with aliased import
	importerAliased := filepath.Join(tmpDir, "aliased.go")
	importerAliasedContent := `package main

import (
	"fmt"
	ds "test/data"
)

func useAliased() {
	store := ds.NewStore("aliased")
	fmt.Println(store.Name)
}
`
	err = os.WriteFile(importerAliased, []byte(importerAliasedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write aliased importer: %v", err)
	}

	// Create importer with standard import
	importerStandard := filepath.Join(tmpDir, "standard.go")
	importerStandardContent := `package main

import (
	"fmt"
	"test/data"
)

func useStandard() {
	store := data.NewStore("standard")
	s := data.Store{Name: "direct"}
	fmt.Println(store, s)
}
`
	err = os.WriteFile(importerStandard, []byte(importerStandardContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write standard importer: %v", err)
	}

	// Create main file
	mainFile := filepath.Join(tmpDir, "main.go")
	mainContent := `package main

func main() {
	useAliased()
	useStandard()
}
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write main: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Move data/store.go to storage/store.go
	targetFile := filepath.Join(tmpDir, "storage", "store.go")
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
	if !strings.Contains(string(targetContent), "package storage") {
		t.Errorf("Target file should have 'package storage', got:\n%s", string(targetContent))
	}

	// Verify aliased importer: path updates but alias preserved, references use alias
	aliasedUpdated, err := os.ReadFile(importerAliased)
	if err != nil {
		t.Fatalf("Failed to read aliased importer: %v", err)
	}
	aliasedStr := string(aliasedUpdated)
	if !strings.Contains(aliasedStr, `ds "test/storage"`) {
		t.Errorf("Aliased importer should have 'ds \"test/storage\"', got:\n%s", aliasedStr)
	}
	if !strings.Contains(aliasedStr, "ds.NewStore") {
		t.Errorf("Aliased importer should still use ds.NewStore (alias unchanged), got:\n%s", aliasedStr)
	}

	// Verify standard importer: both import and references update
	standardUpdated, err := os.ReadFile(importerStandard)
	if err != nil {
		t.Fatalf("Failed to read standard importer: %v", err)
	}
	standardStr := string(standardUpdated)
	if !strings.Contains(standardStr, `"test/storage"`) {
		t.Errorf("Standard importer should import test/storage, got:\n%s", standardStr)
	}
	if !strings.Contains(standardStr, "storage.NewStore") {
		t.Errorf("Standard importer should use storage.NewStore, got:\n%s", standardStr)
	}
	if !strings.Contains(standardStr, "storage.Store{") {
		t.Errorf("Standard importer should use storage.Store{}, got:\n%s", standardStr)
	}

	// Verify code builds
	buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

// TestMoveFileWithCrossPackageDeps verifies that transitive dependencies are handled correctly
func TestMoveFileWithCrossPackageDeps(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_move_file_cross_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create Package A: models/user.go
	modelsDir := filepath.Join(tmpDir, "models")
	err = os.MkdirAll(modelsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create models dir: %v", err)
	}

	userFile := filepath.Join(modelsDir, "user.go")
	userContent := `package models

type User struct {
	ID   int
	Name string
}
`
	err = os.WriteFile(userFile, []byte(userContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write user file: %v", err)
	}

	// Create Package B: handlers/user.go (imports models)
	handlersDir := filepath.Join(tmpDir, "handlers")
	err = os.MkdirAll(handlersDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create handlers dir: %v", err)
	}

	handlerFile := filepath.Join(handlersDir, "user.go")
	handlerContent := `package handlers

import "test/models"

func GetUser() *models.User {
	return &models.User{ID: 1, Name: "Alice"}
}

func ProcessUser(u *models.User) {
	println(u.Name)
}
`
	err = os.WriteFile(handlerFile, []byte(handlerContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write handler file: %v", err)
	}

	// Create Package C: main.go (imports both models and handlers)
	mainFile := filepath.Join(tmpDir, "main.go")
	mainContent := `package main

import (
	"test/handlers"
	"test/models"
)

func main() {
	user := handlers.GetUser()
	handlers.ProcessUser(user)
	
	// Also use models directly
	anotherUser := &models.User{ID: 2, Name: "Bob"}
	println(anotherUser.Name)
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

	// Move models/user.go to entities/user.go
	targetFile := filepath.Join(tmpDir, "entities", "user.go")
	args := []string{"move", userFile, targetFile}
	runMove(moveCmd, args[1:])

	// Verify source file no longer exists
	if _, err := os.Stat(userFile); !os.IsNotExist(err) {
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
	if !strings.Contains(string(targetContent), "package entities") {
		t.Errorf("Target file should have 'package entities', got:\n%s", string(targetContent))
	}

	// Verify Package B (handlers) updated
	handlerUpdated, err := os.ReadFile(handlerFile)
	if err != nil {
		t.Fatalf("Failed to read handler file: %v", err)
	}
	handlerStr := string(handlerUpdated)
	if !strings.Contains(handlerStr, `import "test/entities"`) {
		t.Errorf("Handler should import test/entities, got:\n%s", handlerStr)
	}
	if !strings.Contains(handlerStr, "*entities.User") {
		t.Errorf("Handler should use *entities.User, got:\n%s", handlerStr)
	}
	if !strings.Contains(handlerStr, "entities.User{") {
		t.Errorf("Handler should use entities.User{}, got:\n%s", handlerStr)
	}

	// Verify Package C (main) updated
	mainUpdated, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read main file: %v", err)
	}
	mainStr := string(mainUpdated)
	if !strings.Contains(mainStr, `"test/entities"`) {
		t.Errorf("Main should import test/entities, got:\n%s", mainStr)
	}
	if !strings.Contains(mainStr, "entities.User{") {
		t.Errorf("Main should use entities.User{}, got:\n%s", mainStr)
	}

	// Verify code builds
	buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

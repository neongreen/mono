package typeinfo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPackage_SimplePackage(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"simple.go": `package simple

func Hello() string {
	return "hello"
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("LoadPackage failed: %v", err)
	}

	if pkg.Name != "simple" {
		t.Errorf("Expected package name 'simple', got '%s'", pkg.Name)
	}
	if len(pkg.Syntax) != 1 {
		t.Errorf("Expected 1 file, got %d", len(pkg.Syntax))
	}
}

func TestLoadPackage_MultipleFiles(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"file1.go": `package multi

func Func1() string {
	return "func1"
}
`,
		"file2.go": `package multi

func Func2() string {
	return "func2"
}
`,
		"file3.go": `package multi

func Func3() string {
	return "func3"
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("LoadPackage failed: %v", err)
	}

	if pkg.Name != "multi" {
		t.Errorf("Expected package name 'multi', got '%s'", pkg.Name)
	}
	if len(pkg.Syntax) != 3 {
		t.Errorf("Expected 3 files, got %d", len(pkg.Syntax))
	}
}

func TestLoadPackage_WithTests(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"main.go": `package testpkg

func Add(a, b int) int {
	return a + b
}
`,
		"main_test.go": `package testpkg

import "testing"

func TestAdd(t *testing.T) {
	if Add(2, 3) != 5 {
		t.Error("Add failed")
	}
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("LoadPackage failed: %v", err)
	}

	if pkg.Name != "testpkg" {
		t.Errorf("Expected package name 'testpkg', got '%s'", pkg.Name)
	}
	// Should load at least the main file; test files may or may not be included depending on pattern
	if len(pkg.Syntax) < 1 {
		t.Errorf("Expected at least 1 file, got %d", len(pkg.Syntax))
	}
}

func TestLoadPackage_TypeInfoPresent(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"types.go": `package types

type Person struct {
	Name string
	Age  int
}

func NewPerson(name string, age int) *Person {
	return &Person{Name: name, Age: age}
}
`,
	})
	defer os.RemoveAll(tmpDir)

	pkg, err := LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("LoadPackage failed: %v", err)
	}

	if pkg.Types == nil {
		t.Fatal("Expected Types to be populated")
	}
	if pkg.TypesInfo == nil {
		t.Fatal("Expected TypesInfo to be populated")
	}
	if pkg.TypesInfo.Defs == nil {
		t.Fatal("Expected TypesInfo.Defs to be populated")
	}
	if pkg.TypesInfo.Uses == nil {
		t.Fatal("Expected TypesInfo.Uses to be populated")
	}
}

func TestLoadPackage_NonExistentDir(t *testing.T) {
	_, err := LoadPackage("/nonexistent/directory/path")
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
}

func TestLoadPackage_InvalidCode(t *testing.T) {
	tmpDir := createTempPackage(t, map[string]string{
		"invalid.go": `package invalid

func Broken() {
	// Missing closing brace and other syntax errors
	if true {
		return
`,
	})
	defer os.RemoveAll(tmpDir)

	_, err := LoadPackage(tmpDir)
	if err == nil {
		t.Error("Expected error for invalid code, got nil")
	}
}

func TestLoadPackages_MultiplePatterns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "typeinfo_multi_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create pkg1
	pkg1Dir := filepath.Join(tmpDir, "pkg1")
	if err := os.MkdirAll(pkg1Dir, 0755); err != nil {
		t.Fatalf("Failed to create pkg1: %v", err)
	}
	createFileInDir(t, pkg1Dir, "main.go", `package pkg1

func Func1() string {
	return "pkg1"
}
`)

	// Create pkg2
	pkg2Dir := filepath.Join(tmpDir, "pkg2")
	if err := os.MkdirAll(pkg2Dir, 0755); err != nil {
		t.Fatalf("Failed to create pkg2: %v", err)
	}
	createFileInDir(t, pkg2Dir, "main.go", `package pkg2

func Func2() string {
	return "pkg2"
}
`)

	// Create go.mod in root
	createFileInDir(t, tmpDir, "go.mod", "module test\n\ngo 1.21\n")

	pkgs, err := LoadPackages([]string{"./pkg1", "./pkg2"}, tmpDir)
	if err != nil {
		t.Fatalf("LoadPackages failed: %v", err)
	}

	if len(pkgs) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(pkgs))
	}
}

func TestLoadPackages_DependentPackages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "typeinfo_deps_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create util package
	utilDir := filepath.Join(tmpDir, "util")
	if err := os.MkdirAll(utilDir, 0755); err != nil {
		t.Fatalf("Failed to create util: %v", err)
	}
	createFileInDir(t, utilDir, "util.go", `package util

func Helper() string {
	return "helper"
}
`)

	// Create main package that imports util
	createFileInDir(t, tmpDir, "main.go", `package main

import "test/util"

func main() {
	util.Helper()
}
`)

	// Create go.mod
	createFileInDir(t, tmpDir, "go.mod", "module test\n\ngo 1.21\n")

	pkgs, err := LoadPackages([]string{"."}, tmpDir)
	if err != nil {
		t.Fatalf("LoadPackages failed: %v", err)
	}

	if len(pkgs) != 1 {
		t.Errorf("Expected 1 package, got %d", len(pkgs))
	}

	// Verify imports are loaded
	mainPkg := pkgs[0]
	if len(mainPkg.Imports) == 0 {
		t.Error("Expected imports to be populated")
	}
}

// Helper functions

func createTempPackage(t *testing.T, files map[string]string) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "typeinfo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create go.mod
	gomod := "module test\n\ngo 1.21\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(gomod), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create all files
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to write %s: %v", name, err)
		}
	}

	return tmpDir
}

func createFileInDir(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write %s: %v", name, err)
	}
}

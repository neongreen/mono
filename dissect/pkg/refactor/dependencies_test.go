package refactor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neongreen/mono/dissect/pkg/typeinfo"
)

// TestAnalyzeMoveDependenciesUnit tests the analyzeMoveDependencies function directly
func TestAnalyzeMoveDependenciesUnit(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "deps_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file with an unexported function
	helperFile := filepath.Join(tmpDir, "helper.go")
	helperContent := `package test

func foo() string {
	return "helper"
}
`
	err = os.WriteFile(helperFile, []byte(helperContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write helper file: %v", err)
	}

	// Create a caller file that uses the unexported function
	callerFile := filepath.Join(tmpDir, "caller.go")
	callerContent := `package test

func bar() {
	foo()
}
`
	err = os.WriteFile(callerFile, []byte(callerContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write caller file: %v", err)
	}

	// Initialize go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Load the package
	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	// Analyze dependencies of the caller file
	deps, err := analyzeMoveDependencies(callerFile, pkg)
	if err != nil {
		t.Fatalf("analyzeMoveDependencies failed: %v", err)
	}

	// Assert: Returns exactly 1 UnexportedDependency
	if len(deps) != 1 {
		t.Fatalf("Expected 1 dependency, got %d", len(deps))
	}

	// Assert: Dependency name is "foo"
	if deps[0].Name != "foo" {
		t.Errorf("Expected dependency name 'foo', got '%s'", deps[0].Name)
	}

	// Assert: Dependency kind is "func"
	if deps[0].Kind != "func" {
		t.Errorf("Expected dependency kind 'func', got '%s'", deps[0].Kind)
	}
}

func TestAnalyzeMoveDependenciesMultipleTypes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "deps_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create file with various unexported symbols
	defsFile := filepath.Join(tmpDir, "defs.go")
	defsContent := `package test

func helperFunc() {}
var helperVar = 42
type helperType struct{}
const helperConst = "const"
`
	err = os.WriteFile(defsFile, []byte(defsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write defs file: %v", err)
	}

	// Create caller using all of them
	callerFile := filepath.Join(tmpDir, "caller.go")
	callerContent := `package test

func use() {
	helperFunc()
	_ = helperVar
	_ = helperType{}
	_ = helperConst
}
`
	err = os.WriteFile(callerFile, []byte(callerContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write caller file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	deps, err := analyzeMoveDependencies(callerFile, pkg)
	if err != nil {
		t.Fatalf("analyzeMoveDependencies failed: %v", err)
	}

	// Should find all 4 unexported dependencies
	if len(deps) != 4 {
		t.Fatalf("Expected 4 dependencies, got %d", len(deps))
	}

	// Check that we have each kind
	kinds := make(map[string]int)
	for _, dep := range deps {
		kinds[dep.Kind]++
	}

	if kinds["func"] != 1 {
		t.Errorf("Expected 1 func, got %d", kinds["func"])
	}
	if kinds["var"] != 1 {
		t.Errorf("Expected 1 var, got %d", kinds["var"])
	}
	if kinds["type"] != 1 {
		t.Errorf("Expected 1 type, got %d", kinds["type"])
	}
	if kinds["const"] != 1 {
		t.Errorf("Expected 1 const, got %d", kinds["const"])
	}
}

func TestAnalyzeMoveDependenciesExportedIgnored(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "deps_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	defsFile := filepath.Join(tmpDir, "defs.go")
	defsContent := `package test

func ExportedFunc() {}
func unexportedFunc() {}
`
	err = os.WriteFile(defsFile, []byte(defsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write defs file: %v", err)
	}

	callerFile := filepath.Join(tmpDir, "caller.go")
	callerContent := `package test

func use() {
	ExportedFunc()
	unexportedFunc()
}
`
	err = os.WriteFile(callerFile, []byte(callerContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write caller file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	deps, err := analyzeMoveDependencies(callerFile, pkg)
	if err != nil {
		t.Fatalf("analyzeMoveDependencies failed: %v", err)
	}

	// Should only find the unexported one
	if len(deps) != 1 {
		t.Fatalf("Expected 1 unexported dependency, got %d", len(deps))
	}

	if deps[0].Name != "unexportedFunc" {
		t.Errorf("Expected 'unexportedFunc', got '%s'", deps[0].Name)
	}
}

func TestAnalyzeMoveDependenciesSelfReferencesIgnored(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "deps_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// File that defines and uses its own unexported symbol
	file := filepath.Join(tmpDir, "self.go")
	content := `package test

func helper() {
	// Call itself recursively
	helper()
}
`
	err = os.WriteFile(file, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	deps, err := analyzeMoveDependencies(file, pkg)
	if err != nil {
		t.Fatalf("analyzeMoveDependencies failed: %v", err)
	}

	// Should find no dependencies (self-references don't count)
	if len(deps) != 0 {
		t.Errorf("Expected 0 dependencies for self-reference, got %d", len(deps))
	}
}

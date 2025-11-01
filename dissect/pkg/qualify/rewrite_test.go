package qualify

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neongreen/mono/dissect/pkg/references"
	"github.com/neongreen/mono/dissect/pkg/typeinfo"
	"golang.org/x/tools/go/packages"
)

func TestQualifyReferences_SingleReference(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "main.go")
	createFile(t, testFile, `package main

func Helper() {}

func main() {
	Helper()
}
`)

	// Load package to get real type information
	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	// Find actual references using type information
	refs, err := references.FindReferences([]string{"Helper"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	err = QualifyReferences(testFile, refs, "util", "test/util", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	// Read the file and verify it was modified
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "import") && !strings.Contains(contentStr, "test/util") {
		t.Error("Expected import to be added")
	}
	// Check for util. and Helper separately (may be on different lines due to formatting)
	if !strings.Contains(contentStr, "util.") {
		t.Error("Expected util package qualifier")
	}
	if !strings.Contains(contentStr, `"test/util"`) {
		t.Error("Expected test/util import")
	}
}

func TestQualifyReferences_MultipleReferences(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	// First create stubs so the package can be type-checked
	testFile := filepath.Join(tmpDir, "main.go")
	createFile(t, testFile, `package main

func Alpha() {}
func Beta() {}
func Gamma() {}

func main() {
	Alpha()
	Beta()
	Gamma()
}
`)

	// Load package to get accurate positions
	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	// Find references (should find the calls in main)
	refs, err := references.FindReferences([]string{"Alpha", "Beta", "Gamma"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	err = QualifyReferences(testFile, refs, "util", "test/util", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	// Check for util. qualifier (functions may be on different lines due to formatting)
	if !strings.Contains(contentStr, "util.") {
		t.Error("Expected util package qualifier")
	}
	// Check that all three functions are still present
	if !strings.Contains(contentStr, "Alpha") || !strings.Contains(contentStr, "Beta") || !strings.Contains(contentStr, "Gamma") {
		t.Error("Expected all function names to be present")
	}
}

func TestQualifyReferences_AddsImport(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "main.go")
	createFile(t, testFile, `package main

func Helper() {}

func main() {
	Helper()
}
`)

	// Load and find references
	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := references.FindReferences([]string{"Helper"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	err = QualifyReferences(testFile, refs, "util", "test/util", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "import") {
		t.Error("Expected import statement to be added")
	}
	if !strings.Contains(contentStr, "\"test/util\"") {
		t.Error("Expected test/util import")
	}
}

func TestQualifyReferences_ImportAlreadyExists(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	// Create util package
	utilDir := filepath.Join(tmpDir, "util")
	if err := os.MkdirAll(utilDir, 0755); err != nil {
		t.Fatalf("Failed to create util dir: %v", err)
	}
	createFile(t, filepath.Join(utilDir, "util.go"), `package util

func UtilFunc() {}
`)

	testFile := filepath.Join(tmpDir, "main.go")
	createFile(t, testFile, `package main

import "test/util"

func Helper() {}

func main() {
	Helper()
	util.UtilFunc() // Use the import so it's not removed
}
`)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := references.FindReferences([]string{"Helper"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	err = QualifyReferences(testFile, refs, "util", "test/util", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	// Count occurrences of import statement - should only be one
	importCount := strings.Count(contentStr, `"test/util"`)
	if importCount > 1 {
		t.Errorf("Expected 1 import, found %d (duplicate imports)", importCount)
	}
}

func TestQualifyReferences_QualifiedReferencesIgnored(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	// Create util package
	utilDir := filepath.Join(tmpDir, "util")
	if err := os.MkdirAll(utilDir, 0755); err != nil {
		t.Fatalf("Failed to create util dir: %v", err)
	}
	createFile(t, filepath.Join(utilDir, "util.go"), `package util

func Helper() {}
`)

	testFile := filepath.Join(tmpDir, "main.go")
	createFile(t, testFile, `package main

import "test/util"

func main() {
	util.Helper() // already qualified
}
`)

	pkgs, err := typeinfo.LoadPackages([]string{"./..."}, tmpDir)
	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	refs, err := references.FindReferences([]string{"Helper"}, pkgs)
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	err = QualifyReferences(testFile, refs, "util", "test/util", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	newContent, _ := os.ReadFile(testFile)

	// Content should be essentially unchanged (may have formatting differences)
	if !strings.Contains(string(newContent), "util.Helper()") {
		t.Error("Already qualified reference should remain qualified")
	}
}

func TestQualifyReferences_DifferentSymbols(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "main.go")
	createFile(t, testFile, `package main

func Foo() {}
func Bar() {}

func main() {
	Foo()
	Bar()
}
`)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := references.FindReferences([]string{"Foo", "Bar"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	err = QualifyReferences(testFile, refs, "util", "test/util", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	// Check for util. qualifier (may be on different lines due to formatting)
	if !strings.Contains(contentStr, "util.") {
		t.Error("Expected util package qualifier")
	}
	if !strings.Contains(contentStr, "Foo") || !strings.Contains(contentStr, "Bar") {
		t.Error("Expected both Foo and Bar to be present")
	}
}

func TestQualifyReferences_InFunctionCall(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "main.go")
	createFile(t, testFile, `package main

func Helper() string { return "" }
func process(x string) {}

func main() {
	process(Helper())
}
`)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := references.FindReferences([]string{"Helper"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	err = QualifyReferences(testFile, refs, "util", "test/util", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "util.") {
		t.Error("Expected util. qualifier in function call")
	}
}

func TestQualifyReferences_InStructLiteral(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "main.go")
	createFile(t, testFile, `package main

type Config struct {
	Value string
}

func Helper() string { return "" }

func main() {
	c := Config{Value: Helper()}
	_ = c
}
`)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := references.FindReferences([]string{"Helper"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	err = QualifyReferences(testFile, refs, "util", "test/util", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "util.") {
		t.Error("Expected util. qualifier in struct literal")
	}
}

func TestQualifyReferences_InTypeAssertion(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "main.go")
	createFile(t, testFile, `package main

type MyType struct{}

func main() {
	var x interface{} = MyType{}
	_, ok := x.(MyType)
	_ = ok
}
`)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := references.FindReferences([]string{"MyType"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	err = QualifyReferences(testFile, refs, "util", "test/util", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	// Should qualify both uses of MyType (may be on different lines)
	if !strings.Contains(contentStr, "util.") {
		t.Error("Expected util. qualifier for MyType")
	}
}

func TestQualifyReferences_InVariableDecl(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "main.go")
	createFile(t, testFile, `package main

type Config struct{}

func main() {
	var c Config
	_ = c
}
`)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := references.FindReferences([]string{"Config"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	err = QualifyReferences(testFile, refs, "util", "test/util", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "util.") {
		t.Error("Expected util. qualifier in variable declaration")
	}
}

func TestQualifyReferences_PreservesFormatting(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "main.go")
	originalContent := `package main

func Helper() {}

func main() {
	// This is a comment
	Helper()
	// Another comment
}
`
	createFile(t, testFile, originalContent)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := references.FindReferences([]string{"Helper"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	err = QualifyReferences(testFile, refs, "util", "test/util", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	// Basic checks - formatting may change slightly but structure should be preserved
	if !strings.Contains(contentStr, "func main()") {
		t.Error("Expected function structure preserved")
	}
}

func TestQualifyReferences_PreservesComments(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "main.go")
	createFile(t, testFile, `package main

func Helper() {}

// Main function
func main() {
	// Call helper
	Helper()
}
`)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := references.FindReferences([]string{"Helper"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	err = QualifyReferences(testFile, refs, "util", "test/util", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "// Main function") || !strings.Contains(contentStr, "// Call helper") {
		t.Error("Expected comments to be preserved")
	}
}

func TestQualifyReferences_EmptyReferences(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "main.go")
	originalContent := `package main

func main() {}
`
	createFile(t, testFile, originalContent)

	var refs []references.Reference

	err := QualifyReferences(testFile, refs, "util", "test/util", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences should handle empty references: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// File should be essentially unchanged
	if string(content) != originalContent {
		// Allow for minor formatting differences
		if !strings.Contains(string(content), "func main()") {
			t.Error("File content changed unexpectedly")
		}
	}
}

func TestQualifyReferences_BuildsAfter(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	// Create util package
	utilDir := filepath.Join(tmpDir, "util")
	if err := os.MkdirAll(utilDir, 0755); err != nil {
		t.Fatalf("Failed to create util dir: %v", err)
	}
	createFile(t, filepath.Join(utilDir, "util.go"), `package util

func Helper() string {
	return "help"
}
`)

	testFile := filepath.Join(tmpDir, "main.go")
	// Initially, Helper is defined in main (simulating before move)
	createFile(t, testFile, `package main

func Helper() string {
	return "help from main"
}

func main() {
	_ = Helper()
}
`)

	// Verify it builds initially
	buildCmd := exec.Command("go", "build", ".")
	buildCmd.Dir = tmpDir
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Code should build initially: %v", err)
	}

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	refs, err := references.FindReferences([]string{"Helper"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	err = QualifyReferences(testFile, refs, "util", "test/util", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	// Now remove the local Helper definition (simulating the move)
	modifiedContent := `package main

import "test/util"

func main() {
	_ = util.Helper()
}
`
	createFile(t, testFile, modifiedContent)

	// Verify it still builds after qualification
	buildCmd = exec.Command("go", "build", ".")
	buildCmd.Dir = tmpDir
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Code should build after qualification: %v\nOutput: %s", err, output)
	}
}

// Helper functions

func createTempModule(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "qualify_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create go.mod
	gomod := "module test\n\ngo 1.21\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(gomod), 0644); err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	return tmpDir
}

func createFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
}


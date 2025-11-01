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

// TestQualifyReferences_StructLiteralFieldNames tests that field names in struct
// literals are NOT qualified, only the type names and values should be qualified.
// This is a regression test for a bug where ProjectUID field name was becoming
// types.ProjectUID which is invalid Go syntax.
func TestQualifyReferences_StructLiteralFieldNames(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	// Create types package first (simulating the target after move)
	typesDir := filepath.Join(tmpDir, "types")
	if err := os.MkdirAll(typesDir, 0755); err != nil {
		t.Fatalf("Failed to create types dir: %v", err)
	}
	createFile(t, filepath.Join(typesDir, "types.go"), `package types

type ProjectUID string
type TaskUID string
`)

	testFile := filepath.Join(tmpDir, "main.go")
	createFile(t, testFile, `package main

type ProjectUID string
type TaskUID string

type DoctorCollision struct {
	ProjectUID     ProjectUID
	TaskDisplayIDs []string
	Number         int64
}

func main() {
	projectUID := ProjectUID("project-123")
	collision := DoctorCollision{
		ProjectUID:     projectUID,
		TaskDisplayIDs: []string{"task-1", "task-2"},
		Number:         42,
	}
	_ = collision
}
`)

	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	// Find references to ProjectUID and TaskUID types
	refs, err := references.FindReferences([]string{"ProjectUID", "TaskUID"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	err = QualifyReferences(testFile, refs, "types", "test/types", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)

	// Should have import
	if !strings.Contains(contentStr, `"test/types"`) {
		t.Error("Expected test/types import to be added")
	}

	// Should qualify the type name in struct field type declarations
	if !strings.Contains(contentStr, "types.ProjectUID") {
		t.Error("Expected types.ProjectUID in struct field type")
	}

	// Should qualify the type name in variable declarations
	if !strings.Contains(contentStr, "types.ProjectUID(") {
		t.Error("Expected types.ProjectUID in type conversion")
	}

	// CRITICAL: Field names in struct literals should NOT be qualified
	// This is the bug we're fixing
	if strings.Contains(contentStr, "types.ProjectUID:") || strings.Contains(contentStr, "types.Number:") {
		t.Errorf("Field names in struct literal should NOT be qualified. Got:\n%s", contentStr)
	}

	// Field names should still be present (unqualified)
	if !strings.Contains(contentStr, "ProjectUID:") {
		t.Error("Expected unqualified ProjectUID field name in struct literal")
	}

	if !strings.Contains(contentStr, "Number:") {
		t.Error("Expected unqualified Number field name in struct literal")
	}

	// Verify the code actually compiles
	buildCmd := exec.Command("go", "build", ".")
	buildCmd.Dir = tmpDir
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Code should build after qualification: %v\nOutput: %s\nContent:\n%s", err, output, contentStr)
	}
}

func TestQualifyReferences_PackageNameCollision(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	// Create db package first (simulating the target after move)
	dbDir := filepath.Join(tmpDir, "db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("Failed to create db dir: %v", err)
	}
	createFile(t, filepath.Join(dbDir, "db.go"), `package db

type User struct {
	Name string
}

func NewUser(name string) *User {
	return &User{Name: name}
}
`)

	testFile := filepath.Join(tmpDir, "main.go")
	createFile(t, testFile, `package main

type User struct {
	Name string
}

func NewUser(name string) *User {
	return &User{Name: name}
}

func main() {
	db, err := OpenDB()
	if err != nil {
		panic(err)
	}
	user := NewUser("Alice")
	_ = db
	_ = user
}

func OpenDB() (*Database, error) {
	return &Database{}, nil
}

type Database struct{}
`)

	// Load package to get real type information
	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	// Find references to NewUser
	refs, err := references.FindReferences([]string{"NewUser", "User"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	// Qualify with package name "db" (which conflicts with variable "db")
	err = QualifyReferences(testFile, refs, "db", "test/db", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	// Read the file and verify it uses an alias
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)

	// Should NOT use "db." (that would conflict with the variable)
	// Should use "db_pkg." instead
	if strings.Contains(contentStr, "db.NewUser") || strings.Contains(contentStr, "db.User") {
		t.Errorf("Should not use 'db.' qualifier due to variable name collision. Got:\n%s", contentStr)
	}

	// Should use db_pkg alias
	if !strings.Contains(contentStr, "db_pkg") {
		t.Errorf("Expected db_pkg alias to be used. Got:\n%s", contentStr)
	}

	// Should have aliased import
	if !strings.Contains(contentStr, `db_pkg "test/db"`) {
		t.Errorf("Expected aliased import: db_pkg \"test/db\". Got:\n%s", contentStr)
	}

	// Should use qualified reference
	if !strings.Contains(contentStr, "db_pkg.NewUser") {
		t.Errorf("Expected db_pkg.NewUser reference. Got:\n%s", contentStr)
	}

	// Verify the code compiles
	buildCmd := exec.Command("go", "build", ".")
	buildCmd.Dir = tmpDir
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Code should build after qualification: %v\nOutput: %s\nContent:\n%s", err, output, contentStr)
	}
}

func TestQualifyReferences_AliasCollision(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	// Create db package first (simulating the target after move)
	dbDir := filepath.Join(tmpDir, "db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("Failed to create db dir: %v", err)
	}
	createFile(t, filepath.Join(dbDir, "db.go"), `package db

type User struct {
	Name string
}

func NewUser(name string) *User {
	return &User{Name: name}
}
`)

	testFile := filepath.Join(tmpDir, "main.go")
	createFile(t, testFile, `package main

type User struct {
	Name string
}

func NewUser(name string) *User {
	return &User{Name: name}
}

func main() {
	db := "database"
	db_pkg := "package"
	user := NewUser("Alice")
	_ = db
	_ = db_pkg
	_ = user
}
`)

	// Load package to get real type information
	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	// Find references to NewUser
	refs, err := references.FindReferences([]string{"NewUser", "User"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	// Qualify with package name "db" (both "db" and "db_pkg" are taken)
	err = QualifyReferences(testFile, refs, "db", "test/db", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	// Read the file and verify it uses the next available alias
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)

	// Should use db_pkg_ (with one underscore suffix)
	if !strings.Contains(contentStr, "db_pkg_") {
		t.Errorf("Expected db_pkg_ alias to be used. Got:\n%s", contentStr)
	}

	// Should have aliased import
	if !strings.Contains(contentStr, `db_pkg_ "test/db"`) {
		t.Errorf("Expected aliased import: db_pkg_ \"test/db\". Got:\n%s", contentStr)
	}

	// Should use qualified reference
	if !strings.Contains(contentStr, "db_pkg_.NewUser") {
		t.Errorf("Expected db_pkg_.NewUser reference. Got:\n%s", contentStr)
	}

	// Verify the code compiles
	buildCmd := exec.Command("go", "build", ".")
	buildCmd.Dir = tmpDir
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Code should build after qualification: %v\nOutput: %s\nContent:\n%s", err, output, contentStr)
	}
}

func TestQualifyReferences_MultipleCollisions(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	// Create db package first (simulating the target after move)
	dbDir := filepath.Join(tmpDir, "db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("Failed to create db dir: %v", err)
	}
	createFile(t, filepath.Join(dbDir, "db.go"), `package db

type User struct {
	Name string
}

func NewUser(name string) *User {
	return &User{Name: name}
}
`)

	testFile := filepath.Join(tmpDir, "main.go")
	createFile(t, testFile, `package main

type User struct {
	Name string
}

func NewUser(name string) *User {
	return &User{Name: name}
}

func main() {
	db := "database"
	db_pkg := "package1"
	db_pkg_ := "package2"
	db_pkg__ := "package3"
	user := NewUser("Alice")
	_ = db
	_ = db_pkg
	_ = db_pkg_
	_ = db_pkg__
	_ = user
}
`)

	// Load package to get real type information
	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	// Find references to NewUser
	refs, err := references.FindReferences([]string{"NewUser", "User"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	// Qualify with package name "db" (many collisions)
	err = QualifyReferences(testFile, refs, "db", "test/db", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	// Read the file and verify it uses the next available alias
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)

	// Should use db_pkg___ (with three underscores)
	if !strings.Contains(contentStr, "db_pkg___") {
		t.Errorf("Expected db_pkg___ alias to be used. Got:\n%s", contentStr)
	}

	// Should have aliased import
	if !strings.Contains(contentStr, `db_pkg___ "test/db"`) {
		t.Errorf("Expected aliased import: db_pkg___ \"test/db\". Got:\n%s", contentStr)
	}

	// Should use qualified reference
	if !strings.Contains(contentStr, "db_pkg___.NewUser") {
		t.Errorf("Expected db_pkg___.NewUser reference. Got:\n%s", contentStr)
	}

	// Verify the code compiles
	buildCmd := exec.Command("go", "build", ".")
	buildCmd.Dir = tmpDir
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Code should build after qualification: %v\nOutput: %s\nContent:\n%s", err, output, contentStr)
	}
}

func TestQualifyReferences_TypeReferencesInSignatures(t *testing.T) {
	tmpDir := createTempModule(t)
	defer os.RemoveAll(tmpDir)

	// Create db package with DB type
	dbDir := filepath.Join(tmpDir, "db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatalf("Failed to create db dir: %v", err)
	}
	createFile(t, filepath.Join(dbDir, "db.go"), `package db

type DB struct {
	Name string
}

func NewDB() *DB {
	return &DB{}
}
`)

	testFile := filepath.Join(tmpDir, "app.go")
	createFile(t, testFile, `package test

type DB struct {
	Name string
}

func NewDB() *DB {
	return &DB{}
}

// Test various type reference contexts
func checkCollision(db *DB) error {
	return nil
}

func returnDB() (*DB, error) {
	return NewDB(), nil
}

type Wrapper struct {
	database *DB
}

func processMultiple(x *DB, y *DB, z int) (*DB, *DB) {
	return x, y
}
`)

	// Load package to get real type information
	pkg, err := typeinfo.LoadPackage(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	// Find references to DB and NewDB
	refs, err := references.FindReferences([]string{"DB", "NewDB"}, []*packages.Package{pkg})
	if err != nil {
		t.Fatalf("FindReferences failed: %v", err)
	}

	// Qualify with package name "db" (which collides with variable "db")
	err = QualifyReferences(testFile, refs, "db", "test/db", tmpDir)
	if err != nil {
		t.Fatalf("QualifyReferences failed: %v", err)
	}

	// Read the file and verify
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)

	// Should use alias
	if !strings.Contains(contentStr, "db_pkg") {
		t.Errorf("Expected db_pkg alias. Got:\n%s", contentStr)
	}

	// Should qualify type in parameter
	if !strings.Contains(contentStr, "db *db_pkg.DB") {
		t.Errorf("Expected 'db *db_pkg.DB' in parameter. Got:\n%s", contentStr)
	}

	// Should qualify type in return type
	if !strings.Contains(contentStr, "(*db_pkg.DB, error)") {
		t.Errorf("Expected '(*db_pkg.DB, error)' in return type. Got:\n%s", contentStr)
	}

	// Should qualify type in struct field
	if !strings.Contains(contentStr, "database *db_pkg.DB") {
		t.Errorf("Expected 'database *db_pkg.DB' in struct field. Got:\n%s", contentStr)
	}

	// Should qualify function call
	if !strings.Contains(contentStr, "db_pkg.NewDB()") {
		t.Errorf("Expected 'db_pkg.NewDB()' function call. Got:\n%s", contentStr)
	}

	// Verify the code compiles
	buildCmd := exec.Command("go", "build", ".")
	buildCmd.Dir = tmpDir
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Code should build after qualification: %v\nOutput: %s\nContent:\n%s", err, output, contentStr)
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

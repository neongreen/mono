package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestBatchMove_WithPackageNameCollision tests the tk scenario where
// the target package name conflicts with a variable name
func TestBatchMove_WithPackageNameCollision(t *testing.T) {
	dissect := buildDissectBinary(t)
	tmpDir := createTempModuleForBatch(t)

	// Create main package with a variable named "db"
	createFileForBatch(t, filepath.Join(tmpDir, "main.go"), `package main

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

	// Create types file with User type
	createFileForBatch(t, filepath.Join(tmpDir, "types.go"), `package main

type User struct {
	Name string
}

func NewUser(name string) *User {
	return &User{Name: name}
}
`)

	// Run batch move to move types to db package
	cmd := exec.Command(dissect, "move", "--batch", "types.go -> db/")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dissect move --batch failed: %v\nOutput: %s", err, output)
	}

	// Read main.go
	mainContent, err := os.ReadFile(filepath.Join(tmpDir, "main.go"))
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}

	mainStr := string(mainContent)

	// Should use aliased import (db_pkg) not bare "db"
	if !strings.Contains(mainStr, "db_pkg") {
		t.Errorf("Expected db_pkg alias to be used due to variable collision. Got:\n%s", mainStr)
	}

	// Should have aliased import
	if !strings.Contains(mainStr, `db_pkg "test/batch/db"`) {
		t.Errorf("Expected aliased import. Got:\n%s", mainStr)
	}

	// Should use qualified reference with alias
	if !strings.Contains(mainStr, "db_pkg.NewUser") {
		t.Errorf("Expected db_pkg.NewUser reference. Got:\n%s", mainStr)
	}

	// Should NOT use bare "db." (that would conflict with variable)
	if strings.Contains(mainStr, "db.NewUser") || strings.Contains(mainStr, "db.User") {
		t.Errorf("Should not use 'db.' qualifier. Got:\n%s", mainStr)
	}

	// Verify types.go was moved and has correct package
	dbTypesPath := filepath.Join(tmpDir, "db/types.go")
	if _, err := os.Stat(dbTypesPath); os.IsNotExist(err) {
		t.Error("types.go was not moved to db/")
	}

	dbContent, err := os.ReadFile(dbTypesPath)
	if err != nil {
		t.Fatalf("Failed to read db/types.go: %v", err)
	}

	if !strings.Contains(string(dbContent), "package db") {
		t.Error("Package declaration was not updated to 'package db'")
	}

	// Verify code builds
	buildCmd := exec.Command("go", "build", "./...")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code should build after batch move: %v\nOutput: %s\nmain.go:\n%s", err, output, mainStr)
	}
}

// createTempModule creates a temporary Go module for testing
// buildDissectBinary builds the dissect binary for testing and returns its path
func buildDissectBinary(t *testing.T) string {
	// Build dissect to a temp location
	tmpBinary := filepath.Join(t.TempDir(), "dissect")
	cmd := exec.Command("go", "build", "-o", tmpBinary, ".")
	cmd.Dir = filepath.Join(".") // Current directory (dissect/cmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build dissect: %v\nOutput: %s", err, output)
	}
	return tmpBinary
}

func createTempModuleForBatch(t *testing.T) string {
	tmpDir := t.TempDir()

	// Create go.mod
	goModContent := `module test/batch

go 1.21
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	return tmpDir
}

// createFile creates a file with the given content
func createFileForBatch(t *testing.T, path, content string) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file %s: %v", path, err)
	}
}

// TestBatchMove_BasicDirectory tests moving multiple files to a directory
func TestBatchMove_BasicDirectory(t *testing.T) {
	dissect := buildDissectBinary(t)
	tmpDir := createTempModuleForBatch(t)

	// Create test files
	createFileForBatch(t, filepath.Join(tmpDir, "a.go"), `package main

func FuncA() string {
	return "a"
}
`)

	createFileForBatch(t, filepath.Join(tmpDir, "b.go"), `package main

func FuncB() string {
	return "b"
}
`)

	createFileForBatch(t, filepath.Join(tmpDir, "c.go"), `package main

func FuncC() string {
	return "c"
}
`)

	// Run batch move
	cmd := exec.Command(dissect, "move", "--batch", "a.go,b.go,c.go -> target/")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dissect move --batch failed: %v\nOutput: %s", err, output)
	}

	// Verify files were moved
	for _, file := range []string{"a.go", "b.go", "c.go"} {
		targetPath := filepath.Join(tmpDir, "target", file)
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			t.Errorf("File %s was not moved to target/", file)
		}

		// Check old location is gone
		oldPath := filepath.Join(tmpDir, file)
		if _, err := os.Stat(oldPath); err == nil {
			t.Errorf("File %s still exists at old location", file)
		}

		// Check package was updated
		content, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", targetPath, err)
		}
		if !strings.Contains(string(content), "package target") {
			t.Errorf("Package declaration in %s was not updated to 'package target'", file)
		}
	}

	// Verify code still builds
	buildCmd := exec.Command("go", "build", "./...")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Errorf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

// TestBatchMove_FileRename tests renaming a single file
func TestBatchMove_FileRename(t *testing.T) {
	dissect := buildDissectBinary(t)
	tmpDir := createTempModuleForBatch(t)

	// Create source file
	createFileForBatch(t, filepath.Join(tmpDir, "old.go"), `package main

func Foo() string {
	return "foo"
}
`)

	// Create main file that uses it (same package)
	createFileForBatch(t, filepath.Join(tmpDir, "main.go"), `package main

func main() {
	_ = Foo()
}
`)

	// Run batch move (rename)
	cmd := exec.Command(dissect, "move", "--batch", "old.go -> new.go")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dissect move --batch failed: %v\nOutput: %s", err, output)
	}

	// Verify file was renamed
	newPath := filepath.Join(tmpDir, "new.go")
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("File was not renamed to new.go")
	}

	// Check old location is gone
	oldPath := filepath.Join(tmpDir, "old.go")
	if _, err := os.Stat(oldPath); err == nil {
		t.Error("File old.go still exists at old location")
	}

	// Verify code still builds
	buildCmd := exec.Command("go", "build", "./...")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Errorf("Code doesn't build after rename: %v\nOutput: %s", err, output)
	}
}

// TestBatchMove_MultipleGroups tests multiple groups in one command
func TestBatchMove_MultipleGroups(t *testing.T) {
	dissect := buildDissectBinary(t)
	tmpDir := createTempModuleForBatch(t)

	// Create test files
	createFileForBatch(t, filepath.Join(tmpDir, "db.go"), `package main

type DB struct {
	Name string
}
`)

	createFileForBatch(t, filepath.Join(tmpDir, "util.go"), `package main

func Helper() string {
	return "help"
}
`)

	// Run batch move with multiple groups
	cmd := exec.Command(dissect, "move", "--batch",
		"db.go -> internal/db/",
		"util.go -> internal/utils/")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dissect move --batch failed: %v\nOutput: %s", err, output)
	}

	// Verify files were moved to correct locations
	dbPath := filepath.Join(tmpDir, "internal/db/db.go")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("db.go was not moved to internal/db/")
	}

	utilPath := filepath.Join(tmpDir, "internal/utils/util.go")
	if _, err := os.Stat(utilPath); os.IsNotExist(err) {
		t.Error("util.go was not moved to internal/utils/")
	}

	// Check packages were updated
	dbContent, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("Failed to read db.go: %v", err)
	}
	if !strings.Contains(string(dbContent), "package db") {
		t.Error("Package declaration in db.go was not updated to 'package db'")
	}

	utilContent, err := os.ReadFile(utilPath)
	if err != nil {
		t.Fatalf("Failed to read util.go: %v", err)
	}
	if !strings.Contains(string(utilContent), "package utils") {
		t.Error("Package declaration in util.go was not updated to 'package utils'")
	}

	// Verify code still builds
	buildCmd := exec.Command("go", "build", "./...")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Errorf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

// TestBatchMove_WithMutualReferences tests files that reference each other
func TestBatchMove_WithMutualReferences(t *testing.T) {
	dissect := buildDissectBinary(t)
	tmpDir := createTempModuleForBatch(t)

	// Create files that reference each other
	createFileForBatch(t, filepath.Join(tmpDir, "a.go"), `package main

func A() {
	B()
}
`)

	createFileForBatch(t, filepath.Join(tmpDir, "b.go"), `package main

func B() {
	A()
}
`)

	// Run batch move
	cmd := exec.Command(dissect, "move", "--batch", "a.go,b.go -> internal/pkg/")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dissect move --batch failed: %v\nOutput: %s", err, output)
	}

	// Verify files were moved
	for _, file := range []string{"a.go", "b.go"} {
		targetPath := filepath.Join(tmpDir, "internal/pkg", file)
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			t.Errorf("File %s was not moved to internal/pkg/", file)
		}

		// Check package was updated
		content, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", targetPath, err)
		}
		if !strings.Contains(string(content), "package pkg") {
			t.Errorf("Package declaration in %s was not updated", file)
		}
	}

	// Verify code still builds
	buildCmd := exec.Command("go", "build", "./...")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Errorf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

// TestBatchMove_GlobExpansion tests using glob patterns in sources
func TestBatchMove_GlobExpansion(t *testing.T) {
	dissect := buildDissectBinary(t)
	tmpDir := createTempModuleForBatch(t)

	// Create test files
	createFileForBatch(t, filepath.Join(tmpDir, "db.go"), `package main

type DB struct{}
`)

	createFileForBatch(t, filepath.Join(tmpDir, "db_events.go"), `package main

type Event struct{}
`)

	createFileForBatch(t, filepath.Join(tmpDir, "main.go"), `package main

func main() {}
`)

	// Run batch move with glob
	cmd := exec.Command(dissect, "move", "--batch", "db*.go -> internal/db/")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dissect move --batch failed: %v\nOutput: %s", err, output)
	}

	// Verify only db*.go files were moved
	for _, file := range []string{"db.go", "db_events.go"} {
		targetPath := filepath.Join(tmpDir, "internal/db", file)
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			t.Errorf("File %s was not moved to internal/db/", file)
		}
	}

	// Verify main.go stayed
	mainPath := filepath.Join(tmpDir, "main.go")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		t.Error("main.go was incorrectly moved")
	}

	// Verify code still builds
	buildCmd := exec.Command("go", "build", "./...")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Errorf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

// TestBatchMove_Rollback_SourceNotFound tests validation failure
func TestBatchMove_Rollback_SourceNotFound(t *testing.T) {
	dissect := buildDissectBinary(t)
	tmpDir := createTempModuleForBatch(t)

	// Try to move non-existent file
	cmd := exec.Command(dissect, "move", "--batch", "nonexistent.go -> target/")
	cmd.Dir = tmpDir
	_, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("Expected error for non-existent file, got success")
	}

	// Verify no target directory was created
	targetDir := filepath.Join(tmpDir, "target")
	if _, err := os.Stat(targetDir); err == nil {
		t.Error("Target directory should not have been created")
	}
}

// TestBatchMove_Rollback_MultipleToFile tests validation failure for multiple sources to file
func TestBatchMove_Rollback_MultipleToFile(t *testing.T) {
	dissect := buildDissectBinary(t)
	tmpDir := createTempModuleForBatch(t)

	// Create test files
	createFileForBatch(t, filepath.Join(tmpDir, "a.go"), `package main

func A() {}
`)

	createFileForBatch(t, filepath.Join(tmpDir, "b.go"), `package main

func B() {}
`)

	// Try to move multiple files to a single file
	cmd := exec.Command(dissect, "move", "--batch", "a.go,b.go -> single.go")
	cmd.Dir = tmpDir
	_, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("Expected error for multiple files to single file target, got success")
	}

	// Verify files weren't moved
	if _, err := os.Stat(filepath.Join(tmpDir, "a.go")); os.IsNotExist(err) {
		t.Error("a.go should not have been moved")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "b.go")); os.IsNotExist(err) {
		t.Error("b.go should not have been moved")
	}
}

// TestBatchMove_PreservesComments tests that comments are preserved
func TestBatchMove_UpdatesSymbolReferences(t *testing.T) {
	dissect := buildDissectBinary(t)
	tmpDir := createTempModuleForBatch(t)

	// Create main package with types
	createFileForBatch(t, filepath.Join(tmpDir, "types.go"), `package main

type User struct {
	Name string
}

type Product struct {
	ID int
}

func NewUser(name string) *User {
	return &User{Name: name}
}
`)

	// Create main.go that uses the types
	createFileForBatch(t, filepath.Join(tmpDir, "main.go"), `package main

func main() {
	user := NewUser("Alice")
	product := Product{ID: 123}
	_ = user
	_ = product
}
`)

	// Move types to internal/models/
	cmd := exec.Command(dissect, "move", "--batch", "types.go -> internal/models/")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Batch move failed: %v\nOutput: %s", err, output)
	}

	// Verify main.go was updated with imports and qualified references
	mainContent, err := os.ReadFile(filepath.Join(tmpDir, "main.go"))
	if err != nil {
		t.Fatalf("Failed to read main.go: %v", err)
	}

	mainStr := string(mainContent)

	// Should have import for internal/models
	if !strings.Contains(mainStr, `"test/batch/internal/models"`) {
		t.Errorf("Expected import for test/batch/internal/models, got:\n%s", mainStr)
	}

	// Should qualify type references
	if !strings.Contains(mainStr, "models.NewUser") {
		t.Errorf("Expected qualified function call models.NewUser, got:\n%s", mainStr)
	}

	if !strings.Contains(mainStr, "models.Product") {
		t.Errorf("Expected qualified type models.Product, got:\n%s", mainStr)
	}

	// Verify code builds
	buildCmd := exec.Command("go", "build", "./...")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code should build after batch move: %v\nOutput: %s", err, output)
	}
}

func TestBatchMove_PreservesComments(t *testing.T) {
	dissect := buildDissectBinary(t)
	tmpDir := createTempModuleForBatch(t)

	// Create file with doc comments
	createFileForBatch(t, filepath.Join(tmpDir, "doc.go"), `package main

// MyFunc does something important.
// It has multiple lines of documentation.
func MyFunc() string {
	return "value"
}
`)

	// Run batch move
	cmd := exec.Command(dissect, "move", "--batch", "doc.go -> internal/pkg/")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("dissect move --batch failed: %v\nOutput: %s", err, output)
	}

	// Read moved file
	targetPath := filepath.Join(tmpDir, "internal/pkg/doc.go")
	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("Failed to read moved file: %v", err)
	}

	// Verify comments are present
	contentStr := string(content)
	if !strings.Contains(contentStr, "// MyFunc does something important.") {
		t.Error("Doc comment was not preserved")
	}
	if !strings.Contains(contentStr, "// It has multiple lines of documentation.") {
		t.Error("Multi-line doc comment was not preserved")
	}

	// Verify code still builds
	buildCmd := exec.Command("go", "build", "./...")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Errorf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

func TestBatchMove_UnexportedSymbolError(t *testing.T) {
	dissect := buildDissectBinary(t)
	tmpDir := createTempModuleForBatch(t)

	// Create util file with unexported function
	createFileForBatch(t, filepath.Join(tmpDir, "util.go"), `package main

func helper() string {
	return "help"
}

func PublicFunc() string {
	return helper()
}
`)

	// Create main file that calls the unexported helper function
	createFileForBatch(t, filepath.Join(tmpDir, "main.go"), `package main

import "fmt"

func main() {
	result := helper()
	fmt.Println(result)
}
`)

	// Attempt to move util.go to a different package
	// This should fail because helper() is unexported but referenced from main.go
	cmd := exec.Command(dissect, "move", "--batch", "util.go -> internal/util/")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()

	// Should fail with an error about unexported symbols
	if err == nil {
		t.Fatalf("Expected dissect move to fail with unexported symbol error, but it succeeded.\nOutput: %s", string(output))
	}

	outputStr := string(output)

	// Verify error message contains the expected parts
	expectedParts := []string{
		"Error: Cannot move files - unexported symbols are referenced from other packages",
		"helper",
		"util.go",
		"main.go",
		"dissect move util.go:helper util.go:Helper",
	}

	for _, part := range expectedParts {
		if !strings.Contains(outputStr, part) {
			t.Errorf("Error message missing expected part: %q\nFull output: %s", part, outputStr)
		}
	}
}

func TestBatchMove_UnexportedFieldError(t *testing.T) {
	t.Skip("Field-level unexported symbol detection not yet implemented")
	dissect := buildDissectBinary(t)
	tmpDir := createTempModuleForBatch(t)

	// Create db file with unexported field
	createFileForBatch(t, filepath.Join(tmpDir, "db.go"), `package main

type DB struct {
	data string
}

func NewDB() *DB {
	return &DB{data: "test"}
}
`)

	// Create cmd file that accesses the unexported field
	createFileForBatch(t, filepath.Join(tmpDir, "cmd.go"), `package main

func PrintDBData(db *DB) {
	println(db.data)
}
`)

	// Attempt to move db.go to a different package
	// This should fail because db.data is unexported but referenced from cmd.go
	cmd := exec.Command(dissect, "move", "--batch", "db.go -> internal/db/")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()

	// Should fail with an error about unexported symbols
	if err == nil {
		t.Fatalf("Expected dissect move to fail with unexported field error, but it succeeded.\nOutput: %s", string(output))
	}

	outputStr := string(output)

	// Verify error message contains the expected parts
	expectedParts := []string{
		"Error: Cannot move files - unexported symbols are referenced from other packages",
		"data",
		"db.go",
		"cmd.go",
		"dissect move db.go:DB.data db.go:DB.Data",
	}

	for _, part := range expectedParts {
		if !strings.Contains(outputStr, part) {
			t.Errorf("Error message missing expected part: %q\nFull output: %s", part, outputStr)
		}
	}
}

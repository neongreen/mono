package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMoveVarWithInternalComments tests that moving a var declaration
// preserves ALL comments, including those inside struct literals.
// This is particularly important for cobra.Command definitions which often
// have comments between fields.
func TestMoveVarWithInternalComments(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dissect_var_comments_test_")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test Go module
	goMod := `module example.com/varcommentstest

go 1.24
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a source file with a var that has internal comments
	// This simulates a typical cobra.Command definition
	sourceCode := `package main

import "fmt"

// This is the documentation for myVar
var myVar = struct {
	Name        string
	Description string
	Handler     func()
}{
	Name: "example",
	// This is an internal comment about the Description field
	Description: "A test variable with internal comments",
	// This comment explains the Handler
	Handler: func() {
		// This comment is inside the function
		fmt.Println("handling")
	},
}

var anotherVar = "simple value"

func main() {
	fmt.Println(myVar.Name)
}
`
	sourceFile := filepath.Join(tmpDir, "source.go")
	if err := os.WriteFile(sourceFile, []byte(sourceCode), 0o644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Build the dissect binary
	dissectBinary := filepath.Join(tmpDir, "dissect")
	buildCmd := exec.Command("go", "build", "-o", dissectBinary, "./dissect")
	buildCmd.Dir = findRepoRoot(t)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build dissect: %v\nOutput: %s", err, output)
	}

	// Move myVar to a new file
	targetFile := filepath.Join(tmpDir, "target.go")
	cmd := exec.Command(dissectBinary, "move", "source.go:myVar", "target.go")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to move var: %v\nOutput: %s", err, output)
	}

	// Check that target file was created
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	targetStr := string(targetContent)

	// Check that the var is in the target file
	if !strings.Contains(targetStr, "var myVar") {
		t.Errorf("myVar should be in target file")
	}

	// Check that the Doc comment is preserved
	expectedDocComment := "This is the documentation for myVar"
	if !strings.Contains(targetStr, expectedDocComment) {
		t.Errorf("Doc comment should be preserved in target file. Expected to find: %q\nGot:\n%s", expectedDocComment, targetStr)
	}

	// Check that internal comments are preserved in target file
	internalComment1 := "This is an internal comment about the Description field"
	if !strings.Contains(targetStr, internalComment1) {
		t.Errorf("Internal comment about Description field should be preserved. Expected to find: %q\nGot:\n%s", internalComment1, targetStr)
	}

	internalComment2 := "This comment explains the Handler"
	if !strings.Contains(targetStr, internalComment2) {
		t.Errorf("Internal comment about Handler should be preserved. Expected to find: %q\nGot:\n%s", internalComment2, targetStr)
	}

	internalComment3 := "This comment is inside the function"
	if !strings.Contains(targetStr, internalComment3) {
		t.Errorf("Comment inside function should be preserved. Expected to find: %q\nGot:\n%s", internalComment3, targetStr)
	}

	// Check that source file no longer has the moved var or its comments
	sourceContent, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}
	sourceStr := string(sourceContent)

	if strings.Contains(sourceStr, "var myVar") {
		t.Errorf("myVar should have been removed from source")
	}

	// Check that comments for myVar are not orphaned in source
	if strings.Contains(sourceStr, expectedDocComment) {
		t.Errorf("Doc comment for myVar should not be orphaned in source file. Found: %q\nSource:\n%s", expectedDocComment, sourceStr)
	}

	if strings.Contains(sourceStr, internalComment1) {
		t.Errorf("Internal comment should not be orphaned in source file. Found: %q\nSource:\n%s", internalComment1, sourceStr)
	}

	if strings.Contains(sourceStr, internalComment2) {
		t.Errorf("Internal comment should not be orphaned in source file. Found: %q\nSource:\n%s", internalComment2, sourceStr)
	}

	// Check that anotherVar still exists in source
	if !strings.Contains(sourceStr, "var anotherVar") {
		t.Errorf("anotherVar should still be in source file")
	}

	// Verify the code still builds
	buildCmd = exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

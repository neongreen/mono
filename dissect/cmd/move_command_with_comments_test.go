package main_test

import (
	"dissect/cmd/internal/testutils"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMoveCommandWithComments(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dissect_move_comments_test_")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test Go module
	goMod := `module example.com/commentstest

go 1.24
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a source file with commented functions
	sourceCode := `package main

import "fmt"

// This is a function with a comment above it
// It has multiple lines of comments
func commentedFunction() {
	fmt.Println("This function has comments")
}

// Another function with a single line comment
func anotherFunction() {
	fmt.Println("Another function")
}

func main() {
	commentedFunction()
	anotherFunction()
}
`
	sourceFile := filepath.Join(tmpDir, "source.go")
	if err := os.WriteFile(sourceFile, []byte(sourceCode), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Build the dissect binary
	dissectBinary := filepath.Join(tmpDir, "dissect")
	buildCmd := exec.Command("go", "build", "-o", dissectBinary, "./cmd")
	buildCmd.Dir = findRepoRoot(t)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build dissect: %v\nOutput: %s", err, output)
	}

	// Move commentedFunction to a new file
	targetFile := filepath.Join(tmpDir, "target.go")
	cmd := exec.Command(dissectBinary, "move", "source.go:commentedFunction", "target.go")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to move function: %v\nOutput: %s", err, output)
	}

	// Check that target file was created and contains comments
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	targetStr := string(targetContent)

	// Check that the function is in the target file
	if !testutils.ContainsFunc(targetStr, "commentedFunction") {
		t.Errorf("commentedFunction should be in target file")
	}

	// Check that comments are preserved in target file
	expectedComment := "This is a function with a comment above it"
	if !testutils.ContainsString(targetStr, expectedComment) {
		t.Errorf("Comments should be preserved in target file. Expected to find: %q\nGot:\n%s", expectedComment, targetStr)
	}

	multiLineComment := "It has multiple lines of comments"
	if !testutils.ContainsString(targetStr, multiLineComment) {
		t.Errorf("Multi-line comments should be preserved. Expected to find: %q\nGot:\n%s", multiLineComment, targetStr)
	}

	// Check that source file no longer has the moved function or its comments
	sourceContent, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}
	sourceStr := string(sourceContent)

	if testutils.ContainsFunc(sourceStr, "commentedFunction") {
		t.Errorf("commentedFunction should have been removed from source")
	}

	// Check that comments for commentedFunction are not orphaned in source
	if testutils.ContainsString(sourceStr, expectedComment) {
		t.Errorf("Comments for commentedFunction should not be orphaned in source file. Found: %q\nSource:\n%s", expectedComment, sourceStr)
	}

	// Check that anotherFunction and its comment still exist in source
	if !testutils.ContainsFunc(sourceStr, "anotherFunction") {
		t.Errorf("anotherFunction should still be in source file")
	}

	anotherComment := "Another function with a single line comment"
	if !testutils.ContainsString(sourceStr, anotherComment) {
		t.Errorf("Comments for anotherFunction should still be in source file")
	}

	// Verify the code still builds
	buildCmd = exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

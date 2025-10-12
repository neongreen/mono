package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMoveCommand(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dissect_move_test_")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test Go module
	goMod := `module example.com/movetest

go 1.24
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a source file with multiple functions
	sourceCode := `package main

import "fmt"

func main() {
	Foo()
	Bar()
	Baz()
}

func Foo() {
	fmt.Println("This is Foo")
}

func Bar() {
	fmt.Println("This is Bar")
}

func Baz() {
	fmt.Println("This is Baz")
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

	t.Run("MoveToNewFile", func(t *testing.T) {
		// Move Foo to a new file
		targetFile := filepath.Join(tmpDir, "target.go")
		cmd := exec.Command(dissectBinary, "move", "source.go:Foo", "target.go")
		cmd.Dir = tmpDir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to move function: %v\nOutput: %s", err, output)
		}

		// Check that target file was created
		if _, err := os.Stat(targetFile); os.IsNotExist(err) {
			t.Fatalf("Target file was not created")
		}

		// Check that source file still exists and Foo was removed
		sourceContent, err := os.ReadFile(sourceFile)
		if err != nil {
			t.Fatalf("Failed to read source file: %v", err)
		}
		sourceStr := string(sourceContent)
		if containsFunc(sourceStr, "Foo") {
			t.Errorf("Foo function should have been removed from source")
		}

		// Check that target file contains Foo
		targetContent, err := os.ReadFile(targetFile)
		if err != nil {
			t.Fatalf("Failed to read target file: %v", err)
		}
		targetStr := string(targetContent)
		if !containsFunc(targetStr, "Foo") {
			t.Errorf("Foo function should be in target file")
		}

		// Verify the code still builds
		buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
		buildCmd.Dir = tmpDir
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
		}
	})

	t.Run("MoveMultipleToExistingFile", func(t *testing.T) {
		// Move Bar and Baz to the existing target file
		targetFile := filepath.Join(tmpDir, "target.go")
		cmd := exec.Command(dissectBinary, "move", "source.go:Bar,Baz", "target.go")
		cmd.Dir = tmpDir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to move functions: %v\nOutput: %s", err, output)
		}

		// Check that target file contains all three functions
		targetContent, err := os.ReadFile(targetFile)
		if err != nil {
			t.Fatalf("Failed to read target file: %v", err)
		}
		targetStr := string(targetContent)
		if !containsFunc(targetStr, "Foo") {
			t.Errorf("Foo function should still be in target file")
		}
		if !containsFunc(targetStr, "Bar") {
			t.Errorf("Bar function should be in target file")
		}
		if !containsFunc(targetStr, "Baz") {
			t.Errorf("Baz function should be in target file")
		}

		// Check that source file only has main
		sourceContent, err := os.ReadFile(sourceFile)
		if err != nil {
			t.Fatalf("Failed to read source file: %v", err)
		}
		sourceStr := string(sourceContent)
		if containsFunc(sourceStr, "Bar") {
			t.Errorf("Bar function should have been removed from source")
		}
		if containsFunc(sourceStr, "Baz") {
			t.Errorf("Baz function should have been removed from source")
		}

		// Verify the code still builds
		buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
		buildCmd.Dir = tmpDir
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
		}
	})
}

// Helper function to check if a function exists in code
func containsFunc(code string, funcName string) bool {
	funcDecl := "func " + funcName + "("
	return containsString(code, funcDecl)
}

func containsString(s string, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s string, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

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
	if !containsFunc(targetStr, "commentedFunction") {
		t.Errorf("commentedFunction should be in target file")
	}

	// Check that comments are preserved in target file
	expectedComment := "This is a function with a comment above it"
	if !containsString(targetStr, expectedComment) {
		t.Errorf("Comments should be preserved in target file. Expected to find: %q\nGot:\n%s", expectedComment, targetStr)
	}

	multiLineComment := "It has multiple lines of comments"
	if !containsString(targetStr, multiLineComment) {
		t.Errorf("Multi-line comments should be preserved. Expected to find: %q\nGot:\n%s", multiLineComment, targetStr)
	}

	// Check that source file no longer has the moved function or its comments
	sourceContent, err := os.ReadFile(sourceFile)
	if err != nil {
		t.Fatalf("Failed to read source file: %v", err)
	}
	sourceStr := string(sourceContent)

	if containsFunc(sourceStr, "commentedFunction") {
		t.Errorf("commentedFunction should have been removed from source")
	}

	// Check that comments for commentedFunction are not orphaned in source
	if containsString(sourceStr, expectedComment) {
		t.Errorf("Comments for commentedFunction should not be orphaned in source file. Found: %q\nSource:\n%s", expectedComment, sourceStr)
	}

	// Check that anotherFunction and its comment still exist in source
	if !containsFunc(sourceStr, "anotherFunction") {
		t.Errorf("anotherFunction should still be in source file")
	}

	anotherComment := "Another function with a single line comment"
	if !containsString(sourceStr, anotherComment) {
		t.Errorf("Comments for anotherFunction should still be in source file")
	}

	// Verify the code still builds
	buildCmd = exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
	}
}

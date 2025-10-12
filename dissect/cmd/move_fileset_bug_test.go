package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMoveToFileWithExistingContent tests moving a function to a file that already has content.
// This is the critical test case that exposes the FileSet bug: when targetNode has existing
// declarations with position info from targetFset, but we write using tempFset, it can cause
// panics or corrupt formatting.
func TestMoveToFileWithExistingContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_fileset_bug_test_")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	goMod := `module example.com/filesettest

go 1.24
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Source file with a function to move
	sourceCode := `package main

import "fmt"

// sourceFunc is in the source file
func sourceFunc() {
	fmt.Println("source")
}

func main() {
	sourceFunc()
	targetFunc()
}
`
	sourceFile := filepath.Join(tmpDir, "source.go")
	if err := os.WriteFile(sourceFile, []byte(sourceCode), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Target file that already has content - this is crucial!
	// The existing content has position info from the target FileSet
	targetCode := `package main

import "fmt"

// targetFunc already exists in target
func targetFunc() {
	fmt.Println("target")
}

// anotherFunc also exists
func anotherFunc() {
	fmt.Println("another")
}
`
	targetFile := filepath.Join(tmpDir, "target.go")
	if err := os.WriteFile(targetFile, []byte(targetCode), 0644); err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	dissectBinary := filepath.Join(tmpDir, "dissect")
	buildCmd := exec.Command("go", "build", "-o", dissectBinary, "./cmd")
	buildCmd.Dir = findRepoRoot(t)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build dissect: %v\nOutput: %s", err, output)
	}

	// Move sourceFunc to the target file that already has content
	cmd := exec.Command(dissectBinary, "move", "source.go:sourceFunc", "target.go")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to move function: %v\nOutput: %s", err, output)
	}

	// Read the target file
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	targetStr := string(targetContent)

	// Verify sourceFunc was added
	if !containsFunc(targetStr, "sourceFunc") {
		t.Errorf("sourceFunc should be in target file after move")
	}

	// Verify sourceFunc's comment was preserved
	if !containsString(targetStr, "sourceFunc is in the source file") {
		t.Errorf("sourceFunc's comment should be preserved in target file")
	}

	// Verify existing functions are still there
	if !containsFunc(targetStr, "targetFunc") {
		t.Errorf("targetFunc should still be in target file")
	}
	
	if !containsFunc(targetStr, "anotherFunc") {
		t.Errorf("anotherFunc should still be in target file")
	}

	// Verify existing function comments are preserved
	if !containsString(targetStr, "targetFunc already exists in target") {
		t.Errorf("targetFunc's comment should be preserved")
	}
	
	if !containsString(targetStr, "anotherFunc also exists") {
		t.Errorf("anotherFunc's comment should be preserved")
	}

	// Most importantly, verify the code builds
	buildCmd = exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s\nTarget file:\n%s", err, output, targetStr)
	}
}

// TestMoveMultipleFunctionsToExistingFile tests moving multiple functions sequentially
// to a file with existing content. This stresses the FileSet handling.
func TestMoveMultipleFunctionsToExistingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_multiple_move_test_")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	goMod := `module example.com/multiplemove

go 1.24
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Source file with multiple functions
	sourceCode := `package main

import "fmt"

// firstFunc is the first function
func firstFunc() {
	fmt.Println("first")
}

// secondFunc is the second function
func secondFunc() {
	fmt.Println("second")
}

// thirdFunc is the third function
func thirdFunc() {
	fmt.Println("third")
}

func main() {
	firstFunc()
	secondFunc()
	thirdFunc()
	existingFunc()
}
`
	sourceFile := filepath.Join(tmpDir, "source.go")
	if err := os.WriteFile(sourceFile, []byte(sourceCode), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Target file with existing content
	targetCode := `package main

import "fmt"

// existingFunc was already in target
func existingFunc() {
	fmt.Println("existing")
}
`
	targetFile := filepath.Join(tmpDir, "target.go")
	if err := os.WriteFile(targetFile, []byte(targetCode), 0644); err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	dissectBinary := filepath.Join(tmpDir, "dissect")
	buildCmd := exec.Command("go", "build", "-o", dissectBinary, "./cmd")
	buildCmd.Dir = findRepoRoot(t)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build dissect: %v\nOutput: %s", err, output)
	}

	// Move first function
	cmd := exec.Command(dissectBinary, "move", "source.go:firstFunc", "target.go")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to move firstFunc: %v\nOutput: %s", err, output)
	}

	// Move second function
	cmd = exec.Command(dissectBinary, "move", "source.go:secondFunc", "target.go")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to move secondFunc: %v\nOutput: %s", err, output)
	}

	// Move third function
	cmd = exec.Command(dissectBinary, "move", "source.go:thirdFunc", "target.go")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to move thirdFunc: %v\nOutput: %s", err, output)
	}

	// Read and verify target file
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	targetStr := string(targetContent)

	// Verify all functions are present
	functions := []string{"existingFunc", "firstFunc", "secondFunc", "thirdFunc"}
	for _, fn := range functions {
		if !containsFunc(targetStr, fn) {
			t.Errorf("%s should be in target file", fn)
		}
	}

	// Verify all comments are present
	comments := []string{
		"existingFunc was already in target",
		"firstFunc is the first function",
		"secondFunc is the second function",
		"thirdFunc is the third function",
	}
	for _, comment := range comments {
		if !containsString(targetStr, comment) {
			t.Errorf("Comment %q should be in target file", comment)
		}
	}

	// Verify code builds
	buildCmd = exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after moves: %v\nOutput: %s\nTarget file:\n%s", err, output, targetStr)
	}
}

// TestMoveWithComplexExistingImports tests that existing imports in the target file
// are preserved correctly even with different FileSets
func TestMoveWithComplexExistingImports(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_imports_test_")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	goMod := `module example.com/importstest

go 1.24
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Source file with a function that uses fmt
	sourceCode := `package main

import "fmt"

// printMessage prints a message
func printMessage(msg string) {
	fmt.Println(msg)
}

func main() {
	printMessage("hello")
	processData()
}
`
	sourceFile := filepath.Join(tmpDir, "source.go")
	if err := os.WriteFile(sourceFile, []byte(sourceCode), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Target file with different imports
	targetCode := `package main

import (
	"encoding/json"
	"io"
	"os"
)

// processData processes data using the imports above
func processData() {
	var data map[string]interface{}
	json.Unmarshal([]byte("{}"), &data)
	io.Copy(os.Stdout, os.Stdin)
}
`
	targetFile := filepath.Join(tmpDir, "target.go")
	if err := os.WriteFile(targetFile, []byte(targetCode), 0644); err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}

	dissectBinary := filepath.Join(tmpDir, "dissect")
	buildCmd := exec.Command("go", "build", "-o", dissectBinary, "./cmd")
	buildCmd.Dir = findRepoRoot(t)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build dissect: %v\nOutput: %s", err, output)
	}

	// Move function
	cmd := exec.Command(dissectBinary, "move", "source.go:printMessage", "target.go")
	cmd.Dir = tmpDir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to move function: %v\nOutput: %s", err, output)
	}

	// Read target file
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	targetStr := string(targetContent)

	// Verify function was moved
	if !containsFunc(targetStr, "printMessage") {
		t.Errorf("printMessage should be in target file")
	}

	// Verify comment was preserved
	if !containsString(targetStr, "printMessage prints a message") {
		t.Errorf("printMessage's comment should be preserved")
	}

	// Verify existing function is still there
	if !containsFunc(targetStr, "processData") {
		t.Errorf("processData should still be in target file")
	}

	// Verify all necessary imports are present
	requiredImports := []string{"fmt", "encoding/json", "io", "os"}
	for _, imp := range requiredImports {
		if !containsString(targetStr, `"`+imp+`"`) {
			t.Errorf("Import %q should be in target file", imp)
		}
	}

	// Verify code builds
	buildCmd = exec.Command("go", "build", "-o", "/dev/null", ".")
	buildCmd.Dir = tmpDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Code doesn't build after move: %v\nOutput: %s\nTarget file:\n%s", err, output, targetStr)
	}
}

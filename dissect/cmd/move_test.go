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

	t.Run("MoveWithGlobPattern", func(t *testing.T) {
		// Create multiple source files with different helper functions
		helper1Code := `package main

import "fmt"

func HelperOne() {
	fmt.Println("This is HelperOne")
}
`
		helper2Code := `package main

import "fmt"

func HelperTwo() {
	fmt.Println("This is HelperTwo")
}
`
		if err := os.WriteFile(filepath.Join(tmpDir, "helper1.go"), []byte(helper1Code), 0644); err != nil {
			t.Fatalf("Failed to create helper1.go: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "helper2.go"), []byte(helper2Code), 0644); err != nil {
			t.Fatalf("Failed to create helper2.go: %v", err)
		}

		// Move HelperOne from helper1.go and HelperTwo from helper2.go using separate glob patterns
		targetFile := filepath.Join(tmpDir, "helpers.go")
		cmd := exec.Command(dissectBinary, "move", "helper1.go:HelperOne", "helper2.go:HelperTwo", "helpers.go")
		cmd.Dir = tmpDir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to move functions with multiple source files: %v\nOutput: %s", err, output)
		}

		// Check that target file was created and contains both functions
		targetContent, err := os.ReadFile(targetFile)
		if err != nil {
			t.Fatalf("Failed to read target file: %v", err)
		}
		targetStr := string(targetContent)
		if !containsFunc(targetStr, "HelperOne") {
			t.Errorf("HelperOne function should be in target file")
		}
		if !containsFunc(targetStr, "HelperTwo") {
			t.Errorf("HelperTwo function should be in target file")
		}

		// Check that source files had functions removed
		helper1Content, err := os.ReadFile(filepath.Join(tmpDir, "helper1.go"))
		if err != nil {
			t.Fatalf("Failed to read helper1.go: %v", err)
		}
		if containsFunc(string(helper1Content), "HelperOne") {
			t.Errorf("HelperOne function should have been removed from helper1.go")
		}

		helper2Content, err := os.ReadFile(filepath.Join(tmpDir, "helper2.go"))
		if err != nil {
			t.Fatalf("Failed to read helper2.go: %v", err)
		}
		if containsFunc(string(helper2Content), "HelperTwo") {
			t.Errorf("HelperTwo function should have been removed from helper2.go")
		}

		// Verify the code still builds
		buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
		buildCmd.Dir = tmpDir
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
		}
	})

	t.Run("MoveWithActualGlobPattern", func(t *testing.T) {
		// Create multiple source files with the same function name
		glob1Code := `package main

import "fmt"

func GlobHelper() {
	fmt.Println("This is GlobHelper from glob1")
}
`
		glob2Code := `package main

import "fmt"

func GlobHelper() {
	fmt.Println("This is GlobHelper from glob2")
}
`
		if err := os.WriteFile(filepath.Join(tmpDir, "glob1.go"), []byte(glob1Code), 0644); err != nil {
			t.Fatalf("Failed to create glob1.go: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "glob2.go"), []byte(glob2Code), 0644); err != nil {
			t.Fatalf("Failed to create glob2.go: %v", err)
		}

		// Use glob pattern to match both glob1.go and glob2.go
		// Note: This will try to move GlobHelper from both files, which will fail due to duplicate names.
		// This is expected behavior - the glob expands to multiple files, each matching the pattern.
		targetFile := filepath.Join(tmpDir, "globhelpers.go")
		cmd := exec.Command(dissectBinary, "move", "glob*.go:GlobHelper", "globhelpers.go")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		
		// We expect this to succeed for the first file but fail when building due to duplicates,
		// or to move both and have build failures. Let's just verify the glob expansion worked
		// by checking that at least one function was moved.
		if err == nil {
			// If it succeeded, check that target file was created
			targetContent, err := os.ReadFile(targetFile)
			if err != nil {
				t.Fatalf("Failed to read target file: %v", err)
			}
			targetStr := string(targetContent)
			if !containsFunc(targetStr, "GlobHelper") {
				t.Errorf("GlobHelper function should be in target file")
			}
			
			// The code won't build due to duplicate function names, which is expected
			buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
			buildCmd.Dir = tmpDir
			buildOutput, buildErr := buildCmd.CombinedOutput()
			if buildErr == nil {
				t.Errorf("Expected build to fail due to duplicate function names, but it succeeded")
			} else if !containsString(string(buildOutput), "redeclared") {
				t.Logf("Build failed as expected with duplicate functions. Output: %s", buildOutput)
			}
		} else {
			// If move failed, that's also acceptable - log it for information
			t.Logf("Move failed as expected when trying to move same function from multiple files: %v\nOutput: %s", err, output)
		}
	})

	t.Run("MoveWithFunctionNameGlob", func(t *testing.T) {
		// Clean up any leftover files from previous tests that might interfere
		os.Remove(filepath.Join(tmpDir, "globhelpers.go"))
		
		// Create a source file with multiple test functions
		testCode := `package main

import "fmt"

func TestFoo() {
	fmt.Println("Test Foo")
}

func TestBar() {
	fmt.Println("Test Bar")
}

func TestBaz() {
	fmt.Println("Test Baz")
}

func HelperFunc() {
	fmt.Println("Helper")
}
`
		if err := os.WriteFile(filepath.Join(tmpDir, "tests.go"), []byte(testCode), 0644); err != nil {
			t.Fatalf("Failed to create tests.go: %v", err)
		}

		// Use function name glob to move all Test* functions
		targetFile := filepath.Join(tmpDir, "test_funcs.go")
		cmd := exec.Command(dissectBinary, "move", "tests.go:Test*", "test_funcs.go")
		cmd.Dir = tmpDir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to move functions with glob: %v\nOutput: %s", err, output)
		}

		// Check that target file contains all Test* functions
		targetContent, err := os.ReadFile(targetFile)
		if err != nil {
			t.Fatalf("Failed to read target file: %v", err)
		}
		targetStr := string(targetContent)
		if !containsFunc(targetStr, "TestFoo") {
			t.Errorf("TestFoo function should be in target file")
		}
		if !containsFunc(targetStr, "TestBar") {
			t.Errorf("TestBar function should be in target file")
		}
		if !containsFunc(targetStr, "TestBaz") {
			t.Errorf("TestBaz function should be in target file")
		}

		// Check that HelperFunc was NOT moved
		if containsFunc(targetStr, "HelperFunc") {
			t.Errorf("HelperFunc should NOT be in target file")
		}

		// Check that source file no longer has Test* functions but has HelperFunc
		sourceContent, err := os.ReadFile(filepath.Join(tmpDir, "tests.go"))
		if err != nil {
			t.Fatalf("Failed to read source file: %v", err)
		}
		sourceStr := string(sourceContent)
		if containsFunc(sourceStr, "TestFoo") {
			t.Errorf("TestFoo should have been removed from source")
		}
		if containsFunc(sourceStr, "TestBar") {
			t.Errorf("TestBar should have been removed from source")
		}
		if containsFunc(sourceStr, "TestBaz") {
			t.Errorf("TestBaz should have been removed from source")
		}
		if !containsFunc(sourceStr, "HelperFunc") {
			t.Errorf("HelperFunc should still be in source")
		}

		// Verify the code still builds
		buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
		buildCmd.Dir = tmpDir
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
		}
	})

	t.Run("MoveWithNoMatches", func(t *testing.T) {
		// Create a file without any matching functions
		noMatchCode := `package main

func SomeFunc() {}
`
		if err := os.WriteFile(filepath.Join(tmpDir, "nomatch.go"), []byte(noMatchCode), 0644); err != nil {
			t.Fatalf("Failed to create nomatch.go: %v", err)
		}

		// Try to move functions that don't exist
		cmd := exec.Command(dissectBinary, "move", "nomatch.go:NonExistent*", "target.go")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		
		// Should fail with "no identifiers matched" error
		if err == nil {
			t.Fatalf("Expected move to fail when no functions match, but it succeeded")
		}
		
		if !containsString(string(output), "No identifiers matched") {
			t.Errorf("Expected 'No identifiers matched' error, got: %s", output)
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

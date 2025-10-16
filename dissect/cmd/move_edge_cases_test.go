package main_test

import (
	"dissect/cmd/internal/testutils"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMoveEdgeCases tests edge cases and potential issues with moving declarations
func TestMoveEdgeCases(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dissect_edge_cases_test_")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test Go module
	goMod := `module example.com/edgecases

go 1.24
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Build the dissect binary
	dissectBinary := filepath.Join(tmpDir, "dissect")
	buildCmd := exec.Command("go", "build", "-o", dissectBinary, "./cmd")
	buildCmd.Dir = findRepoRoot(t)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build dissect: %v\nOutput: %s", err, output)
	}

	t.Run("GroupedConstantsBehavior", func(t *testing.T) {
		// When moving one const from a grouped const block, the entire block moves
		sourceCode := `package main

import "fmt"

const (
	Red = "red"
	Blue = "blue"
	Green = "green"
)

func main() {
	fmt.Println(Red, Blue, Green)
}
`
		sourceFile := filepath.Join(tmpDir, "grouped_consts.go")
		if err := os.WriteFile(sourceFile, []byte(sourceCode), 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Move just "Blue" - should move entire const block
		targetFile := filepath.Join(tmpDir, "colors.go")
		cmd := exec.Command(dissectBinary, "move", "grouped_consts.go:Blue", "colors.go")
		cmd.Dir = tmpDir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to move const: %v\nOutput: %s", err, output)
		}

		// Check that entire const block was moved
		targetContent, err := os.ReadFile(targetFile)
		if err != nil {
			t.Fatalf("Failed to read target file: %v", err)
		}
		targetStr := string(targetContent)

		// All three constants should be in the target
		if !testutils.ContainsString(targetStr, "Red") {
			t.Errorf("Red should be in target file (grouped declarations move together)")
		}
		if !testutils.ContainsString(targetStr, "Blue") {
			t.Errorf("Blue should be in target file")
		}
		if !testutils.ContainsString(targetStr, "Green") {
			t.Errorf("Green should be in target file (grouped declarations move together)")
		}

		// Verify the code still builds
		buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
		buildCmd.Dir = tmpDir
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
		}
	})

	t.Run("TypeWithMethodsSeparation", func(t *testing.T) {
		// Moving a type doesn't move its methods - they stay in source
		// This is valid Go - methods can be in different files

		// Create a separate subdirectory for this test
		testDir := filepath.Join(tmpDir, "methods_test")
		if err := os.MkdirAll(testDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(testDir, "go.mod"), []byte(goMod), 0644); err != nil {
			t.Fatalf("Failed to create go.mod: %v", err)
		}

		sourceCode := `package main

import "fmt"

type User struct {
	Name string
}

func (u User) SayHello() {
	fmt.Printf("Hello, %s\n", u.Name)
}

func main() {
	u := User{Name: "Alice"}
	u.SayHello()
}
`
		sourceFile := filepath.Join(testDir, "user_source.go")
		if err := os.WriteFile(sourceFile, []byte(sourceCode), 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Move just the type
		targetFile := filepath.Join(testDir, "user_types.go")
		cmd := exec.Command(dissectBinary, "move", "user_source.go:User", "user_types.go")
		cmd.Dir = testDir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to move type: %v\nOutput: %s", err, output)
		}

		// Check that type is in target
		targetContent, err := os.ReadFile(targetFile)
		if err != nil {
			t.Fatalf("Failed to read target file: %v", err)
		}
		targetStr := string(targetContent)
		if !testutils.ContainsString(targetStr, "type User struct") {
			t.Errorf("User type should be in target file")
		}

		// Check that method stayed in source
		sourceContent, err := os.ReadFile(sourceFile)
		if err != nil {
			t.Fatalf("Failed to read source file: %v", err)
		}
		sourceStr := string(sourceContent)
		if !testutils.ContainsString(sourceStr, "func (u User) SayHello()") {
			t.Errorf("Method should still be in source file")
		}

		// Verify the code still builds (methods can be in different files)
		buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
		buildCmd.Dir = testDir
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
		}
	})

	t.Run("ImportManagementWithTypes", func(t *testing.T) {
		// Test that imports are properly managed when moving types

		// Create a separate subdirectory for this test
		testDir := filepath.Join(tmpDir, "import_test")
		if err := os.MkdirAll(testDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(testDir, "go.mod"), []byte(goMod), 0644); err != nil {
			t.Fatalf("Failed to create go.mod: %v", err)
		}

		sourceCode := `package main

import (
	"fmt"
	"time"
)

type Config struct {
	Timeout time.Duration
}

func main() {
	fmt.Println("test")
}
`
		sourceFile := filepath.Join(testDir, "source.go")
		if err := os.WriteFile(sourceFile, []byte(sourceCode), 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Move Config type
		targetFile := filepath.Join(testDir, "config.go")
		cmd := exec.Command(dissectBinary, "move", "source.go:Config", "config.go")
		cmd.Dir = testDir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to move type: %v\nOutput: %s", err, output)
		}

		// Check that time import is in target file
		targetContent, err := os.ReadFile(targetFile)
		if err != nil {
			t.Fatalf("Failed to read target file: %v", err)
		}
		targetStr := string(targetContent)
		if !testutils.ContainsString(targetStr, "import \"time\"") && !testutils.ContainsString(targetStr, "\"time\"") {
			t.Errorf("time import should be in target file")
		}

		// Check that time import is removed from source file (not used anymore)
		sourceContent, err := os.ReadFile(sourceFile)
		if err != nil {
			t.Fatalf("Failed to read source file: %v", err)
		}
		sourceStr := string(sourceContent)
		if testutils.ContainsString(sourceStr, "\"time\"") {
			t.Errorf("time import should be removed from source file (unused after move)")
		}

		// Verify main function is still in source
		if !testutils.ContainsString(sourceStr, "func main()") {
			t.Fatalf("main function should still be in source file. Source content:\n%s", sourceStr)
		}

		// Verify the code still builds
		buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
		buildCmd.Dir = testDir
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Code doesn't build after move: %v\nOutput: %s\nSource:\n%s\nTarget:\n%s", err, output, sourceStr, targetStr)
		}
	})

	t.Run("DotImportHandling", func(t *testing.T) {
		// Dot imports should be handled without errors (goimports won't add them to target)

		// Create a separate subdirectory for this test
		testDir := filepath.Join(tmpDir, "dot_import_test")
		if err := os.MkdirAll(testDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(testDir, "go.mod"), []byte(goMod), 0644); err != nil {
			t.Fatalf("Failed to create go.mod: %v", err)
		}

		sourceCode := `package main

import . "fmt"

type MyType struct {
	Name string
}

func main() {
	Println("test")
}
`
		sourceFile := filepath.Join(testDir, "dot_import.go")
		if err := os.WriteFile(sourceFile, []byte(sourceCode), 0644); err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// Move type
		targetFile := filepath.Join(testDir, "dot_types.go")
		cmd := exec.Command(dissectBinary, "move", "dot_import.go:MyType", "dot_types.go")
		cmd.Dir = testDir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to move type with dot import: %v\nOutput: %s", err, output)
		}

		// Check that type is in target (without dot import since it's not used)
		targetContent, err := os.ReadFile(targetFile)
		if err != nil {
			t.Fatalf("Failed to read target file: %v", err)
		}
		targetStr := string(targetContent)
		if !testutils.ContainsString(targetStr, "type MyType struct") {
			t.Errorf("MyType should be in target file")
		}

		// Verify the code still builds
		buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
		buildCmd.Dir = testDir
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
		}
	})
}

package main_test

import (
	"dissect/cmd/internal/testutils"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMoveTypes(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dissect_move_types_test_")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test Go module
	goMod := `module example.com/typestest

go 1.24
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a source file with types, interfaces, and functions
	sourceCode := `package main

import "fmt"

// MyType is a custom type
type MyType struct {
	Name string
	Age  int
}

// MyInt is a type alias
type MyInt int

// MyInterface defines behavior
type MyInterface interface {
	DoSomething() string
	GetName() string
}

func (m MyType) DoSomething() string {
	return "doing something"
}

func (m MyType) GetName() string {
	return m.Name
}

func UseTypes() {
	var t MyType
	var i MyInt
	var intf MyInterface
	fmt.Println(t, i, intf)
}

func main() {
	UseTypes()
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

	t.Run("MoveTypeDeclaration", func(t *testing.T) {
		// Move MyType to a new file
		targetFile := filepath.Join(tmpDir, "types.go")
		cmd := exec.Command(dissectBinary, "move", "source.go:MyType", "types.go")
		cmd.Dir = tmpDir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to move type: %v\nOutput: %s", err, output)
		}

		// Check that target file was created
		if _, err := os.Stat(targetFile); os.IsNotExist(err) {
			t.Fatalf("Target file was not created")
		}

		// Check that source file no longer has MyType
		sourceContent, err := os.ReadFile(sourceFile)
		if err != nil {
			t.Fatalf("Failed to read source file: %v", err)
		}
		sourceStr := string(sourceContent)
		if testutils.ContainsString(sourceStr, "type MyType struct") {
			t.Errorf("MyType should have been removed from source")
		}

		// Check that target file contains MyType
		targetContent, err := os.ReadFile(targetFile)
		if err != nil {
			t.Fatalf("Failed to read target file: %v", err)
		}
		targetStr := string(targetContent)
		if !testutils.ContainsString(targetStr, "type MyType struct") {
			t.Errorf("MyType should be in target file")
		}

		// Check that comments are preserved
		if !testutils.ContainsString(targetStr, "MyType is a custom type") {
			t.Errorf("Comments should be preserved in target file")
		}

		// Verify the code still builds
		buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
		buildCmd.Dir = tmpDir
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
		}
	})

	t.Run("MoveInterfaceDeclaration", func(t *testing.T) {
		// Move MyInterface to the types file
		targetFile := filepath.Join(tmpDir, "types.go")
		cmd := exec.Command(dissectBinary, "move", "source.go:MyInterface", "types.go")
		cmd.Dir = tmpDir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to move interface: %v\nOutput: %s", err, output)
		}

		// Check that source file no longer has MyInterface
		sourceContent, err := os.ReadFile(sourceFile)
		if err != nil {
			t.Fatalf("Failed to read source file: %v", err)
		}
		sourceStr := string(sourceContent)
		if testutils.ContainsString(sourceStr, "type MyInterface interface") {
			t.Errorf("MyInterface should have been removed from source")
		}

		// Check that target file contains MyInterface
		targetContent, err := os.ReadFile(targetFile)
		if err != nil {
			t.Fatalf("Failed to read target file: %v", err)
		}
		targetStr := string(targetContent)
		if !testutils.ContainsString(targetStr, "type MyInterface interface") {
			t.Errorf("MyInterface should be in target file")
		}

		// Check that comments are preserved
		if !testutils.ContainsString(targetStr, "MyInterface defines behavior") {
			t.Errorf("Comments should be preserved in target file")
		}

		// Verify the code still builds
		buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
		buildCmd.Dir = tmpDir
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
		}
	})

	t.Run("MoveMultipleTypesAndInterface", func(t *testing.T) {
		// Move MyInt to types.go
		targetFile := filepath.Join(tmpDir, "types.go")
		cmd := exec.Command(dissectBinary, "move", "source.go:MyInt", "types.go")
		cmd.Dir = tmpDir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to move type alias: %v\nOutput: %s", err, output)
		}

		// Check that target file contains all moved types
		targetContent, err := os.ReadFile(targetFile)
		if err != nil {
			t.Fatalf("Failed to read target file: %v", err)
		}
		targetStr := string(targetContent)

		if !testutils.ContainsString(targetStr, "type MyType struct") {
			t.Errorf("MyType should still be in target file")
		}
		if !testutils.ContainsString(targetStr, "type MyInterface interface") {
			t.Errorf("MyInterface should still be in target file")
		}
		if !testutils.ContainsString(targetStr, "type MyInt int") {
			t.Errorf("MyInt should be in target file")
		}

		// Verify the code still builds
		buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
		buildCmd.Dir = tmpDir
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
		}
	})

	t.Run("MoveFunctionAfterTypes", func(t *testing.T) {
		// Move UseTypes function to verify functions still work
		cmd := exec.Command(dissectBinary, "move", "source.go:UseTypes", "use_types.go")
		cmd.Dir = tmpDir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("Failed to move function: %v\nOutput: %s", err, output)
		}

		// Verify the code still builds
		buildCmd := exec.Command("go", "build", "-o", "/dev/null", ".")
		buildCmd.Dir = tmpDir
		if output, err := buildCmd.CombinedOutput(); err != nil {
			t.Fatalf("Code doesn't build after move: %v\nOutput: %s", err, output)
		}
	})
}

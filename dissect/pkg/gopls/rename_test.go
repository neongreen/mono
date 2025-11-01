package gopls

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenameFunction(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.go")
	content := `package main

func oldFunc() {
	println("hello")
}

func main() {
	oldFunc()
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Rename the function
	err := Rename("gopls", testFile, "oldFunc", "newFunc", tmpDir)
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	// Read the updated file
	updated, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	updatedStr := string(updated)

	// Verify the function was renamed
	if strings.Contains(updatedStr, "oldFunc") {
		t.Errorf("Old function name still exists in file")
	}
	if !strings.Contains(updatedStr, "func newFunc()") {
		t.Errorf("New function definition not found")
	}
	if !strings.Contains(updatedStr, "newFunc()") {
		t.Errorf("Function call was not updated")
	}
}

func TestRenameType(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

type OldType struct {
	Value int
}

func useType() OldType {
	return OldType{Value: 42}
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := Rename("gopls", testFile, "OldType", "NewType", tmpDir)
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	updated, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	updatedStr := string(updated)

	if strings.Contains(updatedStr, "OldType") {
		t.Errorf("Old type name still exists")
	}
	if !strings.Contains(updatedStr, "type NewType struct") {
		t.Errorf("New type definition not found")
	}
	if !strings.Contains(updatedStr, "NewType{") {
		t.Errorf("Type usage was not updated")
	}
}

func TestRenameVariable(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

var oldVar = 42

func main() {
	println(oldVar)
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := Rename("gopls", testFile, "oldVar", "newVar", tmpDir)
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	updated, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	updatedStr := string(updated)

	if strings.Contains(updatedStr, "oldVar") {
		t.Errorf("Old variable name still exists")
	}
	if !strings.Contains(updatedStr, "var newVar") {
		t.Errorf("New variable definition not found")
	}
}

func TestRenameMethod(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

type MyStruct struct{}

func (m *MyStruct) oldMethod() {
	println("method")
}

func main() {
	s := &MyStruct{}
	s.oldMethod()
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := Rename("gopls", testFile, "oldMethod", "newMethod", tmpDir)
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	updated, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	updatedStr := string(updated)

	if strings.Contains(updatedStr, "oldMethod") {
		t.Errorf("Old method name still exists")
	}
	if !strings.Contains(updatedStr, "func (m *MyStruct) newMethod()") {
		t.Errorf("New method definition not found")
	}
	if !strings.Contains(updatedStr, "s.newMethod()") {
		t.Errorf("Method call was not updated")
	}
}

func TestRenameInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

func validFunc() {}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Try to rename a symbol that doesn't exist
	err := Rename("gopls", testFile, "nonexistent", "newName", tmpDir)
	if err == nil {
		t.Errorf("Expected error when renaming nonexistent symbol, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestRenameExport(t *testing.T) {
	// Test renaming from unexported to exported
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

func helper() {
	println("helper")
}

func main() {
	helper()
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := Rename("gopls", testFile, "helper", "Helper", tmpDir)
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	updated, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	updatedStr := string(updated)

	if !strings.Contains(updatedStr, "func Helper()") {
		t.Errorf("Function was not exported")
	}
	if !strings.Contains(updatedStr, "Helper()") {
		t.Errorf("Function call was not updated")
	}
}

func TestRenameUnexport(t *testing.T) {
	// Test renaming from exported to unexported
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

func PublicFunc() {
	println("public")
}

func main() {
	PublicFunc()
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := Rename("gopls", testFile, "PublicFunc", "privateFunc", tmpDir)
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	updated, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	updatedStr := string(updated)

	if !strings.Contains(updatedStr, "func privateFunc()") {
		t.Errorf("Function was not unexported")
	}
	if !strings.Contains(updatedStr, "privateFunc()") {
		t.Errorf("Function call was not updated")
	}
}

func TestRenameEmptyName(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package main

func test() {}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := Rename("gopls", testFile, "test", "", tmpDir)
	if err == nil {
		t.Errorf("Expected error for empty new name")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected 'cannot be empty' error, got: %v", err)
	}
}

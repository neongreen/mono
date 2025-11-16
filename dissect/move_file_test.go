package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neongreen/mono/dissect/pkg/refactor"
)

// TestMoveFile tests basic file moving/renaming functionality
func TestMoveFile(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dissect_move_file_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a source file
	sourceFile := filepath.Join(tmpDir, "source.go")
	sourceContent := `package main

import "fmt"

func Foo() {
	fmt.Println("Hello from Foo")
}

func Bar() {
	fmt.Println("Hello from Bar")
}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Initialize go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Move the file
	targetFile := filepath.Join(tmpDir, "target.go")
	args := []string{"move", sourceFile, targetFile}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Run the move command
	runMove(moveCmd, args[1:])

	// Verify source file no longer exists
	if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
		t.Errorf("Source file still exists after move")
	}

	// Verify target file exists
	if _, err := os.Stat(targetFile); err != nil {
		t.Fatalf("Target file does not exist: %v", err)
	}

	// Verify target file has the correct content
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	if string(targetContent) != sourceContent {
		t.Errorf("Target file content doesn't match source.\nExpected:\n%s\nGot:\n%s", sourceContent, string(targetContent))
	}
}

// TestMoveFileToSubdirectory tests moving a file to a subdirectory (package should change)
func TestMoveFileToSubdirectory(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dissect_move_file_subdir_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a source file
	sourceFile := filepath.Join(tmpDir, "admin_cmd.go")
	sourceContent := `package main

func AdminCommand() {
	// admin command logic
}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Initialize go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Move the file to a subdirectory
	targetFile := filepath.Join(tmpDir, "cmd", "admin.go")
	args := []string{"move", sourceFile, targetFile}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Run the move command
	runMove(moveCmd, args[1:])

	// Verify source file no longer exists
	if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
		t.Errorf("Source file still exists after move")
	}

	// Verify target file exists
	if _, err := os.Stat(targetFile); err != nil {
		t.Fatalf("Target file does not exist: %v", err)
	}

	// Verify target file has package changed to match directory
	expectedContent := `package cmd

func AdminCommand() {
	// admin command logic
}
`
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	if string(targetContent) != expectedContent {
		t.Errorf("Target file content doesn't match expected.\nExpected:\n%s\nGot:\n%s", expectedContent, string(targetContent))
	}
}

// TestMoveFileWithImportUpdates tests that imports are updated when moving files between packages
func TestMoveFileWithImportUpdates(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dissect_move_file_imports_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a source file in a helper subdirectory
	helperDir := filepath.Join(tmpDir, "helper")
	err = os.MkdirAll(helperDir, 0o755)
	if err != nil {
		t.Fatalf("Failed to create helper dir: %v", err)
	}

	sourceFile := filepath.Join(helperDir, "util.go")
	sourceContent := `package helper

func Helper() string {
	return "helper"
}
`
	err = os.WriteFile(sourceFile, []byte(sourceContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Create a main file that imports from helper
	mainFile := filepath.Join(tmpDir, "main.go")
	mainContent := `package main

import "test/helper"

func main() {
	helper.Helper()
}
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	// Initialize go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Move the file from helper/ to utils/
	targetFile := filepath.Join(tmpDir, "utils", "helper.go")
	args := []string{"move", sourceFile, targetFile}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Run the move command
	runMove(moveCmd, args[1:])

	// Verify source file no longer exists
	if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
		t.Errorf("Source file still exists after move")
	}

	// Verify target file exists and has correct package
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}

	expectedTargetContent := `package utils

func Helper() string {
	return "helper"
}
`
	if string(targetContent) != expectedTargetContent {
		t.Errorf("Target file package not updated correctly.\nExpected:\n%s\nGot:\n%s", expectedTargetContent, string(targetContent))
	}

	// Verify main file has import updated from test/helper to test/utils
	mainContentBytes, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read main file: %v", err)
	}

	expectedMainContent := `package main

import "test/utils"

func main() {
	utils.Helper()
}
`
	if string(mainContentBytes) != expectedMainContent {
		t.Errorf("Main file imports not updated correctly.\nExpected:\n%s\nGot:\n%s", expectedMainContent, string(mainContentBytes))
	}
}

// TestMoveFileSamePackageToSubdirectory tests moving a file within the same package to a subdirectory
// This should break references from sibling files that use the moved definitions
func TestMoveFileSamePackageToSubdirectory(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "dissect_move_same_package_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create admin_cmd.go that defines AdminCmd variable in package main
	adminFile := filepath.Join(tmpDir, "admin_cmd.go")
	adminContent := `package main

var AdminCmd = "admin"

func AdminHelper() string {
	return "admin helper"
}
`
	err = os.WriteFile(adminFile, []byte(adminContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write admin file: %v", err)
	}

	// Create main.go that uses AdminCmd from admin_cmd.go (same package, no import)
	mainFile := filepath.Join(tmpDir, "main.go")
	mainContent := `package main

import "fmt"

func main() {
	fmt.Println(AdminCmd)
	fmt.Println(AdminHelper())
}
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	// Initialize go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Move admin_cmd.go to cmd/admin.go (creates new package)
	targetFile := filepath.Join(tmpDir, "cmd", "admin.go")
	args := []string{"move", adminFile, targetFile}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// Run the move command
	runMove(moveCmd, args[1:])

	// Verify source file no longer exists
	if _, err := os.Stat(adminFile); !os.IsNotExist(err) {
		t.Errorf("Source file still exists after move")
	}

	// Verify target file exists
	if _, err := os.Stat(targetFile); err != nil {
		t.Fatalf("Target file does not exist: %v", err)
	}

	// Verify target file has package cmd
	targetContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target file: %v", err)
	}
	if !strings.Contains(string(targetContent), "package cmd") {
		t.Errorf("Target file should have 'package cmd', got:\n%s", string(targetContent))
	}

	// The critical test: verify project DOES build after the move
	// The auto-fix should have qualified the references and added the import
	buildCmd := exec.Command("go", "build", ".")
	buildCmd.Dir = tmpDir
	buildOutput, buildErr := buildCmd.CombinedOutput()

	// We EXPECT this to succeed because dissect move should have:
	// 1. Moved the file to cmd/admin.go
	// 2. Updated the package declaration to "package cmd"
	// 3. Found unqualified references to AdminCmd and AdminHelper in main.go
	// 4. Qualified them as cmd.AdminCmd and cmd.AdminHelper
	// 5. Added "import test/cmd" to main.go
	if buildErr != nil {
		t.Errorf("Expected build to succeed after moving file with auto-fix, but it failed: %v\nOutput: %s", buildErr, buildOutput)

		// Print main.go for debugging
		mainContent, _ := os.ReadFile(mainFile)
		t.Logf("main.go after move:\n%s", string(mainContent))
	}

	// Verify main.go was updated with import and qualified references
	updatedMainContent, err := os.ReadFile(mainFile)
	if err != nil {
		t.Fatalf("Failed to read main file: %v", err)
	}
	mainStr := string(updatedMainContent)

	if !strings.Contains(mainStr, `"test/cmd"`) && !strings.Contains(mainStr, `test/cmd`) {
		t.Errorf("Expected main.go to have import for test/cmd, got:\n%s", mainStr)
	}

	// Check for qualified references (may be on different lines due to formatting)
	if !strings.Contains(mainStr, "cmd.") {
		t.Errorf("Expected main.go to have qualified references with cmd., got:\n%s", mainStr)
	}
}

// TestMoveFileBlocksOnUnexportedFuncDeps tests that moving a file with unexported function dependencies is blocked
func TestMoveFileBlocksOnUnexportedFuncDeps(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_deps_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create helper.go with unexported function
	helperFile := filepath.Join(tmpDir, "helper.go")
	helperContent := `package main

func helper() string {
	return "help"
}
`
	err = os.WriteFile(helperFile, []byte(helperContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write helper file: %v", err)
	}

	// Create caller.go that uses helper
	callerFile := filepath.Join(tmpDir, "caller.go")
	callerContent := `package main

func main() {
	helper()
}
`
	err = os.WriteFile(callerFile, []byte(callerContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write caller file: %v", err)
	}

	// Initialize go.mod
	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	targetFile := filepath.Join(tmpDir, "subdir", "caller.go")

	// Call the refactor function directly to get the error
	err = refactor.MoveFileWithImportUpdates(callerFile, targetFile, tmpDir, "goimports")

	// Assert: Error contains "cannot move" and "unexported"
	if err == nil {
		t.Fatal("Expected error when moving file with unexported dependencies, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "cannot move") {
		t.Errorf("Expected error to contain 'cannot move', got: %v", errMsg)
	}

	if !strings.Contains(errMsg, "unexported") {
		t.Errorf("Expected error to contain 'unexported', got: %v", errMsg)
	}

	// Assert: Error mentions "helper"
	if !strings.Contains(errMsg, "helper") {
		t.Errorf("Expected error to mention 'helper' symbol, got: %v", errMsg)
	}

	// Check that source file still exists (move was blocked)
	if _, err := os.Stat(callerFile); os.IsNotExist(err) {
		t.Errorf("Source file was moved despite unexported dependencies")
	}

	// Check that target file was not created
	if _, err := os.Stat(targetFile); !os.IsNotExist(err) {
		t.Errorf("Target file was created despite unexported dependencies")
	}
}

// TestMoveFileBlocksOnUnexportedVarDeps tests that moving a file with unexported variable dependencies is blocked
func TestMoveFileBlocksOnUnexportedVarDeps(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_deps_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config.go with unexported var
	configFile := filepath.Join(tmpDir, "config.go")
	configContent := `package main

var dbPath = "/tmp/db"
`
	err = os.WriteFile(configFile, []byte(configContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create main.go that uses dbPath
	mainFile := filepath.Join(tmpDir, "main.go")
	mainContent := `package main

import "fmt"

func main() {
	fmt.Println(dbPath)
}
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	targetFile := filepath.Join(tmpDir, "cmd", "main.go")

	// Call refactor function directly
	err = refactor.MoveFileWithImportUpdates(mainFile, targetFile, tmpDir, "goimports")

	// Assert: Error mentions "dbPath" variable
	if err == nil {
		t.Fatal("Expected error when moving file with unexported variable dependency, got nil")
	}

	if !strings.Contains(err.Error(), "dbPath") {
		t.Errorf("Expected error to mention 'dbPath', got: %v", err)
	}

	// Check that source file still exists
	if _, err := os.Stat(mainFile); os.IsNotExist(err) {
		t.Errorf("Source file was moved despite unexported variable dependency")
	}

	// Check that target was not created
	if _, err := os.Stat(targetFile); !os.IsNotExist(err) {
		t.Errorf("Target file was created despite unexported variable dependency")
	}
}

// TestMoveFileBlocksOnUnexportedTypeDeps tests that moving a file with unexported type dependencies is blocked
func TestMoveFileBlocksOnUnexportedTypeDeps(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_deps_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create types.go with unexported type
	typesFile := filepath.Join(tmpDir, "types.go")
	typesContent := `package main

type connection struct {
	addr string
}
`
	err = os.WriteFile(typesFile, []byte(typesContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write types file: %v", err)
	}

	// Create client.go that uses connection
	clientFile := filepath.Join(tmpDir, "client.go")
	clientContent := `package main

func dial() *connection {
	return &connection{addr: "localhost"}
}
`
	err = os.WriteFile(clientFile, []byte(clientContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write client file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	targetFile := filepath.Join(tmpDir, "pkg", "client.go")

	// Call refactor function directly
	err = refactor.MoveFileWithImportUpdates(clientFile, targetFile, tmpDir, "goimports")

	// Assert: Error mentions "connection" type
	if err == nil {
		t.Fatal("Expected error when moving file with unexported type dependency, got nil")
	}

	if !strings.Contains(err.Error(), "connection") {
		t.Errorf("Expected error to mention 'connection' type, got: %v", err)
	}

	// Check that source file still exists
	if _, err := os.Stat(clientFile); os.IsNotExist(err) {
		t.Errorf("Source file was moved despite unexported type dependency")
	}

	// Check that target was not created
	if _, err := os.Stat(targetFile); !os.IsNotExist(err) {
		t.Errorf("Target file was created despite unexported type dependency")
	}
}

// TestMoveFileBlocksOnUnexportedFieldDeps tests that moving a file with unexported field dependencies is blocked
func TestMoveFileBlocksOnUnexportedFieldDeps(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_deps_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create database.go with unexported field
	databaseFile := filepath.Join(tmpDir, "database.go")
	databaseContent := `package main

type DB struct {
	connection string
}

func NewDB() *DB {
	return &DB{connection: "localhost"}
}
`
	err = os.WriteFile(databaseFile, []byte(databaseContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write database file: %v", err)
	}

	// Create client.go that accesses the unexported field
	clientFile := filepath.Join(tmpDir, "client.go")
	clientContent := `package main

func GetConnection(db *DB) string {
	return db.connection
}
`
	err = os.WriteFile(clientFile, []byte(clientContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write client file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	targetFile := filepath.Join(tmpDir, "pkg", "database.go")

	// Call refactor function directly
	err = refactor.MoveFileWithImportUpdates(databaseFile, targetFile, tmpDir, "goimports")

	// Assert: Error mentions the unexported field
	if err == nil {
		t.Fatal("Expected error when moving file with unexported field dependency, got nil")
	}

	if !strings.Contains(err.Error(), "connection") {
		t.Errorf("Expected error to mention 'connection' field, got: %v", err)
	}

	// Check that source file still exists
	if _, err := os.Stat(databaseFile); os.IsNotExist(err) {
		t.Errorf("Source file was moved despite unexported field dependency")
	}

	// Check that target was not created
	if _, err := os.Stat(targetFile); !os.IsNotExist(err) {
		t.Errorf("Target file was created despite unexported field dependency")
	}
}

// TestMoveFileAllowsExportedDeps tests that moving a file with only exported dependencies succeeds
func TestMoveFileAllowsExportedDeps(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_deps_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create api.go with exported function
	apiFile := filepath.Join(tmpDir, "api.go")
	apiContent := `package main

func Helper() string {
	return "help"
}
`
	err = os.WriteFile(apiFile, []byte(apiContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write api file: %v", err)
	}

	// Create utils.go that uses Helper
	utilsFile := filepath.Join(tmpDir, "utils.go")
	utilsContent := `package main

func UseHelper() {
	Helper()
}
`
	err = os.WriteFile(utilsFile, []byte(utilsContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write utils file: %v", err)
	}

	// Create main.go that will stay in the main package
	mainFile := filepath.Join(tmpDir, "main.go")
	mainContent := `package main

func main() {
	UseHelper()
}
`
	err = os.WriteFile(mainFile, []byte(mainContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write main file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Move utils.go to subdir (it references Helper which is exported)
	targetFile := filepath.Join(tmpDir, "subdir", "utils.go")

	// Call refactor function directly
	err = refactor.MoveFileWithImportUpdates(utilsFile, targetFile, tmpDir, "goimports")
	// Assert: Move succeeds (exported symbols are fine to reference across packages)
	if err != nil {
		t.Fatalf("Expected move to succeed with exported dependencies, got error: %v", err)
	}

	// Verify move succeeded
	if _, err := os.Stat(utilsFile); !os.IsNotExist(err) {
		t.Errorf("Source file still exists after move with exported deps")
	}

	if _, err := os.Stat(targetFile); err != nil {
		t.Fatalf("Target file does not exist: %v", err)
	}

	// Note: This test doesn't check if the code builds because auto-qualifying
	// references FROM the moved file to the original package is not yet implemented.
	// The test just verifies that moves with exported dependencies are not blocked.
}

// TestMoveFileBlocksOnMultipleUnexportedDeps tests error lists multiple dependencies
func TestMoveFileBlocksOnMultipleUnexportedDeps(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_deps_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create file with 3 unexported functions
	helpersFile := filepath.Join(tmpDir, "helpers.go")
	helpersContent := `package main

func helper1() {}
func helper2() {}
func helper3() {}
`
	err = os.WriteFile(helpersFile, []byte(helpersContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write helpers file: %v", err)
	}

	// Create caller using all 3
	callerFile := filepath.Join(tmpDir, "caller.go")
	callerContent := `package main

func main() {
	helper1()
	helper2()
	helper3()
}
`
	err = os.WriteFile(callerFile, []byte(callerContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write caller file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	targetFile := filepath.Join(tmpDir, "pkg", "caller.go")

	// Call refactor function directly
	err = refactor.MoveFileWithImportUpdates(callerFile, targetFile, tmpDir, "goimports")

	// Assert: Error lists all 3 dependencies
	if err == nil {
		t.Fatal("Expected error when moving file with multiple unexported dependencies, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "helper1") || !strings.Contains(errMsg, "helper2") || !strings.Contains(errMsg, "helper3") {
		t.Errorf("Expected error to list all 3 dependencies (helper1, helper2, helper3), got: %v", errMsg)
	}

	// Check that source file still exists
	if _, err := os.Stat(callerFile); os.IsNotExist(err) {
		t.Errorf("Source file was moved despite multiple unexported dependencies")
	}
}

// TestMoveFileWithinSamePackageIgnoresDeps tests that moves within same package succeed
func TestMoveFileWithinSamePackageIgnoresDeps(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dissect_deps_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create helper and caller in same package
	helperFile := filepath.Join(tmpDir, "helper.go")
	helperContent := `package main

func helper() {}
`
	err = os.WriteFile(helperFile, []byte(helperContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write helper file: %v", err)
	}

	callerFile := filepath.Join(tmpDir, "caller.go")
	callerContent := `package main

func main() {
	helper()
}
`
	err = os.WriteFile(callerFile, []byte(callerContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to write caller file: %v", err)
	}

	err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0o644)
	if err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Move within same package (just rename)
	targetFile := filepath.Join(tmpDir, "renamed.go")

	// Call refactor function directly
	err = refactor.MoveFileWithImportUpdates(callerFile, targetFile, tmpDir, "goimports")
	// Assert: Move succeeds (same package = OK)
	if err != nil {
		t.Fatalf("Expected same-package move to succeed, got error: %v", err)
	}

	// Verify move succeeded (same package, unexported access is OK)
	if _, err := os.Stat(callerFile); !os.IsNotExist(err) {
		t.Errorf("Source file still exists after same-package move")
	}

	if _, err := os.Stat(targetFile); err != nil {
		t.Fatalf("Target file does not exist: %v", err)
	}

	// Verify it builds
	buildCmd := exec.Command("go", "build", ".")
	buildCmd.Dir = tmpDir
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Errorf("Build failed after same-package move: %v\nOutput: %s", err, output)
	}
}

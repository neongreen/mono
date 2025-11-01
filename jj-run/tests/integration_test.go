package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var jjRunBinary string

func init() {
	// Build the jj-run binary before running tests
	cmd := exec.Command("go", "build", "-o", "../jj-run", "../cmd/main.go")
	if err := cmd.Run(); err != nil {
		panic("Failed to build jj-run binary: " + err.Error())
	}
	absPath, err := filepath.Abs("../jj-run")
	if err != nil {
		panic("Failed to get absolute path: " + err.Error())
	}
	jjRunBinary = absPath
}

// runCommand executes a command and returns its output
func runCommand(t *testing.T, dir string, command string, args ...string) (string, string, int) {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("Failed to run command: %v", err)
		}
	}

	return stdout.String(), stderr.String(), exitCode
}

// setupRepo creates a test repository and returns its path
func setupRepo(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "jj-run-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize jj repo
	runCommand(t, tmpDir, "jj", "git", "init", "--colocate", ".")

	return tmpDir
}

// TestBasicFunctionality tests the basic merge functionality (test1)
func TestBasicFunctionality(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not found, skipping test")
	}

	repoDir := setupRepo(t)
	defer os.RemoveAll(repoDir)

	// Set PAGER to cat for non-interactive jj log
	os.Setenv("PAGER", "cat")

	// Create several commits
	oneFile := filepath.Join(repoDir, "one.txt")
	if err := os.WriteFile(oneFile, []byte("First commit\n"), 0644); err != nil {
		t.Fatalf("Failed to write one.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "one .txt file", "one.txt")

	multi1File := filepath.Join(repoDir, "multi1.txt")
	multi2File := filepath.Join(repoDir, "multi2.txt")
	if err := os.WriteFile(multi1File, []byte("Line A\nLine B\n"), 0644); err != nil {
		t.Fatalf("Failed to write multi1.txt: %v", err)
	}
	if err := os.WriteFile(multi2File, []byte("Another file\n"), 0644); err != nil {
		t.Fatalf("Failed to write multi2.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "multiple .txt files", "multi1.txt", "multi2.txt")

	thirdFile := filepath.Join(repoDir, "third.txt")
	if err := os.WriteFile(thirdFile, []byte("Third commit\n"), 0644); err != nil {
		t.Fatalf("Failed to write third.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "another single .txt file", "third.txt")

	// Show commit contents before merging
	stdoutBefore, _, _ := runCommand(t, repoDir, "jj", "log", "-p", "-r", "::")

	// Check that all files exist before merge
	if !strings.Contains(stdoutBefore, "one.txt") {
		t.Errorf("Expected one.txt in log before merge")
	}
	if !strings.Contains(stdoutBefore, "multi1.txt") {
		t.Errorf("Expected multi1.txt in log before merge")
	}
	if !strings.Contains(stdoutBefore, "third.txt") {
		t.Errorf("Expected third.txt in log before merge")
	}

	// Use jj-run to merge all .txt files
	_, stderr, exitCode := runCommand(t, repoDir, jjRunBinary, "-r", "::", `for f in *.txt; do cat "$f" >> merged.txt; rm "$f"; done`)

	if exitCode != 0 {
		t.Logf("jj-run stderr: %s", stderr)
		t.Fatalf("jj-run failed with exit code %d", exitCode)
	}

	// Show commit contents after merging
	stdoutAfter, _, _ := runCommand(t, repoDir, "jj", "log", "-p", "-r", "::")

	// Verify results
	if !strings.Contains(stdoutAfter, "merged.txt") {
		t.Errorf("Expected merged.txt in log after merge, got: %s", stdoutAfter)
	}

	// Original files should not be present
	for _, f := range []string{"one.txt", "multi1.txt", "multi2.txt", "third.txt"} {
		if strings.Contains(stdoutAfter, f) {
			t.Errorf("Original file %s still present in log after merge", f)
		}
	}

	t.Logf("Test passed: basic functionality works")
}

// TestErrorHandlingContinue tests error handling with continue strategy (test2 part 1)
func TestErrorHandlingContinue(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not found, skipping test")
	}

	repoDir := setupRepo(t)
	defer os.RemoveAll(repoDir)

	os.Setenv("PAGER", "cat")

	// Create commits
	oneFile := filepath.Join(repoDir, "one.txt")
	if err := os.WriteFile(oneFile, []byte("First commit\n"), 0644); err != nil {
		t.Fatalf("Failed to write one.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "one .txt file", "one.txt")

	failmeFile := filepath.Join(repoDir, "failme.txt")
	if err := os.WriteFile(failmeFile, []byte("This will fail\n"), 0644); err != nil {
		t.Fatalf("Failed to write failme.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "failme", "failme.txt")

	thirdFile := filepath.Join(repoDir, "third.txt")
	if err := os.WriteFile(thirdFile, []byte("Third commit\n"), 0644); err != nil {
		t.Fatalf("Failed to write third.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "another single .txt file", "third.txt")

	// Run jj-run with a command that fails if failme.txt exists
	_, stderr, exitCode := runCommand(t, repoDir, jjRunBinary, "-r", "::", "-e", "continue", "test -f failme.txt && exit 1")

	// Should exit 0 with -e continue
	if exitCode != 0 {
		t.Errorf("Expected exit code 0 with -e continue, got %d", exitCode)
	}

	// Should report error for failed command
	if !strings.Contains(stderr, "Error while processing change") {
		t.Errorf("Expected error message in stderr, got: %s", stderr)
	}

	if !strings.Contains(stderr, "Command failed with return code 1") {
		t.Errorf("Expected 'Command failed with return code 1' in stderr, got: %s", stderr)
	}

	t.Logf("Test passed: error handling with continue strategy works")
}

// TestErrorHandlingStop tests error handling with stop strategy (test_stop)
func TestErrorHandlingStop(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not found, skipping test")
	}

	repoDir := setupRepo(t)
	defer os.RemoveAll(repoDir)

	os.Setenv("PAGER", "cat")

	// Create commits
	oneFile := filepath.Join(repoDir, "one.txt")
	if err := os.WriteFile(oneFile, []byte("First commit\n"), 0644); err != nil {
		t.Fatalf("Failed to write one.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "one .txt file", "one.txt")

	failmeFile := filepath.Join(repoDir, "failme.txt")
	if err := os.WriteFile(failmeFile, []byte("This will fail\n"), 0644); err != nil {
		t.Fatalf("Failed to write failme.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "failme", "failme.txt")

	thirdFile := filepath.Join(repoDir, "third.txt")
	if err := os.WriteFile(thirdFile, []byte("Third commit\n"), 0644); err != nil {
		t.Fatalf("Failed to write third.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "another single .txt file", "third.txt")

	// Run jj-run with a command that fails if failme.txt exists
	_, stderr, exitCode := runCommand(t, repoDir, jjRunBinary, "-r", "::", "-e", "stop", "test -f failme.txt && exit 1")

	// Should exit nonzero with -e stop
	if exitCode == 0 {
		t.Errorf("Expected nonzero exit code with -e stop on failure, got 0")
	}

	if !strings.Contains(stderr, "Command failed with return code 1") {
		t.Errorf("Expected 'Command failed with return code 1' in stderr, got: %s", stderr)
	}

	if !strings.Contains(stderr, "Stopped on change") {
		t.Errorf("Expected 'Stopped on change' in stderr, got: %s", stderr)
	}

	t.Logf("Test passed: error handling with stop strategy works")
}

// TestErrorHandlingFatal tests error handling with fatal strategy (test_fatal)
func TestErrorHandlingFatal(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not found, skipping test")
	}

	repoDir := setupRepo(t)
	defer os.RemoveAll(repoDir)

	os.Setenv("PAGER", "cat")

	// Create commits
	oneFile := filepath.Join(repoDir, "one.txt")
	if err := os.WriteFile(oneFile, []byte("First commit\n"), 0644); err != nil {
		t.Fatalf("Failed to write one.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "one .txt file", "one.txt")

	failmeFile := filepath.Join(repoDir, "failme.txt")
	if err := os.WriteFile(failmeFile, []byte("This will fail\n"), 0644); err != nil {
		t.Fatalf("Failed to write failme.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "failme", "failme.txt")

	thirdFile := filepath.Join(repoDir, "third.txt")
	if err := os.WriteFile(thirdFile, []byte("Third commit\n"), 0644); err != nil {
		t.Fatalf("Failed to write third.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "another single .txt file", "third.txt")

	// Run jj-run with a command that fails if failme.txt exists
	_, stderr, exitCode := runCommand(t, repoDir, jjRunBinary, "-r", "::", "-e", "fatal", "test -f failme.txt && exit 1")

	// Should exit nonzero with -e fatal
	if exitCode == 0 {
		t.Errorf("Expected nonzero exit code with -e fatal on failure, got 0")
	}

	if !strings.Contains(stderr, "Command failed with return code 1") {
		t.Errorf("Expected 'Command failed with return code 1' in stderr, got: %s", stderr)
	}

	if !strings.Contains(stderr, "Fatal error at change") {
		t.Errorf("Expected 'Fatal error at change' in stderr, got: %s", stderr)
	}

	t.Logf("Test passed: error handling with fatal strategy works")
}

// TestParentRewriting tests that parent rewriting works correctly
func TestParentRewriting(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not found, skipping test")
	}

	repoDir := setupRepo(t)
	defer os.RemoveAll(repoDir)

	os.Setenv("PAGER", "cat")

	// Create a simple commit
	file1 := filepath.Join(repoDir, "file1.txt")
	if err := os.WriteFile(file1, []byte("content1\n"), 0644); err != nil {
		t.Fatalf("Failed to write file1.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "first commit", "file1.txt")

	// Run jj-run with a command that modifies the file
	stdout, stderr, exitCode := runCommand(t, repoDir, jjRunBinary, "-r", "::", "echo 'modified' > file1.txt")

	if exitCode != 0 {
		t.Logf("jj-run stderr: %s", stderr)
		t.Fatalf("jj-run failed with exit code %d", exitCode)
	}

	t.Logf("jj-run stdout: %s", stdout)
	t.Logf("jj-run stderr: %s", stderr)

	// Check that the command reported rewrites
	if !strings.Contains(stderr, "Rewrote") {
		t.Errorf("Expected 'Rewrote' in stderr")
	}

	// Parse the rewrite count - should be "Rewrote 1/1 commits" or similar
	if strings.Contains(stderr, "Rewrote 0/") {
		t.Errorf("Expected at least 1 commit to be rewritten, but got 0")
	}

	// Verify the file was actually modified in the commit
	logOutput, _, _ := runCommand(t, repoDir, "jj", "log", "-p", "-r", "::")

	if !strings.Contains(logOutput, "modified") {
		t.Errorf("Expected 'modified' in the commit content, got: %s", logOutput)
	}

	t.Logf("Test passed: parent rewriting works correctly")
}

// TestWorkspaceCleanupOnStop tests that temporary workspaces are cleaned up when stop error occurs
func TestWorkspaceCleanupOnStop(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not found, skipping test")
	}

	repoDir := setupRepo(t)
	defer os.RemoveAll(repoDir)

	os.Setenv("PAGER", "cat")

	// Create commits
	oneFile := filepath.Join(repoDir, "one.txt")
	if err := os.WriteFile(oneFile, []byte("First commit\n"), 0644); err != nil {
		t.Fatalf("Failed to write one.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "one .txt file", "one.txt")

	failmeFile := filepath.Join(repoDir, "failme.txt")
	if err := os.WriteFile(failmeFile, []byte("This will fail\n"), 0644); err != nil {
		t.Fatalf("Failed to write failme.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "failme", "failme.txt")

	// Get initial workspace list
	workspacesBefore, _, _ := runCommand(t, repoDir, "jj", "workspace", "list")
	t.Logf("Workspaces before: %s", workspacesBefore)

	// Run jj-run with a command that fails (stop strategy)
	_, stderr, exitCode := runCommand(t, repoDir, jjRunBinary, "-r", "::", "-e", "stop", "test -f failme.txt && exit 1")

	// Should exit nonzero with -e stop
	if exitCode == 0 {
		t.Errorf("Expected nonzero exit code with -e stop on failure, got 0")
	}

	if !strings.Contains(stderr, "Stopped on change") {
		t.Errorf("Expected 'Stopped on change' in stderr, got: %s", stderr)
	}

	// Get workspace list after error
	workspacesAfter, _, _ := runCommand(t, repoDir, "jj", "workspace", "list")
	t.Logf("Workspaces after: %s", workspacesAfter)

	// Verify that temporary workspaces were cleaned up
	// The workspaces should be the same (only the default workspace should remain)
	workspaceLinesBefore := strings.Split(strings.TrimSpace(workspacesBefore), "\n")
	workspaceLinesAfter := strings.Split(strings.TrimSpace(workspacesAfter), "\n")

	if len(workspaceLinesAfter) != len(workspaceLinesBefore) {
		t.Errorf("Expected same number of workspaces after cleanup. Before: %d, After: %d",
			len(workspaceLinesBefore), len(workspaceLinesAfter))
		t.Logf("Before: %s", workspacesBefore)
		t.Logf("After: %s", workspacesAfter)
	}

	// Check that no jj-run temporary workspaces remain
	if strings.Contains(workspacesAfter, "jj-run-") {
		t.Errorf("Found temporary jj-run workspace in workspace list after stop error: %s", workspacesAfter)
	}

	t.Logf("Test passed: workspace cleanup on stop error works")
}

// TestWorkspaceCleanupOnFatal tests that temporary workspaces are cleaned up when fatal error occurs
func TestWorkspaceCleanupOnFatal(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not found, skipping test")
	}

	repoDir := setupRepo(t)
	defer os.RemoveAll(repoDir)

	os.Setenv("PAGER", "cat")

	// Create commits
	oneFile := filepath.Join(repoDir, "one.txt")
	if err := os.WriteFile(oneFile, []byte("First commit\n"), 0644); err != nil {
		t.Fatalf("Failed to write one.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "one .txt file", "one.txt")

	failmeFile := filepath.Join(repoDir, "failme.txt")
	if err := os.WriteFile(failmeFile, []byte("This will fail\n"), 0644); err != nil {
		t.Fatalf("Failed to write failme.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "failme", "failme.txt")

	// Get initial workspace list
	workspacesBefore, _, _ := runCommand(t, repoDir, "jj", "workspace", "list")
	t.Logf("Workspaces before: %s", workspacesBefore)

	// Run jj-run with a command that fails (fatal strategy)
	_, stderr, exitCode := runCommand(t, repoDir, jjRunBinary, "-r", "::", "-e", "fatal", "test -f failme.txt && exit 1")

	// Should exit nonzero with -e fatal
	if exitCode == 0 {
		t.Errorf("Expected nonzero exit code with -e fatal on failure, got 0")
	}

	if !strings.Contains(stderr, "Fatal error at change") {
		t.Errorf("Expected 'Fatal error at change' in stderr, got: %s", stderr)
	}

	// Get workspace list after error
	workspacesAfter, _, _ := runCommand(t, repoDir, "jj", "workspace", "list")
	t.Logf("Workspaces after: %s", workspacesAfter)

	// Verify that temporary workspaces were cleaned up
	// The workspaces should be the same (only the default workspace should remain)
	workspaceLinesBefore := strings.Split(strings.TrimSpace(workspacesBefore), "\n")
	workspaceLinesAfter := strings.Split(strings.TrimSpace(workspacesAfter), "\n")

	if len(workspaceLinesAfter) != len(workspaceLinesBefore) {
		t.Errorf("Expected same number of workspaces after cleanup. Before: %d, After: %d",
			len(workspaceLinesBefore), len(workspaceLinesAfter))
		t.Logf("Before: %s", workspacesBefore)
		t.Logf("After: %s", workspacesAfter)
	}

	// Check that no jj-run temporary workspaces remain
	if strings.Contains(workspacesAfter, "jj-run-") {
		t.Errorf("Found temporary jj-run workspace in workspace list after fatal error: %s", workspacesAfter)
	}

	t.Logf("Test passed: workspace cleanup on fatal error works")
}

// TestDirectModeBasic tests basic direct mode functionality
func TestDirectModeBasic(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not found, skipping test")
	}

	repoDir := setupRepo(t)
	defer os.RemoveAll(repoDir)

	os.Setenv("PAGER", "cat")

	// Create several commits
	oneFile := filepath.Join(repoDir, "one.txt")
	if err := os.WriteFile(oneFile, []byte("First commit\n"), 0644); err != nil {
		t.Fatalf("Failed to write one.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "first commit", "one.txt")

	twoFile := filepath.Join(repoDir, "two.txt")
	if err := os.WriteFile(twoFile, []byte("Second commit\n"), 0644); err != nil {
		t.Fatalf("Failed to write two.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "second commit", "two.txt")

	// Use direct mode to add a marker file to each commit
	_, stderr, exitCode := runCommand(t, repoDir, jjRunBinary, "--direct", "-r", "::", "echo 'processed' > .processed")

	if exitCode != 0 {
		t.Logf("jj-run stderr: %s", stderr)
		t.Fatalf("jj-run failed with exit code %d", exitCode)
	}

	// Verify the marker file was added to the commits
	logOutput, _, _ := runCommand(t, repoDir, "jj", "log", "-p", "-r", "::")

	if !strings.Contains(logOutput, ".processed") {
		t.Errorf("Expected .processed file in commits, got: %s", logOutput)
	}

	if !strings.Contains(logOutput, "processed") {
		t.Errorf("Expected 'processed' content in commits, got: %s", logOutput)
	}

	// Check that direct mode message appears
	if !strings.Contains(stderr, "direct mode") {
		t.Errorf("Expected 'direct mode' message in stderr, got: %s", stderr)
	}

	t.Logf("Test passed: direct mode basic functionality works")
}

// TestDirectModeErrorHandling tests error handling in direct mode
func TestDirectModeErrorHandling(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not found, skipping test")
	}

	repoDir := setupRepo(t)
	defer os.RemoveAll(repoDir)

	os.Setenv("PAGER", "cat")

	// Create commits
	oneFile := filepath.Join(repoDir, "one.txt")
	if err := os.WriteFile(oneFile, []byte("First commit\n"), 0644); err != nil {
		t.Fatalf("Failed to write one.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "first commit", "one.txt")

	failmeFile := filepath.Join(repoDir, "failme.txt")
	if err := os.WriteFile(failmeFile, []byte("This will fail\n"), 0644); err != nil {
		t.Fatalf("Failed to write failme.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "failme", "failme.txt")

	twoFile := filepath.Join(repoDir, "two.txt")
	if err := os.WriteFile(twoFile, []byte("Second commit\n"), 0644); err != nil {
		t.Fatalf("Failed to write two.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "second commit", "two.txt")

	// Run direct mode with a command that fails if failme.txt exists (continue strategy)
	_, stderr, exitCode := runCommand(t, repoDir, jjRunBinary, "--direct", "-r", "::", "-e", "continue", "test -f failme.txt && exit 1")

	// Should report error but continue
	if exitCode == 0 {
		t.Errorf("Expected nonzero exit code due to failed command")
	}

	if !strings.Contains(stderr, "Command failed with return code 1") {
		t.Errorf("Expected 'Command failed with return code 1' in stderr, got: %s", stderr)
	}

	// Should process all changes despite the error
	if !strings.Contains(stderr, "Processed") {
		t.Errorf("Expected 'Processed' message in stderr, got: %s", stderr)
	}

	t.Logf("Test passed: direct mode error handling works")
}

// TestDirectModeStop tests stop error strategy in direct mode
func TestDirectModeStop(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not found, skipping test")
	}

	repoDir := setupRepo(t)
	defer os.RemoveAll(repoDir)

	os.Setenv("PAGER", "cat")

	// Create commits
	oneFile := filepath.Join(repoDir, "one.txt")
	if err := os.WriteFile(oneFile, []byte("First commit\n"), 0644); err != nil {
		t.Fatalf("Failed to write one.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "first commit", "one.txt")

	failmeFile := filepath.Join(repoDir, "failme.txt")
	if err := os.WriteFile(failmeFile, []byte("This will fail\n"), 0644); err != nil {
		t.Fatalf("Failed to write failme.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "failme", "failme.txt")

	// Run direct mode with stop strategy
	_, stderr, exitCode := runCommand(t, repoDir, jjRunBinary, "--direct", "-r", "::", "-e", "stop", "test -f failme.txt && exit 1")

	// Should exit nonzero with -e stop
	if exitCode == 0 {
		t.Errorf("Expected nonzero exit code with -e stop on failure")
	}

	if !strings.Contains(stderr, "Stopped on change") {
		t.Errorf("Expected 'Stopped on change' in stderr, got: %s", stderr)
	}

	t.Logf("Test passed: direct mode stop strategy works")
}

// TestDirectModeNoWorkspaces tests that direct mode doesn't create workspaces
func TestDirectModeNoWorkspaces(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not found, skipping test")
	}

	repoDir := setupRepo(t)
	defer os.RemoveAll(repoDir)

	os.Setenv("PAGER", "cat")

	// Create a commit
	oneFile := filepath.Join(repoDir, "one.txt")
	if err := os.WriteFile(oneFile, []byte("First commit\n"), 0644); err != nil {
		t.Fatalf("Failed to write one.txt: %v", err)
	}
	runCommand(t, repoDir, "jj", "commit", "-m", "first commit", "one.txt")

	// Get initial workspace list
	workspacesBefore, _, _ := runCommand(t, repoDir, "jj", "workspace", "list")
	t.Logf("Workspaces before: %s", workspacesBefore)

	// Run direct mode command
	_, stderr, exitCode := runCommand(t, repoDir, jjRunBinary, "--direct", "-r", "::", "echo 'test' > .test")

	if exitCode != 0 {
		t.Logf("jj-run stderr: %s", stderr)
		t.Fatalf("jj-run failed with exit code %d", exitCode)
	}

	// Get workspace list after direct mode
	workspacesAfter, _, _ := runCommand(t, repoDir, "jj", "workspace", "list")
	t.Logf("Workspaces after: %s", workspacesAfter)

	// Verify that no temporary workspaces were created
	workspaceLinesBefore := strings.Split(strings.TrimSpace(workspacesBefore), "\n")
	workspaceLinesAfter := strings.Split(strings.TrimSpace(workspacesAfter), "\n")

	if len(workspaceLinesAfter) != len(workspaceLinesBefore) {
		t.Errorf("Expected same number of workspaces in direct mode. Before: %d, After: %d",
			len(workspaceLinesBefore), len(workspaceLinesAfter))
		t.Logf("Before: %s", workspacesBefore)
		t.Logf("After: %s", workspacesAfter)
	}

	// Check that no jj-run temporary workspaces were created
	if strings.Contains(workspacesAfter, "jj-run-") {
		t.Errorf("Found temporary jj-run workspace in direct mode: %s", workspacesAfter)
	}

	t.Logf("Test passed: direct mode doesn't create workspaces")
}

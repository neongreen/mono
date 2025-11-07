package testutils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// CompareDirectories compares the contents of expected files (from files_out)
// with the actual files generated in a specified directory using git diff --no-index.
// It returns an error if any discrepancies are found.
func CompareDirectories(t *testing.T, expectedFiles map[string]string, actualDirPath string) error {
	t.Helper()

	// 1. Create a temporary directory for expected files
	expectedDirPath, err := os.MkdirTemp("", "expected_files_")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory for expected files: %w", err)
	}
	// Don't clean up the temporary directory - it's useful for debug:
	// defer os.RemoveAll(expectedDirPath)

	// Write expected files into the temporary directory
	for filePath, content := range expectedFiles {
		fullPath := filepath.Join(expectedDirPath, filePath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to write expected file %s: %w", fullPath, err)
		}
	}

	// Run go fmt on the expected files directory to normalize formatting
	fmtCmd := exec.Command("go", "fmt", "./...")
	fmtCmd.Dir = expectedDirPath
	fmtOutput, fmtErr := fmtCmd.CombinedOutput()
	if fmtErr != nil {
		return fmt.Errorf("go fmt failed on expected files directory: %w\nOutput: %s", fmtErr, fmtOutput)
	}

	// 2. Compare directories using git diff --no-index
	cmd := exec.Command("git", "diff", "--no-index", expectedDirPath, actualDirPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// git diff returns non-zero exit code if differences are found
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			// Differences found, return the diff output as an error
			return fmt.Errorf("directories differ:\n%s", string(output))
		}
		// Other error (e.g., git not found, permission issues)
		return fmt.Errorf("failed to run git diff: %w\nOutput: %s", err, string(output))
	}

	// If err is nil, directories are identical
	return nil
}

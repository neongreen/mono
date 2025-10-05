package externaltest

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ProcessFileFunc is the signature for the function that processes a file
type ProcessFileFunc func(absPath string) (status int, exclusionReason string, err error)

// ProjectConfig defines configuration for testing an external project
type ProjectConfig struct {
	Name        string   // Project name (e.g., "google/uuid")
	URL         string   // Git clone URL
	Commit      string   // Specific commit SHA to test
	TargetFiles []string // Files to run dissect on (relative to project root)
	ShowDiff    bool     // Whether to show git diff after dissect
	ProcessFile ProcessFileFunc // Function to process each file (injected dependency)
}

// TestResult contains the results of running dissect on an external project
type TestResult struct {
	ProjectDir     string
	FilesCreated   []string
	FilesBefore    int
	FilesAfter     int
	Diff           string
	BuildPassed    bool
	TestsPassed    bool
	Error          error
}

// Logger interface for testing
type Logger interface {
	Helper()
	Fatalf(format string, args ...interface{})
	Logf(format string, args ...interface{})
}

// RunExternalProjectTest clones a project, runs dissect, and validates the results
func RunExternalProjectTest(t Logger, config ProjectConfig) *TestResult {
	t.Helper()

	result := &TestResult{}

	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "dissect_external_"+sanitizeName(config.Name)+"_")
	if err != nil {
		result.Error = fmt.Errorf("failed to create temporary directory: %w", err)
		return result
	}
	slog.Debug("Temporary directory for external project", "dir", tmpDir, "project", config.Name)

	projectDir := filepath.Join(tmpDir, filepath.Base(config.Name))
	result.ProjectDir = projectDir

	// Clone the project
	slog.Debug("Cloning external project", "url", config.URL, "commit", config.Commit)
	cloneCmd := exec.Command("git", "clone", config.URL, projectDir)
	cloneOutput, cloneErr := cloneCmd.CombinedOutput()
	if cloneErr != nil {
		result.Error = fmt.Errorf("failed to clone project: %w\nOutput: %s", cloneErr, cloneOutput)
		return result
	}

	// Checkout the specific commit
	checkoutCmd := exec.Command("git", "checkout", config.Commit)
	checkoutCmd.Dir = projectDir
	checkoutOutput, checkoutErr := checkoutCmd.CombinedOutput()
	if checkoutErr != nil {
		result.Error = fmt.Errorf("failed to checkout commit: %w\nOutput: %s", checkoutErr, checkoutOutput)
		return result
	}

	// Count files before dissect
	result.FilesBefore = countGoFiles(projectDir)

	// Verify the project builds before dissect
	slog.Debug("Building project before dissect...")
	buildBeforeCmd := exec.Command("go", "build", "./...")
	buildBeforeCmd.Dir = projectDir
	buildBeforeOutput, buildBeforeErr := buildBeforeCmd.CombinedOutput()
	if buildBeforeErr != nil {
		result.Error = fmt.Errorf("project doesn't build before dissect: %w\nOutput: %s", buildBeforeErr, buildBeforeOutput)
		return result
	}

	// Verify tests pass before dissect
	slog.Debug("Running tests before dissect...")
	testBeforeCmd := exec.Command("go", "test", "./...")
	testBeforeCmd.Dir = projectDir
	testBeforeOutput, testBeforeErr := testBeforeCmd.CombinedOutput()
	if testBeforeErr != nil {
		result.Error = fmt.Errorf("tests don't pass before dissect: %w\nOutput: %s", testBeforeErr, testBeforeOutput)
		return result
	}
	slog.Debug("Tests passed before dissect", "output", string(testBeforeOutput))

	// Run dissect on target files
	if config.ProcessFile != nil {
		for _, targetFile := range config.TargetFiles {
			fullPath := filepath.Join(projectDir, targetFile)
			slog.Debug("Running dissect on target file", "file", targetFile)
			status, exclusionReason, err := config.ProcessFile(fullPath)
			if err != nil {
				result.Error = fmt.Errorf("dissect failed on %s: %w", targetFile, err)
				return result
			}
			slog.Debug("Dissect result", "file", targetFile, "status", status, "exclusionReason", exclusionReason)
		}
	}

	// Count files after dissect
	result.FilesAfter = countGoFiles(projectDir)
	result.FilesCreated = findNewGoFiles(projectDir)

	// Get git diff if requested
	if config.ShowDiff {
		diffCmd := exec.Command("git", "diff")
		diffCmd.Dir = projectDir
		diffOutput, _ := diffCmd.CombinedOutput()
		result.Diff = string(diffOutput)
		if result.Diff != "" {
			slog.Info("Git diff after dissect", "project", config.Name, "diff", result.Diff)
		}
	}

	// Verify the project still builds after dissect
	slog.Debug("Building project after dissect...")
	buildAfterCmd := exec.Command("go", "build", "./...")
	buildAfterCmd.Dir = projectDir
	buildAfterOutput, buildAfterErr := buildAfterCmd.CombinedOutput()
	if buildAfterErr != nil {
		result.Error = fmt.Errorf("project doesn't build after dissect: %w\nOutput: %s", buildAfterErr, buildAfterOutput)
		result.BuildPassed = false
		return result
	}
	result.BuildPassed = true

	// Verify tests still pass after dissect
	slog.Debug("Running tests after dissect...")
	testAfterCmd := exec.Command("go", "test", "./...")
	testAfterCmd.Dir = projectDir
	testAfterOutput, testAfterErr := testAfterCmd.CombinedOutput()
	if testAfterErr != nil {
		result.Error = fmt.Errorf("tests don't pass after dissect: %w\nOutput: %s", testAfterErr, testAfterOutput)
		result.TestsPassed = false
		return result
	}
	result.TestsPassed = true
	slog.Debug("Tests passed after dissect", "output", string(testAfterOutput))

	// Log the files created by dissect
	slog.Debug("Listing files after dissect...", "project", config.Name)
	walkErr := filepath.Walk(projectDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			relativePath, _ := filepath.Rel(projectDir, path)
			slog.Debug("Go file after dissect", "path", relativePath)
		}
		return nil
	})
	if walkErr != nil {
		slog.Error("Failed to walk project directory", "error", walkErr)
	}

	slog.Debug("External project test completed successfully", "project", config.Name,
		"files_before", result.FilesBefore, "files_after", result.FilesAfter,
		"files_created", len(result.FilesCreated))

	return result
}

// countGoFiles counts the number of .go files in a directory
func countGoFiles(dir string) int {
	count := 0
	filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			count++
		}
		return nil
	})
	return count
}

// findNewGoFiles finds newly created .go files (unstaged in git)
func findNewGoFiles(dir string) []string {
	var newFiles []string
	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard", "*.go")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return newFiles
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line != "" {
			newFiles = append(newFiles, line)
		}
	}
	return newFiles
}

// sanitizeName removes special characters from a name for use in filenames
func sanitizeName(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(name, "/", "_"), " ", "_")
}

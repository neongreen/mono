package externaltest

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ProcessFileFunc is the signature for the function that processes a file
type ProcessFileFunc func(absPath string) (status int, exclusionReason string, err error)

// ProjectConfig defines configuration for testing an external project
type ProjectConfig struct {
	Name        string          // Project name (e.g., "google/uuid")
	URL         string          // Git clone URL
	Commit      string          // Specific commit SHA to test
	ShowDiff    bool            // Whether to show git diff after dissect
	ProcessFile ProcessFileFunc // Function to process each file (injected dependency)
}

// TestResult contains the results of running dissect on an external project
type TestResult struct {
	ProjectDir   string
	FilesCreated []string
	FilesBefore  int
	FilesAfter   int
	Diff         string
	BuildPassed  bool
	TestsPassed  bool
	Error        error
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
	
	// Timing tracking
	type timing struct {
		operation string
		duration  time.Duration
	}
	var timings []timing
	
	logTiming := func(op string, start time.Time) {
		duration := time.Since(start)
		timings = append(timings, timing{operation: op, duration: duration})
		t.Logf("[TIMING] %s: %v", op, duration)
	}

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
	cloneStart := time.Now()
	slog.Debug("Cloning external project", "url", config.URL, "commit", config.Commit)
	cloneCmd := exec.Command("git", "clone", config.URL, projectDir)
	cloneOutput, cloneErr := cloneCmd.CombinedOutput()
	if cloneErr != nil {
		result.Error = fmt.Errorf("failed to clone project: %w\nOutput: %s", cloneErr, cloneOutput)
		return result
	}
	logTiming("git clone", cloneStart)

	// Checkout the specific commit
	checkoutStart := time.Now()
	checkoutCmd := exec.Command("git", "checkout", config.Commit)
	checkoutCmd.Dir = projectDir
	checkoutOutput, checkoutErr := checkoutCmd.CombinedOutput()
	if checkoutErr != nil {
		result.Error = fmt.Errorf("failed to checkout commit: %w\nOutput: %s", checkoutErr, checkoutOutput)
		return result
	}
	logTiming("git checkout", checkoutStart)

	// Count files before dissect
	result.FilesBefore = countGoFiles(projectDir)

	// Verify the project builds before dissect
	buildBeforeStart := time.Now()
	slog.Debug("Building project before dissect...")
	buildBeforeCmd := exec.Command("go", "build", "./...")
	buildBeforeCmd.Dir = projectDir
	buildBeforeOutput, buildBeforeErr := buildBeforeCmd.CombinedOutput()
	if buildBeforeErr != nil {
		result.Error = fmt.Errorf("project doesn't build before dissect: %w\nOutput: %s", buildBeforeErr, buildBeforeOutput)
		return result
	}
	logTiming("go build (before)", buildBeforeStart)

	// Verify tests pass before dissect
	testBeforeStart := time.Now()
	slog.Debug("Running tests before dissect...")
	testBeforeCmd := exec.Command("go", "test", "./...")
	testBeforeCmd.Dir = projectDir
	testBeforeOutput, testBeforeErr := testBeforeCmd.CombinedOutput()
	if testBeforeErr != nil {
		result.Error = fmt.Errorf("tests don't pass before dissect: %w\nOutput: %s", testBeforeErr, testBeforeOutput)
		return result
	}
	slog.Debug("Tests passed before dissect", "output", string(testBeforeOutput))
	logTiming("go test (before)", testBeforeStart)

	// Find all Go files in the project using go list
	goListStart := time.Now()
	slog.Debug("Finding all Go files in project...")
	goListCmd := exec.Command("go", "list", "-test", "-json", "./...")
	goListCmd.Dir = projectDir
	goListOutput, goListErr := goListCmd.CombinedOutput()
	if goListErr != nil {
		result.Error = fmt.Errorf("failed to list Go files: %w\nOutput: %s", goListErr, goListOutput)
		return result
	}
	logTiming("go list", goListStart)

	// Parse go list output to get all Go files
	// Process all non-test files in the main package (skip test packages, internal packages, cmd packages)
	var allGoFiles []string
	decoder := json.NewDecoder(strings.NewReader(string(goListOutput)))
	for {
		var pkg struct {
			Dir        string   `json:"Dir"`
			ImportPath string   `json:"ImportPath"`
			GoFiles    []string `json:"GoFiles"`
		}
		if err := decoder.Decode(&pkg); err != nil {
			if err.Error() == "EOF" {
				break
			}
			// Skip any decode errors and continue
			continue
		}

		// Skip test packages, cmd packages, and deeply nested internal packages
		if strings.Contains(pkg.ImportPath, ".test]") ||
			strings.Contains(pkg.ImportPath, "/cmd/") ||
			strings.Count(pkg.ImportPath, "/internal/") > 1 {
			continue
		}

		for _, goFile := range pkg.GoFiles {
			// Skip files with absolute paths
			if !filepath.IsAbs(goFile) {
				fullPath := filepath.Join(pkg.Dir, goFile)
				allGoFiles = append(allGoFiles, fullPath)
			}
		}
	}

	slog.Debug("Found Go files", "count", len(allGoFiles))

	// Run dissect on all Go files
	dissectStart := time.Now()
	if config.ProcessFile != nil && len(allGoFiles) > 0 {
		for _, goFile := range allGoFiles {
			relativePath, _ := filepath.Rel(projectDir, goFile)
			slog.Debug("Running dissect on file", "file", relativePath)
			status, exclusionReason, err := config.ProcessFile(goFile)
			if err != nil {
				result.Error = fmt.Errorf("dissect failed on %s: %w", relativePath, err)
				return result
			}
			slog.Debug("Dissect result", "file", relativePath, "status", status, "exclusionReason", exclusionReason)
		}
	}
	logTiming("dissect processing", dissectStart)

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
	buildAfterStart := time.Now()
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
	logTiming("go build (after)", buildAfterStart)

	// Verify tests still pass after dissect
	testAfterStart := time.Now()
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
	logTiming("go test (after)", testAfterStart)
	
	// Log timing summary
	t.Logf("[TIMING SUMMARY] %s:", config.Name)
	var total time.Duration
	for _, tm := range timings {
		t.Logf("  %s: %v", tm.operation, tm.duration)
		total += tm.duration
	}
	t.Logf("  TOTAL: %v", total)

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

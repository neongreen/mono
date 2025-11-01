package main_test

import (
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang-cz/devslog"
	"github.com/pelletier/go-toml/v2"

	main "github.com/neongreen/mono/dissect/cmd"
	"github.com/neongreen/mono/dissect/pkg/externaltest"
	"github.com/neongreen/mono/dissect/pkg/goutils"
	"github.com/neongreen/mono/dissect/pkg/testutils"
)

// Run this to init logging
func init() {
	// Init logging
	slog.SetDefault(slog.New(devslog.NewHandler(os.Stdout, &devslog.Options{HandlerOptions: &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}, NewLineAfterLog: true})))

	// Add go/bin to PATH so gopls can be found
	goBinPath := filepath.Join(os.Getenv("HOME"), "go", "bin")
	currentPath := os.Getenv("PATH")
	if !strings.Contains(currentPath, goBinPath) {
		os.Setenv("PATH", goBinPath+":"+currentPath)
	}

	// Ensure gopls is installed (required by dissect)
	checkGoplsCmd := exec.Command("gopls", "version")
	if err := checkGoplsCmd.Run(); err != nil {
		// gopls not found, install it
		installGoplsCmd := exec.Command("go", "install", "golang.org/x/tools/gopls@latest")
		if installErr := installGoplsCmd.Run(); installErr != nil {
			panic("Failed to install gopls: " + installErr.Error())
		}
	}

	// Ensure goimports is installed (required by dissect)
	checkGoimportsCmd := exec.Command("goimports", "-h")
	if err := checkGoimportsCmd.Run(); err != nil {
		// goimports not found, install it
		installGoimportsCmd := exec.Command("go", "install", "golang.org/x/tools/cmd/goimports@latest")
		if installErr := installGoimportsCmd.Run(); installErr != nil {
			panic("Failed to install goimports: " + installErr.Error())
		}
	}
}

// testFiles represents the structure of our TOML test files.
type testFiles struct {
	FilesIn  map[string]string `toml:"files_in"`
	FilesOut map[string]string `toml:"files_out"` // Add FilesOut
}

// loadTestFilesFromTOML reads a TOML file, parses it, and returns a testFiles struct.
func loadTestFilesFromTOML(t *testing.T, tomlPath string) testFiles { // Change return type
	t.Helper()
	data, err := os.ReadFile(tomlPath)
	if err != nil {
		t.Fatalf("Failed to read TOML file %s: %v", tomlPath, err)
	}

	var tf testFiles
	if err := toml.Unmarshal(data, &tf); err != nil {
		t.Fatalf("Failed to unmarshal TOML file %s: %v", tomlPath, err)
	}
	return tf
}

func runDissectIntegrationTest(t *testing.T, tomlFileName string) {
	t.Helper()

	// Set up shared Go cache directories for faster subsequent test runs
	// This allows all tests to share downloaded dependencies and build artifacts
	cacheDir := filepath.Join(os.TempDir(), "dissect_test_cache")
	goModCache := filepath.Join(cacheDir, "mod")
	goBuildCache := filepath.Join(cacheDir, "build")

	// Create cache directories if they don't exist
	if err := os.MkdirAll(goModCache, 0755); err != nil {
		t.Fatalf("Failed to create Go module cache directory: %v", err)
	}
	if err := os.MkdirAll(goBuildCache, 0755); err != nil {
		t.Fatalf("Failed to create Go build cache directory: %v", err)
	}

	// Create a temporary directory for the test project
	tmpProjectDir, err := os.MkdirTemp("", "dissect_"+tomlFileName+"_")
	if err != nil {
		t.Fatalf("Failed to create temporary project directory: %v", err)
	}
	slog.Debug("Temporary project directory", "dir", tmpProjectDir)
	// Don't clean up the temporary directory - it's useful for debug:
	// defer os.RemoveAll(tmpProjectDir)

	// Load test files from TOML
	repoRoot := findRepoRoot(t)
	tomlPath := filepath.Join(repoRoot, "dissect", "tests", tomlFileName)
	testData := loadTestFilesFromTOML(t, tomlPath)

	// Write files to the temporary project directory
	for filePath, content := range testData.FilesIn {
		fullPath := filepath.Join(tmpProjectDir, filePath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
		slog.Debug("Created file from test data", "file", fullPath)
	}

	// Run go mod tidy to initialize the module with shared cache
	slog.Debug("Running go mod tidy in temporary project directory...", "dir", tmpProjectDir)
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpProjectDir
	tidyCmd.Env = append(os.Environ(),
		"GOMODCACHE="+goModCache,
		"GOCACHE="+goBuildCache,
	)
	tidyOutput, tidyErr := tidyCmd.CombinedOutput()
	if tidyErr != nil {
		t.Fatalf("go mod tidy failed: %v\nOutput: %s", tidyErr, tidyOutput)
	}

	// Run dissect
	for filePath := range testData.FilesIn {
		if strings.HasSuffix(filePath, ".go") {
			slog.Debug("Found Go file for dissect", "file", filePath)
			main.ProcessFile(filepath.Join(tmpProjectDir, filePath))
		} else {
			slog.Debug("Skipping non-Go file for dissect", "file", filePath)
		}
	}

	// Run go fmt on the temporary project directory to normalize formatting
	slog.Debug("Running go fmt on temporary project directory...", "dir", tmpProjectDir)
	fmtCmd := exec.Command("go", "fmt", "./...")
	fmtCmd.Dir = tmpProjectDir
	fmtCmd.Env = append(os.Environ(),
		"GOMODCACHE="+goModCache,
		"GOCACHE="+goBuildCache,
	)
	fmtOutput, fmtErr := fmtCmd.CombinedOutput()
	if fmtErr != nil {
		t.Fatalf("go fmt failed: %v\nOutput: %s", fmtErr, fmtOutput)
	}

	// Normalize imports in the actual output files
	slog.Debug("Normalizing imports in actual output files...", "dir", tmpProjectDir)
	walkErr := filepath.Walk(tmpProjectDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			normalizedContent, normalizeErr := goutils.NormalizeImports(string(content))
			if normalizeErr != nil {
				return normalizeErr
			}
			if writeErr := os.WriteFile(path, []byte(normalizedContent), 0644); writeErr != nil {
				return writeErr
			}
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("Failed to normalize imports in output directory: %v", walkErr)
	}

	// Normalize imports in the expected output files
	for filePath, content := range testData.FilesOut {
		if strings.HasSuffix(filePath, ".go") {
			normalizedContent, normalizeErr := goutils.NormalizeImports(content)
			if normalizeErr != nil {
				t.Fatalf("Failed to normalize imports in expected file %s: %v", filePath, normalizeErr)
			}
			testData.FilesOut[filePath] = normalizedContent
		}
	}

	// Compare the actual output directory with the expected files
	if err := testutils.CompareDirectories(t, testData.FilesOut, tmpProjectDir); err != nil {
		t.Errorf("Directory comparison failed for %s: %v", tomlFileName, err)
	}

	// After the test, slog.debug all files in the out dir and their relative paths
	defer func() {
		slog.Debug("--- Start of output directory state ---")
		walkErr := filepath.Walk(tmpProjectDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				relativePath, _ := filepath.Rel(tmpProjectDir, path)
				content, readErr := os.ReadFile(path)
				if readErr != nil {
					slog.Error("Failed to read file in output directory", "path", relativePath, "error", readErr)
					return nil // Continue walking
				}
				slog.Debug("File in output directory", "path", relativePath, "content", "\n"+string(content))
			}
			return nil
		})
		if walkErr != nil {
			slog.Error("Failed to walk output directory", "error", walkErr)
		}
		slog.Debug("--- End of output directory state ---")
	}()

	// Verify that the project still builds with shared cache
	buildCmd := exec.Command("go", "build", "-o", "/dev/null", "./...") // Declare buildCmd with :=
	buildCmd.Dir = tmpProjectDir
	buildCmd.Env = append(os.Environ(),
		"GOMODCACHE="+goModCache,
		"GOCACHE="+goBuildCache,
	)
	buildOutput, buildErr := buildCmd.CombinedOutput() // Declare buildOutput, buildErr with :=
	if buildErr != nil {
		t.Fatalf("go build failed: %v\nOutput: %s", buildErr, buildOutput)
	}
}

func TestAllDissectIntegration(t *testing.T) {
	// Get the repository root
	repoRoot := findRepoRoot(t)
	testsDir := filepath.Join(repoRoot, "dissect", "tests")

	// Find all .toml files in the tests directory
	var testFiles []string
	err := filepath.Walk(testsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".toml") {
			testFiles = append(testFiles, info.Name())
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to find test files: %v", err)
	}

	for _, fileName := range testFiles {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel() // Run tests in parallel if possible
			runDissectIntegrationTest(t, fileName)
		})
	}
}

// TestExternalProjects tests dissect on real external Go projects.
// This test suite runs dissect on multiple predefined projects and validates correctness.
func TestExternalProjects(t *testing.T) {
	// Skip this test by default as it's optional and takes longer
	if testing.Short() {
		t.Skip("Skipping external project tests in short mode")
	}

	// Get show diff flag from environment
	showDiff := os.Getenv("DISSECT_SHOW_DIFF") == "1" || os.Getenv("DISSECT_SHOW_DIFF") == "true"

	// Run test for each known project
	for projectName := range externaltest.KnownProjects {
		t.Run(projectName, func(t *testing.T) {
			t.Parallel() // Run external project tests concurrently
			config, ok := externaltest.GetProject(projectName)
			if !ok {
				t.Fatalf("Project %s not found", projectName)
			}

			// Override ShowDiff if environment variable is set
			if showDiff {
				config.ShowDiff = true
			}

			// Inject ProcessFile dependency
			config.ProcessFile = func(absPath string) (int, string, error) {
				status, exclusionReason, err := main.ProcessFile(absPath)
				return int(status), exclusionReason, err
			}

			result := externaltest.RunExternalProjectTest(t, config)
			if result.Error != nil {
				t.Fatalf("External project test failed for %s: %v", projectName, result.Error)
			}

			// Log summary
			t.Logf("✓ %s: %d files before, %d files after, %d new files created",
				projectName, result.FilesBefore, result.FilesAfter, len(result.FilesCreated))

			if config.ShowDiff && result.Diff != "" {
				t.Logf("Git diff for %s:\n%s", projectName, result.Diff)
			}
		})
	}
}

func findRepoRoot(t *testing.T) string {
	cmd := exec.Command("go", "env", "GOMOD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to execute 'go env GOMOD': %v\nOutput: %s", err, output)
	}

	goModPath := strings.TrimSpace(string(output))
	if goModPath == "" {
		t.Fatalf("go env GOMOD returned empty output")
	}

	// The module root is the directory containing the go.mod file
	moduleRoot := filepath.Dir(goModPath)
	slog.Debug("Go module root found using 'go env GOMOD'", "root", moduleRoot)

	return moduleRoot
}

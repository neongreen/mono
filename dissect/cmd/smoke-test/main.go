package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/golang-cz/devslog"

	"dissect/pkg/externaltest"
)

func main() {
	// Parse command line flags
	projectURL := flag.String("url", "", "Git clone URL for the project")
	projectCommit := flag.String("commit", "HEAD", "Git commit SHA to checkout (default: HEAD)")
	targetFiles := flag.String("files", "", "Comma-separated list of files to run dissect on")
	showDiff := flag.Bool("diff", false, "Show git diff after running dissect")
	listProjects := flag.Bool("list", false, "List available predefined projects")
	projectName := flag.String("project", "", "Name of predefined project to test (e.g., 'google/uuid')")

	flag.Parse()

	// Init logging
	slog.SetDefault(slog.New(devslog.NewHandler(os.Stdout, &devslog.Options{
		HandlerOptions: &slog.HandlerOptions{Level: slog.LevelInfo, AddSource: false},
		NewLineAfterLog: true,
	})))

	// List projects if requested
	if *listProjects {
		fmt.Println("Available predefined projects:")
		for _, name := range externaltest.GetProjectNames() {
			config, _ := externaltest.GetProject(name)
			fmt.Printf("  %s\n", name)
			fmt.Printf("    URL: %s\n", config.URL)
			fmt.Printf("    Commit: %s\n", config.Commit)
			fmt.Printf("    Target files: %v\n", config.TargetFiles)
		}
		return
	}

	var config externaltest.ProjectConfig

	// Use predefined project if specified
	if *projectName != "" {
		var ok bool
		config, ok = externaltest.GetProject(*projectName)
		if !ok {
			fmt.Fprintf(os.Stderr, "Error: Unknown project '%s'. Use -list to see available projects.\n", *projectName)
			os.Exit(1)
		}
		fmt.Printf("Using predefined project: %s\n", *projectName)
	} else if *projectURL != "" {
		// Use custom project
		if *targetFiles == "" {
			fmt.Fprintf(os.Stderr, "Error: -files is required when using -url\n")
			flag.Usage()
			os.Exit(1)
		}

		files := strings.Split(*targetFiles, ",")
		for i, f := range files {
			files[i] = strings.TrimSpace(f)
		}

		// Extract project name from URL
		name := extractProjectName(*projectURL)

		config = externaltest.ProjectConfig{
			Name:        name,
			URL:         *projectURL,
			Commit:      *projectCommit,
			TargetFiles: files,
			ShowDiff:    *showDiff,
		}
		fmt.Printf("Testing custom project: %s\n", name)
	} else {
		fmt.Fprintf(os.Stderr, "Error: Either -project or -url must be specified\n")
		flag.Usage()
		os.Exit(1)
	}

	// Override ShowDiff if specified
	if *showDiff {
		config.ShowDiff = true
	}

	fmt.Printf("Cloning and testing project...\n")

	// Run the test (without t *testing.T, we'll use a mock)
	result := runSmokeTest(config)

	// Print results
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("SMOKE TEST RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	if result.Error != nil {
		fmt.Printf("❌ FAILED: %v\n", result.Error)
		os.Exit(1)
	}

	fmt.Printf("✓ Project directory: %s\n", result.ProjectDir)
	fmt.Printf("✓ Go files before: %d\n", result.FilesBefore)
	fmt.Printf("✓ Go files after: %d\n", result.FilesAfter)
	fmt.Printf("✓ New files created: %d\n", len(result.FilesCreated))
	if len(result.FilesCreated) > 0 {
		fmt.Println("\nNew files:")
		for _, file := range result.FilesCreated {
			fmt.Printf("  - %s\n", file)
		}
	}
	fmt.Printf("✓ Build passed: %v\n", result.BuildPassed)
	fmt.Printf("✓ Tests passed: %v\n", result.TestsPassed)

	if config.ShowDiff && result.Diff != "" {
		fmt.Println("\n" + strings.Repeat("-", 60))
		fmt.Println("GIT DIFF")
		fmt.Println(strings.Repeat("-", 60))
		fmt.Println(result.Diff)
	}

	fmt.Println("\n✓ Smoke test passed successfully!")
	fmt.Printf("\nProject preserved at: %s\n", result.ProjectDir)
}

// runSmokeTest runs the external test without *testing.T
func runSmokeTest(config externaltest.ProjectConfig) *externaltest.TestResult {
	// Inject a ProcessFile function that just logs (actual dissect functionality would need to be linked)
	config.ProcessFile = func(absPath string) (int, string, error) {
		// This is a placeholder - in a real implementation, you would need to
		// either build dissect as a library or call it as a separate binary
		fmt.Printf("Would process file: %s\n", absPath)
		fmt.Println("Note: To use dissect functionality, import and call ProcessFile from dissect/cmd package")
		return 0, "", nil
	}

	// Create a mock logger for the test
	logger := &mockLogger{}
	return externaltest.RunExternalProjectTest(logger, config)
}

// mockLogger is a minimal implementation of Logger for non-test usage
type mockLogger struct{}

func (m *mockLogger) Helper()                                  {}
func (m *mockLogger) Fatalf(format string, args ...interface{}) { fmt.Fprintf(os.Stderr, format+"\n", args...); os.Exit(1) }
func (m *mockLogger) Logf(format string, args ...interface{})  { fmt.Printf(format+"\n", args...) }

// extractProjectName extracts a project name from a git URL
func extractProjectName(url string) string {
	// Remove .git suffix
	name := strings.TrimSuffix(url, ".git")
	// Extract last two path components (owner/repo)
	parts := strings.Split(name, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	return name
}

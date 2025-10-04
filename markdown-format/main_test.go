package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestFormatMarkdownFromFiles tests the markdown formatter using input/output file pairs from testdata directory.
// This approach makes it easy to add new test cases by simply adding new .input.md and .output.md file pairs.
func TestFormatMarkdownFromFiles(t *testing.T) {
	testdataDir := "testdata"

	// Find all input files
	inputFiles, err := filepath.Glob(filepath.Join(testdataDir, "*.input.md"))
	if err != nil {
		t.Fatalf("Failed to find input files: %v", err)
	}

	if len(inputFiles) == 0 {
		t.Skip("No test files found in testdata directory")
	}

	for _, inputFile := range inputFiles {
		// Derive test name and output file from input file
		baseName := filepath.Base(inputFile)
		testName := strings.TrimSuffix(baseName, ".input.md")
		outputFile := filepath.Join(testdataDir, testName+".output.md")

		t.Run(testName, func(t *testing.T) {
			// Read input file
			input, err := os.ReadFile(inputFile)
			if err != nil {
				t.Fatalf("Failed to read input file %s: %v", inputFile, err)
			}

			// Read expected output file
			expectedOutput, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file %s: %v", outputFile, err)
			}

			// Format the markdown
			actualOutput, err := formatMarkdown(input)
			if err != nil {
				t.Fatalf("formatMarkdown() error = %v", err)
			}

			// Compare outputs
			if diff := cmp.Diff(string(expectedOutput), string(actualOutput)); diff != "" {
				t.Errorf("formatMarkdown() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

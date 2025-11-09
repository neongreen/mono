// cobralint-sarif converts cobralint text output to SARIF format
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/owenrumney/go-sarif/v2/sarif"
)

// Diagnostic represents a parsed linter output line
type Diagnostic struct {
	File    string
	Line    int
	Column  int
	Message string
}

func main() {
	// Read from stdin
	diags, err := parseLinterOutput(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing input: %v\n", err)
		os.Exit(1)
	}

	// Generate SARIF
	if err := outputSARIF(diags); err != nil {
		fmt.Fprintf(os.Stderr, "error generating SARIF: %v\n", err)
		os.Exit(1)
	}
}

// parseLinterOutput parses text output from the analyzer
// Format: /path/to/file.go:line:column: message
func parseLinterOutput(r io.Reader) ([]Diagnostic, error) {
	var diags []Diagnostic
	// Pattern: filepath:line:column: message
	re := regexp.MustCompile(`^(.+?):(\d+):(\d+):\s*(.+)$`)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) == 5 {
			lineNum, _ := strconv.Atoi(matches[2])
			colNum, _ := strconv.Atoi(matches[3])
			diags = append(diags, Diagnostic{
				File:    matches[1],
				Line:    lineNum,
				Column:  colNum,
				Message: matches[4],
			})
		}
	}

	return diags, scanner.Err()
}

func outputSARIF(diags []Diagnostic) error {
	// Create SARIF report
	report, err := sarif.New(sarif.Version210)
	if err != nil {
		return fmt.Errorf("creating SARIF report: %w", err)
	}

	run := sarif.NewRunWithInformationURI("cobralint", "https://github.com/neongreen/mono/tree/main/linters/cobralint")
	run.Tool.Driver.WithVersion("1.0.0")

	// Add rule definitions
	run.AddRule("require-json-flag").
		WithDescription("All Cobra commands should have a --json flag for machine-readable output").
		WithHelpURI("https://github.com/neongreen/mono/tree/main/linters/cobralint#require-json-flag")

	// Add diagnostics to SARIF report
	for _, diag := range diags {
		// Convert to relative path from working directory
		relPath, err := filepath.Rel(".", diag.File)
		if err != nil {
			relPath = diag.File
		}

		result := run.CreateResultForRule("require-json-flag").
			WithMessage(sarif.NewTextMessage(diag.Message)).
			WithLevel("warning")

		// Add location
		location := sarif.NewPhysicalLocation().
			WithArtifactLocation(sarif.NewSimpleArtifactLocation(relPath)).
			WithRegion(sarif.NewSimpleRegion(diag.Line, diag.Column))

		result.AddLocation(sarif.NewLocation().WithPhysicalLocation(location))
	}

	report.AddRun(run)

	// Output SARIF JSON
	sarifJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling SARIF: %w", err)
	}

	fmt.Println(string(sarifJSON))
	return nil
}

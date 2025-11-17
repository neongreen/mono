package main

import (
	"encoding/json"
	"fmt"
	"go/token"
	"os"
	"path/filepath"

	"github.com/neongreen/mono/linters/cobralint"
	"github.com/owenrumney/go-sarif/v2/sarif"
	"github.com/spf13/cobra"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
)

var (
	formatFlag string
)

var rootCmd = &cobra.Command{
	Use:   "cobralint [flags] [packages]",
	Short: "Linter for Cobra commands",
	Long:  "Enforces conventions for Cobra commands, such as requiring a --json flag for machine-readable output.",
	Args:  cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		patterns := args
		if len(patterns) == 0 {
			patterns = []string{"."}
		}

		switch formatFlag {
		case "text":
			return runWithTextOutput(patterns)
		case "sarif":
			return runWithSARIF(patterns)
		default:
			return fmt.Errorf("unknown format: %s (supported: text, sarif)", formatFlag)
		}
	},
}

func init() {
	rootCmd.Flags().StringVar(&formatFlag, "format", "text", "output format: text or sarif")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runWithTextOutput(patterns []string) error {
	// Load packages
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return fmt.Errorf("loading packages: %w", err)
	}

	// Check for package errors
	var hasErrors bool
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			hasErrors = true
		}
	})
	if hasErrors {
		return fmt.Errorf("packages had errors")
	}

	// Run analyzer and print text output
	var foundIssues bool
	for _, pkg := range pkgs {
		pass := createPass(pkg, func(d analysis.Diagnostic) {
			foundIssues = true
			position := pkg.Fset.Position(d.Pos)
			fmt.Fprintf(os.Stderr, "%s: %s\n", position, d.Message)
		})

		if err := runAnalyzer(pass); err != nil {
			return err
		}
	}

	if foundIssues {
		os.Exit(3) // Exit with code 3 to indicate issues found (like go vet)
	}
	return nil
}

func runWithSARIF(patterns []string) error {
	// Load packages
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return fmt.Errorf("loading packages: %w", err)
	}

	// Check for package errors
	var hasErrors bool
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			hasErrors = true
		}
	})
	if hasErrors {
		return fmt.Errorf("packages had errors")
	}

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

	// Run analyzer and collect diagnostics
	var allDiags []analysisResult
	for _, pkg := range pkgs {
		pass := createPass(pkg, func(d analysis.Diagnostic) {
			allDiags = append(allDiags, analysisResult{
				fset: pkg.Fset,
				diag: d,
			})
		})

		if err := runAnalyzer(pass); err != nil {
			return err
		}
	}

	// Convert diagnostics to SARIF results
	for _, result := range allDiags {
		position := result.fset.Position(result.diag.Pos)

		// Convert to relative path from working directory
		relPath, err := filepath.Rel(".", position.Filename)
		if err != nil {
			relPath = position.Filename
		}

		sarifResult := run.CreateResultForRule("require-json-flag").
			WithMessage(sarif.NewTextMessage(result.diag.Message)).
			WithLevel("warning")

		// Add location
		location := sarif.NewPhysicalLocation().
			WithArtifactLocation(sarif.NewSimpleArtifactLocation(relPath)).
			WithRegion(sarif.NewRegion().
				WithStartLine(position.Line).
				WithStartColumn(position.Column))

		sarifResult.AddLocation(sarif.NewLocation().WithPhysicalLocation(location))
	}

	report.AddRun(run)

	// Output SARIF JSON
	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling SARIF: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

// createPass creates an analysis.Pass for the given package
func createPass(pkg *packages.Package, reportFunc func(analysis.Diagnostic)) *analysis.Pass {
	return &analysis.Pass{
		Analyzer:   cobralint.Analyzer,
		Fset:       pkg.Fset,
		Files:      pkg.Syntax,
		OtherFiles: pkg.OtherFiles,
		Pkg:        pkg.Types,
		TypesInfo:  pkg.TypesInfo,
		ResultOf:   make(map[*analysis.Analyzer]any),
		Report:     reportFunc,
	}
}

// runAnalyzer runs the cobralint analyzer with its required dependencies
func runAnalyzer(pass *analysis.Pass) error {
	// Populate ResultOf with required analyzers
	for _, req := range cobralint.Analyzer.Requires {
		reqPass := &analysis.Pass{
			Analyzer:   req,
			Fset:       pass.Fset,
			Files:      pass.Files,
			OtherFiles: pass.OtherFiles,
			Pkg:        pass.Pkg,
			TypesInfo:  pass.TypesInfo,
			ResultOf:   make(map[*analysis.Analyzer]any),
			Report:     func(d analysis.Diagnostic) {}, // Ignore reports from required analyzers
		}
		result, err := req.Run(reqPass)
		if err != nil {
			return fmt.Errorf("running required analyzer %s: %w", req.Name, err)
		}
		pass.ResultOf[req] = result
	}

	// Run the analyzer
	_, err := cobralint.Analyzer.Run(pass)
	if err != nil {
		return fmt.Errorf("running analyzer: %w", err)
	}

	return nil
}

type analysisResult struct {
	fset *token.FileSet
	diag analysis.Diagnostic
}

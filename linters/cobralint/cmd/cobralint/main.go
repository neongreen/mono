package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/token"
	"os"
	"path/filepath"

	"github.com/neongreen/mono/linters/cobralint"
	"github.com/owenrumney/go-sarif/v2/sarif"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
	"golang.org/x/tools/go/packages"
)

var (
	formatFlag = flag.String("format", "text", "output format: text or sarif")
)

func main() {
	// Parse our custom flags first
	flag.Parse()

	// If format is text, use the default singlechecker behavior
	if *formatFlag == "text" {
		singlechecker.Main(cobralint.Analyzer)
		return
	}

	// For SARIF output, run analyzer manually
	if *formatFlag == "sarif" {
		if err := runWithSARIF(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Fprintf(os.Stderr, "unknown format: %s (supported: text, sarif)\n", *formatFlag)
	os.Exit(1)
}

func runWithSARIF() error {
	// Load packages to analyze
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
	}

	patterns := flag.Args()
	if len(patterns) == 0 {
		patterns = []string{"."}
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

	// Run analyzer on each package and collect diagnostics
	var allDiags []analysisResult
	for _, pkg := range pkgs {
		pass := &analysis.Pass{
			Analyzer:   cobralint.Analyzer,
			Fset:       pkg.Fset,
			Files:      pkg.Syntax,
			OtherFiles: pkg.OtherFiles,
			Pkg:        pkg.Types,
			TypesInfo:  pkg.TypesInfo,
			ResultOf:   make(map[*analysis.Analyzer]interface{}),
			Report: func(d analysis.Diagnostic) {
				allDiags = append(allDiags, analysisResult{
					fset: pkg.Fset,
					diag: d,
				})
			},
		}

		// Populate ResultOf with required analyzers
		for _, req := range cobralint.Analyzer.Requires {
			reqPass := &analysis.Pass{
				Analyzer:   req,
				Fset:       pkg.Fset,
				Files:      pkg.Syntax,
				OtherFiles: pkg.OtherFiles,
				Pkg:        pkg.Types,
				TypesInfo:  pkg.TypesInfo,
				ResultOf:   make(map[*analysis.Analyzer]interface{}),
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

type analysisResult struct {
	fset *token.FileSet
	diag analysis.Diagnostic
}

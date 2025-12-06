package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/neongreen/mono/claim/internal/index"
	"github.com/neongreen/mono/claim/internal/lint"
	"github.com/neongreen/mono/claim/internal/scan"
)

var (
	lintRoot        string
	lintLineNumbers bool
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint claims for common issues",
	Long: `Checks claims for common issues like:
- References to line numbers (fragile, break when code changes)

By default, all lint rules are enabled. Use flags to disable specific rules.`,
	RunE:         runLint,
	SilenceUsage: true,
}

func init() {
	lintCmd.Flags().StringVar(&lintRoot, "root", ".", "Root directory to scan")
	lintCmd.Flags().BoolVar(&lintLineNumbers, "line-numbers", true, "Warn about line number references")
	RootCmd.AddCommand(lintCmd)
}

func runLint(cmd *cobra.Command, args []string) error {
	// Scan files
	files, err := scan.ScanFiles(lintRoot)
	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	// Build index
	idx, err := index.Build(files)
	if err != nil {
		return fmt.Errorf("failed to build index: %w", err)
	}

	// Configure lint options
	opts := lint.Options{
		WarnLineNumbers: lintLineNumbers,
	}

	// Lint all claims
	var allIssues []lint.Issue
	for _, claim := range idx.Claims {
		issues := lint.LintClaim(claim, opts)
		allIssues = append(allIssues, issues...)
	}

	// Print results
	if len(allIssues) == 0 {
		fmt.Println("No lint issues found.")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d lint issue(s):\n\n", len(allIssues))
	for _, issue := range allIssues {
		fmt.Fprintf(os.Stderr, "[%s] %s: %s\n",
			issue.Severity,
			issue.ClaimID,
			issue.Rule,
		)
		fmt.Fprintf(os.Stderr, "  %s\n\n", issue.Message)
	}

	return fmt.Errorf("lint found %d issue(s)", len(allIssues))
}
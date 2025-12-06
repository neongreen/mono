package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/neongreen/mono/claim/internal/check"
	"github.com/neongreen/mono/claim/internal/index"
	"github.com/neongreen/mono/claim/internal/runner"
	"github.com/neongreen/mono/claim/internal/scan"
)

var (
	checkRoot        string
	checkLensFile    string
	checkMaxRefDepth int
	checkDebugPrompt bool
	checkLLM         string
	checkAll         bool
)

var checkCmd = &cobra.Command{
	Use:   "check --claim <id> OR check --all",
	Short: "Check a specific claim or all claims using Claude",
	Long: `Checks whether a claim is properly proven by its bullets.
Uses Claude with structured output to verify the claim.

Use --claim to check a single claim, or --all to check all claims in the directory.`,
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().StringVar(&checkRoot, "root", ".", "Root directory to scan")
	checkCmd.Flags().StringVar(&checkLensFile, "lens-file", "", "Additional lens file to load")
	checkCmd.Flags().IntVar(&checkMaxRefDepth, "max-ref-depth", 3, "Maximum depth for referenced claims")
	checkCmd.Flags().BoolVar(&checkDebugPrompt, "debug-prompt", false, "Print the prompt sent to Claude")
	checkCmd.Flags().StringVar(&checkLLM, "llm", "claude", "LLM to use (claude or codex)")
	checkCmd.Flags().String("claim", "", "Claim ID to check")
	checkCmd.Flags().BoolVar(&checkAll, "all", false, "Check all claims")
	RootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	claimID, _ := cmd.Flags().GetString("claim")

	// Validate flags
	if !checkAll && claimID == "" {
		return fmt.Errorf("either --claim or --all must be specified")
	}
	if checkAll && claimID != "" {
		return fmt.Errorf("cannot specify both --claim and --all")
	}

	// Scan files
	files, err := scan.ScanFiles(checkRoot)
	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	// Build index
	idx, err := index.Build(files)
	if err != nil {
		return fmt.Errorf("failed to build index: %w", err)
	}

	// Load lenses
	lenses := idx.Lenses
	if checkLensFile != "" {
		extraLenses, err := scan.LoadLensFile(checkLensFile)
		if err != nil {
			return fmt.Errorf("failed to load lens file: %w", err)
		}
		for name, content := range extraLenses {
			lenses[name] = content
		}
	}

	// Create runner based on LLM choice
	var r runner.Runner
	switch checkLLM {
	case "claude":
		command := os.Getenv("CLAIM_CLAUDE_CMD")
		if command == "" {
			command = "claude"
		}
		r = &runner.ClaudeRunner{
			Command: command,
			Verbose: debugFlag,
		}
	case "codex":
		command := os.Getenv("CLAIM_CODEX_CMD")
		if command == "" {
			command = "codex"
		}
		r = &runner.CodexRunner{
			Command: command,
			Verbose: debugFlag,
		}
	default:
		return fmt.Errorf("unknown LLM: %s (must be 'claude' or 'codex')", checkLLM)
	}

	// Get list of claims to check
	var claimIDs []string
	if checkAll {
		for id := range idx.Claims {
			claimIDs = append(claimIDs, id)
		}
		fmt.Fprintf(os.Stderr, "Checking %d claims...\n\n", len(claimIDs))
	} else {
		claimIDs = []string{claimID}
	}

	ctx := context.Background()
	hasFailures := false
	var lastResult *runner.ClaimResult

	// Check each claim
	for i, id := range claimIDs {
		if checkAll && i > 0 {
			fmt.Println("\n" + strings.Repeat("=", 80) + "\n")
		}

		// Run check
		result, err := check.Check(ctx, idx, lenses, id, checkMaxRefDepth, checkDebugPrompt, false, r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking claim %s: %v\n", id, err)
			hasFailures = true
			continue
		}

		// Get the claim for display
		claim, _ := idx.GetClaim(id)

		// Print report
		check.PrintReport(os.Stdout, claim, result)

		// Track failures for --all
		if result.Result != "proven" {
			hasFailures = true
		}
		
		// Remember last result for single-claim mode
		lastResult = result
	}

	// Exit based on results
	if checkAll {
		if hasFailures {
			return fmt.Errorf("some claims need improvement")
		}
		return nil
	}

	// Single claim exit logic
	//
	// @claim[exit-codes-match-semantics]: Exit codes reflect whether the claim is proven or not
	// - Exit code 0 means success: the claim was proven
	// - Exit code 1 means failure: the claim is unproven or sorry
	//   - Unproven means the bullets don't prove the claim
	//   - Sorry means the claim contains @sorry bullets (accepted without proof)
	// - Exit code 1 also covers tool errors like unexpected result values
	// - This matches standard Unix convention: 0 = success, non-zero = failure
	switch lastResult.Result {
	case "proven":
		// Exit 0 for proven (success)
		return nil
	case "unproven", "sorry":
		// Exit 1 for unproven/sorry (needs work)
		return fmt.Errorf("claim %s is %s - needs improvement", claimID, lastResult.Result)
	default:
		return fmt.Errorf("unexpected result: %s", lastResult.Result)
	}
}

package cmd

import (
	"context"
	"fmt"
	"os"

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
)

var checkCmd = &cobra.Command{
	Use:   "check --claim <id>",
	Short: "Check a specific claim using Claude",
	Long: `Checks whether a claim is properly proven by its bullets.
Uses Claude with structured output to verify the claim.`,
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().StringVar(&checkRoot, "root", ".", "Root directory to scan")
	checkCmd.Flags().StringVar(&checkLensFile, "lens-file", "", "Additional lens file to load")
	checkCmd.Flags().IntVar(&checkMaxRefDepth, "max-ref-depth", 3, "Maximum depth for referenced claims")
	checkCmd.Flags().BoolVar(&checkDebugPrompt, "debug-prompt", false, "Print the prompt sent to Claude")
	checkCmd.Flags().StringVar(&checkLLM, "llm", "claude", "LLM to use (claude or codex)")
	checkCmd.Flags().String("claim", "", "Claim ID to check (required)")
	checkCmd.MarkFlagRequired("claim")
	RootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	claimID, _ := cmd.Flags().GetString("claim")

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

	// Run check
	ctx := context.Background()
	// Don't anonymize in check command (user wants to see real filenames)
	result, err := check.Check(ctx, idx, lenses, claimID, checkMaxRefDepth, checkDebugPrompt, false, r)
	if err != nil {
		return err
	}

	// Print report
	check.PrintReport(os.Stdout, result)

	// Exit based on result
	switch result.Result {
	case "proven":
		// Exit 2 for proven (indicates test should fail)
		return fmt.Errorf("claim %s was proven (expected unproven)", claimID)
	case "unproven", "sorry":
		// Exit 0 for expected results
		return nil
	default:
		return fmt.Errorf("unexpected result: %s", result.Result)
	}
}

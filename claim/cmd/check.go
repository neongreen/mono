package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/neongreen/mono/claim/internal/check"
	claimcontext "github.com/neongreen/mono/claim/internal/context"
	"github.com/neongreen/mono/claim/internal/index"
	"github.com/neongreen/mono/claim/internal/runner"
	"github.com/neongreen/mono/claim/internal/scan"
)

var (
	checkRoot         string
	checkLensFile     string
	checkDebugPrompt  bool
	checkLLM          string
	checkAll          bool
	checkContextModel string
)

var checkCmd = &cobra.Command{
	Use:   "check <claim-id> OR check --all",
	Short: "Check a specific claim or all claims using Claude",
	Long: `Checks whether a claim is properly proven by its @proof block.
Uses Claude with structured output to verify the claim.

Examples:
  claim check my-claim-id    # Check a specific claim
  claim check --all          # Check all claims`,
	Args:         cobra.MaximumNArgs(1),
	RunE:         runCheck,
	SilenceUsage: true,
}

func init() {
	checkCmd.Flags().StringVar(&checkRoot, "root", ".", "Root directory to scan")
	checkCmd.Flags().StringVar(&checkLensFile, "lens-file", "", "Additional lens file to load")
	checkCmd.Flags().BoolVar(&checkDebugPrompt, "debug-prompt", false, "Print the prompt sent to Claude")
	checkCmd.Flags().StringVar(&checkLLM, "llm", "claude", "LLM to use (claude or codex)")
	checkCmd.Flags().BoolVar(&checkAll, "all", false, "Check all claims")
	checkCmd.Flags().StringVar(&checkContextModel, "context-model", "haiku", "Model for @context resolution (default: haiku)")
	RootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	var claimID string

	// Handle positional argument
	if len(args) > 0 {
		claimID = args[0]
	}

	// Validate flags
	if !checkAll && claimID == "" {
		return fmt.Errorf("specify a claim ID or use --all")
	}
	if checkAll && claimID != "" {
		return fmt.Errorf("cannot specify both a claim ID and --all")
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
			WorkDir: checkRoot,
		}
	case "codex":
		command := os.Getenv("CLAIM_CODEX_CMD")
		if command == "" {
			command = "codex"
		}
		r = &runner.CodexRunner{
			Command: command,
			Verbose: debugFlag,
			WorkDir: checkRoot,
		}
	default:
		return fmt.Errorf("unknown LLM: %s (must be 'claude' or 'codex')", checkLLM)
	}

	ctx := context.Background()

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

	hasFailures := false
	var lastResult string

	// Create artifacts directory
	timestamp := time.Now().Format("20060102-150405")
	artifactsDir := filepath.Join("/tmp", "claim-artifacts", timestamp)
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return fmt.Errorf("failed to create artifacts dir: %w", err)
	}

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	fmt.Printf("%s %s\n\n", labelStyle.Render("Artifacts:"), artifactsDir)

	// Check each claim
	for i, id := range claimIDs {
		if checkAll && i > 0 {
			fmt.Println("\n" + strings.Repeat("=", 80) + "\n")
		}

		// Create context resolver
		contextResolver := &claimcontext.Resolver{
			Command: "claude",
			Model:   checkContextModel,
			WorkDir: checkRoot,
		}

		opts := check.CheckProofOptions{
			DebugPrompt:     checkDebugPrompt,
			ProgressWriter:  os.Stdout,
			ContextResolver: contextResolver,
			ArtifactsDir:    artifactsDir,
		}

		result, err := check.CheckProof(ctx, idx, lenses, id, r, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n%s %s\n",
				lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("Error:"),
				err.Error())
			if !checkAll {
				os.Exit(1)
			}
			hasFailures = true
			continue
		}

		if result.Result != "proven" {
			hasFailures = true
		}
		lastResult = result.Result
	}

	// Exit based on results
	if checkAll {
		if hasFailures {
			return fmt.Errorf("some claims failed verification")
		}
		return nil
	}

	// Single claim exit
	if lastResult == "unproven" {
		os.Exit(1)
	}
	return nil
}

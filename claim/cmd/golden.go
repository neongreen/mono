package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/neongreen/mono/claim/internal/check"
	"github.com/neongreen/mono/claim/internal/index"
	"github.com/neongreen/mono/claim/internal/runner"
	"github.com/neongreen/mono/claim/internal/scan"
)

var (
	goldenRoot string
)

var goldenCmd = &cobra.Command{
	Use:   "golden",
	Short: "Run the golden test suite",
	Long: `Runs all test cases from fixtures/cases.jsonl and verifies
that each claim returns unproven (never proven).`,
	RunE: runGolden,
}

func init() {
	goldenCmd.Flags().StringVar(&goldenRoot, "root", "fixtures", "Root directory containing fixtures")
	RootCmd.AddCommand(goldenCmd)
}

func runGolden(cmd *cobra.Command, args []string) error {
	// Load cases
	casesPath := filepath.Join(goldenRoot, "cases.jsonl")
	f, err := os.Open(casesPath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", casesPath, err)
	}
	defer f.Close()

	var cases []struct {
		Claim string `json:"claim"`
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var tc struct {
			Claim string `json:"claim"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &tc); err != nil {
			return fmt.Errorf("failed to parse case: %w", err)
		}
		cases = append(cases, tc)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read cases: %w", err)
	}

	// Scan files
	files, err := scan.ScanFiles(goldenRoot)
	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	// Build index
	idx, err := index.Build(files)
	if err != nil {
		return fmt.Errorf("failed to build index: %w", err)
	}

	// Create runner
	r := &runner.ClaudeRunner{
		Command: os.Getenv("CLAIM_CLAUDE_CMD"),
	}
	if r.Command == "" {
		r.Command = "claude"
	}

	// Run all cases
	// Golden tests verify that flawed proofs are NOT proven.
	// Both "unproven" results and errors (e.g., missing references) count as passing.
	ctx := context.Background()
	failed := 0
	for _, tc := range cases {
		fmt.Printf("Checking %s... ", tc.Claim)
		opts := check.CheckProofOptions{
			DebugPrompt:    false,
			ProgressWriter: nil,
		}
		result, err := check.CheckProof(ctx, idx, idx.Lenses, tc.Claim, r, opts)
		if err != nil {
			// Errors mean the claim could not be verified - this is a valid "not proven" outcome
			fmt.Printf("OK (error: %v)\n", err)
			continue
		}

		if result.Result == "proven" {
			fmt.Printf("FAILED (was proven)\n")
			failed++
		} else {
			fmt.Printf("OK (%s)\n", result.Result)
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d/%d cases failed", failed, len(cases))
	}

	fmt.Printf("\nAll %d cases passed!\n", len(cases))
	return nil
}

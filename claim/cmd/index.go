package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/neongreen/mono/claim/internal/index"
	"github.com/neongreen/mono/claim/internal/scan"
)

var (
	indexJSON bool
)

var indexCmd = &cobra.Command{
	Use:   "index [root]",
	Short: "Scan and index claims in the codebase",
	Long: `Scans repository files for @claim and @lens blocks and builds an index.
Reports claim counts, duplicate IDs, and optionally outputs JSON.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runIndex,
}

func init() {
	indexCmd.Flags().BoolVar(&indexJSON, "json", false, "Output index as JSON")
	RootCmd.AddCommand(indexCmd)
}

func runIndex(cmd *cobra.Command, args []string) error {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	// Scan files
	files, err := scan.ScanFiles(root)
	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	// Build index
	idx, err := index.Build(files)
	if err != nil {
		return fmt.Errorf("failed to build index: %w", err)
	}

	// Check for duplicates
	duplicates := idx.FindDuplicates()
	if len(duplicates) > 0 {
		for id, locations := range duplicates {
			fmt.Fprintf(os.Stderr, "Error: duplicate claim ID %q found in:\n", id)
			for _, loc := range locations {
				fmt.Fprintf(os.Stderr, "  %s:%d\n", loc.File, loc.Line)
			}
		}
		return fmt.Errorf("found %d duplicate claim IDs", len(duplicates))
	}

	if indexJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(idx)
	}

	// Print summary
	fmt.Printf("Found %d claims and %d lenses\n", len(idx.Claims), len(idx.Lenses))
	return nil
}

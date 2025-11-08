package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/import/beads"
	"github.com/spf13/cobra"
)

var importBeadsCmd = &cobra.Command{
	Use:   "import-beads [path]",
	Short: "Import issues from beads JSONL format",
	Long: `Import issues from a beads .beads/issues.jsonl file into tk.

AUTO-DETECTS PREFIXES:
This command scans the beads file and creates one tk project per prefix.
For example, if beads has mono-1, mono-2, foo-1, it creates two projects.

ALIAS NAMING:
Projects get alias: <--prefix><beads-prefix>
Default --prefix is "bd-", so:
  - beads "mono" → tk alias "bd-mono"
  - beads "foo" → tk alias "bd-foo"

LIMITATION - ALIAS UNIQUENESS:
tk's multi-machine sync requires that one node cannot create duplicate aliases.
If you already have a project with alias "bd-mono", the import will fail.
Use --prefix to avoid clashes (e.g., --prefix=beads-).

CLASH DETECTION:
Before import, checks if any resulting aliases already exist from this node.
If clash detected, import aborts with clear error message.

PRESERVES:
- Exact task numbering (mono-123 in beads → bd-mono-123 in tk)
- Titles and descriptions
- Status (open, in_progress, closed)
- Priority (0-4) as metadata
- Labels as metadata
- Dependencies (blocks, parent-child, related, discovered-from)

Examples:
  tk import-beads                                    # Import from .beads/issues.jsonl (prefix: bd-)
  tk import-beads /path/to/repo                      # Auto-finds .beads/issues.jsonl
  tk import-beads .beads/issues.jsonl --prefix=old-  # Use different prefix
  tk import-beads --dry-run                          # Preview what would be imported
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		aliasPrefix, _ := cmd.Flags().GetString("prefix")

		// Determine the path to the beads file
		beadsPath, err := resolveBeadsPath(args)
		if err != nil {
			return err
		}

		// Check if file exists
		if _, err := os.Stat(beadsPath); os.IsNotExist(err) {
			return fmt.Errorf("beads file not found: %s", beadsPath)
		}

		// Open database
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Prepare import options
		opts := beads.ImportOptions{
			BeadsPath:   beadsPath,
			AliasPrefix: aliasPrefix,
			DryRun:      dryRun,
		}

		// If dry run, show preview
		if dryRun {
			return showDryRunPreview(db, opts)
		}

		// Perform import
		result, err := beads.Import(db, opts)
		if err != nil {
			return err
		}

		// Display results
		displayImportResults(result)
		return nil
	},
}

// resolveBeadsPath determines the path to the beads file from arguments
func resolveBeadsPath(args []string) (string, error) {
	if len(args) == 0 {
		// Default to current directory
		return ".beads/issues.jsonl", nil
	}

	arg := args[0]
	// Check if it's a file or directory
	info, err := os.Stat(arg)
	if err != nil {
		return "", fmt.Errorf("path not found: %w", err)
	}

	if info.IsDir() {
		return filepath.Join(arg, ".beads", "issues.jsonl"), nil
	}

	return arg, nil
}

// showDryRunPreview shows a preview of what would be imported
func showDryRunPreview(db *database.DB, opts beads.ImportOptions) error {
	// Read issues
	issues, err := beads.ReadBeadsFile(opts.BeadsPath)
	if err != nil {
		return fmt.Errorf("failed to read beads file: %w", err)
	}

	if len(issues) == 0 {
		fmt.Println("No issues found in beads file")
		return nil
	}

	// Group by prefix
	prefixGroups := beads.ExtractPrefixesFromBeads(issues)

	fmt.Println("\nDry run mode - no changes will be made")
	fmt.Printf("\nFound %d issues across %d prefix(es):\n", len(issues), len(prefixGroups))

	for prefix, group := range prefixGroups {
		alias := opts.AliasPrefix + prefix
		fmt.Printf("  %s: %d issues → will create project with alias '%s'\n", prefix, len(group), alias)
	}

	fmt.Println("\nAll issues:")
	for _, issue := range issues {
		fmt.Printf("  %s: %s (status: %s, type: %s)\n",
			issue.ID, issue.Title, issue.Status, issue.Type)
	}

	return nil
}

// displayImportResults displays the results of the import operation
func displayImportResults(result *beads.ImportResult) {
	// Show renumbering summary
	if len(result.RenumberedIssues) > 0 {
		fmt.Printf("\nRenumbered %d issues with non-numeric IDs:\n", len(result.RenumberedIssues))
		for _, msg := range result.RenumberedIssues {
			fmt.Printf("  %s\n", msg)
		}
	}

	fmt.Printf("\nImported %d issues (%d skipped)\n", result.TotalImported, result.TotalSkipped)

	if result.RelationsImported > 0 {
		fmt.Printf("Imported %d relationships\n", result.RelationsImported)
	}

	// Show created projects
	if len(result.ProjectsCreated) > 0 {
		fmt.Println("\nCreated projects:")
		for prefix, projectUID := range result.ProjectsCreated {
			fmt.Printf("  %s (UID: %s)\n", prefix, projectUID)
		}
	}
}

func init() {
	importBeadsCmd.Flags().Bool("dry-run", false, "Preview import without making changes")
	importBeadsCmd.Flags().StringP("prefix", "p", "bd-", "Prefix to prepend to beads prefixes for tk aliases (e.g., 'bd-' creates 'bd-mono' from 'mono')")
}

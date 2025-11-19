package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/sanitycheck"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Read-only diagnostic command; JSON output is for scripting
var debugSanitycheckCmd = &cobra.Command{
	Use:   "debug-sanitycheck",
	Short: "Compare reducer state with database projections",
	Long: `Compares the state built from events (using the reducer) with the current database projection tables.

This is a diagnostic tool to detect inconsistencies between the event log and projection tables.
If differences are found, a detailed report is written to ~/.tk/state-diff.json.

The sanity check is also run automatically before every tk command (silently).
This command allows you to run it explicitly and see detailed output.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		comparison, err := sanitycheck.CompareState(db, cfg)
		if err != nil {
			return fmt.Errorf("failed to run sanity check: %w", err)
		}

		if jsonOutput {
			output, err := json.MarshalIndent(comparison, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal comparison: %w", err)
			}
			fmt.Println(string(output))
		} else {
			sanitycheck.PrintComparison(comparison)
		}

		if len(comparison.Differences) > 0 {
			return fmt.Errorf("found %d difference(s) between reducer and database", len(comparison.Differences))
		}

		return nil
	},
}

func init() {
	debugSanitycheckCmd.Flags().Bool("json", false, "Output as JSON")
}

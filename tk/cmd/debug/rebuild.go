package debug

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
// RebuildCmd is the rebuild subcommand for debug
var RebuildCmd = &cobra.Command{
	Use:   "rebuild-projections",
	Short: "Rebuild all projection tables from events",
	Long: `Rebuilds all projection tables (tasks, task_numbers, projects, project_aliases) from events.

This clears all projection tables and replays all events in Lamport timestamp order,
ensuring deterministic projection regardless of the order events were originally ingested.

Use this when:
  - Projection tables are inconsistent after sync
  - After fixing projection bugs
  - To verify projection determinism

WARNING: This will temporarily clear all projection tables. The operation is atomic
(uses transactions), but avoid running other tk commands simultaneously.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		verify, _ := cmd.Flags().GetBool("verify")

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		if dryRun {
			fmt.Println("Dry run - would rebuild projections from events")

			// Count events
			events, err := db.GetEvents()
			if err != nil {
				return fmt.Errorf("failed to get events: %w", err)
			}

			fmt.Printf("Would replay %d events in Lamport timestamp order\n", len(events))
			fmt.Println("\nProjection tables that would be cleared:")
			fmt.Println("  - task_numbers")
			fmt.Println("  - tasks")
			fmt.Println("  - project_aliases")
			fmt.Println("  - projects")

			return nil
		}

		if verify {
			fmt.Println("Verifying projection determinism...")
			if err := db.VerifyProjectionDeterminism(); err != nil {
				return fmt.Errorf("verification failed: %w", err)
			}
			fmt.Println("✓ Projections are deterministic")
			return nil
		}

		fmt.Println("Rebuilding projections from events...")

		events, err := db.GetEvents()
		if err != nil {
			return fmt.Errorf("failed to get events: %w", err)
		}

		fmt.Printf("Replaying %d events in Lamport timestamp order...\n", len(events))

		if err := db.RebuildProjections(); err != nil {
			return fmt.Errorf("failed to rebuild projections: %w", err)
		}

		fmt.Println("✓ Projections rebuilt successfully")
		fmt.Println("\nIt's recommended to run 'tk debug doctor' to verify database health")

		return nil
	},
}

func init() {
	RebuildCmd.Flags().Bool("dry-run", false, "Show what would be done without making changes")
	RebuildCmd.Flags().Bool("verify", false, "Verify projection determinism without rebuilding")
}

package migrate

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/deprecation"
	"github.com/spf13/cobra"
)

var ScanDeprecatedCmd = &cobra.Command{
	Use:   "scan-deprecated",
	Short: "Scan event log for deprecated field usage",
	Long: `Scans all events in the database and reports which deprecated fields are still in use.

This helps determine when it's safe to remove deprecated code after a migration.

Example:
  tk migrate scan-deprecated

Shows:
  - Which deprecated fields are present in events
  - How many events use each field
  - When the field was last seen

When all counts reach 0, it's safe to remove the deprecated code.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Get all events
		fmt.Println("Scanning events for deprecated field usage...")
		events, err := db.GetEvents()
		if err != nil {
			return fmt.Errorf("failed to get events: %w", err)
		}

		fmt.Printf("Found %d events to scan\n\n", len(events))

		// Scan all events
		tracker, err := deprecation.ScanAllEvents(events)
		if err != nil {
			return fmt.Errorf("failed to scan events: %w", err)
		}

		// Print summary
		fmt.Print(tracker.PrintSummary())

		// Show next steps
		stats := tracker.GetStats()
		if len(stats) > 0 {
			fmt.Println("\nNext steps:")
			fmt.Println("  1. Run 'tk migrate v4-to-v5' to rewrite events (not implemented yet)")
			fmt.Println("  2. Run this scan again to verify fields are removed")
			fmt.Println("  3. When count reaches 0, remove code listed by 'mise run deprecations'")
		} else {
			fmt.Println("\nAll deprecated fields have been removed from events!")
			fmt.Println("Run 'mise run deprecations' to see code that can be deleted.")
		}

		return nil
	},
}

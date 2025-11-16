package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var debugEventsShowCmd = &cobra.Command{
	Use:   "debug-events-show",
	Short: "Show detailed information about a specific event",
	Long: `Show detailed information about a specific event including its full payload.

This is a debug command to help understand what data is stored in an event.

Examples:
  tk events show ev-1-abc123
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eventID := args[0]
		jsonOutput, _ := cmd.Flags().GetBool("json")

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		events, err := db.GetEvents()
		if err != nil {
			return err
		}

		var found *types.Event
		for i, e := range events {
			if e.ID == eventID {
				found = &events[i]
				break
			}
		}

		if found == nil {
			return fmt.Errorf("event not found: %s", eventID)
		}

		// Always output JSON for now (backward compatibility)
		_ = jsonOutput
		output, err := json.MarshalIndent(found, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		fmt.Println(string(output))
		return nil
	},
}

func init() {
	debugEventsShowCmd.Flags().Bool("json", false, "Output as JSON")
}

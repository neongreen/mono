package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var debugEventsLsCmd = &cobra.Command{
	Use:   "debug-events-ls",
	Short: "List all events in the database",
	Long: `List all events in the database with their basic information.

This is a debug command to help understand what events exist in the database.

Examples:
  tk events list
  tk events list --limit 10
  tk events list --kind prefix.created
  tk events list --json
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		kindFilter, _ := cmd.Flags().GetString("kind")
		verbose, _ := cmd.Flags().GetBool("verbose")
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

		// Filter by kind if specified
		if kindFilter != "" {
			var filtered []types.Event
			for _, e := range events {
				if e.Kind == kindFilter {
					filtered = append(filtered, e)
				}
			}
			events = filtered
		}

		// Apply limit if specified
		if limit > 0 && len(events) > limit {
			events = events[:limit]
		}

		if len(events) == 0 {
			if jsonOutput {
				fmt.Println("[]")
			} else {
				fmt.Println("No events found")
			}
			return nil
		}

		// JSON output mode
		if jsonOutput {
			output, err := json.MarshalIndent(events, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal events: %w", err)
			}
			fmt.Println(string(output))
			return nil
		}

		// Create table
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		if verbose {
			t.AppendHeader(table.Row{"ID", "TS", "Kind", "Actor", "Role", "Created At", "Payload"})
		} else {
			t.AppendHeader(table.Row{"ID", "TS", "Kind", "Actor", "Created At"})
		}
		t.SetStyle(table.StyleLight)
		t.Style().Options.SeparateRows = false
		t.Style().Options.DrawBorder = false

		for _, e := range events {
			createdAtStr := e.CreatedAt.Format("2006-01-02 15:04:05")
			if verbose {
				// Pretty print payload JSON
				var payloadMap map[string]any
				if err := json.Unmarshal(e.Payload, &payloadMap); err == nil {
					payloadJSON, _ := json.MarshalIndent(payloadMap, "", "  ")
					t.AppendRow(table.Row{e.ID, e.TS, e.Kind, e.Actor, e.Role, createdAtStr, string(payloadJSON)})
				} else {
					t.AppendRow(table.Row{e.ID, e.TS, e.Kind, e.Actor, e.Role, createdAtStr, string(e.Payload)})
				}
			} else {
				t.AppendRow(table.Row{e.ID, e.TS, e.Kind, e.Actor, createdAtStr})
			}
		}

		t.Render()
		fmt.Printf("\nTotal events: %d\n", len(events))
		return nil
	},
}

func init() {
	debugEventsLsCmd.Flags().Int("limit", 0, "Limit the number of events to show")
	debugEventsLsCmd.Flags().String("kind", "", "Filter events by kind")
	debugEventsLsCmd.Flags().Bool("verbose", false, "Show full event details including payload")
	debugEventsLsCmd.Flags().Bool("json", false, "Output events as JSON")
}

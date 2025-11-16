package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

var debugEventsStatsCmd = &cobra.Command{
	Use:   "debug-events-stats",
	Short: "Show statistics about events in the database",
	Long: `Show statistics about events in the database, grouped by kind.

This is a debug command to help understand the distribution of events.

Examples:
  tk events stats
`,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		if len(events) == 0 {
			if jsonOutput {
				output := map[string]any{
					"total":    0,
					"by_kind":  map[string]int{},
					"by_actor": map[string]int{},
					"by_role":  map[string]int{},
				}
				data, _ := json.MarshalIndent(output, "", "  ")
				fmt.Println(string(data))
			} else {
				fmt.Println("No events found")
			}
			return nil
		}

		// Count events by kind
		kindCounts := make(map[string]int)
		actorCounts := make(map[string]int)
		roleCounts := make(map[string]int)

		for _, e := range events {
			kindCounts[e.Kind]++
			actorCounts[e.Actor]++
			roleCounts[e.Role]++
		}

		if jsonOutput {
			output := map[string]any{
				"total":    len(events),
				"by_kind":  kindCounts,
				"by_actor": actorCounts,
				"by_role":  roleCounts,
			}
			data, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal stats: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		// Display statistics
		fmt.Printf("Total events: %d\n\n", len(events))

		// Events by kind
		fmt.Println("Events by kind:")
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Kind", "Count"})
		t.SetStyle(table.StyleLight)
		t.Style().Options.SeparateRows = false
		t.Style().Options.DrawBorder = false

		for kind, count := range kindCounts {
			t.AppendRow(table.Row{kind, count})
		}
		t.Render()

		// Events by actor
		fmt.Println("\nEvents by actor:")
		t2 := table.NewWriter()
		t2.SetOutputMirror(os.Stdout)
		t2.AppendHeader(table.Row{"Actor", "Count"})
		t2.SetStyle(table.StyleLight)
		t2.Style().Options.SeparateRows = false
		t2.Style().Options.DrawBorder = false

		for actor, count := range actorCounts {
			t2.AppendRow(table.Row{actor, count})
		}
		t2.Render()

		// Events by role
		fmt.Println("\nEvents by role:")
		t3 := table.NewWriter()
		t3.SetOutputMirror(os.Stdout)
		t3.AppendHeader(table.Row{"Role", "Count"})
		t3.SetStyle(table.StyleLight)
		t3.Style().Options.SeparateRows = false
		t3.Style().Options.DrawBorder = false

		for role, count := range roleCounts {
			t3.AppendRow(table.Row{role, count})
		}
		t3.Render()

		return nil
	},
}

func init() {
	StatsCmd.Flags().Bool("json", false, "Output as JSON")
}

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Debug commands for inspecting events",
}

var eventsListCmd = &cobra.Command{
	Use:   "list",
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

		db, err := openExistingDB()
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
			var filtered []Event
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
				var payloadMap map[string]interface{}
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

var eventsShowCmd = &cobra.Command{
	Use:   "show [event-id]",
	Short: "Show detailed information about a specific event",
	Long: `Show detailed information about a specific event including its full payload.

This is a debug command to help understand what data is stored in an event.

Examples:
  tk events show ev-1-abc123
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eventID := args[0]

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		events, err := db.GetEvents()
		if err != nil {
			return err
		}

		var found *Event
		for i, e := range events {
			if e.ID == eventID {
				found = &events[i]
				break
			}
		}

		if found == nil {
			return fmt.Errorf("event not found: %s", eventID)
		}

		// Pretty print the event as JSON
		output, err := json.MarshalIndent(found, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		fmt.Println(string(output))
		return nil
	},
}

var eventsStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show statistics about events in the database",
	Long: `Show statistics about events in the database, grouped by kind.

This is a debug command to help understand the distribution of events.

Examples:
  tk events stats
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		events, err := db.GetEvents()
		if err != nil {
			return err
		}

		if len(events) == 0 {
			fmt.Println("No events found")
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
	eventsListCmd.Flags().Int("limit", 0, "Limit the number of events to show")
	eventsListCmd.Flags().String("kind", "", "Filter events by kind")
	eventsListCmd.Flags().Bool("verbose", false, "Show full event details including payload")
	eventsListCmd.Flags().Bool("json", false, "Output events as JSON")

	eventsCmd.AddCommand(eventsListCmd)
	eventsCmd.AddCommand(eventsShowCmd)
	eventsCmd.AddCommand(eventsStatsCmd)
}

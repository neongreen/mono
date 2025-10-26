package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var prefixCmd = &cobra.Command{
	Use:   "prefix",
	Short: "Manage task prefixes",
}

var prefixCreateCmd = &cobra.Command{
	Use:   "create [prefix] [description]",
	Short: "Create a new task prefix",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		prefix := args[0]
		description := args[1]

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		currentUser, err := getCurrentUser()
		if err != nil {
			return err
		}

		// Normalize prefix (will be lowercased in CreatePrefix)
		normalizedPrefix := strings.ToLower(prefix)

		if err := db.CreatePrefix(prefix, description, currentUser); err != nil {
			return err
		}

		// Show the normalized prefix in the output
		fmt.Printf("Created prefix %q: %s\n", normalizedPrefix, description)
		return nil
	},
}

var prefixListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all task prefixes",
	Long: `List task prefixes.

By default, shows only prefixes created on this machine (node).
Use --all to show prefixes from all nodes, including synced prefixes.

State values:
  explicit    - Created with 'tk prefix create' (has full metadata)
  discovered  - Found in task IDs but not explicitly created (no metadata)
  removed     - Marked as removed with 'tk prefix remove'

Source values (--all mode only):
  local       - Prefix created on this machine
  synced      - Prefix received from another machine via sync

Examples:
  tk prefix list                # Show local prefixes only
  tk prefix list --all          # Show all prefixes including synced
  tk prefix list --all --verbose # Show with creation timestamps
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		verbose, _ := cmd.Flags().GetBool("verbose")

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Get current node ID to distinguish local vs remote prefixes
		currentNode, err := db.GetOrCreateNodeID()
		if err != nil {
			return err
		}

		var prefixes []Prefix
		if all {
			prefixes, err = db.GetAllPrefixes()
		} else {
			prefixes, err = db.GetPrefixes()
		}
		if err != nil {
			return err
		}

		if len(prefixes) == 0 {
			fmt.Println("No prefixes found")
			return nil
		}

		// Create table
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		if all && verbose {
			t.AppendHeader(table.Row{"Prefix", "Node", "Source", "State", "Description", "Created By", "Created At"})
		} else if all {
			t.AppendHeader(table.Row{"Prefix", "Node", "Source", "State", "Description", "Created By"})
		} else {
			t.AppendHeader(table.Row{"Prefix", "State", "Description", "Created By"})
		}
		t.SetStyle(table.StyleLight)
		t.Style().Options.SeparateRows = false
		t.Style().Options.DrawBorder = false

		for _, p := range prefixes {
			// Determine state based on data, not string matching
			state := "explicit"
			if p.CreatedAt.IsZero() {
				// Discovered prefixes have no creation time
				state = "discovered"
			}
			if p.Removed {
				// Removed prefixes override discovered/explicit
				state = "removed"
			}

			// Determine source (local vs synced)
			source := "local"
			if p.Node != currentNode {
				source = "synced"
			}

			if all && verbose {
				createdAtStr := ""
				if !p.CreatedAt.IsZero() {
					createdAtStr = p.CreatedAt.Format("2006-01-02 15:04:05")
				}
				t.AppendRow(table.Row{p.Prefix, p.Node, source, state, p.Description, p.CreatedBy, createdAtStr})
			} else if all {
				t.AppendRow(table.Row{p.Prefix, p.Node, source, state, p.Description, p.CreatedBy})
			} else {
				t.AppendRow(table.Row{p.Prefix, state, p.Description, p.CreatedBy})
			}
		}

		t.Render()
		return nil
	},
}

var prefixDescribeCmd = &cobra.Command{
	Use:   "describe [prefix]",
	Short: "Show details about a prefix",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prefix := strings.ToLower(args[0])

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Get all prefixes to find the one we want
		prefixes, err := db.GetAllPrefixes()
		if err != nil {
			return err
		}

		nodeID, err := db.GetOrCreateNodeID()
		if err != nil {
			return err
		}

		var found *Prefix
		for _, p := range prefixes {
			if p.Prefix == prefix && p.Node == nodeID {
				found = &p
				break
			}
		}

		if found == nil {
			return fmt.Errorf("prefix %q not found for this node", prefix)
		}

		// Display prefix details
		fmt.Printf("Prefix: %s\n", found.Prefix)
		fmt.Printf("Node: %s\n", found.Node)
		fmt.Printf("Description: %s\n", found.Description)
		fmt.Printf("Created By: %s\n", found.CreatedBy)
		if !found.CreatedAt.IsZero() {
			fmt.Printf("Created At: %s\n", found.CreatedAt.Format("2006-01-02 15:04:05"))
		}

		// Count tasks with this prefix
		taskIDs, err := db.GetTaskIDsByPrefixes([]string{prefix})
		if err != nil {
			return err
		}
		fmt.Printf("Task Count: %d\n", len(taskIDs))

		return nil
	},
}

var prefixRemoveCmd = &cobra.Command{
	Use:   "remove [prefix]",
	Short: "Mark a prefix as removed",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		prefix := strings.ToLower(args[0])

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		currentUser, err := getCurrentUser()
		if err != nil {
			return err
		}

		if err := db.RemovePrefix(prefix, currentUser); err != nil {
			return err
		}

		fmt.Printf("Marked prefix %q as removed\n", prefix)
		return nil
	},
}

func init() {
	prefixListCmd.Flags().Bool("all", false, "Show prefixes from all nodes")
	prefixListCmd.Flags().Bool("verbose", false, "Show additional details like creation time")
	prefixCmd.AddCommand(prefixCreateCmd)
	prefixCmd.AddCommand(prefixListCmd)
	prefixCmd.AddCommand(prefixDescribeCmd)
	prefixCmd.AddCommand(prefixRemoveCmd)
}

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
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

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
		if all {
			t.AppendHeader(table.Row{"Prefix", "Node", "Description", "Created By"})
		} else {
			t.AppendHeader(table.Row{"Prefix", "Description", "Created By"})
		}
		t.SetStyle(table.StyleLight)
		t.Style().Options.SeparateRows = false
		t.Style().Options.DrawBorder = false

		for _, p := range prefixes {
			if all {
				t.AppendRow(table.Row{p.Prefix, p.Node, p.Description, p.CreatedBy})
			} else {
				t.AppendRow(table.Row{p.Prefix, p.Description, p.CreatedBy})
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

func init() {
	prefixListCmd.Flags().Bool("all", false, "Show prefixes from all nodes")
	prefixCmd.AddCommand(prefixCreateCmd)
	prefixCmd.AddCommand(prefixListCmd)
	prefixCmd.AddCommand(prefixDescribeCmd)
}

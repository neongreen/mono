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

func init() {
	prefixListCmd.Flags().Bool("all", false, "Show prefixes from all nodes")
	prefixCmd.AddCommand(prefixCreateCmd)
	prefixCmd.AddCommand(prefixListCmd)
}

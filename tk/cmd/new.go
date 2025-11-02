package main

import "github.com/spf13/cobra"

var newCmd = &cobra.Command{
	Use:   "new [title]",
	Short: "Create a new task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]

		db, err := OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		return createTask(db, cmd, title)
	},
}

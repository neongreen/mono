package cmd

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [title]",
	Short: "Create a new task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		projectFlag, _ := cmd.Flags().GetString("project")
		currentUser, err := getCurrentUser()
		if err != nil {
			return err
		}

		result, err := database.CreateTask(db, database.CreateTaskParams{
			ProjectRef:  projectFlag,
			Title:       args[0],
			CurrentUser: currentUser,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Created task %s: %s\n", result.DisplayID, args[0])
		return nil
	},
}

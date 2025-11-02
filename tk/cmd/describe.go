package cmd

import (
	"github.com/neongreen/mono/tk/internal/database"
	"strings"

	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:     "describe <task> <title>",
	Aliases: []string{"desc", "d"},
	Short:   "Change task title",
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		title := strings.Join(args[1:], " ")

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		currentUser, err := getCurrentUser()
		if err != nil {
			return err
		}

		return editTaskTitle(db, taskRef, title, currentUser)
	},
}

package cmd

import (
	"fmt"
	"strings"

	"github.com/neongreen/mono/tk/internal/utils"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/tasks"
	"github.com/neongreen/mono/tk/internal/types"
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

		currentUser, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Resolve task reference
		taskUID, err := database.ResolveTaskReference(db, types.NewTaskRef(taskRef))
		if err != nil {
			return err
		}

		// Edit task title using business logic
		if err := tasks.EditTitle(db, taskUID, title, currentUser); err != nil {
			return err
		}

		// Display success message
		displayID, err := database.RenderTaskDisplayID(db, taskUID)
		if err != nil {
			displayID = taskRef
		}
		fmt.Printf("Updated title for %s\n", displayID)
		return nil
	},
}

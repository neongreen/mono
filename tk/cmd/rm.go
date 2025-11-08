package cmd

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/tasks"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:     "rm [task-id]",
	Aliases: []string{"delete", "del", "remove"},
	Short:   "Delete a task",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		taskUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(taskRef))
		if err != nil {
			return err
		}

		displayID, err := database.RenderTaskDisplayID(db, taskUUID)
		if err != nil {
			displayID = taskRef
		}

		currentUser, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		if err := tasks.Delete(db, taskUUID, currentUser); err != nil {
			return err
		}

		fmt.Printf("Deleted task %s\n", displayID)
		return nil
	},
}

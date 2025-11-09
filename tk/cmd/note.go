package cmd

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/tasks"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

var noteCmd = &cobra.Command{
	Use:   "note [task-id] [text]",
	Short: "Add a note to a task",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		text := args[1]

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

		// Get task title for display
		var taskTitle string
		err = db.Db.QueryRow(`SELECT title FROM tasks WHERE task_uid = ?`, taskUUID).Scan(&taskTitle)
		if err != nil {
			taskTitle = "" // If we can't get title, just use ID
		}

		currentUser, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		if err := tasks.AddNote(db, taskUUID, text, currentUser, &clock.RealClock{}); err != nil {
			return err
		}

		if taskTitle != "" {
			fmt.Printf("Added note to task %s: %s\n", displayID, taskTitle)
		} else {
			fmt.Printf("Added note to task %s\n", displayID)
		}
		return nil
	},
}

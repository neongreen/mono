package cmd

import (
        "github.com/neongreen/mono/tk/internal/utils"
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/tasks"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var markCmd = &cobra.Command{
	Use:   "mark [task-id] [state]",
	Short: "Set task status",
	Args: func(cmd *cobra.Command, args []string) error {
		unset, _ := cmd.Flags().GetBool("unset")
		if unset {
			return cobra.ExactArgs(1)(cmd, args)
		}
		return cobra.ExactArgs(2)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		unset, _ := cmd.Flags().GetBool("unset")

		var state string
		if unset {
			state = ""
		} else {
			state = args[1]
		}

		axis, _ := cmd.Flags().GetString("axis")
		role, _ := cmd.Flags().GetString("role")

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Resolve task reference
		taskUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(taskRef))
		if err != nil {
			return err
		}

		currentUser, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Mark the task using business logic
		opts := tasks.MarkOptions{
			Axis:  axis,
			State: state,
			Role:  role,
		}

		if err := tasks.Mark(db, taskUUID, opts, currentUser); err != nil {
			return err
		}

		// Display success message
		displayID, err := database.RenderTaskDisplayID(db, taskUUID)
		if err != nil {
			displayID = taskRef
		}

		if unset {
			fmt.Printf("Unset status for task %s (axis: %s)\n", displayID, axis)
		} else {
			fmt.Printf("Set status for task %s: %s=%s\n", displayID, axis, state)
		}

		return nil
	},
}

func init() {
	markCmd.Flags().String("axis", "generic", "Status axis to set")
	markCmd.Flags().String("role", "human", "Role setting the status")
	markCmd.Flags().Bool("unset", false, "Unset status instead of setting it")
}

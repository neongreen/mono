package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
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

		taskUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(taskRef))
		if err != nil {
			return err
		}

		displayID, err := database.RenderTaskDisplayID(db, taskUUID)
		if err != nil {
			displayID = taskRef
		}

		currentUser, err := getCurrentUser()
		if err != nil {
			return err
		}

		eventID, err := database.GenerateEventID(db)
		if err != nil {
			return err
		}

		lamportTS, err := db.GetNextLamportTS()
		if err != nil {
			return err
		}

		payload := types.TaskStatusSetPayload{
			TaskUUID: taskUUID,
			TaskID:   taskRef,
			Axis:     axis,
			State:    state,
			Role:     role,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		now := time.Now()
		event := types.Event{
			ID:        eventID,
			TS:        lamportTS,
			CreatedAt: now,
			Actor:     currentUser,
			Role:      role,
			Kind:      "task.status.set",
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return err
		}

		if unset {
			fmt.Printf("Unset status for task %s (axis: %s)\n", displayID, axis)
		} else {
			fmt.Printf("Set status for task %s: %s=%s\n", displayID, axis, state)
		}
		return nil
	},
}

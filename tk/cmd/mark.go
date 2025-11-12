package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/status"
	"github.com/neongreen/mono/tk/internal/tasks"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

var markCmd = &cobra.Command{
	Use:   "mark [task-id] [state]",
	Short: "Set task status",
	Args: func(cmd *cobra.Command, args []string) error {
		unset, _ := cmd.Flags().GetBool("unset")
		statusFlag, _ := cmd.Flags().GetString("status")

		// If using --status or --unset, only need task-id
		if unset || statusFlag != "" {
			return cobra.ExactArgs(1)(cmd, args)
		}
		// Otherwise need task-id and state
		return cobra.ExactArgs(2)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		unset, _ := cmd.Flags().GetBool("unset")
		statusFlag, _ := cmd.Flags().GetString("status")
		axis, _ := cmd.Flags().GetString("axis")
		customStatus, _ := cmd.Flags().GetBool("custom-status")
		role, _ := cmd.Flags().GetString("role")
		noteText, _ := cmd.Flags().GetString("m")

		var state string

		// Determine the state to set based on flags and arguments
		if unset {
			state = ""
		} else if statusFlag != "" {
			// Using --status flag
			if axis != "generic" && cmd.Flags().Changed("axis") {
				return fmt.Errorf("cannot use both --status and --axis flags")
			}
			// Validate status unless using custom-status flag
			// TODO: Fetch existing custom statuses from project (tk-364)
			if err := status.ValidateStatus(statusFlag, customStatus, nil); err != nil {
				return err
			}
			state = status.NormalizeStatus(statusFlag)
			axis = "generic"
		} else {
			// Using positional argument for status
			state = args[1]
			// Validate status unless using custom-status flag
			// TODO: Fetch existing custom statuses from project (tk-364)
			if err := status.ValidateStatus(state, customStatus, nil); err != nil {
				return err
			}
			state = status.NormalizeStatus(state)
		}

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

		if err := tasks.Mark(db, taskUUID, opts, currentUser, &clock.RealClock{}); err != nil {
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

		// If -m flag provided, add a note
		if noteText != "" {
			eventID, err := database.GenerateEventID(db)
			if err != nil {
				return fmt.Errorf("failed to generate event ID for note: %w", err)
			}

			lamportTS, err := db.GetNextLamportTS()
			if err != nil {
				return fmt.Errorf("failed to get lamport timestamp for note: %w", err)
			}

			payload := types.TaskNoteAddPayload{
				TaskUUID: taskUUID,
				TaskID:   taskRef,
				Markdown: noteText,
			}
			payloadJSON, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("failed to marshal note payload: %w", err)
			}

			event := types.Event{
				ID:        eventID,
				TS:        lamportTS,
				CreatedAt: (&clock.RealClock{}).Now(),
				Actor:     currentUser,
				Role:      role, // Use same role as status change
				Kind:      "task.note.add",
				Payload:   payloadJSON,
			}

			if err := db.InsertEvent(event); err != nil {
				return fmt.Errorf("failed to insert note event: %w", err)
			}

			fmt.Printf("Added note to task %s\n", displayID)
		}

		return nil
	},
}

func init() {
	markCmd.Flags().String("status", "", "Status to set (next, wip, done, closed)")
	markCmd.Flags().Bool("custom-status", false, "Allow setting a custom status not in the predefined list")
	markCmd.Flags().String("axis", "generic", "Status axis to set")
	markCmd.Flags().String("role", "human", "Role setting the status")
	markCmd.Flags().Bool("unset", false, "Unset status instead of setting it")
	markCmd.Flags().StringP("m", "m", "", "Add a note when marking status")

	// Hide --axis flag from help but keep it functional
	markCmd.Flags().MarkHidden("axis")
}

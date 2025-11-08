package cmd

import (
        "github.com/neongreen/mono/tk/internal/utils"
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
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

		eventID, err := database.GenerateEventID(db)
		if err != nil {
			return err
		}

		lamportTS, err := db.GetNextLamportTS()
		if err != nil {
			return err
		}

		payload := types.TaskDeletePayload{
			TaskUUID: taskUUID,
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
			Role:      "human",
			Kind:      string(types.EventKindTaskDelete),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return err
		}

		if err := db.ProjectTaskDeleteEvent(event); err != nil {
			return fmt.Errorf("failed to project task.delete event: %w", err)
		}

		fmt.Printf("Deleted task %s\n", displayID)
		return nil
	},
}

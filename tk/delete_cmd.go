package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/types"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [task-id]",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]

		db, err := OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		taskUUID, err := ResolveTaskReference(db, taskRef)
		if err != nil {
			return err
		}

		displayID, err := RenderTaskDisplayID(db, taskUUID)
		if err != nil {
			displayID = taskRef
		}

		currentUser, err := getCurrentUser()
		if err != nil {
			return err
		}

		eventID, err := GenerateEventID(db)
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

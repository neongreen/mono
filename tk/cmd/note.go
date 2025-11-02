package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
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

		payload := types.TaskNoteAddPayload{
			TaskUUID: taskUUID,
			TaskID:   taskRef,
			Markdown: text,
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
			Kind:      "task.note.add",
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return err
		}

		fmt.Printf("Added note to task %s\n", displayID)
		return nil
	},
}

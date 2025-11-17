package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var relateRmCmd = &cobra.Command{
	Use:   "relate-rm",
	Short: "Remove a relation between two tasks",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcTaskID := args[0]
		relationType := args[1]
		dstTaskID := args[2]

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Resolve both task IDs to UUIDs
		srcUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(srcTaskID))
		if err != nil {
			return fmt.Errorf("failed to resolve source task %q: %w", srcTaskID, err)
		}

		dstUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(dstTaskID))
		if err != nil {
			return fmt.Errorf("failed to resolve target task %q: %w", dstTaskID, err)
		}

		// Get current user
		currentUser, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Generate event ID
		eventID, err := database.GenerateEventID(db)
		if err != nil {
			return err
		}

		// Get next Lamport timestamp
		lamportTS, err := db.GetNextLamportTS()
		if err != nil {
			return err
		}

		// Create relation.remove event
		payload := types.RelationRemovePayload{
			Src:  srcUUID,
			Type: relationType,
			Dst:  dstUUID,
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
			Kind:      "relation.remove",
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return err
		}

		srcDisplay, err := database.RenderTaskDisplayID(db, srcUUID)
		if err != nil {
			srcDisplay = srcTaskID
		}
		dstDisplay, err := database.RenderTaskDisplayID(db, dstUUID)
		if err != nil {
			dstDisplay = dstTaskID
		}

		fmt.Printf("Removed relation: %s %s %s\n", srcDisplay, relationType, dstDisplay)
		return nil
	},
}

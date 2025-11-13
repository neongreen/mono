package group

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

var RmCmd = &cobra.Command{
	Use:   "rm <group-id>",
	Short: "Remove a group (soft delete)",
	Long: `Remove a group container and all its members.

This is a soft delete - the container and members are marked as removed
but remain in the database for event history.

Example:
  tk group rm q-1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		groupID := args[0]

		// Verify group exists and is actually a group
		var primitive string
		var removed int
		var name string
		err = db.Db.QueryRow(`
			SELECT primitive, removed, name
			FROM containers
			WHERE id = ?
		`, groupID).Scan(&primitive, &removed, &name)
		if err != nil {
			return fmt.Errorf("group %q not found", groupID)
		}

		if primitive != string(types.PrimitiveGroup) {
			return fmt.Errorf("%q is a %s, not a group", groupID, primitive)
		}

		if removed == 1 {
			return fmt.Errorf("group %q has already been removed", groupID)
		}

		// Get current user
		actor, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Create container.remove event
		payload := types.RemoveContainerPayload{
			ID: groupID,
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		eventID, err := database.GenerateEventID(db)
		if err != nil {
			return fmt.Errorf("failed to generate event ID: %w", err)
		}

		ts, err := db.GetNextLamportTS()
		if err != nil {
			return fmt.Errorf("failed to get next lamport timestamp: %w", err)
		}

		event := types.Event{
			ID:        eventID,
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      "human",
			Kind:      string(types.EventKindContainerRemove),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event
		if err := db.ProjectContainerRemoveEvent(event); err != nil {
			return fmt.Errorf("failed to project event: %w", err)
		}

		fmt.Printf("Removed group %s: %s\n", groupID, name)
		return nil
	},
}

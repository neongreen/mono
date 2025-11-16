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
var queueRmCmd = &cobra.Command{
	Use:   "queue-rm",
	Short: "Remove a queue (soft delete)",
	Long: `Remove a queue container and all its members.

This is a soft delete - the container and members are marked as removed
but remain in the database for event history.

Example:
  tk queue rm q-1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		queueID := args[0]

		// Verify queue exists and is actually a queue
		var primitive string
		var removed int
		var name string
		err = db.Db.QueryRow(`
			SELECT primitive, removed, name
			FROM containers
			WHERE id = ?
		`, queueID).Scan(&primitive, &removed, &name)
		if err != nil {
			return fmt.Errorf("queue %q not found", queueID)
		}

		if primitive != string(types.PrimitiveQueue) {
			return fmt.Errorf("%q is a %s, not a queue", queueID, primitive)
		}

		if removed == 1 {
			return fmt.Errorf("queue %q has already been removed", queueID)
		}

		// Get current user
		actor, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Create container.remove event
		payload := types.RemoveContainerPayload{
			ID: queueID,
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

		fmt.Printf("Removed queue %s: %s\n", queueID, name)
		return nil
	},
}

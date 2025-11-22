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
var queuePushCmd = &cobra.Command{
	Use:   "queue-push",
	Short: "Push an item onto a queue",
	Long: `Push an item onto the tail of a queue.

Example:
  tk queue push q-1 tk-123`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		queueID := args[0]
		itemRef := args[1]

		// Resolve task reference to task UID
		taskUID, err := database.ResolveTaskReference(db, types.NewTaskRef(itemRef))
		if err != nil {
			return fmt.Errorf("failed to resolve task %q: %w", itemRef, err)
		}

		// Verify queue exists and is actually a queue
		var primitive string
		var removed int
		err = db.Db.QueryRow(`
			SELECT primitive, removed
			FROM containers
			WHERE id = ?
		`, queueID).Scan(&primitive, &removed)
		if err != nil {
			return fmt.Errorf("queue %q not found", queueID)
		}

		if primitive != string(types.PrimitiveQueue) {
			return fmt.Errorf("%q is a %s, not a queue", queueID, primitive)
		}

		if removed == 1 {
			return fmt.Errorf("queue %q has been removed", queueID)
		}

		// Get current user
		actor, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Create queue.push event
		payload := types.QueuePushPayload{
			ContainerID: queueID,
			ItemID:      types.TaskUID(taskUID),
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
			Kind:      string(types.EventKindQueuePush),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event
		if err := db.RebuildProjections(); err != nil {
			return fmt.Errorf("failed to project event: %w", err)
		}

		fmt.Printf("Pushed %s onto queue %s\n", itemRef, queueID)
		return nil
	},
}

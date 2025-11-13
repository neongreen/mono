package queue

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

var PopCmd = &cobra.Command{
	Use:   "pop <queue-id>",
	Short: "Pop an item from a queue",
	Long: `Pop an item from the head of a queue (first in, first out).

Example:
  tk queue pop q-1`,
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

		// Find the head item (smallest position, not removed)
		var headItemID string
		err = db.Db.QueryRow(`
			SELECT item_id
			FROM container_members
			WHERE container_id = ? AND removed = 0
			ORDER BY position ASC
			LIMIT 1
		`, queueID).Scan(&headItemID)

		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("queue %q is empty", queueID)
		}
		if err != nil {
			return fmt.Errorf("failed to find head item: %w", err)
		}

		// Get current user
		actor, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Create queue.pop event
		payload := types.QueuePopPayload{
			ContainerID: queueID,
			ItemID:      headItemID,
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
			Kind:      string(types.EventKindQueuePop),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event
		if err := db.ProjectQueuePopEvent(event); err != nil {
			return fmt.Errorf("failed to project event: %w", err)
		}

		fmt.Printf("Popped %s from queue %s\n", headItemID, queueID)
		return nil
	},
}

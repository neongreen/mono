package stack

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

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var PopCmd = &cobra.Command{
	Use:   "pop <stack-id>",
	Short: "Pop an item from a stack",
	Long: `Pop an item from the head of a stack (first in, first out).

Example:
  tk stack pop q-1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		stackID := args[0]

		// Verify stack exists and is actually a stack
		var primitive string
		var removed int
		err = db.Db.QueryRow(`
			SELECT primitive, removed
			FROM containers
			WHERE id = ?
		`, stackID).Scan(&primitive, &removed)
		if err != nil {
			return fmt.Errorf("stack %q not found", stackID)
		}

		if primitive != string(types.PrimitiveStack) {
			return fmt.Errorf("%q is a %s, not a stack", stackID, primitive)
		}

		if removed == 1 {
			return fmt.Errorf("stack %q has been removed", stackID)
		}

		// Find the tail item (largest position, not removed) - LIFO!
		var tailItemID string
		err = db.Db.QueryRow(`
			SELECT item_id
			FROM container_members
			WHERE container_id = ? AND removed = 0
			ORDER BY position DESC
			LIMIT 1
		`, stackID).Scan(&tailItemID)

		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("stack %q is empty", stackID)
		}
		if err != nil {
			return fmt.Errorf("failed to find head item: %w", err)
		}

		// Get current user
		actor, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Create stack.pop event (tailItemID is already a task UID from database)
		payload := types.StackPopPayload{
			ContainerID: stackID,
			ItemID:      types.TaskUID(tailItemID),
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
			Kind:      string(types.EventKindStackPop),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event
		if err := db.ProjectStackPopEvent(event); err != nil {
			return fmt.Errorf("failed to project event: %w", err)
		}

		// Render task UID as display ID for user output
		displayID, err := database.RenderTaskDisplayID(db, tailItemID)
		if err != nil {
			displayID = tailItemID
		}

		fmt.Printf("Popped %s from stack %s\n", displayID, stackID)
		return nil
	},
}

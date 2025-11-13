package stack

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

var PushCmd = &cobra.Command{
	Use:   "push <stack-id> <item-id>",
	Short: "Push an item onto a stack",
	Long: `Push an item onto the tail of a stack.

Example:
  tk stack push q-1 tk-123`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		stackID := args[0]
		itemRef := args[1]

		// Resolve task reference to task UID
		taskUID, err := database.ResolveTaskReference(db, types.NewTaskRef(itemRef))
		if err != nil {
			return fmt.Errorf("failed to resolve task %q: %w", itemRef, err)
		}

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

		// Get current user
		actor, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Create stack.push event
		payload := types.StackPushPayload{
			ContainerID: stackID,
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
			Kind:      string(types.EventKindStackPush),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event
		if err := db.ProjectStackPushEvent(event); err != nil {
			return fmt.Errorf("failed to project event: %w", err)
		}

		fmt.Printf("Pushed %s onto stack %s\n", itemRef, stackID)
		return nil
	},
}

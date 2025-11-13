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

var RemoveCmd = &cobra.Command{
	Use:   "remove <group-id> <item-id>",
	Short: "Remove an item from a group",
	Long: `Remove an item from a group.

Example:
  tk group remove g-1 tk-123`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		groupID := args[0]
		itemRef := args[1]

		// Resolve task reference to task UID
		taskUID, err := database.ResolveTaskReference(db, types.NewTaskRef(itemRef))
		if err != nil {
			return fmt.Errorf("failed to resolve task %q: %w", itemRef, err)
		}

		// Verify group exists
		var primitive string
		var removed int
		err = db.Db.QueryRow(`
			SELECT primitive, removed
			FROM containers
			WHERE id = ?
		`, groupID).Scan(&primitive, &removed)
		if err != nil {
			return fmt.Errorf("group %q not found", groupID)
		}

		if primitive != string(types.PrimitiveGroup) {
			return fmt.Errorf("%q is a %s, not a group", groupID, primitive)
		}

		if removed == 1 {
			return fmt.Errorf("group %q has been removed", groupID)
		}

		// Get current user
		actor, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Create group.remove event
		payload := types.GroupRemovePayload{
			ContainerID: groupID,
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
			Kind:      string(types.EventKindGroupRemove),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event
		if err := db.ProjectGroupRemoveEvent(event); err != nil {
			return fmt.Errorf("failed to project event: %w", err)
		}

		fmt.Printf("Removed %s from group %s\n", itemRef, groupID)
		return nil
	},
}

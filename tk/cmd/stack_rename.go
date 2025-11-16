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
var stackRenameCmd = &cobra.Command{
	Use:   "stack-rename",
	Short: "Rename a stack",
	Long: `Rename a stack container.

Example:
  tk stack rename q-1 "December 2025 Sprint"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		stackID := args[0]
		newName := args[1]

		// Verify stack exists and is actually a stack
		var primitive string
		var removed int
		var oldName string
		err = db.Db.QueryRow(`
			SELECT primitive, removed, name
			FROM containers
			WHERE id = ?
		`, stackID).Scan(&primitive, &removed, &oldName)
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

		// Create container.rename event
		payload := types.RenameContainerPayload{
			ID:   stackID,
			Name: newName,
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
			Kind:      string(types.EventKindContainerRename),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event
		if err := db.ProjectContainerRenameEvent(event); err != nil {
			return fmt.Errorf("failed to project event: %w", err)
		}

		fmt.Printf("Renamed stack %s: %q → %q\n", stackID, oldName, newName)
		return nil
	},
}

package project

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var RmCmd = &cobra.Command{
	Use:   "rm <project-ref>",
	Short: "Delete a project (project-ref can be a UID, alias, or name)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Resolve the project reference to a ProjectUID
		ref := types.NewProjectRef(args[0])
		projectUID, err := database.ResolveProjectRef(db, ref)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}

		// Get current user
		actor, err := getCurrentUser()
		if err != nil {
			return err
		}

		// Create project.delete event
		payload := types.ProjectDeletePayload{
			ProjectUID: projectUID.String(),
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
			Kind:      string(types.EventKindProjectDelete),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event (removes the project from the table)
		if err := db.ProjectProjectDeleteEvent(event); err != nil {
			return fmt.Errorf("failed to project project deletion: %w", err)
		}

		fmt.Printf("Deleted project %s\n", projectUID)
		return nil
	},
}

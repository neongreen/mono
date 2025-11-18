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
var projectRenameCmd = &cobra.Command{
	Use:   "project-rename",
	Short: "Rename a project",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		projectRef := types.NewProjectRef(args[0])
		newName := args[1]

		// Validate new project name
		if err := types.ValidateProjectName(newName); err != nil {
			return err
		}

		// Resolve project reference to ProjectUID
		projectUID, err := database.ResolveProjectRef(db, projectRef)
		if err != nil {
			return err
		}

		// Get current user
		actor, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Create project.name.set event
		payload := types.ProjectNameSetPayload{
			ProjectUID: projectUID,
			Name:       newName,
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
			Kind:      string(types.EventKindProjectNameSet),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event into projects table
		if err := db.RebuildProjections(); err != nil {
			return fmt.Errorf("failed to project event: %w", err)
		}

		fmt.Printf("Renamed project %s to: %s\n", projectUID, newName)
		return nil
	},
}

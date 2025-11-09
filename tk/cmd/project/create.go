package project

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

var CreateCmd = &cobra.Command{
	Use:   "create <name> [description]",
	Short: "Create a new project",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Check database version - projects are v4+ only
		version, err := db.GetDBVersion()
		if err != nil {
			return err
		}
		if version < 4 {
			return fmt.Errorf("projects require database v4 or higher, but current version is v%d", version)
		}

		name := args[0]
		description := ""
		if len(args) > 1 {
			description = args[1]
		}

		// Validate project name
		if err := types.ValidateProjectName(name); err != nil {
			return err
		}

		// Get current user
		actor, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Generate new project UID
		projectUID := types.NewProjectUID()

		// Create project.created event
		payload := types.ProjectCreatedPayload{
			ProjectUID:  string(projectUID),
			Type:        "local",
			Name:        name,
			Description: description,
			CreatedBy:   actor,
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
			Kind:      string(types.EventKindProjectCreated),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event into projects table
		if err := db.ProjectProjectCreatedEvent(event); err != nil {
			return fmt.Errorf("failed to project event: %w", err)
		}

		fmt.Printf("Created project %s: %s\n", projectUID, name)
		return nil
	},
}

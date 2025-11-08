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

var AliasAddCmd = &cobra.Command{
	Use:   "add <project-ref> <alias>",
	Short: "Add an alias for a project (project-ref can be a UID, alias, or name)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Resolve the project reference to a ProjectUID
		projectRef := types.NewProjectRef(args[0])
		projectUID, err := database.ResolveProjectRef(db, projectRef)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}

		alias := args[1]

		// Get current user and node
		actor, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		nodeID, err := db.GetOrCreateNodeID()
		if err != nil {
			return err
		}

		// Create project.alias.add event
		payload := types.ProjectAliasAddPayload{
			ProjectUID: projectUID.String(),
			Alias:      alias,
			Node:       nodeID,
			AddedBy:    actor,
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
			Kind:      string(types.EventKindProjectAliasAdd),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the alias
		if err := db.ProjectProjectAliasAddEvent(event); err != nil {
			return fmt.Errorf("failed to project alias: %w", err)
		}

		fmt.Printf("Added alias '%s' for project %s\n", alias, projectUID)
		return nil
	},
}

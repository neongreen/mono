package project

import (
	"encoding/json"
	"fmt"
	"os/user"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var AliasRemoveCmd = &cobra.Command{
	Use:   "remove <alias>",
	Short: "Remove an alias for a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		alias := args[0]

		// Get current node
		nodeID, err := db.GetOrCreateNodeID()
		if err != nil {
			return err
		}

		// Find the project UID for this alias
		var projectUID string
		err = db.Db.QueryRow(`
			SELECT project_uid FROM project_aliases
			WHERE alias = ? AND node = ?
		`, alias, nodeID).Scan(&projectUID)
		if err != nil {
			return fmt.Errorf("alias not found: %w", err)
		}

		// Get current user
		actor, err := getCurrentUser()
		if err != nil {
			return err
		}

		// Create project.alias.remove event
		payload := types.ProjectAliasRemovePayload{
			ProjectUID: projectUID,
			Alias:      alias,
			Node:       nodeID,
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
			Kind:      string(types.EventKindProjectAliasRemove),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event (removes the alias from the table)
		if err := db.ProjectProjectAliasRemoveEvent(event); err != nil {
			return fmt.Errorf("failed to project alias removal: %w", err)
		}

		fmt.Printf("Removed alias '%s' for project %s\n", alias, projectUID)
		return nil
	},
}

// getCurrentUser returns the current user identifier
func getCurrentUser() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	return currentUser.Username, nil
}

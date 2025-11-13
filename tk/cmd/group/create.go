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

var CreateCmd = &cobra.Command{
	Use:   "create <kind> <name>",
	Short: "Create a new group",
	Long: `Create a new group container.

The kind must be defined first using: tk schema add-kind group <kind>

Example:
  tk schema add-kind group sprint --description "Sprint work"
  tk group create sprint "November 2025 Sprint"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Check database version
		version, err := db.GetDBVersion()
		if err != nil {
			return err
		}
		if version < 6 {
			return fmt.Errorf("containers require database v6 or higher, but current version is v%d", version)
		}

		kindName := args[0]
		containerName := args[1]

		// Verify kind exists and is a group
		var primitive string
		var deprecated int
		err = db.Db.QueryRow(`
			SELECT primitive, deprecated
			FROM container_kinds
			WHERE name = ?
		`, kindName).Scan(&primitive, &deprecated)
		if err != nil {
			return fmt.Errorf("kind %q not found. Define it first with: tk schema add-kind group %s", kindName, kindName)
		}

		if primitive != string(types.PrimitiveGroup) {
			return fmt.Errorf("kind %q is a %s, not a group", kindName, primitive)
		}

		if deprecated == 1 {
			return fmt.Errorf("kind %q is deprecated", kindName)
		}

		// Generate container ID
		containerID, err := db.GenerateContainerID(types.PrimitiveGroup)
		if err != nil {
			return fmt.Errorf("failed to generate container ID: %w", err)
		}

		// Get current user
		actor, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Create container.create event
		payload := types.CreateContainerPayload{
			ID:        containerID,
			Primitive: types.PrimitiveGroup,
			Kind:      kindName,
			Name:      containerName,
			CreatedBy: actor,
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
			Kind:      string(types.EventKindContainerCreate),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event into containers table
		if err := db.ProjectContainerCreateEvent(event); err != nil {
			return fmt.Errorf("failed to project event: %w", err)
		}

		fmt.Printf("Created group %s: %s\n", containerID, containerName)
		return nil
	},
}

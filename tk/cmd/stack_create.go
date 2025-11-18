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
var stackCreateCmd = &cobra.Command{
	Use:   "stack-create",
	Short: "Create a new stack",
	Long: `Create a new stack container.

If no kind is specified, uses the default "general" kind.

Examples:
  tk stack create "My Stack"
  tk stack create later "November 2025 Sprint"

To define custom kinds:
  tk schema add-kind stack later --description "Return to later"`,
	Args: cobra.RangeArgs(1, 2),
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

		// Parse arguments: either "name" or "kind name"
		var kindName, containerName string
		if len(args) == 1 {
			kindName = "general"
			containerName = args[0]
		} else {
			kindName = args[0]
			containerName = args[1]
		}

		// Ensure kind exists, auto-creating "general" if needed
		var primitive string
		var deprecated int
		err = db.Db.QueryRow(`
			SELECT primitive, deprecated
			FROM container_kinds
			WHERE name = ?
		`, kindName).Scan(&primitive, &deprecated)
		if err != nil {
			// If "general" kind doesn't exist, auto-create it
			if kindName == "general" {
				if err := db.CreateDefaultKind(types.PrimitiveStack, "general", "General purpose stack"); err != nil {
					return fmt.Errorf("failed to create default kind: %w", err)
				}
				// Re-query after creation
				err = db.Db.QueryRow(`
					SELECT primitive, deprecated
					FROM container_kinds
					WHERE name = ?
				`, kindName).Scan(&primitive, &deprecated)
				if err != nil {
					return fmt.Errorf("failed to verify default kind: %w", err)
				}
			} else {
				return fmt.Errorf("kind %q not found. Define it first with: tk schema add-kind stack %s", kindName, kindName)
			}
		}

		if primitive != string(types.PrimitiveStack) {
			return fmt.Errorf("kind %q is a %s, not a stack", kindName, primitive)
		}

		if deprecated == 1 {
			return fmt.Errorf("kind %q is deprecated", kindName)
		}

		// Generate container ID
		containerID, err := db.GenerateContainerID(types.PrimitiveStack)
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
			Primitive: types.PrimitiveStack,
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
		if err := db.RebuildProjections(); err != nil {
			return fmt.Errorf("failed to project event: %w", err)
		}

		fmt.Printf("Created stack %s: %s\n", containerID, containerName)
		return nil
	},
}

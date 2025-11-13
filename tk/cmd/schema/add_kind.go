package schema

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

var (
	addKindDescription string
)

var AddKindCmd = &cobra.Command{
	Use:   "add-kind <primitive> <name>",
	Short: "Define a new container kind",
	Long: `Define a new container kind with a primitive type.

Primitive types:
  queue  - Ordered FIFO (first in, first out)
  stack  - Ordered LIFO (last in, first out)
  group  - Unordered set

Examples:
  tk schema add-kind queue sprint --description "Timeboxed work period"
  tk schema add-kind stack later --description "Return to later"
  tk schema add-kind group today --description "Today's focus"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Check database version - containers are v6+
		version, err := db.GetDBVersion()
		if err != nil {
			return err
		}
		if version < 6 {
			return fmt.Errorf("containers require database v6 or higher, but current version is v%d", version)
		}

		primitiveStr := args[0]
		kindName := args[1]

		// Validate primitive
		var primitive types.ContainerPrimitive
		switch primitiveStr {
		case "queue":
			primitive = types.PrimitiveQueue
		case "stack":
			primitive = types.PrimitiveStack
		case "group":
			primitive = types.PrimitiveGroup
		default:
			return fmt.Errorf("invalid primitive %q: must be queue, stack, or group", primitiveStr)
		}

		// Get current user
		actor, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Create container.kind.define event
		payload := types.DefineContainerKindPayload{
			Name:        kindName,
			Primitive:   primitive,
			Description: addKindDescription,
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
			Kind:      string(types.EventKindContainerKindDefine),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event into container_kinds table
		if err := db.ProjectContainerKindDefineEvent(event); err != nil {
			return fmt.Errorf("failed to project event: %w", err)
		}

		fmt.Printf("Defined %s kind: %s\n", primitive, kindName)
		if addKindDescription != "" {
			fmt.Printf("Description: %s\n", addKindDescription)
		}
		return nil
	},
}

func init() {
	AddKindCmd.Flags().StringVar(&addKindDescription, "description", "", "Description of the container kind")
}

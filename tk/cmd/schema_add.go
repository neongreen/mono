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

var (
	addDescription string
	addLLMHint     string
)

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var schemaAddCmd = &cobra.Command{
	Use:   "schema-add",
	Short: "Define a new schema kind",
	Long: `Define a new container kind or item kind.

Kind types:
  queue  - FIFO container (first in, first out)
  stack  - LIFO container (last in, first out)
  group  - Unordered container
  item   - Custom item type (task, decision, resource, etc.)

Examples:
  tk schema add queue sprint --description "Timeboxed work period"
  tk schema add item decision --description "Architecture decisions" --llm-hint "Use for key technical choices"
  tk schema add item resource --description "Links and documentation"`,
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

		kindType := args[0]
		kindName := args[1]

		// Handle item kinds vs container kinds
		if kindType == "item" {
			return createItemKind(db, kindName, addDescription, addLLMHint)
		}

		// Handle container kinds (queue/stack/group)
		if version < 6 {
			return fmt.Errorf("containers require database v6 or higher, but current version is v%d", version)
		}

		var primitive types.ContainerPrimitive
		switch kindType {
		case "queue":
			primitive = types.PrimitiveQueue
		case "stack":
			primitive = types.PrimitiveStack
		case "group":
			primitive = types.PrimitiveGroup
		default:
			return fmt.Errorf("invalid kind type %q: must be queue, stack, group, or item", kindType)
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
			Description: addDescription,
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
		if addDescription != "" {
			fmt.Printf("Description: %s\n", addDescription)
		}
		return nil
	},
}

// createItemKind creates a new item kind via event
func createItemKind(db *database.DB, kindName, description, llmHint string) error {
	// Check database version - item kinds are v7+
	version, err := db.GetDBVersion()
	if err != nil {
		return err
	}
	if version < 7 {
		return fmt.Errorf("item kinds require database v7 or higher, but current version is v%d", version)
	}

	// Get current user
	actor, err := utils.GetCurrentUser()
	if err != nil {
		return err
	}

	// Create item_kind.define event
	payload := types.DefineItemKindPayload{
		Name:        kindName,
		Description: description,
		LLMHint:     llmHint,
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
		Kind:      string(types.EventKindItemKindDefine),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	// Project the event into item_kinds table
	if err := db.ProjectItemKindDefineEvent(event); err != nil {
		return fmt.Errorf("failed to project event: %w", err)
	}

	fmt.Printf("Defined item kind: %s\n", kindName)
	if description != "" {
		fmt.Printf("Description: %s\n", description)
	}
	if llmHint != "" {
		fmt.Printf("LLM hint: %s\n", llmHint)
	}
	return nil
}

func init() {
	schemaAddCmd.Flags().StringVar(&addDescription, "description", "", "Description of the kind")
	schemaAddCmd.Flags().StringVar(&addLLMHint, "llm-hint", "", "LLM guidance for when to use this kind (item kinds only)")
}

// Keep old command for backward compatibility (deprecated)
// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var schemaAddCmd = &cobra.Command{
	Use:   "schema-add",
	Deprecated: "use 'tk schema add' instead",
	Hidden:     true,
	RunE:       AddCmd.RunE,
}

package meta

import (
	"encoding/json"
	"fmt"
	"os/user"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var SetCmd = &cobra.Command{
	Use:   "set <task> <key> <json-value>",
	Short: "Set metadata value (creates claim)",
	Long: `Set a metadata value for a task, creating a claim.

The value must be valid JSON. Examples:
  tk meta set tk-1 priority 1
  tk meta set tk-1 labels '["bug","urgent"]'
  tk meta set tk-1 sla '{"days":7}'
  tk meta set tk-1 assignee '"alice"'

Multiple actors can set the same key; the effective value is resolved
by authority (human > qa > rel > agent > bot).`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		key := args[1]
		valueStr := args[2]
		role, _ := cmd.Flags().GetString("role")

		// Validate JSON
		if !json.Valid([]byte(valueStr)) {
			return fmt.Errorf("invalid JSON value: %s", valueStr)
		}

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Resolve task
		taskUID, err := database.ResolveTaskReference(db, types.NewTaskRef(taskRef))
		if err != nil {
			return fmt.Errorf("failed to resolve task %q: %w", taskRef, err)
		}

		// Get current user
		actor, err := getCurrentUser()
		if err != nil {
			return err
		}

		// Default role to human if not specified
		if role == "" {
			role = "human"
		}

		// Generate event ID
		eventID, err := database.GenerateEventID(db)
		if err != nil {
			return err
		}

		// Get next Lamport timestamp
		ts, err := db.GetNextLamportTS()
		if err != nil {
			return err
		}

		// Create task.meta.set event
		payload := types.TaskMetaSetPayload{
			TaskUUID: taskUID,
			TaskID:   taskRef,
			Key:      key,
			Value:    json.RawMessage(valueStr),
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		event := types.Event{
			ID:        eventID,
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      role,
			Kind:      string(types.EventKindTaskMetaSet),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		fmt.Printf("Set %s=%s for task %s (role: %s)\n", key, valueStr, taskRef, role)
		return nil
	},
}

func init() {
	SetCmd.Flags().String("role", "", "Role making the claim (human, agent, bot, qa, rel)")
}

func getCurrentUser() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	return currentUser.Username, nil
}

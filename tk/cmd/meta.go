package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/reducer"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var metaCmd = &cobra.Command{
	Use:   "meta",
	Short: "Manage task metadata",
	Long: `Manage task metadata with claims and authority resolution.

Metadata keys can have competing values from different roles (human, agent, qa, etc).
The effective value is resolved using the same authority lattice as status axes.`,
}

var metaSetCmd = &cobra.Command{
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

var metaGetCmd = &cobra.Command{
	Use:   "get <task> <key>",
	Short: "Get effective metadata value",
	Long:  `Get the effective metadata value for a key, after authority resolution.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		key := args[1]

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

		// Build reducer to get current state
		events, err := db.GetEvents()
		if err != nil {
			return fmt.Errorf("failed to get events: %w", err)
		}

		r := reducer.NewReducer()
		for _, e := range events {
			if err := r.Apply(e); err != nil {
				return fmt.Errorf("failed to apply event: %w", err)
			}
		}

		task, ok := r.GetTask(taskUID)
		if !ok {
			return fmt.Errorf("task not found: %s", taskRef)
		}

		if task.Metadata == nil {
			return fmt.Errorf("no metadata found for task %s", taskRef)
		}

		metaStatus, ok := task.Metadata[key]
		if !ok {
			return fmt.Errorf("metadata key %q not found for task %s", key, taskRef)
		}

		// Print effective value as raw JSON
		fmt.Println(string(metaStatus.Effective))
		return nil
	},
}

var metaListCmd = &cobra.Command{
	Use:   "list <task>",
	Short: "List all metadata for a task",
	Long:  `List all metadata keys and their effective values for a task.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		jsonOutput, _ := cmd.Flags().GetBool("json")

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

		// Build reducer
		events, err := db.GetEvents()
		if err != nil {
			return fmt.Errorf("failed to get events: %w", err)
		}

		r := reducer.NewReducer()
		for _, e := range events {
			if err := r.Apply(e); err != nil {
				return fmt.Errorf("failed to apply event: %w", err)
			}
		}

		task, ok := r.GetTask(taskUID)
		if !ok {
			return fmt.Errorf("task not found: %s", taskRef)
		}

		if len(task.Metadata) == 0 {
			if jsonOutput {
				fmt.Println("{}")
			} else {
				fmt.Printf("No metadata for task %s\n", taskRef)
			}
			return nil
		}

		if jsonOutput {
			// Output all metadata as JSON object
			output := make(map[string]json.RawMessage)
			for key, status := range task.Metadata {
				output[key] = status.Effective
			}
			data, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
		} else {
			// Human-readable output
			fmt.Printf("Metadata for %s:\n\n", taskRef)
			for key, status := range task.Metadata {
				fmt.Printf("  %s: %s\n", key, string(status.Effective))
			}
		}

		return nil
	},
}

var metaClaimsCmd = &cobra.Command{
	Use:   "claims <task> <key>",
	Short: "Show all competing claims for a metadata key",
	Long:  `Show all competing claims for a metadata key, including tentative claims.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		key := args[1]

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

		// Build reducer
		events, err := db.GetEvents()
		if err != nil {
			return fmt.Errorf("failed to get events: %w", err)
		}

		r := reducer.NewReducer()
		for _, e := range events {
			if err := r.Apply(e); err != nil {
				return fmt.Errorf("failed to apply event: %w", err)
			}
		}

		task, ok := r.GetTask(taskUID)
		if !ok {
			return fmt.Errorf("task not found: %s", taskRef)
		}

		if task.Metadata == nil {
			return fmt.Errorf("no metadata found for task %s", taskRef)
		}

		metaStatus, ok := task.Metadata[key]
		if !ok {
			return fmt.Errorf("metadata key %q not found for task %s", key, taskRef)
		}

		// Print effective value
		fmt.Printf("Effective: %s\n\n", string(metaStatus.Effective))

		// Print all claims
		if len(metaStatus.Claims) > 1 {
			fmt.Println("Claims:")
			for _, claim := range metaStatus.Claims {
				tentativeStr := ""
				if claim.Tentative {
					tentativeStr = " (tentative)"
				}
				fmt.Printf("  %s by %s (ts: %d)%s\n",
					string(claim.Value),
					claim.Role,
					claim.TS,
					tentativeStr)
			}
		}

		return nil
	},
}

func init() {
	metaSetCmd.Flags().String("role", "", "Role making the claim (human, agent, bot, qa, rel)")
	metaListCmd.Flags().Bool("json", false, "Output as JSON")

	metaCmd.AddCommand(metaSetCmd)
	metaCmd.AddCommand(metaGetCmd)
	metaCmd.AddCommand(metaListCmd)
	metaCmd.AddCommand(metaClaimsCmd)

	rootCmd.AddCommand(metaCmd)
}

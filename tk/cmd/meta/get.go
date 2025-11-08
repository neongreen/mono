package meta

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/reducer"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var GetCmd = &cobra.Command{
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

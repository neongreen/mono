package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/reducer"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var metaLsCmd = &cobra.Command{
	Use:   "meta-ls",
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

func init() {
	metaLsCmd.Flags().Bool("json", false, "Output as JSON")
}

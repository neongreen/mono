package cmd

import (
	"fmt"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var relateLsCmd = &cobra.Command{
	Use:   "relate-ls",
	Short: "List all relations for a task",
	Long: `List all relations for a task in a format that can be copy-pasted to remove.

Examples:
  tk relate ls tk-161
  tk relate ls tk-vsc-74`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Load config
		config, err := config_pkg.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Use cached reducer for performance
		reducer, err := db.GetCachedReducerWithConfig(config)
		if err != nil {
			return err
		}

		// Resolve task ID to UUID
		taskUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(taskID))
		if err != nil {
			return fmt.Errorf("failed to resolve task %q: %w", taskID, err)
		}

		task, ok := reducer.GetTask(taskUUID)
		if !ok {
			return fmt.Errorf("task %q not found", taskID)
		}

		if task.Relations == nil {
			fmt.Printf("No relations for task %s\n", taskID)
			return nil
		}

		taskDisplayID, err := database.RenderTaskDisplayID(db, taskUUID)
		if err != nil {
			taskDisplayID = taskID
		}

		fmt.Printf("Relations for %s:\n", taskDisplayID)

		printed := false

		// Print blocks relations
		if len(task.Relations.Blocks.Out) > 0 {
			for _, target := range task.Relations.Blocks.Out {
				targetDisplay, _ := database.RenderTaskDisplayID(db, target.TaskUUID)
				fmt.Printf("  %s blocks %s\n", taskDisplayID, targetDisplay)
				printed = true
			}
		}
		if len(task.Relations.Blocks.In) > 0 {
			for _, source := range task.Relations.Blocks.In {
				sourceDisplay, _ := database.RenderTaskDisplayID(db, source.TaskUUID)
				fmt.Printf("  %s blocks %s\n", sourceDisplay, taskDisplayID)
				printed = true
			}
		}

		// Print subtask relations
		if len(task.Relations.Subtask.Children) > 0 {
			for _, childUUID := range task.Relations.Subtask.Children {
				childDisplay, _ := database.RenderTaskDisplayID(db, childUUID)
				fmt.Printf("  %s subtask %s\n", taskDisplayID, childDisplay)
				printed = true
			}
		}
		if task.Relations.Subtask.Parent != "" {
			parentDisplay, _ := database.RenderTaskDisplayID(db, task.Relations.Subtask.Parent)
			fmt.Printf("  %s subtask %s\n", parentDisplay, taskDisplayID)
			printed = true
		}

		// Print related relations
		if len(task.Relations.Related.Out) > 0 {
			for _, target := range task.Relations.Related.Out {
				targetDisplay, _ := database.RenderTaskDisplayID(db, target.TaskUUID)
				fmt.Printf("  %s related %s\n", taskDisplayID, targetDisplay)
				printed = true
			}
		}

		// Print duplicate_of relations
		if len(task.Relations.Duplicate.Out) > 0 {
			for _, target := range task.Relations.Duplicate.Out {
				targetDisplay, _ := database.RenderTaskDisplayID(db, target.TaskUUID)
				fmt.Printf("  %s duplicate_of %s\n", taskDisplayID, targetDisplay)
				printed = true
			}
		}

		// Print supersedes relations
		if len(task.Relations.Supersedes.Out) > 0 {
			for _, target := range task.Relations.Supersedes.Out {
				targetDisplay, _ := database.RenderTaskDisplayID(db, target.TaskUUID)
				fmt.Printf("  %s supersedes %s\n", taskDisplayID, targetDisplay)
				printed = true
			}
		}

		if !printed {
			fmt.Printf("  (none)\n")
		}

		return nil
	},
}

func init() {
	relateLsCmd.Flags().Bool("json", false, "Output as JSON")
}

package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var conflictsCmd = &cobra.Command{
	Use:   "conflicts [task-id]",
	Short: "Show relation conflicts for a task or all tasks",
	Long: `Show relation conflicts including cycles in blocks and subtask relations.

If a task ID is provided, shows conflicts for that task only.
Otherwise, shows all conflicts in the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Load config
		config, err := LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get all events and build reducer
		events, err := db.GetEvents()
		if err != nil {
			return err
		}

		reducer, err := BuildFromEventsWithConfig(events, config)
		if err != nil {
			return err
		}

		// Check for cycles in blocks and subtasks
		blocksCycles := reducer.relations.DetectCycles("blocks")
		subtaskCycles := reducer.relations.DetectCycles("subtask")

		if len(blocksCycles) == 0 && len(subtaskCycles) == 0 {
			fmt.Println("No conflicts detected")
			return nil
		}

		// Display cycles
		if len(blocksCycles) > 0 {
			fmt.Printf("Blocks cycles detected (%d):\n\n", len(blocksCycles))
			for i, cycle := range blocksCycles {
				fmt.Printf("Cycle %d:\n", i+1)

				// Convert UUIDs to task IDs for display
				var taskIDs []string
				for _, uuid := range cycle {
					task, ok := reducer.GetTask(uuid)
					if ok {
						taskIDs = append(taskIDs, task.TaskID)
					} else {
						taskIDs = append(taskIDs, uuid)
					}
				}

				fmt.Printf("  %s\n", strings.Join(taskIDs, " → "))
				fmt.Println()

				// Show fix hint
				if len(taskIDs) > 0 {
					fmt.Printf("  Fix: tk relate remove %s blocks %s\n\n", taskIDs[len(taskIDs)-1], taskIDs[0])
				}
			}
		}

		if len(subtaskCycles) > 0 {
			fmt.Printf("Subtask cycles detected (%d):\n\n", len(subtaskCycles))
			for i, cycle := range subtaskCycles {
				fmt.Printf("Cycle %d:\n", i+1)

				// Convert UUIDs to task IDs for display
				var taskIDs []string
				for _, uuid := range cycle {
					task, ok := reducer.GetTask(uuid)
					if ok {
						taskIDs = append(taskIDs, task.TaskID)
					} else {
						taskIDs = append(taskIDs, uuid)
					}
				}

				fmt.Printf("  %s\n", strings.Join(taskIDs, " → "))
				fmt.Println()

				// Show fix hint
				if len(taskIDs) > 0 {
					fmt.Printf("  Fix: tk relate remove %s subtask %s\n\n", taskIDs[len(taskIDs)-1], taskIDs[0])
				}
			}
		}

		return nil
	},
}

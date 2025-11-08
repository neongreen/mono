package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	conflicts_pkg "github.com/neongreen/mono/tk/cmd/conflicts"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/reducer"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var conflictsCmd = &cobra.Command{
	Use:   "conflicts [task-id]",
	Short: "Show relation conflicts for a task or all tasks",
	Long: `Show relation conflicts including cycles in blocks and subtask relations.

If a task ID is provided, shows conflicts for that task only.
Otherwise, shows all conflicts in the database.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

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

		// Check for cycles in blocks and subtasks
		blocksCycles := reducer.Relations().DetectCycles("blocks")
		subtaskCycles := reducer.Relations().DetectCycles("subtask")

		if jsonOutput {
			type CycleOutput struct {
				Type    string   `json:"type"`
				TaskIDs []string `json:"task_ids"`
				Fix     string   `json:"fix"`
			}

			output := struct {
				Conflicts []CycleOutput `json:"conflicts"`
			}{
				Conflicts: []CycleOutput{},
			}

			for _, cycle := range blocksCycles {
				var taskIDs []string
				for _, uuid := range cycle {
					task, ok := reducer.GetTask(uuid)
					if ok {
						taskIDs = append(taskIDs, task.TaskID)
					} else {
						taskIDs = append(taskIDs, uuid)
					}
				}

				fix := ""
				if len(taskIDs) > 0 {
					fix = fmt.Sprintf("tk relate remove %s blocks %s", taskIDs[len(taskIDs)-1], taskIDs[0])
				}

				output.Conflicts = append(output.Conflicts, CycleOutput{
					Type:    "blocks",
					TaskIDs: taskIDs,
					Fix:     fix,
				})
			}

			for _, cycle := range subtaskCycles {
				var taskIDs []string
				for _, uuid := range cycle {
					task, ok := reducer.GetTask(uuid)
					if ok {
						taskIDs = append(taskIDs, task.TaskID)
					} else {
						taskIDs = append(taskIDs, uuid)
					}
				}

				fix := ""
				if len(taskIDs) > 0 {
					fix = fmt.Sprintf("tk relate remove %s subtask %s", taskIDs[len(taskIDs)-1], taskIDs[0])
				}

				output.Conflicts = append(output.Conflicts, CycleOutput{
					Type:    "subtask",
					TaskIDs: taskIDs,
					Fix:     fix,
				})
			}

			data, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal output: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		if len(blocksCycles) == 0 && len(subtaskCycles) == 0 {
			fmt.Println("No conflicts detected")
			return nil
		}

		// Display cycles
		displayCycles(reducer, "blocks", blocksCycles)
		displayCycles(reducer, "subtask", subtaskCycles)

		return nil
	},
}

func init() {
	conflictsCmd.Flags().Bool("json", false, "Output as JSON")
	conflictsCmd.AddCommand(conflicts_pkg.NumbersCmd)
}

// displayCycles displays detected cycles for a given relation type
func displayCycles(reducer *reducer.Reducer, cycleType string, cycles [][]string) {
	if len(cycles) == 0 {
		return
	}

	titleCaser := cases.Title(language.English)
	fmt.Printf("%s cycles detected (%d):\n\n", titleCaser.String(cycleType), len(cycles))
	for i, cycle := range cycles {
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
			fmt.Printf("  Fix: tk relate remove %s %s %s\n\n", taskIDs[len(taskIDs)-1], cycleType, taskIDs[0])
		}
	}
}

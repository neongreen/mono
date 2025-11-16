package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

var relateBlockersCmd = &cobra.Command{
	Use:   "relate-blockers [task-id]",
	Short: "List blockers for a task",
	Long:  `List all direct and transitive blockers for a task.`,
	Args:  cobra.ExactArgs(1),
	RunE:  blockersCmdImpl,
}

var blockersCmd = &cobra.Command{
	Use:   "blockers [task-id]",
	Short: "List blockers for a task (alias for relate-blockers)",
	Args:  cobra.ExactArgs(1),
	RunE:  blockersCmdImpl,
}

func blockersCmdImpl(cmd *cobra.Command, args []string) error {
	taskID := args[0]
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

	// Resolve task ID to UUID
	taskUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(taskID))
	if err != nil {
		return err
	}

	if _, ok := reducer.GetTask(taskUUID); !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	displayID, err := database.RenderTaskDisplayID(db, taskUUID)
	if err != nil {
		displayID = taskID
	}

	// Get transitive blockers
	maxDepth := 10
	blockers := utils.GetTransitiveBlockers(reducer.Relations(), taskUUID, reducer.Tasks(), config.Blocking.BlockingAxis, config.Blocking.DoneStates, maxDepth)

	if jsonOutput {
		type BlockerOutput struct {
			TaskID   string `json:"task_id"`
			Distance int    `json:"distance"`
			Title    string `json:"title"`
		}

		output := struct {
			TaskID   string          `json:"task_id"`
			Blockers []BlockerOutput `json:"blockers"`
		}{
			TaskID:   displayID,
			Blockers: []BlockerOutput{},
		}

		for _, blocker := range blockers {
			blockerDisplay, err := database.RenderTaskDisplayID(db, blocker.TaskUUID)
			if err != nil {
				blockerDisplay = blocker.TaskDisplayID
			}
			output.Blockers = append(output.Blockers, BlockerOutput{
				TaskID:   blockerDisplay,
				Distance: blocker.Distance,
				Title:    blocker.Title,
			})
		}

		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal output: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	if len(blockers) == 0 {
		fmt.Printf("Task %s has no blockers\n", displayID)
		return nil
	}

	fmt.Printf("Blockers for %s:\n\n", displayID)

	// Print blockers in a table
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Distance", "Task ID", "Title"})
	t.SetStyle(table.StyleLight)

	for _, blocker := range blockers {
		blockerDisplay, err := database.RenderTaskDisplayID(db, blocker.TaskUUID)
		if err != nil {
			blockerDisplay = blocker.TaskDisplayID
		}
		t.AppendRow(table.Row{blocker.Distance, blockerDisplay, blocker.Title})
	}

	fmt.Println(t.Render())

	return nil
}

func init() {
	relateBlockersCmd.Flags().Bool("json", false, "Output as JSON")
	blockersCmd.Flags().Bool("json", false, "Output as JSON")
}

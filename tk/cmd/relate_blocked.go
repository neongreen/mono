package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var relateBlockedCmd = &cobra.Command{
	Use:   "relate-blocked",
	Short: "List all blocked tasks",
	Long:  `List all tasks that are currently blocked.`,
	RunE:  blockedCmdImpl,
}

var blockedCmd = &cobra.Command{
	Use:   "blocked",
	Short: "List all blocked tasks (alias for relate-blocked)",
	RunE:  blockedCmdImpl,
}

func blockedCmdImpl(cmd *cobra.Command, args []string) error {
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

	// Filter blocked tasks
	allTasks := reducer.GetAllTasks()
	var blockedTasks []*types.Task
	for _, task := range allTasks {
		if task.Blocked {
			blockedTasks = append(blockedTasks, task)
		}
	}

	if jsonOutput {
		type BlockedTaskOutput struct {
			TaskID   string   `json:"task_id"`
			Title    string   `json:"title"`
			Blockers []string `json:"blockers"`
		}

		output := struct {
			Count int                 `json:"count"`
			Tasks []BlockedTaskOutput `json:"tasks"`
		}{
			Count: len(blockedTasks),
			Tasks: []BlockedTaskOutput{},
		}

		for _, task := range blockedTasks {
			var blockerIDs []string
			for _, b := range task.Blockers {
				// Render the display ID for each blocker
				blockerID, err := database.RenderTaskDisplayID(db, b.TaskUUID)
				if err != nil || blockerID == "" {
					blockerID = b.TaskUUID
				}
				blockerIDs = append(blockerIDs, blockerID)
			}
			output.Tasks = append(output.Tasks, BlockedTaskOutput{
				TaskID:   task.TaskDisplayID,
				Title:    task.Title,
				Blockers: blockerIDs,
			})
		}

		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal output: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	if len(blockedTasks) == 0 {
		fmt.Println("No blocked tasks")
		return nil
	}

	fmt.Printf("Blocked tasks (%d):\n\n", len(blockedTasks))

	// Print blocked tasks in a table
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Task ID", "Title", "Blockers"})
	t.SetStyle(table.StyleLight)

	for _, task := range blockedTasks {
		blockerSummary := ""
		if len(task.Blockers) > 0 {
			var blockerIDs []string
			for _, b := range task.Blockers {
				// Render the display ID for each blocker
				blockerID, err := database.RenderTaskDisplayID(db, b.TaskUUID)
				if err != nil || blockerID == "" {
					blockerID = b.TaskUUID
				}
				blockerIDs = append(blockerIDs, blockerID)
			}
			blockerSummary = strings.Join(blockerIDs, ", ")
		}
		t.AppendRow(table.Row{task.TaskDisplayID, task.Title, blockerSummary})
	}

	fmt.Println(t.Render())

	return nil
}

func init() {
	relateBlockedCmd.Flags().Bool("json", false, "Output as JSON")
	blockedCmd.Flags().Bool("json", false, "Output as JSON")
}

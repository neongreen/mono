package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var blockersCmd = &cobra.Command{
	Use:   "blockers [task-id]",
	Short: "List blockers for a task",
	Long:  `List all direct and transitive blockers for a task.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]
		jsonOutput, _ := cmd.Flags().GetBool("json")

		db, err := OpenExistingDB()
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
		taskUUID, err := database.ResolveTaskReference(db, taskID)
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
				blockerDisplay, err := database.RenderTaskDisplayID(db, blocker.TaskID)
				if err != nil {
					blockerDisplay = blocker.TaskID
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
			blockerDisplay, err := database.RenderTaskDisplayID(db, blocker.TaskID)
			if err != nil {
				blockerDisplay = blocker.TaskID
			}
			t.AppendRow(table.Row{blocker.Distance, blockerDisplay, blocker.Title})
		}

		fmt.Println(t.Render())

		return nil
	},
}

var blockedCmd = &cobra.Command{
	Use:   "blocked",
	Short: "List all blocked tasks",
	Long:  `List all tasks that are currently blocked.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		db, err := OpenExistingDB()
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
					blockerIDs = append(blockerIDs, b.TaskID)
				}
				output.Tasks = append(output.Tasks, BlockedTaskOutput{
					TaskID:   task.TaskID,
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
					blockerIDs = append(blockerIDs, b.TaskID)
				}
				blockerSummary = strings.Join(blockerIDs, ", ")
			}
			t.AppendRow(table.Row{task.TaskID, task.Title, blockerSummary})
		}

		fmt.Println(t.Render())

		return nil
	},
}

func init() {
	blockersCmd.Flags().Bool("json", false, "Output as JSON")
	blockedCmd.Flags().Bool("json", false, "Output as JSON")
}

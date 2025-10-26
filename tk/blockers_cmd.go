package main

import (
	"fmt"
	"strings"

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

		// Resolve task ID to UUID
		taskUUID, err := db.ResolveTaskIDToUUID(taskID)
		if err != nil {
			return err
		}

		task, ok := reducer.GetTask(taskUUID)
		if !ok {
			return fmt.Errorf("task not found: %s", taskID)
		}

		// Get transitive blockers
		maxDepth := 10
		blockers := reducer.relations.GetTransitiveBlockers(taskUUID, reducer.tasks, config.Blocking.BlockingAxis, config.Blocking.DoneStates, maxDepth)

		if len(blockers) == 0 {
			fmt.Printf("Task %s has no blockers\n", task.TaskID)
			return nil
		}

		fmt.Printf("Blockers for %s:\n\n", task.TaskID)

		// Print blockers in a table
		t := table.NewWriter()
		t.AppendHeader(table.Row{"Distance", "Task ID", "Title"})
		t.SetStyle(table.StyleLight)

		for _, blocker := range blockers {
			t.AppendRow(table.Row{blocker.Distance, blocker.TaskID, blocker.Title})
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

		// Filter blocked tasks
		allTasks := reducer.GetAllTasks()
		var blockedTasks []*Task
		for _, task := range allTasks {
			if task.Blocked {
				blockedTasks = append(blockedTasks, task)
			}
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

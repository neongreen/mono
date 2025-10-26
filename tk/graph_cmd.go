package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var graphCmd = &cobra.Command{
	Use:   "graph [task-id]",
	Short: "Show a graph of task relations",
	Long:  `Show an ASCII tree of task relations (blocks or subtasks).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]
		relationType, _ := cmd.Flags().GetString("type")
		depth, _ := cmd.Flags().GetInt("depth")

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

		fmt.Printf("Task: %s - %s\n\n", task.TaskID, task.Title)

		// Build and print the tree
		visited := make(map[string]bool)
		printRelationTree(reducer, task, relationType, 0, depth, "", visited)

		return nil
	},
}

func printRelationTree(reducer *Reducer, task *Task, relationType string, currentDepth, maxDepth int, prefix string, visited map[string]bool) {
	if currentDepth > maxDepth {
		return
	}

	// Prevent infinite loops
	if visited[task.TaskUUID] {
		fmt.Printf("%s%s - %s (already shown above)\n", prefix, task.TaskID, task.Title)
		return
	}
	visited[task.TaskUUID] = true

	// Print current task
	if currentDepth > 0 {
		fmt.Printf("%s%s - %s", prefix, task.TaskID, task.Title)
		if task.Blocked {
			fmt.Printf(" ⛔")
		}
		fmt.Println()
	}

	// Get related tasks based on type
	var targets []RelationTarget
	switch relationType {
	case "blocks":
		targets = reducer.relations.GetOutgoingRelations(task.TaskUUID, "blocks")
	case "subtask":
		targets = reducer.relations.GetOutgoingRelations(task.TaskUUID, "subtask")
	default:
		targets = reducer.relations.GetOutgoingRelations(task.TaskUUID, relationType)
	}

	// Print children
	for i, target := range targets {
		childTask, ok := reducer.GetTask(target.TaskUUID)
		if !ok {
			continue
		}

		isLast := i == len(targets)-1
		var connector string
		var newPrefix string

		if isLast {
			connector = "└── "
			newPrefix = prefix + "    "
		} else {
			connector = "├── "
			newPrefix = prefix + "│   "
		}

		// Pass the full prefix with connector for printing this child,
		// and newPrefix for its children
		fullPrefix := prefix + connector
		if !visited[childTask.TaskUUID] {
			fmt.Printf("%s%s - %s", fullPrefix, childTask.TaskID, childTask.Title)
			if childTask.Blocked {
				fmt.Printf(" ⛔")
			}
			fmt.Println()
		}
		printRelationTreeImpl(reducer, childTask, relationType, currentDepth+1, maxDepth, newPrefix, visited)
	}
}

// Helper function for recursive printing
func printRelationTreeImpl(reducer *Reducer, task *Task, relationType string, currentDepth, maxDepth int, prefix string, visited map[string]bool) {
	if currentDepth > maxDepth {
		return
	}

	// Prevent infinite loops
	if visited[task.TaskUUID] {
		return
	}
	visited[task.TaskUUID] = true

	// Get related tasks based on type
	var targets []RelationTarget
	switch relationType {
	case "blocks":
		targets = reducer.relations.GetOutgoingRelations(task.TaskUUID, "blocks")
	case "subtask":
		targets = reducer.relations.GetOutgoingRelations(task.TaskUUID, "subtask")
	default:
		targets = reducer.relations.GetOutgoingRelations(task.TaskUUID, relationType)
	}

	// Print children
	for i, target := range targets {
		childTask, ok := reducer.GetTask(target.TaskUUID)
		if !ok {
			continue
		}

		isLast := i == len(targets)-1
		var connector string
		var newPrefix string

		if isLast {
			connector = "└── "
			newPrefix = prefix + "    "
		} else {
			connector = "├── "
			newPrefix = prefix + "│   "
		}

		fullPrefix := prefix + connector
		fmt.Printf("%s%s - %s", fullPrefix, childTask.TaskID, childTask.Title)
		if childTask.Blocked {
			fmt.Printf(" ⛔")
		}
		fmt.Println()

		printRelationTreeImpl(reducer, childTask, relationType, currentDepth+1, maxDepth, newPrefix, visited)
	}
}

func init() {
	graphCmd.Flags().String("type", "blocks", "Relation type to graph (blocks, subtask, related)")
	graphCmd.Flags().Int("depth", 10, "Maximum depth to traverse")
}

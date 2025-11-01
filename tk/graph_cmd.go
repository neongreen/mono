package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/neongreen/mono/tk/internal/reducer"
	"github.com/neongreen/mono/tk/internal/types"
)

var graphCmd = &cobra.Command{
	Use:   "graph [task-id]",
	Short: "Show a graph of task relations",
	Long:  `Show an ASCII tree of task relations (blocks or subtasks).`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		relationType, _ := cmd.Flags().GetString("type")
		depth, _ := cmd.Flags().GetInt("depth")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		db, err := OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Load config
		config, err := LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Use cached reducer for performance
		reducer, err := db.GetCachedReducerWithConfig(config)
		if err != nil {
			return err
		}

		// Resolve task ID to UUID
		taskUUID, err := ResolveTaskReference(db, taskRef)
		if err != nil {
			return err
		}

		task, ok := reducer.GetTask(taskUUID)
		if !ok {
			return fmt.Errorf("task not found: %s", taskRef)
		}

		displayID, err := RenderTaskDisplayID(db, taskUUID)
		if err != nil {
			displayID = taskRef
		}

		if jsonOutput {
			// Build graph structure
			type GraphNode struct {
				TaskID   string       `json:"task_id"`
				Title    string       `json:"title"`
				Blocked  bool         `json:"blocked"`
				Children []*GraphNode `json:"children,omitempty"`
			}

			visited := make(map[string]bool)
			var buildGraph func(t *types.Task, currentDepth, maxDepth int) *GraphNode
			buildGraph = func(t *types.Task, currentDepth, maxDepth int) *GraphNode {
				if currentDepth > maxDepth || visited[t.TaskUUID] {
					return nil
				}
				visited[t.TaskUUID] = true

				taskDisplay, err := RenderTaskDisplayID(db, t.TaskUUID)
				if err != nil {
					taskDisplay = t.TaskID
				}

				node := &GraphNode{
					TaskID:  taskDisplay,
					Title:   t.Title,
					Blocked: t.Blocked,
				}

				var targets []types.RelationTarget
				switch relationType {
				case "blocks":
					targets = reducer.Relations().GetOutgoingRelations(t.TaskUUID, "blocks")
				case "subtask":
					targets = reducer.Relations().GetOutgoingRelations(t.TaskUUID, "subtask")
				default:
					targets = reducer.Relations().GetOutgoingRelations(t.TaskUUID, relationType)
				}

				for _, target := range targets {
					if childTask, ok := reducer.GetTask(target.TaskUUID); ok {
						if childNode := buildGraph(childTask, currentDepth+1, maxDepth); childNode != nil {
							node.Children = append(node.Children, childNode)
						}
					}
				}

				return node
			}

			graph := buildGraph(task, 0, depth)
			data, err := json.MarshalIndent(graph, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal graph: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("Task: %s - %s\n\n", displayID, task.Title)

		// Build and print the tree
		visited := make(map[string]bool)
		printRelationTree(db, reducer, task, relationType, 0, depth, "", visited)

		return nil
	},
}

func printRelationTree(db *DB, reducer *reducer.Reducer, task *types.Task, relationType string, currentDepth, maxDepth int, prefix string, visited map[string]bool) {
	if currentDepth > maxDepth {
		return
	}

	// Prevent infinite loops
	if visited[task.TaskUUID] {
		display, err := RenderTaskDisplayID(db, task.TaskUUID)
		if err != nil {
			display = task.TaskID
		}
		fmt.Printf("%s%s - %s (already shown above)\n", prefix, display, task.Title)
		return
	}
	visited[task.TaskUUID] = true

	// Print current task
	if currentDepth > 0 {
		display, err := RenderTaskDisplayID(db, task.TaskUUID)
		if err != nil {
			display = task.TaskID
		}
		fmt.Printf("%s%s - %s", prefix, display, task.Title)
		if task.Blocked {
			fmt.Printf(" ⛔")
		}
		fmt.Println()
	}

	// Get related tasks based on type
	var targets []types.RelationTarget
	switch relationType {
	case "blocks":
		targets = reducer.Relations().GetOutgoingRelations(task.TaskUUID, "blocks")
	case "subtask":
		targets = reducer.Relations().GetOutgoingRelations(task.TaskUUID, "subtask")
	default:
		targets = reducer.Relations().GetOutgoingRelations(task.TaskUUID, relationType)
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
			childDisplay, err := RenderTaskDisplayID(db, childTask.TaskUUID)
			if err != nil {
				childDisplay = childTask.TaskID
			}
			fmt.Printf("%s%s - %s", fullPrefix, childDisplay, childTask.Title)
			if childTask.Blocked {
				fmt.Printf(" ⛔")
			}
			fmt.Println()
		}
		printRelationTreeImpl(db, reducer, childTask, relationType, currentDepth+1, maxDepth, newPrefix, visited)
	}
}

// Helper function for recursive printing
func printRelationTreeImpl(db *DB, reducer *reducer.Reducer, task *types.Task, relationType string, currentDepth, maxDepth int, prefix string, visited map[string]bool) {
	if currentDepth > maxDepth {
		return
	}

	// Prevent infinite loops
	if visited[task.TaskUUID] {
		return
	}
	visited[task.TaskUUID] = true

	// Get related tasks based on type
	var targets []types.RelationTarget
	switch relationType {
	case "blocks":
		targets = reducer.Relations().GetOutgoingRelations(task.TaskUUID, "blocks")
	case "subtask":
		targets = reducer.Relations().GetOutgoingRelations(task.TaskUUID, "subtask")
	default:
		targets = reducer.Relations().GetOutgoingRelations(task.TaskUUID, relationType)
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
		childDisplay, err := RenderTaskDisplayID(db, childTask.TaskUUID)
		if err != nil {
			childDisplay = childTask.TaskID
		}
		fmt.Printf("%s%s - %s", fullPrefix, childDisplay, childTask.Title)
		if childTask.Blocked {
			fmt.Printf(" ⛔")
		}
		fmt.Println()

		printRelationTreeImpl(db, reducer, childTask, relationType, currentDepth+1, maxDepth, newPrefix, visited)
	}
}

func init() {
	graphCmd.Flags().String("type", "blocks", "Relation type to graph (blocks, subtask, related)")
	graphCmd.Flags().Int("depth", 10, "Maximum depth to traverse")
	graphCmd.Flags().Bool("json", false, "Output as JSON")
}

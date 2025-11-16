package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/lib/pathlang"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/pathlang_resolver"
	"github.com/neongreen/mono/tk/internal/reducer"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query <path-expression>",
	Short: "Query tasks and projects using path syntax",
	Long: `Query tasks and projects using path syntax.

Examples:
  tk query /foo                # Show project 'foo'
  tk query /foo-13             # Show task 'foo-13'
  tk query /foo/tasks          # List all tasks in project 'foo'
  tk query /foo-13/subtasks    # Show subtasks of 'foo-13'
  tk query /foo-13/blockers    # Show tasks blocking 'foo-13'
  tk query /foo-13/notes       # Show notes for 'foo-13'
  tk query /foo-13/json        # Show task as JSON

Path syntax:
  /project-alias           # Project by alias
  /project-alias-number    # Task by display ID
  /project/tasks           # All tasks in project
  /task/subtasks           # Child tasks
  /task/blockers           # Blocking tasks
  /task/notes              # Task notes
  /resource/json           # JSON representation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pathStr := args[0]

		// Parse the path
		path, err := pathlang.Parse(pathStr)
		if err != nil {
			return fmt.Errorf("failed to parse path %q: %w", pathStr, err)
		}

		// Open database
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

		// Get reducer
		reducer, err := db.GetCachedReducerWithConfig(config)
		if err != nil {
			return err
		}

		// Create resolver
		resolver := pathlang_resolver.NewTkResolver(db, reducer)

		// Evaluate path
		ctx := context.Background()
		nodes, err := pathlang.Eval(ctx, resolver, path)
		if err != nil {
			return fmt.Errorf("failed to evaluate path: %w", err)
		}

		if len(nodes) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		// Display results
		return displayResults(db, reducer, nodes, pathStr)
	},
}

// displayResults outputs results in appropriate format
func displayResults(db *database.DB, reducer *reducer.Reducer, nodes []pathlang.Node, currentPath string) error {
	// Check if we're displaying JSON
	if len(nodes) == 1 {
		node := nodes[0].(*pathlang_resolver.Node)
		if node.Type == pathlang_resolver.NodeTypeJSON {
			return displayAsJSON(db, node)
		}
	}

	// Otherwise display normally
	for i, n := range nodes {
		if i > 0 {
			fmt.Println()
		}

		node := n.(*pathlang_resolver.Node)

		switch node.Type {
		case pathlang_resolver.NodeTypeProject:
			displayProject(db, node, currentPath)
		case pathlang_resolver.NodeTypeTask:
			displayTask(db, node, currentPath)
		case pathlang_resolver.NodeTypeTasks:
			// This shouldn't happen as tasks is a collection
			fmt.Println("Tasks collection")
		case pathlang_resolver.NodeTypeNotes:
			displayNotes(node, currentPath)
		default:
			fmt.Printf("Unknown node type: %s\n", node.Type)
		}
	}
	return nil
}

// displayAsJSON outputs a resource as JSON
func displayAsJSON(db *database.DB, node *pathlang_resolver.Node) error {
	var data interface{}

	switch node.Type {
	case pathlang_resolver.NodeTypeJSON:
		// Determine what to serialize based on what's in the node
		if node.Task != nil {
			data = node.Task
		} else if node.ProjectUID != "" {
			// Query project info
			var name string
			var projType string
			err := db.Db.QueryRow(`
				SELECT name, COALESCE(type, 'local') FROM projects WHERE project_uid = ?
			`, node.ProjectUID).Scan(&name, &projType)
			if err != nil {
				return fmt.Errorf("failed to query project: %w", err)
			}
			data = map[string]interface{}{
				"project_uid": node.ProjectUID,
				"name":        name,
				"type":        projType,
			}
		}
	}

	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func displayProject(db *database.DB, node *pathlang_resolver.Node, currentPath string) {
	if node.ProjectUID == "" {
		fmt.Println("Project (no UID)")
		return
	}

	// Query project info from database
	var name string
	var projType string
	err := db.Db.QueryRow(`
		SELECT name, COALESCE(type, 'local') FROM projects WHERE project_uid = ?
	`, node.ProjectUID).Scan(&name, &projType)

	if err != nil {
		fmt.Printf("Project UID: %s (error loading details: %v)\n", node.ProjectUID, err)
		return
	}

	// Try to get preferred alias
	alias, _ := database.PreferredAliasForProject(db, types.ProjectUID(node.ProjectUID))

	if alias != "" {
		fmt.Printf("Project: %s (%s)\n", boldText(alias), name)
	} else {
		fmt.Printf("Project: %s\n", boldText(name))
	}
	fmt.Printf("UID: %s\n", node.ProjectUID)
	if projType != "" && projType != "local" {
		fmt.Printf("Type: %s\n", projType)
	}

	// Show available sub-resources
	fmt.Println("\nAvailable paths:")
	// Use the access alias if available
	pathName := node.AccessAlias
	if pathName == "" {
		pathName = alias
	}
	if pathName == "" {
		pathName = name
	}
	fmt.Printf("  /%s/tasks  - List all tasks in this project\n", pathName)
	fmt.Printf("  /%s/json   - JSON representation\n", pathName)
}

func displayTask(db *database.DB, node *pathlang_resolver.Node, currentPath string) {
	if node.Task == nil {
		fmt.Println("Task (no data)")
		return
	}

	// Get display ID
	displayID, err := database.RenderTaskDisplayID(db, node.TaskUID)
	if err != nil {
		displayID = node.TaskUID
	}

	fmt.Printf("Task: %s\n", boldText(displayID))
	if node.Task.Title != "" {
		fmt.Printf("Title: %s\n", node.Task.Title)
	}

	status := getTaskStatus(node.Task)
	if status != "" {
		fmt.Printf("Status: %s\n", colorizeStatus(status))
	}

	// Show available sub-resources
	fmt.Println("\nAvailable paths:")
	fmt.Printf("  /%s/notes     - View notes\n", displayID)
	fmt.Printf("  /%s/json      - JSON representation\n", displayID)

	// Show subtasks if they exist
	if node.Task.Relations != nil && len(node.Task.Relations.Subtask.Children) > 0 {
		fmt.Printf("  /%s/subtasks  - View %d subtask(s)\n", displayID, len(node.Task.Relations.Subtask.Children))
	}

	// Show blockers if they exist
	if len(node.Task.Blockers) > 0 {
		fmt.Printf("  /%s/blockers  - View %d blocker(s)\n", displayID, len(node.Task.Blockers))
	}
}

func displayNotes(node *pathlang_resolver.Node, currentPath string) {
	if node.Task == nil || len(node.Task.Notes) == 0 {
		fmt.Println("No notes")
		return
	}

	// Get display ID for the task
	displayID := node.TaskUID
	// Try to get a better display ID if we have the full task info
	// (we should, since we're showing notes)

	fmt.Printf("Notes for task %s:\n", displayID)
	for i, note := range node.Task.Notes {
		if i > 0 {
			fmt.Println()
		}
		if note.Actor != "" || !note.Timestamp.IsZero() {
			fmt.Printf("[")
			if note.Actor != "" {
				fmt.Printf("%s", note.Actor)
			}
			if !note.Timestamp.IsZero() {
				if note.Actor != "" {
					fmt.Printf(" - ")
				}
				fmt.Printf("%s", note.Timestamp.Format("2006-01-02 15:04:05"))
			}
			fmt.Printf("]\n")
		}
		if note.Markdown != "" {
			fmt.Println(note.Markdown)
		}
	}

	// Show that JSON is available
	fmt.Println("\nAvailable paths:")
	fmt.Printf("  %s/json  - JSON representation\n", currentPath)
}

func getTaskStatus(task *types.Task) string {
	if task == nil || task.Axes == nil {
		return ""
	}
	axis, ok := task.Axes["generic"]
	if !ok {
		return ""
	}
	return axis.Effective
}

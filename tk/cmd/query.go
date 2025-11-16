package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/lib/pathlang"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/pathlang_resolver"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query <path-expression>",
	Short: "Query tasks and projects using path syntax",
	Long: `Query tasks and projects using path syntax.

Examples:
  tk query /foo          # Show project 'foo'
  tk query /foo-13       # Show task 'foo-13'
  tk query /foo-13/subtasks    # Show subtasks of 'foo-13'
  tk query /foo-13/blockers    # Show tasks blocking 'foo-13'
  tk query /foo-13/notes       # Show notes for 'foo-13'

Path syntax:
  /project-alias           # Project by alias
  /project-alias-number    # Task by display ID
  /task/subtasks           # Child tasks
  /task/blockers           # Blocking tasks
  /task/notes              # Task notes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
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
			if jsonOutput {
				fmt.Println("[]")
			} else {
				fmt.Println("No results found.")
			}
			return nil
		}

		// Display results
		if jsonOutput {
			return displayResultsJSON(db, nodes)
		}
		return displayResultsHuman(db, reducer, nodes)
	},
}

// displayResultsJSON outputs results in JSON format
func displayResultsJSON(db *database.DB, nodes []pathlang.Node) error {
	var results []map[string]any
	
	for _, n := range nodes {
		node := n.(*pathlang_resolver.Node)
		result := map[string]any{
			"type": string(node.Type),
		}
		
		switch node.Type {
		case pathlang_resolver.NodeTypeProject:
			// Query project info
			var name string
			var projType string
			err := db.Db.QueryRow(`
				SELECT name, COALESCE(type, 'local') FROM projects WHERE project_uid = ?
			`, node.ProjectUID).Scan(&name, &projType)
			if err == nil {
				result["project_uid"] = node.ProjectUID
				result["name"] = name
				if projType != "" && projType != "local" {
					result["type_str"] = projType
				}
			}
		case pathlang_resolver.NodeTypeTask:
			if node.Task != nil {
				result["task_uid"] = node.TaskUID
				result["project_uid"] = node.ProjectUID
				result["title"] = node.Task.Title
				result["status"] = getTaskStatus(node.Task)
			}
		case pathlang_resolver.NodeTypeNotes:
			if node.Task != nil {
				result["task_uid"] = node.TaskUID
				var notes []string
				for _, note := range node.Task.Notes {
					notes = append(notes, note.Markdown)
				}
				result["notes"] = notes
			}
		}
		
		results = append(results, result)
	}
	
	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

// displayResultsHuman outputs results in human-readable format
func displayResultsHuman(db *database.DB, reducer interface{}, nodes []pathlang.Node) error {
	for i, n := range nodes {
		if i > 0 {
			fmt.Println()
		}
		
		node := n.(*pathlang_resolver.Node)
		
		switch node.Type {
		case pathlang_resolver.NodeTypeProject:
			displayProject(db, node)
		case pathlang_resolver.NodeTypeTask:
			displayTask(db, node)
		case pathlang_resolver.NodeTypeNotes:
			displayNotes(node)
		default:
			fmt.Printf("Unknown node type: %s\n", node.Type)
		}
	}
	return nil
}

func displayProject(db *database.DB, node *pathlang_resolver.Node) {
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
}

func displayTask(db *database.DB, node *pathlang_resolver.Node) {
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
}

func displayNotes(node *pathlang_resolver.Node) {
	if node.Task == nil || len(node.Task.Notes) == 0 {
		fmt.Println("No notes")
		return
	}
	
	fmt.Printf("Notes for task %s:\n", node.TaskUID)
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

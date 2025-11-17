package cmd

import (
	"fmt"
	"strings"

	"github.com/neongreen/mono/lib/cli"
	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/pathlang_resolver"
	"github.com/neongreen/mono/tk/internal/reducer"
	"github.com/neongreen/mono/tk/internal/tasks"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
)

// handleTaskAction handles actions on task resources
func handleTaskAction(db *database.DB, reducer *reducer.Reducer, node *pathlang_resolver.Node, action string, args []string) error {
	switch action {
	case "status":
		return taskActionStatus(db, reducer, node, args)
	case "note":
		return taskActionNote(db, reducer, node, args)
	default:
		return fmt.Errorf("unknown task action: %s", action)
	}
}

// handleProjectAction handles actions on project resources
func handleProjectAction(db *database.DB, reducer *reducer.Reducer, node *pathlang_resolver.Node, action string, args []string) error {
	switch action {
	case "status":
		return projectActionStatus(db, reducer, node, args)
	case "info":
		return projectActionInfo(db, reducer, node, args)
	default:
		return fmt.Errorf("unknown project action: %s", action)
	}
}

// handleNotesAction handles actions on notes resources
func handleNotesAction(db *database.DB, reducer *reducer.Reducer, node *pathlang_resolver.Node, action string, args []string) error {
	switch action {
	case "add":
		return notesActionAdd(db, reducer, node, args)
	case "list":
		return notesActionList(db, reducer, node, args)
	default:
		return fmt.Errorf("unknown notes action: %s", action)
	}
}

// taskActionStatus displays the status of a task
func taskActionStatus(db *database.DB, reducer *reducer.Reducer, node *pathlang_resolver.Node, args []string) error {
	if node.Task == nil {
		return fmt.Errorf("task data not available")
	}

	displayID, err := database.RenderTaskDisplayID(db, node.TaskUID)
	if err != nil {
		displayID = node.TaskUID
	}

	fmt.Printf("Task: %s\n", cli.Header(displayID))
	if node.Task.Title != "" {
		fmt.Printf("Title: %s\n", node.Task.Title)
	}

	status := getTaskStatus(node.Task)
	if status != "" {
		fmt.Printf("Status: %s\n", colorizeStatus(status))
	}

	return nil
}

// taskActionNote adds a note to a task
func taskActionNote(db *database.DB, reducer *reducer.Reducer, node *pathlang_resolver.Node, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("note text required")
	}

	noteText := strings.Join(args, " ")

	currentUser, err := utils.GetCurrentUser()
	if err != nil {
		return err
	}

	if err := tasks.AddNote(db, node.TaskUID, noteText, currentUser, &clock.RealClock{}); err != nil {
		return fmt.Errorf("failed to add note: %w", err)
	}

	displayID, err := database.RenderTaskDisplayID(db, node.TaskUID)
	if err != nil {
		displayID = node.TaskUID
	}

	fmt.Printf("%s Note added to task %s\n", cli.Success("✓"), cli.Key(displayID))
	return nil
}

// projectActionStatus displays the status of a project
func projectActionStatus(db *database.DB, reducer *reducer.Reducer, node *pathlang_resolver.Node, args []string) error {
	// Query project info
	var name string
	var projType string
	err := db.Db.QueryRow(`
		SELECT name, COALESCE(type, 'local') FROM projects WHERE project_uid = ?
	`, node.ProjectUID).Scan(&name, &projType)
	if err != nil {
		return fmt.Errorf("failed to query project: %w", err)
	}

	// Try to get preferred alias
	alias, _ := database.PreferredAliasForProject(db, types.ProjectUID(node.ProjectUID))

	fmt.Printf("Project: %s\n", cli.Header(name))
	if alias != "" {
		fmt.Printf("Alias: %s\n", cli.Key(alias))
	}
	fmt.Printf("UID: %s\n", node.ProjectUID)
	if projType != "" && projType != "local" {
		fmt.Printf("Type: %s\n", projType)
	}

	// Count tasks in project
	var taskCount int
	err = db.Db.QueryRow(`
		SELECT COUNT(*) FROM task_display_ids WHERE project_uid = ?
	`, node.ProjectUID).Scan(&taskCount)
	if err == nil {
		fmt.Printf("Tasks: %d\n", taskCount)
	}

	return nil
}

// projectActionInfo displays detailed information about a project
func projectActionInfo(db *database.DB, reducer *reducer.Reducer, node *pathlang_resolver.Node, args []string) error {
	return projectActionStatus(db, reducer, node, args)
}

// notesActionAdd adds a note to a task
func notesActionAdd(db *database.DB, reducer *reducer.Reducer, node *pathlang_resolver.Node, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("note text required")
	}

	if node.Task == nil {
		return fmt.Errorf("task data not available")
	}

	noteText := strings.Join(args, " ")

	currentUser, err := utils.GetCurrentUser()
	if err != nil {
		return err
	}

	if err := tasks.AddNote(db, node.TaskUID, noteText, currentUser, &clock.RealClock{}); err != nil {
		return fmt.Errorf("failed to add note: %w", err)
	}

	displayID, err := database.RenderTaskDisplayID(db, node.TaskUID)
	if err != nil {
		displayID = node.TaskUID
	}

	fmt.Printf("%s Note added to task %s\n", cli.Success("✓"), cli.Key(displayID))
	return nil
}

// notesActionList lists all notes for a task
func notesActionList(db *database.DB, reducer *reducer.Reducer, node *pathlang_resolver.Node, args []string) error {
	if node.Task == nil || len(node.Task.Notes) == 0 {
		fmt.Println("No notes")
		return nil
	}

	displayID, err := database.RenderTaskDisplayID(db, node.TaskUID)
	if err != nil {
		displayID = node.TaskUID
	}

	fmt.Printf("Notes for task %s:\n", cli.Key(displayID))
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

	return nil
}

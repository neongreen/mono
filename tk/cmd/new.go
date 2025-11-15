package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/tasks"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var newCmd = &cobra.Command{
	Use:     "new [title]",
	Aliases: []string{"add"},
	Short:   "Create a new task",
	Long: `Create a new item (task, decision, resource, etc.).

Item kinds define what type of item you're creating. Default is 'task'.
Use 'tk schema list' to see available item kinds, or define custom ones with 'tk schema add item <name>'.

Examples:
  tk new "Fix bug"                           # Creates a task
  tk new "Use PostgreSQL" --kind decision    # Creates a decision
  tk new "API docs" --kind resource          # Creates a resource
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("title is required")
		}
		if len(args) > 1 {
			// User likely forgot quotes around multi-word title
			return fmt.Errorf("multi-word titles must be quoted. Example: tk new \"My task title\"")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		projectRef, _ := cmd.Flags().GetString("project")
		parentRef, _ := cmd.Flags().GetString("parent")
		itemKind, _ := cmd.Flags().GetString("kind")
		title := args[0]

		// Auto-detect project from "project: title" format only if -p was not explicitly specified
		if !cmd.Flags().Changed("project") {
			if idx := strings.Index(title, ": "); idx > 0 {
				prefix := title[:idx]
				restOfTitle := title[idx+2:]

				// Check if prefix is a valid project
				if _, err := database.ResolveProjectRef(db, types.NewProjectRef(prefix)); err == nil {
					projectRef = prefix
					title = restOfTitle
				}
			}
		}

		// Resolve project reference to UID
		projectUID, err := database.ResolveProjectRef(db, types.NewProjectRef(projectRef))
		if err != nil {
			// Provide helpful error message based on whether other projects exist
			return handleProjectNotFoundError(db, projectRef, cmd.Flags().Changed("project"))
		}

		currentUser, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Resolve parent task BEFORE creating the new task (if --parent is specified)
		var parentUUID string
		if parentRef != "" {
			parentUUID, err = database.ResolveTaskReference(db, types.NewTaskRef(parentRef))
			if err != nil {
				return fmt.Errorf("failed to resolve parent task %q: %w", parentRef, err)
			}
		}

		result, err := tasks.Create(db, tasks.CreateParams{
			ProjectUID: projectUID,
			Title:      title,
			ItemKind:   itemKind,
		}, currentUser, &clock.RealClock{})
		if err != nil {
			return err
		}

		fmt.Printf("Created task %s: %s\n", result.DisplayID, args[0])

		// If --parent is specified, create a parent relation
		if parentRef != "" {

			// Create relation.add event (parent relation: parentUUID is parent of childUUID)
			// parent(a,b) means a is parent of b, which is subtask(a,b)
			eventID, err := database.GenerateEventID(db)
			if err != nil {
				return err
			}

			lamportTS, err := db.GetNextLamportTS()
			if err != nil {
				return err
			}

			payload := types.RelationAddPayload{
				Src:  parentUUID,
				Type: "subtask",
				Dst:  string(result.TaskUID),
				Note: "",
			}
			payloadJSON, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("failed to marshal relation payload: %w", err)
			}

			event := types.Event{
				ID:        eventID,
				TS:        lamportTS,
				CreatedAt: time.Now(),
				Actor:     currentUser,
				Role:      "human",
				Kind:      "relation.add",
				Payload:   payloadJSON,
			}

			if err := db.InsertEvent(event); err != nil {
				return fmt.Errorf("failed to insert relation event: %w", err)
			}

			parentDisplay, err := database.RenderTaskDisplayID(db, parentUUID)
			if err != nil {
				parentDisplay = parentRef
			}

			var parentTitle string
			err = db.Db.QueryRow(`SELECT title FROM tasks WHERE task_uid = ?`, parentUUID).Scan(&parentTitle)
			if err != nil {
				parentTitle = ""
			}

			if parentTitle != "" {
				fmt.Printf("Set parent: %s (%s)\n", parentDisplay, parentTitle)
			} else {
				fmt.Printf("Set parent: %s\n", parentDisplay)
			}
		}

		return nil
	},
}

// handleProjectNotFoundError provides a helpful error message when a project is not found.
// It checks if other projects exist and tailors the message accordingly.
func handleProjectNotFoundError(db *database.DB, projectRef string, projectFlagWasExplicit bool) error {
	// Get all existing projects
	rows, err := db.Db.Query(`SELECT name FROM projects ORDER BY name`)
	if err != nil {
		return fmt.Errorf("project %q not found", projectRef)
	}
	defer rows.Close()

	var existingProjects []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		existingProjects = append(existingProjects, name)
	}

	// If no projects exist at all, suggest creating "me" as the default
	if len(existingProjects) == 0 {
		return fmt.Errorf(`project %q not found.

No projects exist yet. You can:
  1. Create the default "me" project (recommended for personal tasks):
     tk project create me "Personal tasks"

     This will become your default project.

  2. Create a custom project:
     tk project create <name> "<description>"

     Then use -p <name> when creating tasks, or set it as default in your config.`, projectRef)
	}

	// If trying to use "me" but it doesn't exist and other projects do exist
	if projectRef == "me" && !projectFlagWasExplicit {
		return fmt.Errorf(`project "me" not found (no -p flag was specified, so "me" is used by default).

Available projects: %s

You can:
  1. Specify which project to use with -p flag:
     tk new -p <project> "Task title"

  2. Create the "me" project to use as your default:
     tk project create me "Personal tasks"

     After creating "me", commands without -p will use it automatically.`, strings.Join(existingProjects, ", "))
	}

	// User explicitly specified a project that doesn't exist
	if len(existingProjects) > 0 {
		return fmt.Errorf(`project %q not found.

Available projects: %s

Create it with:
  tk project create %s "<description>"

Or use an existing project with -p:
  tk new -p <project> "Task title"`, projectRef, strings.Join(existingProjects, ", "), projectRef)
	}

	return fmt.Errorf("project %q not found", projectRef)
}

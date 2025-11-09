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

var newCmd = &cobra.Command{
	Use:     "new [title]",
	Aliases: []string{"add"},
	Short:   "Create a new task",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		projectRef, _ := cmd.Flags().GetString("project")
		parentRef, _ := cmd.Flags().GetString("parent")
		title := args[0]

		// Auto-detect project from "project: title" format if -p not specified
		if projectRef == "tk" { // Default project
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
			return fmt.Errorf("project/alias %q not found. Create it first with: tk project create %s", projectRef, projectRef)
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

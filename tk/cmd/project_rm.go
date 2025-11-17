package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var projectRmCmd = &cobra.Command{
	Use:   "project-rm",
	Short: "Delete a project (project-ref can be a UID, alias, or name)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Resolve the project reference to a ProjectUID
		ref := types.NewProjectRef(args[0])
		projectUID, err := database.ResolveProjectRef(db, ref)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}

		// Check if project has tasks
		force, _ := cmd.Flags().GetBool("force")
		var taskCount int
		err = db.Db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE project_uid = ?`, projectUID.String()).Scan(&taskCount)
		if err != nil {
			return fmt.Errorf("failed to count tasks in project: %w", err)
		}

		if taskCount > 0 && !force {
			return fmt.Errorf("cannot delete project %s: it has %d task(s). Use --force to delete anyway", ref, taskCount)
		}

		// Get current user
		actor, err := utils.GetCurrentUser()
		if err != nil {
			return err
		}

		// Create project.delete event
		payload := types.ProjectDeletePayload{
			ProjectUID: projectUID,
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		eventID, err := database.GenerateEventID(db)
		if err != nil {
			return fmt.Errorf("failed to generate event ID: %w", err)
		}

		ts, err := db.GetNextLamportTS()
		if err != nil {
			return fmt.Errorf("failed to get next lamport timestamp: %w", err)
		}

		event := types.Event{
			ID:        eventID,
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      "human",
			Kind:      string(types.EventKindProjectDelete),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event (removes the project from the table)
		if err := db.ProjectProjectDeleteEvent(event); err != nil {
			return fmt.Errorf("failed to project project deletion: %w", err)
		}

		if taskCount > 0 {
			fmt.Printf("Deleted project %s (and %d task(s))\n", projectUID, taskCount)
		} else {
			fmt.Printf("Deleted project %s\n", projectUID)
		}
		return nil
	},
}

func init() {
	projectRmCmd.Flags().Bool("force", false, "Force deletion even if project has tasks")
}

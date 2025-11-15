package migrate

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
var FixRelocateBugCmd = &cobra.Command{
	Use:   "fix-relocate-bug",
	Short: "Fix orphaned tasks from corrupt task.relocate events (tk-281)",
	Long: `Scans for synthetic projects created by the projection layer and offers to
clean them up by emitting proper correction events.

Synthetic projects are created when task.created or task.relocate events
reference project names instead of valid project UUIDs. This was a bug
in earlier versions (fixed in v5).

This command will:
  1. Find all synthetic projects
  2. For each synthetic project, offer to:
     - Recreate the project (emit project.created event)
     - Move tasks to an existing project
     - Leave as-is (keep synthetic)

This is OPTIONAL cleanup. Synthetic projects are fully functional, this
just makes the event log cleaner.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Find synthetic projects
		rows, err := db.Db.Query(`
			SELECT project_uid, name
			FROM projects
			WHERE is_synthetic = 1
			ORDER BY name
		`)
		if err != nil {
			return fmt.Errorf("failed to query synthetic projects: %w", err)
		}
		defer rows.Close()

		type SyntheticProject struct {
			UID  string
			Name string
		}

		var synthetics []SyntheticProject
		for rows.Next() {
			var uid, name string
			if err := rows.Scan(&uid, &name); err != nil {
				return fmt.Errorf("failed to scan synthetic project: %w", err)
			}
			synthetics = append(synthetics, SyntheticProject{UID: uid, Name: name})
		}
		if err := rows.Err(); err != nil {
			return err
		}

		if len(synthetics) == 0 {
			fmt.Println("✓ No synthetic projects found")
			return nil
		}

		fmt.Printf("Found %d synthetic project(s):\n\n", len(synthetics))

		// Process each synthetic project
		for _, sp := range synthetics {
			// Count tasks in this project
			var taskCount int
			err := db.Db.QueryRow(`
				SELECT COUNT(*) FROM tasks WHERE project_uid = ?
			`, sp.UID).Scan(&taskCount)
			if err != nil {
				return fmt.Errorf("failed to count tasks: %w", err)
			}

			fmt.Printf("Project: %s (%d tasks)\n", sp.Name, taskCount)
			fmt.Printf("  Current state: Synthetic (no project.created event)\n")
			fmt.Printf("  What would you like to do?\n")
			fmt.Printf("    1. Recreate project '%s' (emit project.created event)\n", sp.Name)
			fmt.Printf("    2. Skip (leave as synthetic)\n")
			fmt.Printf("\nChoice [1]: ")

			var choice string
			fmt.Scanln(&choice)
			if choice == "" {
				choice = "1"
			}

			switch choice {
			case "1":
				if err := recreateProject(db, sp.UID, sp.Name); err != nil {
					return fmt.Errorf("failed to recreate project: %w", err)
				}
				fmt.Printf("✓ Created project '%s'\n\n", sp.Name)

			case "2":
				fmt.Printf("Skipped\n\n")
				continue

			default:
				fmt.Printf("Invalid choice, skipping\n\n")
				continue
			}
		}

		return nil
	},
}

func recreateProject(db *database.DB, syntheticUID, name string) error {
	// Get current user
	actor, err := utils.GetCurrentUser()
	if err != nil {
		return err
	}

	// Create a NEW project.created event
	// Note: We create a NEW project UID (not reuse the synthetic one)
	// The synthetic project will be replaced on next projection rebuild
	newProjectUID := types.NewProjectUID()

	payload := types.ProjectCreatedPayload{
		ProjectUID:  newProjectUID.String(),
		Type:        "local",
		Name:        name,
		Description: fmt.Sprintf("Recreated from synthetic project (cleanup from tk-281)"),
		CreatedBy:   actor,
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
		return fmt.Errorf("failed to get lamport timestamp: %w", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindProjectCreated),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	if err := db.ProjectProjectCreatedEvent(event); err != nil {
		return fmt.Errorf("failed to project event: %w", err)
	}

	// Now move all tasks from synthetic project to new project
	// Get all tasks in synthetic project
	taskRows, err := db.Db.Query(`
		SELECT task_uid FROM tasks WHERE project_uid = ?
	`, syntheticUID)
	if err != nil {
		return fmt.Errorf("failed to query tasks: %w", err)
	}
	defer taskRows.Close()

	var taskUIDs []string
	for taskRows.Next() {
		var taskUID string
		if err := taskRows.Scan(&taskUID); err != nil {
			return fmt.Errorf("failed to scan task: %w", err)
		}
		taskUIDs = append(taskUIDs, taskUID)
	}
	if err := taskRows.Err(); err != nil {
		return err
	}

	// Create task.relocate events for each task
	for _, taskUID := range taskUIDs {
		// Get current number
		var oldNumber int64
		err := db.Db.QueryRow(`
			SELECT number FROM task_numbers WHERE task_uid = ?
		`, taskUID).Scan(&oldNumber)
		if err != nil {
			return fmt.Errorf("failed to get task number: %w", err)
		}

		relocatePayload := types.TaskRelocatePayload{
			TaskUID:        types.TaskUID(taskUID),
			FromProjectUID: types.ProjectUID(syntheticUID),
			ToProjectUID:   newProjectUID,
			NumberPolicy: types.NumberPolicyPayload{
				Mode:   "force",
				Number: oldNumber,
			},
		}

		relocateJSON, err := json.Marshal(relocatePayload)
		if err != nil {
			return fmt.Errorf("failed to marshal relocate payload: %w", err)
		}

		relocateEventID, err := database.GenerateEventID(db)
		if err != nil {
			return fmt.Errorf("failed to generate event ID: %w", err)
		}

		relocateTS, err := db.GetNextLamportTS()
		if err != nil {
			return fmt.Errorf("failed to get lamport timestamp: %w", err)
		}

		relocateEvent := types.Event{
			ID:        relocateEventID,
			TS:        relocateTS,
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      "human",
			Kind:      string(types.EventKindTaskRelocate),
			Payload:   relocateJSON,
		}

		if err := db.InsertEvent(relocateEvent); err != nil {
			return fmt.Errorf("failed to insert relocate event: %w", err)
		}

		if err := db.ProjectTaskRelocateEvent(relocateEvent); err != nil {
			return fmt.Errorf("failed to project relocate event: %w", err)
		}
	}

	// Delete the synthetic project (it will be replaced by the real one)
	_, err = db.Db.Exec(`DELETE FROM projects WHERE project_uid = ?`, syntheticUID)
	if err != nil {
		return fmt.Errorf("failed to delete synthetic project: %w", err)
	}

	return nil
}

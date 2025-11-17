package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show recent task status changes",
	Long: `Display a timeline of recent task status changes.

Shows when tasks were marked as done, moved to wip, etc.

Examples:
  # Show last 20 status changes
  tk history

  # Show last 50 status changes
  tk history --limit 50`,
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Query for recent task.status.set events
		query := `
			SELECT
				json_extract(payload, '$.task_uuid') as task_uuid,
				json_extract(payload, '$.state') as state,
				created_at
			FROM events
			WHERE kind = 'task.status.set'
				AND json_extract(payload, '$.axis') = 'generic'
			ORDER BY created_at DESC
			LIMIT ?
		`

		rows, err := db.Query(query, limit)
		if err != nil {
			return fmt.Errorf("failed to query status changes: %w", err)
		}
		defer rows.Close()

		type StatusChange struct {
			TaskUUID  string
			State     string
			Timestamp time.Time
		}

		var changes []StatusChange
		for rows.Next() {
			var change StatusChange
			var createdAtNs int64

			if err := rows.Scan(&change.TaskUUID, &change.State, &createdAtNs); err != nil {
				return fmt.Errorf("failed to scan row: %w", err)
			}

			change.Timestamp = time.Unix(0, createdAtNs)
			changes = append(changes, change)
		}

		if len(changes) == 0 {
			fmt.Println("No status changes found")
			return nil
		}

		// Get display IDs and titles for all tasks
		taskInfo := make(map[string]struct {
			DisplayID string
			Title     string
		})

		for _, change := range changes {
			if _, exists := taskInfo[change.TaskUUID]; !exists {
				displayID, err := database.RenderTaskDisplayID(db, change.TaskUUID)
				if err != nil {
					displayID = change.TaskUUID[:8]
				}

				var title string
				err = db.QueryRow(`
					SELECT title FROM tasks WHERE task_uid = ?
				`, change.TaskUUID).Scan(&title)
				if err != nil {
					title = ""
				}

				taskInfo[change.TaskUUID] = struct {
					DisplayID string
					Title     string
				}{displayID, title}
			}
		}

		// Print timeline
		for _, change := range changes {
			info := taskInfo[change.TaskUUID]
			timestamp := change.Timestamp.Format("15:04:05")

			// Colorize status
			stateStr := colorizeStatus(change.State)

			fmt.Printf("%s  %-12s → %-8s  %s\n",
				color.New(color.Faint).Sprint(timestamp),
				info.DisplayID,
				stateStr,
				truncate(info.Title, 60),
			)
		}

		return nil
	},
}

func init() {
	historyCmd.Flags().Int("limit", 20, "Number of recent changes to show")
	historyCmd.Flags().Bool("json", false, "Output as JSON")
}

// truncate truncates a string to maxLen, adding "..." if needed
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

var statuslineCmd = &cobra.Command{
	Use:   "statusline",
	Short: "Display compact task activity status line",
	Long: `Output a compact one-line summary of recent task activity.

Designed for use in status bars, shell prompts, or Claude Code status line.

Shows recent status changes with symbols:
  ✓ = marked done
  ○ = moved to wip/next
  × = closed

Also shows count of currently active (wip) tasks.

Examples:
  # Show last 5 changes (default)
  tk statusline

  # Show last 10 changes
  tk statusline --limit 10

  # In Claude Code settings.json:
  {
    "statusLine": {
      "command": "tk statusline"
    }
  }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		noColor, _ := cmd.Flags().GetBool("no-color")

		if noColor {
			color.NoColor = true
		}

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

		// Get display IDs for tasks
		taskDisplayIDs := make(map[string]string)
		for _, change := range changes {
			if _, exists := taskDisplayIDs[change.TaskUUID]; !exists {
				displayID, err := database.RenderTaskDisplayID(db, change.TaskUUID)
				if err != nil {
					displayID = change.TaskUUID[:8]
				}
				taskDisplayIDs[change.TaskUUID] = displayID
			}
		}

		// Count active (wip) tasks using reducer for accurate current state
		config, err := config_pkg.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		reducer, err := db.GetCachedReducerWithConfig(config)
		if err != nil {
			return fmt.Errorf("failed to get reducer: %w", err)
		}

		wipCount := 0
		for _, task := range reducer.GetAllTasks() {
			if axis, ok := task.Axes["generic"]; ok {
				if axis.Effective == "wip" {
					wipCount++
				}
			}
		}

		// Build output parts
		var parts []string
		for _, change := range changes {
			displayID := taskDisplayIDs[change.TaskUUID]
			symbol := getStatusSymbol(change.State)
			coloredOutput := colorizeStatusChange(symbol, displayID, change.State)
			parts = append(parts, coloredOutput)
		}

		// Add wip count if any
		if wipCount > 0 {
			wipPart := fmt.Sprintf("[%d wip]", wipCount)
			if !noColor {
				wipPart = color.New(color.FgYellow, color.Faint).Sprint(wipPart)
			}
			parts = append(parts, wipPart)
		}

		// Output single line
		if len(parts) > 0 {
			fmt.Println(strings.Join(parts, " "))
		}

		return nil
	},
}

func init() {
	statuslineCmd.Flags().Int("limit", 5, "Number of recent changes to show")
	statuslineCmd.Flags().Bool("no-color", false, "Disable colored output")
}

// getStatusSymbol returns a symbol for the given status
func getStatusSymbol(status string) string {
	switch status {
	case "done":
		return "✓"
	case "wip", "next":
		return "○"
	case "closed":
		return "×"
	default:
		return "↻"
	}
}

// colorizeStatusChange colorizes a status change based on the status
func colorizeStatusChange(symbol, displayID, status string) string {
	if color.NoColor {
		return symbol + displayID
	}

	var c *color.Color
	switch status {
	case "done":
		c = color.New(color.FgGreen)
	case "wip":
		c = color.New(color.FgYellow)
	case "next":
		c = color.New(color.FgCyan)
	case "closed":
		c = color.New(color.FgRed, color.Faint)
	default:
		c = color.New(color.FgBlue)
	}

	return c.Sprint(symbol + displayID)
}

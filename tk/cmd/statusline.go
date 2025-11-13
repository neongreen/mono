package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/termutil"
	"github.com/spf13/cobra"
)

var statuslineCmd = &cobra.Command{
	Use:   "statusline",
	Short: "Display active WIP tasks with recent updates",
	Long: `Output a compact summary of active (wip) tasks sorted by last update.

Designed for use in status bars, shell prompts, or Claude Code status line.

Shows each wip task with:
  - Task ID
  - Truncated title (first ~40 chars)
  - Time since last update (e.g., "2h ago", "5m ago")

Examples:
  # Show last 5 wip tasks (default)
  tk statusline

  # Show last 3 wip tasks
  tk statusline --limit 3

  # In Claude Code settings.json:
  {
    "statusLine": {
      "type": "command",
      "command": "FORCE_COLOR=1 tk statusline"
    }
  }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		noColor, _ := cmd.Flags().GetBool("no-color")

		// Force colors when FORCE_COLOR or CLICOLOR_FORCE is set (for statusline use)
		if os.Getenv("FORCE_COLOR") != "" || os.Getenv("CLICOLOR_FORCE") != "" {
			color.NoColor = false
		}

		if noColor {
			color.NoColor = true
		}

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		config, err := config_pkg.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		reducer, err := db.GetCachedReducerWithConfig(config)
		if err != nil {
			return fmt.Errorf("failed to get reducer: %w", err)
		}

		// Get all wip tasks
		var wipTasks []*struct {
			UUID      string
			DisplayID string
			Title     string
			UpdatedAt time.Time
		}

		for _, task := range reducer.GetAllTasks() {
			if axis, ok := task.Axes["generic"]; ok {
				if axis.Effective == "wip" {
					displayID, err := database.RenderTaskDisplayID(db, task.TaskUUID)
					if err != nil {
						displayID = task.TaskUUID[:8]
					}

					wipTasks = append(wipTasks, &struct {
						UUID      string
						DisplayID string
						Title     string
						UpdatedAt time.Time
					}{
						UUID:      task.TaskUUID,
						DisplayID: displayID,
						Title:     task.Title,
						UpdatedAt: task.UpdatedAt,
					})
				}
			}
		}

		// Sort by most recently updated
		for i := 0; i < len(wipTasks)-1; i++ {
			for j := i + 1; j < len(wipTasks); j++ {
				if wipTasks[j].UpdatedAt.After(wipTasks[i].UpdatedAt) {
					wipTasks[i], wipTasks[j] = wipTasks[j], wipTasks[i]
				}
			}
		}

		// Limit to requested number
		if len(wipTasks) > limit {
			wipTasks = wipTasks[:limit]
		}

		// Get terminal width for dynamic truncation
		termWidth := termutil.GetTerminalWidth()

		// Output each task on its own line
		now := time.Now()
		for _, task := range wipTasks {
			// Format relative time
			relativeTime := formatRelativeTime(now.Sub(task.UpdatedAt))

			// Calculate available width for title
			// Format: "ID<space>TITLE<space>TIME"
			// Reserve: ID width + 2 spaces + TIME width
			reservedWidth := len(task.DisplayID) + 2 + len(relativeTime)
			availableWidth := termWidth - reservedWidth

			// Don't truncate if terminal is very wide or if we have enough space
			// Only truncate if title would exceed available width
			title := task.Title
			if termWidth > 0 && len(title) > availableWidth && availableWidth > 0 {
				title = title[:availableWidth-3] + "..."
			}

			// Colorize output
			var line string
			if !color.NoColor {
				idColor := color.New(color.FgYellow, color.Bold)
				titleColor := color.New(color.FgWhite)
				timeColor := color.New(color.FgCyan, color.Faint)
				line = fmt.Sprintf("%s %s %s",
					idColor.Sprint(task.DisplayID),
					titleColor.Sprint(title),
					timeColor.Sprint(relativeTime))
			} else {
				line = fmt.Sprintf("%s %s %s", task.DisplayID, title, relativeTime)
			}

			fmt.Println(line)
		}

		return nil
	},
}

func init() {
	statuslineCmd.Flags().Int("limit", 5, "Number of wip tasks to show")
	statuslineCmd.Flags().Bool("no-color", false, "Disable colored output")
}

// formatRelativeTime formats a duration as a human-readable relative time
func formatRelativeTime(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		return fmt.Sprintf("%dh ago", hours)
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%dd ago", days)
}

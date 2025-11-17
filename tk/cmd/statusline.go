package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/termutil"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Display command; outputs formatted text, not structured data
var statuslineCmd = &cobra.Command{
	Use:   "statusline",
	Short: "Display recently updated tasks",
	Long: `Output a compact summary of recently updated tasks (any status).

Designed for use in status bars, shell prompts, or Claude Code status line.

Shows each task with:
  - Task ID
  - Truncated title (dynamically sized to terminal width)
  - Time since last update (e.g., "2h ago", "5m ago")

Examples:
  # Show last 5 recently updated tasks (default)
  tk statusline

  # Show last 3 tasks
  tk statusline --limit 3

  # Only show tasks updated in last 24 hours
  tk statusline --max-age 24h

  # In Claude Code settings.json:
  {
    "statusLine": {
      "type": "command",
      "command": "FORCE_COLOR=1 tk statusline --limit 3 --max-age 24h"
    }
  }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		noColor, _ := cmd.Flags().GetBool("no-color")
		maxAgeStr, _ := cmd.Flags().GetString("max-age")

		// Force colors when FORCE_COLOR or CLICOLOR_FORCE is set (for statusline use)
		if os.Getenv("FORCE_COLOR") != "" || os.Getenv("CLICOLOR_FORCE") != "" {
			color.NoColor = false
		}

		if noColor {
			color.NoColor = true
		}

		// Parse max-age duration
		var maxAge time.Duration
		if maxAgeStr != "" {
			var err error
			maxAge, err = time.ParseDuration(maxAgeStr)
			if err != nil {
				return fmt.Errorf("invalid --max-age duration: %w", err)
			}
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

		// Get all tasks (any status)
		var recentTasks []*struct {
			UUID      string
			DisplayID string
			Title     string
			Status    string
			UpdatedAt time.Time
		}

		now := time.Now()
		for _, task := range reducer.GetAllTasks() {
			// Skip tasks older than max-age if specified
			if maxAge > 0 && now.Sub(task.UpdatedAt) > maxAge {
				continue
			}

			displayID, err := database.RenderTaskDisplayID(db, task.TaskUUID)
			if err != nil {
				displayID = task.TaskUUID[:8]
			}

			// Get task status
			status := ""
			if axis, ok := task.Axes["generic"]; ok {
				status = axis.Effective
			}

			recentTasks = append(recentTasks, &struct {
				UUID      string
				DisplayID string
				Title     string
				Status    string
				UpdatedAt time.Time
			}{
				UUID:      task.TaskUUID,
				DisplayID: displayID,
				Title:     task.Title,
				Status:    status,
				UpdatedAt: task.UpdatedAt,
			})
		}

		// Sort by most recently updated
		for i := 0; i < len(recentTasks)-1; i++ {
			for j := i + 1; j < len(recentTasks); j++ {
				if recentTasks[j].UpdatedAt.After(recentTasks[i].UpdatedAt) {
					recentTasks[i], recentTasks[j] = recentTasks[j], recentTasks[i]
				}
			}
		}

		// Limit to requested number
		if len(recentTasks) > limit {
			recentTasks = recentTasks[:limit]
		}

		// Get terminal width for dynamic truncation
		termWidth := termutil.GetTerminalWidth()

		// Output each task on its own line
		for _, task := range recentTasks {
			// Format relative time
			relativeTime := formatRelativeTime(now.Sub(task.UpdatedAt))

			// Calculate available width for title
			// Format: "ID<space>TITLE<space>TIME"
			// Reserve: ID width + 2 spaces + TIME width
			reservedWidth := len(task.DisplayID) + 2 + len(relativeTime)
			availableWidth := termWidth - reservedWidth

			// For multi-line titles, show only first line
			title := task.Title
			if idx := strings.IndexAny(title, "\n\r"); idx >= 0 {
				title = title[:idx]
			}

			// Don't truncate if terminal is very wide or if we have enough space
			// Only truncate if title would exceed available width
			if termWidth > 0 && len(title) > availableWidth && availableWidth > 0 {
				title = title[:availableWidth-3] + "..."
			}

			// Colorize output based on status
			var line string
			if !color.NoColor {
				// Choose color based on task status
				var idColor *color.Color
				switch task.Status {
				case "done":
					idColor = color.New(color.FgGreen, color.Bold)
				case "wip":
					idColor = color.New(color.FgYellow, color.Bold)
				case "next":
					idColor = color.New(color.FgCyan, color.Bold)
				case "closed":
					idColor = color.New(color.FgRed, color.Faint)
				default:
					idColor = color.New(color.FgWhite, color.Bold)
				}

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
	statuslineCmd.Flags().Int("limit", 5, "Number of tasks to show")
	statuslineCmd.Flags().String("max-age", "2h", "Only show tasks updated within this duration (e.g., 24h, 7d)")
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

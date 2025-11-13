package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/fatih/color"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

// colorizeStatus returns a colored status string based on the status value
func colorizeStatus(status string) string {
	switch status {
	case "next":
		return blueStatus(status)
	case "wip":
		return yellowStatus(status)
	case "done", "fixed":
		return greenStatus(status)
	default:
		return status
	}
}

// getStatusStyle returns a lipgloss style for the given status
func getStatusStyle(status string) lipgloss.Style {
	switch status {
	case "next":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Blue
	case "wip":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
	case "done", "fixed":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green
	case "closed":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
	default:
		return lipgloss.NewStyle()
	}
}

// renderTaskTable renders a table of tasks using lipgloss/table
func renderTaskTable(db *database.DB, tasks []*types.Task, showAliases bool, termWidth int, widths *ColumnWidths) {
	// Build table rows
	var rows [][]string

	for _, task := range tasks {
		displayID, err := database.RenderTaskDisplayID(db, task.TaskUUID)
		if err != nil {
			displayID = task.TaskDisplayID
		}

		status := ""
		if axis, ok := task.Axes["generic"]; ok {
			status = colorizeStatus(axis.Effective)
		}

		// Extract priority
		priority := ""
		if meta, ok := task.Metadata["priority"]; ok {
			var p any
			if err := json.Unmarshal(meta.Effective, &p); err == nil {
				priority = fmt.Sprintf("%v", p)
			}
		}

		// Extract labels
		labelsStr := ""
		if meta, ok := task.Metadata["labels"]; ok {
			var labels []any
			if err := json.Unmarshal(meta.Effective, &labels); err == nil && len(labels) > 0 {
				var labelStrs []string
				for _, l := range labels {
					labelStrs = append(labelStrs, fmt.Sprintf("%v", l))
				}
				labelsStr = strings.Join(labelStrs, ", ")
			}
		}

		// Display empty titles as "(empty)"
		title := task.Title
		if title == "" {
			title = "(empty)"
		}

		// Truncate title to fit terminal width
		// Reserve space for other columns: ID(~10) + STATUS(~8) + P(~3) + LABELS(~15) + borders/padding(~15)
		if termWidth > 0 {
			reservedWidth := 50
			maxTitleWidth := termWidth - reservedWidth
			if maxTitleWidth < 30 {
				maxTitleWidth = 30 // Minimum title width
			}
			if len(title) > maxTitleWidth {
				title = title[:maxTitleWidth-3] + "..."
			}
		}

		row := []string{displayID, status, priority, labelsStr, title}
		if showAliases {
			// Insert aliases as second column
			aliasesStr := ""
			if len(task.Aliases) > 0 {
				var shortAliases []string
				for _, alias := range task.Aliases {
					shortAliases = append(shortAliases, database.FormatTaskID(db, alias))
				}
				aliasesStr = strings.Join(shortAliases, ", ")
			}
			row = []string{displayID, aliasesStr, status, priority, labelsStr, title}
		}
		rows = append(rows, row)
	}

	// Create lipgloss table
	headers := []string{"ID", "STATUS", "P", "LABELS", "TITLE"}
	if showAliases {
		headers = []string{"ID", "ALIASES", "STATUS", "P", "LABELS", "TITLE"}
	}

	// Create table with padding
	re := lipgloss.NewRenderer(os.Stdout)

	var baseStyle = re.NewStyle().Padding(0, 1)
	var headerStyle = baseStyle.Bold(true).Foreground(lipgloss.Color("240"))

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		BorderLeft(false).
		BorderRight(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return headerStyle
			}
			return baseStyle
		}).
		Headers(headers...).
		Rows(rows...)

	fmt.Println(t.Render())
}

// outputTasksJSON outputs tasks as a JSON array
// NOTE: This always returns a flat array of tasks, not grouped JSON.
// VSCode extension and other clients should group tasks on the client side.
// See tk-vscode/src/extension.ts:fetchTk for the client-side grouping logic.
func outputTasksJSON(db *database.DB, tasks []*types.Task) error {
	for _, task := range tasks {
		displayID, err := database.RenderTaskDisplayID(db, task.TaskUUID)
		if err != nil {
			displayID = task.TaskDisplayID
		}
		task.TaskDisplayID = displayID

		// Populate project UUID from database
		var projectUID string
		err = db.Db.QueryRow(`SELECT project_uid FROM tasks WHERE task_uid = ?`, task.TaskUUID).Scan(&projectUID)
		if err == nil {
			task.ProjectUUID = projectUID
		}
	}

	jsonOutput, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	fmt.Println(string(jsonOutput))
	return nil
}

var (
	// Color formatters for status display
	blueStatus   = color.New(color.FgBlue).SprintFunc()
	yellowStatus = color.New(color.FgYellow).SprintFunc()
	greenStatus  = color.New(color.FgGreen).SprintFunc()
	redText      = color.New(color.FgRed).SprintFunc()
	boldText     = color.New(color.Bold).SprintFunc()
)

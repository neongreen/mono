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

// padRight pads a string to the specified width with trailing spaces
func padRight(s string, width int) string {
	// Account for ANSI color codes in length calculation
	visibleLen := len(stripAnsiCodes(s))
	if visibleLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visibleLen)
}

// truncateOrPad truncates or pads a string to the specified width
func truncateOrPad(s string, width int) string {
	visibleLen := len(stripAnsiCodes(s))
	if visibleLen > width {
		// Truncate with ellipsis
		if width <= 3 {
			return s[:width]
		}
		return s[:width-3] + "..."
	}
	return padRight(s, width)
}

// stripAnsiCodes removes ANSI color codes for length calculation
func stripAnsiCodes(s string) string {
	// Simple regex to strip ANSI escape codes
	// This handles most common color codes
	var result strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			i++ // skip '['
			continue
		}
		if inEscape {
			if (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z') {
				inEscape = false
			}
			continue
		}
		result.WriteString(string(s[i]))
	}
	return result.String()
}

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

// renderTaskTable renders a table of tasks using lipgloss/table.
// If widths is nil, it will calculate widths from the given tasks.
// Pass a non-nil widths to use consistent column widths across multiple tables.
func renderTaskTable(db *database.DB, tasks []*types.Task, showAliases bool, termWidth int, widths *ColumnWidths) {
	// First pass: collect all raw cell values
	type cellData struct {
		displayID string
		aliases   string
		status    string
		priority  string
		labels    string
		title     string
	}

	cellValues := make([]cellData, len(tasks))

	for i, task := range tasks {
		displayID, err := database.RenderTaskDisplayID(db, task.TaskUUID)
		if err != nil {
			displayID = task.TaskDisplayID
		}
		cellValues[i].displayID = displayID

		status := ""
		if axis, ok := task.Axes["generic"]; ok {
			status = colorizeStatus(axis.Effective)
		}
		cellValues[i].status = status

		// Extract priority
		priority := ""
		if meta, ok := task.Metadata["priority"]; ok {
			var p any
			if err := json.Unmarshal(meta.Effective, &p); err == nil {
				priority = fmt.Sprintf("%v", p)
			}
		}
		cellValues[i].priority = priority

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
		cellValues[i].labels = labelsStr

		// Extract aliases
		aliasesStr := ""
		if showAliases && len(task.Aliases) > 0 {
			var shortAliases []string
			for _, alias := range task.Aliases {
				shortAliases = append(shortAliases, database.FormatTaskID(db, alias))
			}
			aliasesStr = strings.Join(shortAliases, ", ")
		}
		cellValues[i].aliases = aliasesStr

		// Display empty titles as "(empty)"
		title := task.Title
		if title == "" {
			title = "(empty)"
		}

		// Take only first line of title
		if idx := strings.Index(title, "\n"); idx != -1 {
			title = title[:idx]
		}
		cellValues[i].title = title
	}

	// Calculate optimal column widths if not provided
	var calculatedWidths ColumnWidths
	if widths == nil {
		displayIDs := make(map[string]string)
		for i, task := range tasks {
			displayIDs[task.TaskUUID] = cellValues[i].displayID
		}
		constraints := DefaultColumnConstraints(termWidth, showAliases)
		calculatedWidths = CalculateColumnWidths(tasks, displayIDs, constraints)
		widths = &calculatedWidths
	}

	// Build padded rows using calculated widths
	var rows [][]string
	for _, cell := range cellValues {
		var row []string
		if showAliases {
			row = []string{
				padRight(cell.displayID, widths.ID),
				padRight(cell.aliases, widths.Aliases),
				padRight(cell.status, widths.Status),
				padRight(cell.priority, widths.Priority),
				padRight(cell.labels, widths.Labels),
				truncateOrPad(cell.title, widths.Title),
			}
		} else {
			row = []string{
				padRight(cell.displayID, widths.ID),
				padRight(cell.status, widths.Status),
				padRight(cell.priority, widths.Priority),
				padRight(cell.labels, widths.Labels),
				truncateOrPad(cell.title, widths.Title),
			}
		}
		rows = append(rows, row)
	}

	// Create headers with same padding
	var headers []string
	if showAliases {
		headers = []string{
			padRight("ID", widths.ID),
			padRight("ALIASES", widths.Aliases),
			padRight("STATUS", widths.Status),
			padRight("P", widths.Priority),
			padRight("LABELS", widths.Labels),
			padRight("TITLE", widths.Title),
		}
	} else {
		headers = []string{
			padRight("ID", widths.ID),
			padRight("STATUS", widths.Status),
			padRight("P", widths.Priority),
			padRight("LABELS", widths.Labels),
			padRight("TITLE", widths.Title),
		}
	}

	// Create table with padding
	re := lipgloss.NewRenderer(os.Stdout)

	var baseStyle = re.NewStyle().Padding(0, 1)
	var headerStyle = baseStyle.Bold(true).Foreground(lipgloss.Color("240"))

	t := table.New().
		Width(termWidth).
		Wrap(true).
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

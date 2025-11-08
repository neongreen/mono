package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

// colorizeStatus returns a colored status string based on the status value
func colorizeStatus(status string) string {
	switch status {
	case "wip":
		return yellowStatus(status)
	case "done", "fixed":
		return greenStatus(status)
	default:
		return status
	}
}

// renderTaskTable renders a table of tasks with the specified configuration
// If widths is nil, it will calculate widths based on the provided tasks
func renderTaskTable(db *database.DB, tasks []*types.Task, showAliases bool, termWidth int, widths *ColumnWidths) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	if showAliases {
		t.AppendHeader(table.Row{"ID", "Aliases", "Status", "P", "Labels", "Title"})
	} else {
		t.AppendHeader(table.Row{"ID", "Status", "P", "Labels", "Title"})
	}

	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.Style().Options.DrawBorder = false

	// Calculate widths if not provided
	if widths == nil {
		// Build displayID map for width calculation
		displayIDs := make(map[string]string)
		for _, task := range tasks {
			displayID, err := database.RenderTaskDisplayID(db, task.TaskUUID)
			if err == nil {
				displayIDs[task.TaskUUID] = displayID
			}
		}

		// Calculate optimal column widths based on actual data
		constraints := DefaultColumnConstraints(termWidth, showAliases)
		calculatedWidths := CalculateColumnWidths(tasks, displayIDs, constraints)
		widths = &calculatedWidths
	}

	// Configure columns with calculated widths
	// Set both WidthMin and WidthMax to force fixed widths for consistency across groups
	if showAliases {
		t.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, AutoMerge: false, WidthMin: widths.ID, WidthMax: widths.ID},                                          // ID
			{Number: 2, AutoMerge: false, WidthMin: widths.Aliases, WidthMax: widths.Aliases},                                // Aliases
			{Number: 3, AutoMerge: false, WidthMin: widths.Status, WidthMax: widths.Status},                                  // Status
			{Number: 4, AutoMerge: false, WidthMin: widths.Priority, WidthMax: widths.Priority},                              // P
			{Number: 5, AutoMerge: false, WidthMin: widths.Labels, WidthMax: widths.Labels, WidthMaxEnforcer: text.WrapSoft}, // Labels
			{Number: 6, AutoMerge: false, WidthMin: widths.Title, WidthMax: widths.Title, WidthMaxEnforcer: text.WrapSoft},   // Title
		})
	} else {
		t.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, AutoMerge: false, WidthMin: widths.ID, WidthMax: widths.ID},                                          // ID
			{Number: 2, AutoMerge: false, WidthMin: widths.Status, WidthMax: widths.Status},                                  // Status
			{Number: 3, AutoMerge: false, WidthMin: widths.Priority, WidthMax: widths.Priority},                              // P
			{Number: 4, AutoMerge: false, WidthMin: widths.Labels, WidthMax: widths.Labels, WidthMaxEnforcer: text.WrapSoft}, // Labels
			{Number: 5, AutoMerge: false, WidthMin: widths.Title, WidthMax: widths.Title, WidthMaxEnforcer: text.WrapSoft},   // Title
		})
	}

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

		if showAliases {

			aliasesStr := ""
			if len(task.Aliases) > 0 {
				var shortAliases []string
				for _, alias := range task.Aliases {
					shortAliases = append(shortAliases, database.FormatTaskID(db, alias))
				}
				aliasesStr = strings.Join(shortAliases, ", ")
			}
			t.AppendRow(table.Row{displayID, aliasesStr, status, priority, labelsStr, task.Title})
		} else {
			t.AppendRow(table.Row{displayID, status, priority, labelsStr, task.Title})
		}
	}

	t.Render()
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
	yellowStatus = color.New(color.FgYellow).SprintFunc()
	greenStatus  = color.New(color.FgGreen).SprintFunc()
	redText      = color.New(color.FgRed).SprintFunc()
	boldText     = color.New(color.Bold).SprintFunc()
)

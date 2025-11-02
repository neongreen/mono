package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
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
func renderTaskTable(db *database.DB, tasks []*types.Task, showAliases bool, termWidth int) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	if showAliases {
		t.AppendHeader(table.Row{"ID", "Aliases", "Status", "Title"})
	} else {
		t.AppendHeader(table.Row{"ID", "Status", "Title"})
	}

	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.Style().Options.DrawBorder = false

	if showAliases {

		titleMaxWidth := termWidth - 60
		if titleMaxWidth < 20 {
			titleMaxWidth = 20
		}
		t.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, AutoMerge: false},
			{Number: 2, AutoMerge: false},
			{Number: 3, AutoMerge: false},
			{Number: 4, AutoMerge: false, WidthMax: titleMaxWidth, WidthMaxEnforcer: text.WrapSoft},
		})
	} else {

		titleMaxWidth := termWidth - 30
		if titleMaxWidth < 20 {
			titleMaxWidth = 20
		}
		t.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, AutoMerge: false},
			{Number: 2, AutoMerge: false},
			{Number: 3, AutoMerge: false, WidthMax: titleMaxWidth, WidthMaxEnforcer: text.WrapSoft},
		})
	}

	for _, task := range tasks {
		displayID, err := database.RenderTaskDisplayID(db, task.TaskUUID)
		if err != nil {
			displayID = task.TaskID
		}

		status := ""
		if axis, ok := task.Axes["generic"]; ok {
			status = colorizeStatus(axis.Effective)
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
			t.AppendRow(table.Row{displayID, aliasesStr, status, task.Title})
		} else {
			t.AppendRow(table.Row{displayID, status, task.Title})
		}
	}

	t.Render()
}

// outputTasksJSON outputs tasks as JSON, respecting grouping
func outputTasksJSON(db *database.DB, tasks []*types.Task, groupBy string) error {

	for _, task := range tasks {
		displayID, err := database.RenderTaskDisplayID(db, task.TaskUUID)
		if err != nil {
			displayID = task.TaskID
		}
		task.TaskID = displayID
	}

	type GroupedOutput struct {
		Group string        `json:"group"`
		Tasks []*types.Task `json:"tasks"`
	}

	type Output struct {
		Groups []GroupedOutput `json:"groups,omitempty"`
		Tasks  []*types.Task   `json:"tasks,omitempty"`
	}

	var output Output

	switch groupBy {
	case "prefix", "project":

		grouped := make(map[string][]*types.Task)
		var groupOrder []string

		// First, get all projects to ensure we include empty ones
		allProjects, err := database.GetAllProjectDisplayNames(db)
		if err != nil {
			return fmt.Errorf("failed to get projects: %w", err)
		}

		// Initialize all projects in the grouped map
		for _, displayName := range allProjects {
			grouped[displayName] = []*types.Task{}
			groupOrder = append(groupOrder, displayName)
		}

		// Now add tasks to their respective groups
		for _, task := range tasks {
			var groupKey string

			projectAlias, err := database.GetProjectAliasForTask(db, task.TaskUUID)
			if err != nil {
				groupKey = task.TaskUUID
			} else {
				groupKey = projectAlias
			}

			// If this is a new group (shouldn't happen if we got all projects), add it
			if _, exists := grouped[groupKey]; !exists {
				groupOrder = append(groupOrder, groupKey)
			}
			grouped[groupKey] = append(grouped[groupKey], task)
		}

		// Sort projects alphabetically
		sort.Strings(groupOrder)

		output.Groups = make([]GroupedOutput, 0, len(groupOrder))
		for _, groupKey := range groupOrder {
			output.Groups = append(output.Groups, GroupedOutput{
				Group: groupKey,
				Tasks: grouped[groupKey],
			})
		}

	case "status":

		grouped := make(map[string][]*types.Task)
		var groupOrder []string

		for _, task := range tasks {
			status := ""
			if axis, ok := task.Axes["generic"]; ok {
				status = axis.Effective
			}
			if status == "" {
				status = "(no status)"
			}

			if _, exists := grouped[status]; !exists {
				groupOrder = append(groupOrder, status)
			}
			grouped[status] = append(grouped[status], task)
		}

		output.Groups = make([]GroupedOutput, 0, len(groupOrder))
		for _, status := range groupOrder {
			output.Groups = append(output.Groups, GroupedOutput{
				Group: status,
				Tasks: grouped[status],
			})
		}

	case "none":

		output.Tasks = tasks

	default:
		return fmt.Errorf("invalid --group value: %s (must be project, status, or none)", groupBy)
	}

	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	fmt.Println(string(jsonOutput))
	return nil
}

var (
	dbPath string
	// Color formatters for status display
	yellowStatus = color.New(color.FgYellow).SprintFunc()
	greenStatus  = color.New(color.FgGreen).SprintFunc()
)

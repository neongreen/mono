package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/query"
	"github.com/neongreen/mono/tk/internal/status"
	"github.com/neongreen/mono/tk/internal/termutil"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv("FORCE_COLOR") != "" || os.Getenv("CLICOLOR_FORCE") != "" {
			color.NoColor = false
		}

		axisFilter, _ := cmd.Flags().GetString("axis")
		statusFilter, _ := cmd.Flags().GetString("status")
		sortBy, _ := cmd.Flags().GetString("sort")
		projectFilter, _ := cmd.Flags().GetStringSlice("project")
		showAliases, _ := cmd.Flags().GetBool("aliases")
		groupBy, _ := cmd.Flags().GetString("group")
		blockedOnly, _ := cmd.Flags().GetBool("blocked")
		unblockedOnly, _ := cmd.Flags().GetBool("unblocked")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		// Handle --status flag: convert to axis filter format
		// Supports comma-separated values (OR logic) and negation with !
		if statusFilter != "" {
			if axisFilter != "" && cmd.Flags().Changed("axis") {
				return fmt.Errorf("cannot use both --status and --axis flags")
			}
			// Parse status filter (handles multiple values and negation)
			parsedFilter, err := parseStatusFilter(statusFilter)
			if err != nil {
				return err
			}
			axisFilter = parsedFilter
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
			return err
		}

		tasks := reducer.GetAllTasks()

		// Build task UID set for project filtering
		var taskUIDSet map[string]bool
		if len(projectFilter) > 0 {
			taskIDs, err := db.GetTaskIDsByProjects(projectFilter)
			if err != nil {
				return err
			}
			taskUIDSet = make(map[string]bool)
			for _, id := range taskIDs {
				taskUIDSet[id] = true
			}
		}

		// Apply filters using query package
		filterOpts := query.FilterOptions{
			Projects:      projectFilter,
			AxisFilter:    axisFilter,
			BlockedOnly:   blockedOnly,
			UnblockedOnly: unblockedOnly,
		}
		tasks = query.FilterTasks(tasks, taskUIDSet, filterOpts)

		types.SortTasks(tasks, sortBy)

		if jsonOutput {
			// JSON output doesn't support grouping, always output flat list
			return outputTasksJSON(db, tasks)
		}

		termWidth := termutil.GetTerminalWidth()

		switch groupBy {
		case "project", "prefix":
			// Group tasks by project
			getProjectKey := func(task *types.Task) string {
				projectAlias, err := database.GetProjectAliasForTask(db, task.TaskUUID)
				if err != nil {
					return task.TaskUUID
				}
				return projectAlias
			}
			grouped, groupOrder := query.GroupTasks(tasks, groupBy, getProjectKey)

			// Only show all projects (including empty ones) when no project filter is specified
			if len(projectFilter) == 0 {
				allProjects, err := database.GetAllProjectDisplayNames(db)
				if err != nil {
					return fmt.Errorf("failed to get projects: %w", err)
				}

				for _, displayName := range allProjects {
					if _, exists := grouped[displayName]; !exists {
						grouped[displayName] = []*types.Task{}
						groupOrder = append(groupOrder, displayName)
					}
				}
			}

			sort.Strings(groupOrder)

			// Calculate column widths once for ALL tasks to ensure consistency across groups
			displayIDs := make(map[string]string)
			for _, task := range tasks {
				displayID, err := database.RenderTaskDisplayID(db, task.TaskUUID)
				if err == nil {
					displayIDs[task.TaskUUID] = displayID
				}
			}
			constraints := DefaultColumnConstraints(termWidth, showAliases)
			widths := CalculateColumnWidths(tasks, displayIDs, constraints)

			for i, groupKey := range groupOrder {
				if i > 0 {
					fmt.Println()
				}
				fmt.Printf("Project: %s\n", groupKey)
				renderTaskTable(db, grouped[groupKey], showAliases, termWidth, &widths)
			}

		case "status":
			// Group tasks by status
			getStatusKey := func(task *types.Task) string {
				status := ""
				if axis, ok := task.Axes["generic"]; ok {
					status = axis.Effective
				}
				if status == "" {
					status = "(no status)"
				}
				return status
			}
			grouped, groupOrder := query.GroupTasks(tasks, groupBy, getStatusKey)

			// Calculate column widths once for ALL tasks to ensure consistency across groups
			displayIDs := make(map[string]string)
			for _, task := range tasks {
				displayID, err := database.RenderTaskDisplayID(db, task.TaskUUID)
				if err == nil {
					displayIDs[task.TaskUUID] = displayID
				}
			}
			constraints := DefaultColumnConstraints(termWidth, showAliases)
			widths := CalculateColumnWidths(tasks, displayIDs, constraints)

			for i, status := range groupOrder {
				if i > 0 {
					fmt.Println()
				}
				fmt.Printf("Status: %s\n", colorizeStatus(status))
				renderTaskTable(db, grouped[status], showAliases, termWidth, &widths)
			}

		case "none":

			renderTaskTable(db, tasks, showAliases, termWidth, nil)

		default:
			return fmt.Errorf("invalid --group value: %s (must be project, status, or none)", groupBy)
		}

		return nil
	},
}

// parseStatusFilter converts status filter to axis filter format
// Handles single status, comma-separated statuses (OR), and negation with !
// Examples:
//   "wip" -> "generic:wip"
//   "todo,wip" -> handled by query package (OR logic) - TODO: tk-359
//   "!done" -> negation handled by query package - TODO: tk-360
func parseStatusFilter(statusFilter string) (string, error) {
	statusFilter = strings.TrimSpace(statusFilter)
	if statusFilter == "" {
		return "", nil
	}

	// Check for negation (! prefix)
	if strings.HasPrefix(statusFilter, "!") {
		// Negation support is tracked in tk-360
		return "", fmt.Errorf("status negation with ! is not yet implemented (tracked in tk-360)")
	}

	// Check for comma-separated values (OR logic)
	if strings.Contains(statusFilter, ",") {
		// Multi-status support is tracked in tk-359
		return "", fmt.Errorf("comma-separated statuses are not yet implemented (tracked in tk-359)")
	}

	// Single status - validate and convert to axis format
	normalized := status.NormalizeStatus(statusFilter)
	if !status.IsValidPredefinedStatus(normalized) {
		return "", fmt.Errorf(
			"invalid status: %q\nValid statuses are: %s",
			statusFilter,
			strings.Join(status.PredefinedStatuses, ", "),
		)
	}

	return fmt.Sprintf("generic:%s", normalized), nil
}

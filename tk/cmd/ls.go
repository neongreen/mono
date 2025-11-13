package cmd

import (
	"fmt"
	"os"
	"regexp"
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
		grepPattern, _ := cmd.Flags().GetString("grep")
		limit, _ := cmd.Flags().GetInt("limit")

		// Validate grep pattern if provided
		if grepPattern != "" {
			if _, err := regexp.Compile(grepPattern); err != nil {
				return fmt.Errorf("invalid regex pattern: %w", err)
			}
		}

		// If --sort was explicitly provided but --group was not, default to "none"
		if cmd.Flags().Changed("sort") && !cmd.Flags().Changed("group") {
			groupBy = "none"
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

		// Extract existing custom statuses from all tasks for validation
		existingCustomStatuses := status.GetExistingCustomStatusesFromTasks(tasks)

		// Handle --status flag: convert to axis filter format
		// Supports comma-separated values (OR logic) and negation with !
		var axisFilters []string
		if statusFilter != "" {
			if axisFilter != "" && cmd.Flags().Changed("axis") {
				return fmt.Errorf("cannot use both --status and --axis flags")
			}
			// Parse status filter (handles multiple values and negation)
			parsedFilters, err := parseStatusFilter(statusFilter, existingCustomStatuses)
			if err != nil {
				return err
			}
			axisFilters = parsedFilters
		} else if axisFilter != "" {
			// Backward compatibility: use single axis filter
			axisFilters = []string{axisFilter}
		}

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
			AxisFilters:   axisFilters,
			BlockedOnly:   blockedOnly,
			UnblockedOnly: unblockedOnly,
			GrepPattern:   grepPattern,
		}
		tasks = query.FilterTasks(tasks, taskUIDSet, filterOpts)

		types.SortTasks(tasks, sortBy)

		// Apply limit if specified
		if limit > 0 && len(tasks) > limit {
			tasks = tasks[:limit]
		}

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

			// Only show all projects (including empty ones) when no filters are active
			// If any filter is active (project, status, blocked, grep), only show projects with matches
			hasActiveFilters := len(projectFilter) > 0 || statusFilter != "" || axisFilter != "" ||
				blockedOnly || unblockedOnly || grepPattern != ""

			if len(projectFilter) == 0 && !hasActiveFilters {
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
//   "wip" -> []string{"generic:wip"}
//   "todo,wip" -> []string{"generic:todo", "generic:wip"}
//   "!done" -> negation handled by query package - TODO: tk-360
func parseStatusFilter(statusFilter string, existingCustomStatuses []string) ([]string, error) {
	statusFilter = strings.TrimSpace(statusFilter)
	if statusFilter == "" {
		return nil, nil
	}

	// Check for negation (! prefix)
	if strings.HasPrefix(statusFilter, "!") {
		// Negation support is tracked in tk-360
		return nil, fmt.Errorf("status negation with ! is not yet implemented (tracked in tk-360)")
	}

	// Split by comma for OR logic
	statusParts := strings.Split(statusFilter, ",")
	var axisFilters []string

	for _, s := range statusParts {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		// Validate each status with existing custom statuses from the project
		if err := status.ValidateStatus(s, false, existingCustomStatuses); err != nil {
			return nil, err
		}

		normalized := status.NormalizeStatus(s)
		axisFilters = append(axisFilters, fmt.Sprintf("generic:%s", normalized))
	}

	if len(axisFilters) == 0 {
		return nil, nil
	}

	return axisFilters, nil
}

package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
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
		sortBy, _ := cmd.Flags().GetString("sort")
		projectFilter, _ := cmd.Flags().GetStringSlice("project")
		showAliases, _ := cmd.Flags().GetBool("aliases")
		groupBy, _ := cmd.Flags().GetString("group")
		blockedOnly, _ := cmd.Flags().GetBool("blocked")
		unblockedOnly, _ := cmd.Flags().GetBool("unblocked")
		jsonOutput, _ := cmd.Flags().GetBool("json")

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

		if len(projectFilter) > 0 {
			taskIDs, err := db.GetTaskIDsByProjects(projectFilter)
			if err != nil {
				return err
			}

			// Filter tasks by project
			var filtered []*types.Task
			taskUIDSet := make(map[string]bool)
			for _, id := range taskIDs {
				taskUIDSet[id] = true
			}
			for _, task := range tasks {
				if taskUIDSet[task.TaskUUID] {
					filtered = append(filtered, task)
				}
			}
			tasks = filtered
		}

		if axisFilter != "" {
			parts := strings.Split(axisFilter, ":")
			if len(parts) != 2 {
				return fmt.Errorf("invalid axis filter format, expected axis:state")
			}
			axisName := parts[0]
			stateName := parts[1]

			var filtered []*types.Task
			for _, task := range tasks {
				if axis, ok := task.Axes[axisName]; ok {
					if axis.Effective == stateName {
						filtered = append(filtered, task)
					}
				}
			}
			tasks = filtered
		}

		if blockedOnly {
			var filtered []*types.Task
			for _, task := range tasks {
				if task.Blocked {
					filtered = append(filtered, task)
				}
			}
			tasks = filtered
		} else if unblockedOnly {
			var filtered []*types.Task
			for _, task := range tasks {
				if !task.Blocked {
					filtered = append(filtered, task)
				}
			}
			tasks = filtered
		}

		types.SortTasks(tasks, sortBy)

		if jsonOutput {
			return outputTasksJSON(db, tasks, groupBy)
		}

		termWidth := termutil.GetTerminalWidth()

		switch groupBy {
		case "project", "prefix":

			grouped := make(map[string][]*types.Task)
			var groupOrder []string // To maintain consistent order

			// Only show all projects (including empty ones) when no project filter is specified
			if len(projectFilter) == 0 {
				allProjects, err := database.GetAllProjectDisplayNames(db)
				if err != nil {
					return fmt.Errorf("failed to get projects: %w", err)
				}

				for _, displayName := range allProjects {
					grouped[displayName] = []*types.Task{}
					groupOrder = append(groupOrder, displayName)
				}
			}

			for _, task := range tasks {

				projectAlias, err := database.GetProjectAliasForTask(db, task.TaskUUID)
				var groupKey string
				if err != nil {
					groupKey = task.TaskUUID
				} else {
					groupKey = projectAlias
				}

				if _, exists := grouped[groupKey]; !exists {
					groupOrder = append(groupOrder, groupKey)
				}
				grouped[groupKey] = append(grouped[groupKey], task)
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

package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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

		db, err := OpenExistingDB()
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

		termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			termWidth = 80
		}

		switch groupBy {
		case "project", "prefix":

			grouped := make(map[string][]*types.Task)
			var groupOrder []string // To maintain consistent order

			allProjects, err := database.GetAllProjectDisplayNames(db)
			if err != nil {
				return fmt.Errorf("failed to get projects: %w", err)
			}

			for _, displayName := range allProjects {
				grouped[displayName] = []*types.Task{}
				groupOrder = append(groupOrder, displayName)
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

			for i, groupKey := range groupOrder {
				if i > 0 {
					fmt.Println()
				}
				fmt.Printf("Project: %s\n", groupKey)
				renderTaskTable(db, grouped[groupKey], showAliases, termWidth)
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

			for i, status := range groupOrder {
				if i > 0 {
					fmt.Println()
				}
				fmt.Printf("Status: %s\n", colorizeStatus(status))
				renderTaskTable(db, grouped[status], showAliases, termWidth)
			}

		case "none":

			renderTaskTable(db, tasks, showAliases, termWidth)

		default:
			return fmt.Errorf("invalid --group value: %s (must be project, status, or none)", groupBy)
		}

		return nil
	},
}

package cmd

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/neongreen/mono/lib/setlang"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/query"
	"github.com/neongreen/mono/tk/internal/status"
	"github.com/neongreen/mono/tk/internal/taskset"
	"github.com/neongreen/mono/tk/internal/termutil"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List tasks",
	Long: `List all items (tasks, decisions, resources, etc.).

By default, shows all item kinds. Use --kind to filter by specific kinds.
Use 'tk schema list' to see available item kinds.

Filtering:
  Flags (traditional):
    tk ls --status wip --kind decision --project tk

  Query language (new - inspired by jj revsets):
    tk ls --query "status(wip) & kind(decision) & project(tk)"
    tk ls --query "blocked() | status(wip)"
    tk ls --query "project(tk) ~ status(done)"

  Available functions:
    status(X)    - Filter by status (wip, done, next, closed)
    kind(X)      - Filter by item kind (task, decision, resource)
    project(X)   - Filter by project name
    blocked()    - All blocked tasks
    unblocked()  - All unblocked tasks
    author(X)    - Filter by creator
    title("X")   - Substring match in title
    all          - All tasks

  Operators:
    &  - AND (intersection)
    |  - OR (union)
    ~  - NOT/difference (A ~ B = A but not B)
    () - Grouping

Examples:
  tk ls                              # Show all items
  tk ls --kind decision              # Show only decisions (flags)
  tk ls --query "kind(decision)"     # Same using query
  tk ls --query "status(wip) & project(tk) ~ blocked()"  # Complex query
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv("FORCE_COLOR") != "" || os.Getenv("CLICOLOR_FORCE") != "" {
			color.NoColor = false
		}

		axisFilter, _ := cmd.Flags().GetString("axis")
		statusFilter, _ := cmd.Flags().GetString("status")
		kindFilter, _ := cmd.Flags().GetString("kind")
		queryExpr, _ := cmd.Flags().GetString("query")
		sortBy, _ := cmd.Flags().GetString("sort")
		projectFilter, _ := cmd.Flags().GetStringSlice("project")
		showAliases, _ := cmd.Flags().GetBool("aliases")
		groupBy, _ := cmd.Flags().GetString("group")
		blockedOnly, _ := cmd.Flags().GetBool("blocked")
		unblockedOnly, _ := cmd.Flags().GetBool("unblocked")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		grepPattern, _ := cmd.Flags().GetString("grep")
		limit, _ := cmd.Flags().GetInt("limit")
		inContainer, _ := cmd.Flags().GetString("in")

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

		// Handle --query flag if provided (taskset query language)
		if queryExpr != "" {
			// Build project UID to name map for query evaluation
			projectUIDToName := make(map[string]string)
			rows, err := db.Db.Query(`SELECT project_uid, name FROM projects`)
			if err != nil {
				return fmt.Errorf("failed to query projects: %w", err)
			}
			for rows.Next() {
				var uid, name string
				if err := rows.Scan(&uid, &name); err != nil {
					rows.Close()
					return fmt.Errorf("failed to scan project: %w", err)
				}
				projectUIDToName[uid] = name
			}
			rows.Close()

			// Create taskset context
			ctx := taskset.NewTaskContext(tasks, projectUIDToName)

			// Evaluate query
			resultSet, err := setlang.Eval(ctx, queryExpr)
			if err != nil {
				return fmt.Errorf("query evaluation failed: %w", err)
			}

			// Convert result set to task UID map for filtering
			queryTaskUIDs := make(map[string]bool)
			resultItems := resultSet.Items()
			for _, uid := range resultItems {
				queryTaskUIDs[uid] = true
			}

			// Debug: print query results if debug enabled
			if debugFlag {
				fmt.Fprintf(os.Stderr, "Query '%s' matched %d tasks\n", queryExpr, len(queryTaskUIDs))
			}

			// Filter tasks to only those matching query
			var filteredTasks []*types.Task
			for _, task := range tasks {
				if queryTaskUIDs[task.TaskUUID] {
					filteredTasks = append(filteredTasks, task)
				}
			}
			tasks = filteredTasks
		}

		// Extract existing custom statuses from all tasks for validation
		existingCustomStatuses := status.GetExistingCustomStatusesFromTasks(tasks)

		// Handle --status flag: convert to axis filter format
		// Supports comma-separated values (OR logic) and negation with !
		var axisFilters []string
		var negatedAxisState string
		if statusFilter != "" {
			if axisFilter != "" && cmd.Flags().Changed("axis") {
				return fmt.Errorf("cannot use both --status and --axis flags")
			}
			// Parse status filter (handles multiple values and negation)
			parsedFilters, negatedState, err := parseStatusFilter(statusFilter, existingCustomStatuses)
			if err != nil {
				return err
			}
			axisFilters = parsedFilters
			negatedAxisState = negatedState
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

		// Build task UID set for container filtering (v6+)
		if inContainer != "" {
			version, _ := db.GetDBVersion()
			if version < 6 {
				return fmt.Errorf("--in flag requires database v6 or higher (containers)")
			}

			// Query tasks in this container
			rows, err := db.Db.Query(`
				SELECT item_id
				FROM container_members
				WHERE container_id = ? AND removed = 0
			`, inContainer)
			if err != nil {
				return fmt.Errorf("failed to query container members: %w", err)
			}
			defer rows.Close()

			containerTaskUIDs := make(map[string]bool)
			for rows.Next() {
				var itemID string
				if err := rows.Scan(&itemID); err != nil {
					return fmt.Errorf("failed to scan member: %w", err)
				}
				containerTaskUIDs[itemID] = true
			}

			if len(containerTaskUIDs) == 0 {
				return fmt.Errorf("container %q not found or empty", inContainer)
			}

			// Merge with project filter if present
			if taskUIDSet != nil {
				// Intersect: only tasks in BOTH project AND container
				for uid := range taskUIDSet {
					if !containerTaskUIDs[uid] {
						delete(taskUIDSet, uid)
					}
				}
			} else {
				taskUIDSet = containerTaskUIDs
			}
		}

		// Parse item kind filter (supports comma-separated values for OR logic)
		var itemKinds []string
		if kindFilter != "" {
			itemKinds = strings.Split(kindFilter, ",")
			// Trim whitespace from each kind
			for i := range itemKinds {
				itemKinds[i] = strings.TrimSpace(itemKinds[i])
			}
		}

		// Apply filters using query package
		filterOpts := query.FilterOptions{
			Projects:         projectFilter,
			AxisFilters:      axisFilters,
			NegatedAxisState: negatedAxisState,
			BlockedOnly:      blockedOnly,
			UnblockedOnly:    unblockedOnly,
			GrepPattern:      grepPattern,
			ItemKinds:        itemKinds,
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

		// Pre-calculate column widths from ALL tasks for consistency across groups
		displayIDs := make(map[string]string)
		for _, task := range tasks {
			displayID, err := database.RenderTaskDisplayID(db, task.TaskUUID)
			if err == nil {
				displayIDs[task.TaskUUID] = displayID
			}
		}
		constraints := DefaultColumnConstraints(termWidth, showAliases)
		widths := CalculateColumnWidths(tasks, displayIDs, constraints)

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

			for i, status := range groupOrder {
				if i > 0 {
					fmt.Println()
				}
				fmt.Printf("Status: %s\n", colorizeStatus(status))
				renderTaskTable(db, grouped[status], showAliases, termWidth, &widths)
			}

		case "none":

			renderTaskTable(db, tasks, showAliases, termWidth, &widths)

		default:
			return fmt.Errorf("invalid --group value: %s (must be project, status, or none)", groupBy)
		}

		return nil
	},
}

// parseStatusFilter converts status filter to axis filter format
// Handles single status, comma-separated statuses (OR), and negation with !
// Examples:
//
//	"wip" -> (filters: []string{"generic:wip"}, negated: "")
//	"todo,wip" -> (filters: []string{"generic:todo", "generic:wip"}, negated: "")
//	"!done" -> (filters: nil, negated: "generic:done")
//
// Returns (axisFilters, negatedAxisState, error)
func parseStatusFilter(statusFilter string, existingCustomStatuses []string) ([]string, string, error) {
	statusFilter = strings.TrimSpace(statusFilter)
	if statusFilter == "" {
		return nil, "", nil
	}

	// Check for negation (! prefix)
	if after, ok := strings.CutPrefix(statusFilter, "!"); ok {
		negatedStatus := after
		negatedStatus = strings.TrimSpace(negatedStatus)

		// Validate negated status
		if err := status.ValidateStatus(negatedStatus, false, existingCustomStatuses); err != nil {
			return nil, "", fmt.Errorf("invalid negated status: %w", err)
		}

		normalized := status.NormalizeStatus(negatedStatus)
		return nil, fmt.Sprintf("generic:%s", normalized), nil
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
			return nil, "", err
		}

		normalized := status.NormalizeStatus(s)
		axisFilters = append(axisFilters, fmt.Sprintf("generic:%s", normalized))
	}

	if len(axisFilters) == 0 {
		return nil, "", nil
	}

	return axisFilters, "", nil
}

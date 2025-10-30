package main

import (
	"encoding/json"
	"fmt"
	"os"

	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	dbPath string
	// Color formatters for status display
	yellowStatus = color.New(color.FgYellow).SprintFunc()
	greenStatus  = color.New(color.FgGreen).SprintFunc()
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "tk",
	Short: "tk - system-wide event-sourced task tracker",
	Long:  `tk is a command-line tool that tracks tasks system-wide using an append-only event log in a single SQLite database.`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new tk database",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := GetDBPath()
		if err != nil {
			return err
		}

		if DBExists(path) {
			return fmt.Errorf("database already exists at %s", path)
		}

		db, err := OpenDB(path)
		if err != nil {
			return err
		}
		defer db.Close()

		if err := db.InitDB(); err != nil {
			return err
		}

		fmt.Printf("Database initialized at %s\n", path)
		return nil
	},
}

var dbPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the current database path",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := GetDBPath()
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	},
}

var newCmd = &cobra.Command{
	Use:   "new [title]",
	Short: "Create a new task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Always use project-based path
		return createTask(db, cmd, title)
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Manage task status and sync status",
}

var statusSetCmd = &cobra.Command{
	Use:   "set [task-id] [state]",
	Short: "Set task status",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		state := args[1]

		axis, _ := cmd.Flags().GetString("axis")
		role, _ := cmd.Flags().GetString("role")

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		taskUUID, err := ResolveTaskReference(db, taskRef)
		if err != nil {
			return err
		}

		displayID, err := RenderTaskDisplayID(db, taskUUID)
		if err != nil {
			displayID = taskRef
		}

		currentUser, err := getCurrentUser()
		if err != nil {
			return err
		}

		eventID, err := GenerateEventID(db)
		if err != nil {
			return err
		}

		// Get next Lamport timestamp from DB
		lamportTS, err := db.GetNextLamportTS()
		if err != nil {
			return err
		}

		payload := TaskStatusSetPayload{
			TaskUUID: taskUUID,
			TaskID:   taskRef,
			Axis:     axis,
			State:    state,
			Role:     role,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		now := time.Now()
		event := Event{
			ID:        eventID,
			TS:        lamportTS,
			CreatedAt: now,
			Actor:     currentUser,
			Role:      role,
			Kind:      "task.status.set",
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return err
		}

		fmt.Printf("Set status for task %s: %s=%s\n", displayID, axis, state)
		return nil
	},
}

var noteCmd = &cobra.Command{
	Use:   "note [task-id] [text]",
	Short: "Add a note to a task",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]
		text := args[1]

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		taskUUID, err := ResolveTaskReference(db, taskRef)
		if err != nil {
			return err
		}

		displayID, err := RenderTaskDisplayID(db, taskUUID)
		if err != nil {
			displayID = taskRef
		}

		currentUser, err := getCurrentUser()
		if err != nil {
			return err
		}

		eventID, err := GenerateEventID(db)
		if err != nil {
			return err
		}

		// Get next Lamport timestamp from DB
		lamportTS, err := db.GetNextLamportTS()
		if err != nil {
			return err
		}

		payload := TaskNoteAddPayload{
			TaskUUID: taskUUID,
			TaskID:   taskRef,
			Markdown: text,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		now := time.Now()
		event := Event{
			ID:        eventID,
			TS:        lamportTS,
			CreatedAt: now,
			Actor:     currentUser,
			Role:      "human",
			Kind:      "task.note.add",
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return err
		}

		fmt.Printf("Added note to task %s\n", displayID)
		return nil
	},
}

var viewCmd = &cobra.Command{
	Use:   "view [task-id]",
	Short: "View task details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskRef := args[0]

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Load config for relation processing
		config, err := LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Build reducer to get task and all its IDs (current + aliases)
		// Use cached reducer for performance
		reducer, err := db.GetCachedReducerWithConfig(config)
		if err != nil {
			return err
		}

		taskUUID, err := ResolveTaskReference(db, taskRef)
		if err != nil {
			return err
		}

		task, ok := reducer.GetTask(taskUUID)
		if !ok {
			return fmt.Errorf("task not found: %s", taskRef)
		}

		displayID, err := RenderTaskDisplayID(db, taskUUID)
		if err != nil {
			displayID = taskRef
		}

		taskCopy := *task
		taskCopy.TaskID = displayID

		output, err := json.MarshalIndent(taskCopy, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal task: %w", err)
		}

		fmt.Println(string(output))
		return nil
	},
}

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Respect color environment variables
		if os.Getenv("FORCE_COLOR") != "" || os.Getenv("CLICOLOR_FORCE") != "" {
			color.NoColor = false
		}

		axisFilter, _ := cmd.Flags().GetString("axis")
		sortBy, _ := cmd.Flags().GetString("sort")
		prefixFilter, _ := cmd.Flags().GetStringSlice("prefix")
		showAliases, _ := cmd.Flags().GetBool("aliases")
		groupBy, _ := cmd.Flags().GetString("group")
		blockedOnly, _ := cmd.Flags().GetBool("blocked")
		unblockedOnly, _ := cmd.Flags().GetBool("unblocked")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Load config for relation processing
		config, err := LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Use cached reducer for performance
		reducer, err := db.GetCachedReducerWithConfig(config)
		if err != nil {
			return err
		}

		tasks := reducer.GetAllTasks()

		// Filter by project alias if specified
		if len(prefixFilter) > 0 {
			// Filter by project alias (--prefix flag filters by project alias)
			taskIDs, err := db.GetTaskIDsByPrefixes(prefixFilter)
			if err != nil {
				return err
			}

			// Filter tasks by project alias
			var filtered []*Task
			taskIDSet := make(map[string]bool)
			for _, id := range taskIDs {
				taskIDSet[id] = true
			}
			for _, task := range tasks {
				if taskIDSet[task.TaskID] {
					filtered = append(filtered, task)
				}
			}
			tasks = filtered
		}

		// Filter by axis if specified
		if axisFilter != "" {
			parts := strings.Split(axisFilter, ":")
			if len(parts) != 2 {
				return fmt.Errorf("invalid axis filter format, expected axis:state")
			}
			axisName := parts[0]
			stateName := parts[1]

			var filtered []*Task
			for _, task := range tasks {
				if axis, ok := task.Axes[axisName]; ok {
					if axis.Effective == stateName {
						filtered = append(filtered, task)
					}
				}
			}
			tasks = filtered
		}

		// Filter by blocked status if specified
		if blockedOnly {
			var filtered []*Task
			for _, task := range tasks {
				if task.Blocked {
					filtered = append(filtered, task)
				}
			}
			tasks = filtered
		} else if unblockedOnly {
			var filtered []*Task
			for _, task := range tasks {
				if !task.Blocked {
					filtered = append(filtered, task)
				}
			}
			tasks = filtered
		}

		// Sort tasks based on the --sort flag
		sortTasks(tasks, sortBy)

		// JSON output mode
		if jsonOutput {
			return outputTasksJSON(db, tasks, groupBy)
		}

		// Get terminal width for wrapping
		termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			termWidth = 80 // default width if terminal size cannot be determined
		}

		// Group and render tasks based on groupBy flag
		switch groupBy {
		case "prefix":
			// Group tasks by project
			grouped := make(map[string][]*Task)
			var groupOrder []string // To maintain consistent order

			for _, task := range tasks {
				// Group by project alias
				projectAlias, err := getProjectAliasForTask(db, task.TaskUUID)
				var groupKey string
				if err != nil {
					groupKey = task.TaskUUID // Fallback to UID
				} else {
					groupKey = projectAlias
				}

				if _, exists := grouped[groupKey]; !exists {
					groupOrder = append(groupOrder, groupKey)
				}
				grouped[groupKey] = append(grouped[groupKey], task)
			}

			// Render a table for each project group
			for i, groupKey := range groupOrder {
				if i > 0 {
					fmt.Println() // Add blank line between tables
				}
				fmt.Printf("Project: %s\n", groupKey)
				renderTaskTable(db, grouped[groupKey], showAliases, termWidth)
			}

		case "status":
			// Group tasks by status
			grouped := make(map[string][]*Task)
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

			// Render a table for each status group
			for i, status := range groupOrder {
				if i > 0 {
					fmt.Println() // Add blank line between tables
				}
				fmt.Printf("Status: %s\n", colorizeStatus(status))
				renderTaskTable(db, grouped[status], showAliases, termWidth)
			}

		case "none":
			// No grouping - render single table
			renderTaskTable(db, tasks, showAliases, termWidth)

		default:
			return fmt.Errorf("invalid --group value: %s (must be prefix, status, or none)", groupBy)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "Database commands",
	}
	dbCmd.AddCommand(dbPathCmd)
	rootCmd.AddCommand(dbCmd)

	newCmd.Flags().String("project", "tk", "Project alias or UID to use")
	rootCmd.AddCommand(newCmd)

	statusSetCmd.Flags().String("axis", "generic", "Status axis")
	statusSetCmd.Flags().String("role", "human", "Actor role")
	statusCmd.AddCommand(statusSetCmd)
	statusCmd.AddCommand(statusSyncCmd)
	rootCmd.AddCommand(statusCmd)

	rootCmd.AddCommand(noteCmd)
	rootCmd.AddCommand(viewCmd)

	lsCmd.Flags().String("axis", "", "Filter by axis:state")
	lsCmd.Flags().String("sort", "created", "Sort order: created, id, or title (default: created)")
	lsCmd.Flags().StringSlice("prefix", []string{}, "Filter by project alias (can be specified multiple times)")
	lsCmd.Flags().Bool("aliases", false, "Show task aliases")
	lsCmd.Flags().String("group", "prefix", "Group tasks by: prefix, status, or none (default: prefix)")
	lsCmd.Flags().Bool("blocked", false, "Show only blocked tasks")
	lsCmd.Flags().Bool("unblocked", false, "Show only unblocked tasks")
	lsCmd.Flags().Bool("json", false, "Output tasks as JSON")
	rootCmd.AddCommand(lsCmd)

	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(mvCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(idCmd)
	rootCmd.AddCommand(relateCmd)
	rootCmd.AddCommand(dupCmd)
	rootCmd.AddCommand(blockersCmd)
	rootCmd.AddCommand(blockedCmd)
	rootCmd.AddCommand(graphCmd)
	rootCmd.AddCommand(conflictsCmd)
	rootCmd.AddCommand(nodeCmd)
	rootCmd.AddCommand(remoteCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(ingestCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(eventsCmd)
	rootCmd.AddCommand(adminCmd)
	rootCmd.AddCommand(projectCmd)
}

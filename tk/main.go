package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
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

		// Check database version
		version, err := db.GetDBVersion()
		if err != nil {
			return err
		}

		if version >= v4SpecVersion {
			// V4 path: use --project flag
			return createTaskV4(db, cmd, title)
		} else {
			// V1/V2 path: use --prefix flag (legacy)
			return createTaskLegacy(db, cmd, title)
		}
	},
}

func createTaskV4(db *DB, cmd *cobra.Command, title string) error {
	projectFlag, _ := cmd.Flags().GetString("project")

	// Get current user and node
	currentUser, err := getCurrentUser()
	if err != nil {
		return err
	}

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return err
	}

	// Resolve project
	var projectUID string

	// Check if it's a project UID or alias
	if strings.HasPrefix(projectFlag, "prj_") {
		// It's a project UID
		projectUID = projectFlag
	} else {
		// It's an alias, look it up
		err = db.db.QueryRow(`
			SELECT project_uid FROM project_aliases 
			WHERE alias = ? AND node = ?
		`, projectFlag, nodeID).Scan(&projectUID)
		if err != nil {
			return fmt.Errorf("project/alias %q not found. Create it first with: tk project create <name> --alias %s", projectFlag, projectFlag)
		}
	}

	// Generate task UID
	taskUID := NewTaskUID()

	// Compute proposed number (max + 1)
	var maxNumber int64
	err = db.db.QueryRow(`
		SELECT COALESCE(MAX(number), 0) FROM task_numbers 
		WHERE project_uid = ?
	`, projectUID).Scan(&maxNumber)
	if err != nil {
		return fmt.Errorf("failed to get max number: %w", err)
	}
	proposedNumber := maxNumber + 1

	// Create task.created (v4) event
	payload := TaskCreatedV4Payload{
		TaskUID:        string(taskUID),
		ProjectUID:     projectUID,
		ProposedNumber: proposedNumber,
		CreatedNode:    nodeID,
		Title:          title,
		CreatedBy:      currentUser,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	event := Event{
		ID:        generateEventID(db),
		TS:        getNextLamportTimestamp(db),
		CreatedAt: time.Now(),
		Actor:     currentUser,
		Role:      "human",
		Kind:      string(EventKindTaskCreated),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return err
	}

	// Project the event into tasks and task_numbers tables
	if err := db.ProjectTaskCreatedV4Event(event); err != nil {
		return fmt.Errorf("failed to project task: %w", err)
	}

	// Create task.number.set event
	numberPayload := TaskNumberSetPayload{
		TaskUID:    string(taskUID),
		ProjectUID: projectUID,
		Number:     proposedNumber,
		Reason:     "initial",
	}
	numberPayloadJSON, err := json.Marshal(numberPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal number payload: %w", err)
	}

	numberEvent := Event{
		ID:        generateEventID(db),
		TS:        getNextLamportTimestamp(db),
		CreatedAt: time.Now(),
		Actor:     currentUser,
		Role:      "human",
		Kind:      string(EventKindTaskNumberSet),
		Payload:   numberPayloadJSON,
	}

	if err := db.InsertEvent(numberEvent); err != nil {
		return fmt.Errorf("failed to insert number event: %w", err)
	}

	// Project the number event
	if err := db.ProjectTaskNumberSetEvent(numberEvent); err != nil {
		return fmt.Errorf("failed to project task number: %w", err)
	}

	// Display the task
	displayID := fmt.Sprintf("%s-%d", projectFlag, proposedNumber)
	fmt.Printf("Created task %s: %s\n", displayID, title)
	return nil
}

func createTaskLegacy(db *DB, cmd *cobra.Command, title string) error {
	prefix, _ := cmd.Flags().GetString("prefix")

	// Check if prefix exists
	exists, err := db.PrefixExists(prefix)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("prefix %q does not exist. Create it first with: tk prefix create %s <description>", prefix, prefix)
	}

	currentUser, err := getCurrentUser()
	if err != nil {
		return err
	}

	taskUUID := GenerateTaskUUID()
	taskID, err := GenerateTaskID(db, prefix)
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

	payload := TaskCreatedPayload{
		TaskUUID:  taskUUID,
		TaskID:    taskID,
		Title:     title,
		CreatedBy: currentUser,
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
		Kind:      "task.created",
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return err
	}

	// Get all task IDs for formatting (including the one we just created)
	allTaskIDs, err := db.GetAllTaskIDs()
	if err != nil {
		return err
	}

	displayID := FormatTaskID(taskID, allTaskIDs)
	fmt.Printf("Created task %s: %s\n", displayID, title)
	return nil
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
		// TODO: Consider caching reducer or adding task_index table for performance
		events, err := db.GetEvents()
		if err != nil {
			return err
		}

		reducer, err := BuildFromEventsWithConfig(events, config)
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

		events, err := db.GetEvents()
		if err != nil {
			return err
		}

		// Use BuildFromEventsWithConfig to get relations
		reducer, err := BuildFromEventsWithConfig(events, config)
		if err != nil {
			return err
		}

		tasks := reducer.GetAllTasks()

		// Get task IDs for filtering and formatting
		var taskIDs []string
		if len(prefixFilter) > 0 {
			taskIDs, err = db.GetTaskIDsByPrefixes(prefixFilter)
			if err != nil {
				return err
			}

			// Filter tasks by prefix
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
		} else {
			taskIDs, err = db.GetAllTaskIDs()
			if err != nil {
				return err
			}
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

		// Get terminal width for wrapping
		termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			termWidth = 80 // default width if terminal size cannot be determined
		}

		// Group and render tasks based on groupBy flag
		switch groupBy {
		case "prefix":
			// Group tasks by prefix/project
			grouped := make(map[string][]*Task)
			var groupOrder []string // To maintain consistent order

			// Check DB version to determine grouping strategy
			dbVersion, err := db.GetDBVersion()
			if err != nil {
				return fmt.Errorf("failed to get DB version: %w", err)
			}

			for _, task := range tasks {
				var groupKey string
				if dbVersion >= 4 {
					// V4: Group by project alias
					projectAlias, err := getProjectAliasForTask(db, task.TaskUUID)
					if err != nil {
						groupKey = task.TaskUUID // Fallback to UID
					} else {
						groupKey = projectAlias
					}
				} else {
					// V1/V2: Group by prefix from task ID
					groupKey = extractPrefix(task.TaskID)
				}

				if _, exists := grouped[groupKey]; !exists {
					groupOrder = append(groupOrder, groupKey)
				}
				grouped[groupKey] = append(grouped[groupKey], task)
			}

			// Render a table for each prefix/project group
			for i, groupKey := range groupOrder {
				if i > 0 {
					fmt.Println() // Add blank line between tables
				}
				if dbVersion >= 4 {
					fmt.Printf("Project: %s\n", groupKey)
				} else {
					fmt.Printf("Prefix: %s\n", groupKey)
				}
				renderTaskTable(db, grouped[groupKey], taskIDs, showAliases, termWidth)
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
				renderTaskTable(db, grouped[status], taskIDs, showAliases, termWidth)
			}

		case "none":
			// No grouping - render single table
			renderTaskTable(db, tasks, taskIDs, showAliases, termWidth)

		default:
			return fmt.Errorf("invalid --group value: %s (must be prefix, status, or none)", groupBy)
		}

		return nil
	},
}

// sortTasks sorts tasks based on the specified sort order
func sortTasks(tasks []*Task, sortBy string) {
	switch sortBy {
	case "created":
		// Sort by creation time (oldest first)
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
		})
	case "id":
		// Sort by task ID (lexicographic)
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].TaskID < tasks[j].TaskID
		})
	case "title":
		// Sort by title (lexicographic)
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Title < tasks[j].Title
		})
	default:
		// Default: sort by creation time (oldest first)
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
		})
	}
}

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

// extractPrefix extracts the prefix from a TaskID (format: prefix-number-node)
func extractPrefix(taskID string) string {
	parts := strings.Split(taskID, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// getProjectAliasForTask returns the preferred project alias for a task (v4)
func getProjectAliasForTask(db *DB, taskUID string) (string, error) {
	// Get project UID for this task
	var projectUID string
	err := db.db.QueryRow(`
		SELECT project_uid FROM tasks WHERE task_uid = ?
	`, taskUID).Scan(&projectUID)
	if err != nil {
		return "", fmt.Errorf("failed to get project for task %s: %w", taskUID, err)
	}

	// Get preferred alias for this project
	alias, err := preferredAliasForProject(db, projectUID)
	if err != nil {
		return "", err
	}

	if alias == "" {
		// No alias, return project UID
		return projectUID, nil
	}

	return alias, nil
}

// renderTaskTable renders a table of tasks with the specified configuration
func renderTaskTable(db *DB, tasks []*Task, taskIDs []string, showAliases bool, termWidth int) {
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

	// Configure column widths and wrapping
	if showAliases {
		// Reserve more space for aliases column
		titleMaxWidth := termWidth - 60
		if titleMaxWidth < 20 {
			titleMaxWidth = 20 // minimum width
		}
		t.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, AutoMerge: false}, // ID column
			{Number: 2, AutoMerge: false}, // Aliases column
			{Number: 3, AutoMerge: false}, // Status column
			{Number: 4, AutoMerge: false, WidthMax: titleMaxWidth, WidthMaxEnforcer: text.WrapSoft}, // Title column with wrapping
		})
	} else {
		// Reserve space for ID (~10 chars), Status (~10 chars), separators (~10 chars)
		titleMaxWidth := termWidth - 30
		if titleMaxWidth < 20 {
			titleMaxWidth = 20 // minimum width
		}
		t.SetColumnConfigs([]table.ColumnConfig{
			{Number: 1, AutoMerge: false}, // ID column
			{Number: 2, AutoMerge: false}, // Status column
			{Number: 3, AutoMerge: false, WidthMax: titleMaxWidth, WidthMaxEnforcer: text.WrapSoft}, // Title column with wrapping
		})
	}

	for _, task := range tasks {
		displayID, err := RenderTaskDisplayID(db, task.TaskUUID)
		if err != nil {
			displayID = task.TaskID
		}

		// Get status from generic axis (or empty if not present)
		status := ""
		if axis, ok := task.Axes["generic"]; ok {
			status = colorizeStatus(axis.Effective)
		}

		if showAliases {
			// Format aliases
			aliasesStr := ""
			if len(task.Aliases) > 0 {
				var shortAliases []string
				for _, alias := range task.Aliases {
					shortAliases = append(shortAliases, FormatTaskID(alias, taskIDs))
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

func init() {
	rootCmd.AddCommand(initCmd)

	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "Database commands",
	}
	dbCmd.AddCommand(dbPathCmd)
	rootCmd.AddCommand(dbCmd)

	newCmd.Flags().String("prefix", "tk", "Task prefix to use (v1/v2)")
	newCmd.Flags().String("project", "tk", "Project alias or UID to use (v4)")
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
	lsCmd.Flags().StringSlice("prefix", []string{}, "Filter by prefix (can be specified multiple times)")
	lsCmd.Flags().Bool("aliases", false, "Show task aliases")
	lsCmd.Flags().String("group", "prefix", "Group tasks by: prefix, status, or none (default: prefix)")
	lsCmd.Flags().Bool("blocked", false, "Show only blocked tasks")
	lsCmd.Flags().Bool("unblocked", false, "Show only unblocked tasks")
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
	rootCmd.AddCommand(prefixCmd)
	rootCmd.AddCommand(eventsCmd)
	rootCmd.AddCommand(adminCmd)
	rootCmd.AddCommand(projectCmd)
}

func openExistingDB() (*DB, error) {
	path, err := GetDBPath()
	if err != nil {
		return nil, err
	}

	db, err := OpenDB(path)
	if err != nil {
		return nil, err
	}

	// Always ensure schema is up to date (handles both new DBs and migrations)
	if err := db.InitDB(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	// Check if v4 migration is needed
	needsMigration, err := db.NeedsMigrationToV4()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to check migration status: %w", err)
	}

	if needsMigration {
		fmt.Println("Migrating database to v4...")
		fmt.Printf("Creating backup at %s%s\n", path, v4BackupSuffix)

		if err := db.MigrateToV4(path); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to migrate to v4: %w", err)
		}

		fmt.Println("Migration to v4 complete!")
		fmt.Println("Running post-migration health check...")
		report, err := runDoctor(db)
		if err != nil {
			fmt.Printf("Doctor check failed: %v\n", err)
		} else {
			printDoctorReport(os.Stdout, report)
			if report.ProblemCount() > 0 {
				fmt.Println("Resolve the issues above. You can rerun 'tk doctor' at any time.")
			}
		}
	}

	return db, nil
}

func getCurrentUser() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	return currentUser.Username, nil
}

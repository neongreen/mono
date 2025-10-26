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
		prefix, _ := cmd.Flags().GetString("prefix")

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

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
		taskID := args[0]
		state := args[1]

		axis, _ := cmd.Flags().GetString("axis")
		role, _ := cmd.Flags().GetString("role")

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Resolve task ID to UUID (handles aliases and reprefixed tasks)
		taskUUID, err := db.ResolveTaskIDToUUID(taskID)
		if err != nil {
			return err
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
			TaskID:   taskID, // Use the input ID for legacy compatibility
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

		fmt.Printf("Set status for task %s: %s=%s\n", taskID, axis, state)
		return nil
	},
}

var noteCmd = &cobra.Command{
	Use:   "note [task-id] [text]",
	Short: "Add a note to a task",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]
		text := args[1]

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Resolve task ID to UUID (handles aliases and reprefixed tasks)
		taskUUID, err := db.ResolveTaskIDToUUID(taskID)
		if err != nil {
			return err
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
			TaskID:   taskID, // Use the input ID for legacy compatibility
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

		fmt.Printf("Added note to task %s\n", taskID)
		return nil
	},
}

var viewCmd = &cobra.Command{
	Use:   "view [task-id]",
	Short: "View task details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Build reducer to get task and all its IDs (current + aliases)
		// TODO: Consider caching reducer or adding task_index table for performance
		events, err := db.GetEvents()
		if err != nil {
			return err
		}

		reducer, err := BuildFromEvents(events)
		if err != nil {
			return err
		}

		// Resolve task ID to UUID (handles aliases and reprefixed tasks)
		taskUUID, err := db.ResolveTaskIDToUUID(taskID)
		if err != nil {
			return err
		}

		task, ok := reducer.GetTask(taskUUID)
		if !ok {
			return fmt.Errorf("task not found: %s", taskID)
		}

		output, err := json.MarshalIndent(task, "", "  ")
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

		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		events, err := db.GetEvents()
		if err != nil {
			return err
		}

		reducer, err := BuildFromEvents(events)
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

		// Sort tasks based on the --sort flag
		sortTasks(tasks, sortBy)

		// Get terminal width for wrapping
		termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			termWidth = 80 // default width if terminal size cannot be determined
		}

		// Create table
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)

		if showAliases {
			t.AppendHeader(table.Row{"ID", "Aliases", "Status", "Title"})
		} else {
			t.AppendHeader(table.Row{"ID", "Status", "Title"})
		}

		t.SetStyle(table.StyleLight)
		t.Style().Options.SeparateRows = false
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
			displayID := FormatTaskID(task.TaskID, taskIDs)

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

func init() {
	rootCmd.AddCommand(initCmd)

	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "Database commands",
	}
	dbCmd.AddCommand(dbPathCmd)
	rootCmd.AddCommand(dbCmd)

	newCmd.Flags().String("prefix", "tk", "Task prefix to use")
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
	rootCmd.AddCommand(lsCmd)

	rootCmd.AddCommand(mvCmd)
	rootCmd.AddCommand(nodeCmd)
	rootCmd.AddCommand(remoteCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(ingestCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(prefixCmd)
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

	return db, nil
}

func getCurrentUser() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	return currentUser.Username, nil
}

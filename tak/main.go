package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"strings"

	"github.com/spf13/cobra"
)

var (
	dbPath string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "tak",
	Short: "tak - system-wide event-sourced task tracker",
	Long:  `tak is a command-line tool that tracks tasks system-wide using an append-only event log in a single SQLite database.`,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new tak database",
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

		currentUser, err := getCurrentUser()
		if err != nil {
			return err
		}

		taskID := GenerateULID("tak")
		eventID := GenerateEventID()

		payload := TaskCreatedPayload{
			TaskID:    taskID,
			Title:     title,
			CreatedBy: currentUser,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		event := Event{
			ID:      eventID,
			TS:      GetNextLamportTS(),
			Actor:   currentUser,
			Role:    "human",
			Kind:    "task.created",
			Payload: payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return err
		}

		fmt.Printf("Created task %s: %s\n", taskID, title)
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Manage task status",
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

		currentUser, err := getCurrentUser()
		if err != nil {
			return err
		}

		eventID := GenerateEventID()

		payload := TaskStatusSetPayload{
			TaskID: taskID,
			Axis:   axis,
			State:  state,
			Role:   role,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		event := Event{
			ID:      eventID,
			TS:      GetNextLamportTS(),
			Actor:   currentUser,
			Role:    role,
			Kind:    "task.status.set",
			Payload: payloadJSON,
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

		currentUser, err := getCurrentUser()
		if err != nil {
			return err
		}

		eventID := GenerateEventID()

		payload := TaskNoteAddPayload{
			TaskID:   taskID,
			Markdown: text,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		event := Event{
			ID:      eventID,
			TS:      GetNextLamportTS(),
			Actor:   currentUser,
			Role:    "human",
			Kind:    "task.note.add",
			Payload: payloadJSON,
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

		events, err := db.GetEventsByTaskID(taskID)
		if err != nil {
			return err
		}

		if len(events) == 0 {
			return fmt.Errorf("task not found: %s", taskID)
		}

		reducer, err := BuildFromEvents(events)
		if err != nil {
			return err
		}

		task, ok := reducer.GetTask(taskID)
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
		axisFilter, _ := cmd.Flags().GetString("axis")

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

		for _, task := range tasks {
			fmt.Printf("%s: %s\n", task.TaskID, task.Title)
			for axisName, axis := range task.Axes {
				fmt.Printf("  %s: %s\n", axisName, axis.Effective)
			}
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

	rootCmd.AddCommand(newCmd)

	statusSetCmd.Flags().String("axis", "generic", "Status axis")
	statusSetCmd.Flags().String("role", "human", "Actor role")
	statusCmd.AddCommand(statusSetCmd)
	rootCmd.AddCommand(statusCmd)

	rootCmd.AddCommand(noteCmd)
	rootCmd.AddCommand(viewCmd)

	lsCmd.Flags().String("axis", "", "Filter by axis:state")
	rootCmd.AddCommand(lsCmd)
}

func openExistingDB() (*DB, error) {
	path, err := GetDBPath()
	if err != nil {
		return nil, err
	}

	// Check if database exists, if not, create it
	dbExists := DBExists(path)

	db, err := OpenDB(path)
	if err != nil {
		return nil, err
	}

	// If database didn't exist, initialize the schema
	if !dbExists {
		if err := db.InitDB(); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to initialize database schema: %w", err)
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

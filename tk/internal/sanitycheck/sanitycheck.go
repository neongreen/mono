package sanitycheck

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/reducer"
	"github.com/neongreen/mono/tk/internal/types"
)

// StateComparison holds the differences between reducer state and database state
type StateComparison struct {
	Timestamp         time.Time    `json:"timestamp"`
	EventCount        int          `json:"event_count"`
	ReducerTaskCount  int          `json:"reducer_task_count"`
	DatabaseTaskCount int          `json:"database_task_count"`
	Differences       []Difference `json:"differences,omitempty"`
}

// Difference represents a single difference found
type Difference struct {
	Type        string `json:"type"` // "missing_in_db", "missing_in_reducer", "field_mismatch"
	TaskUID     string `json:"task_uid"`
	Field       string `json:"field,omitempty"`
	ReducerVal  string `json:"reducer_value,omitempty"`
	DatabaseVal string `json:"database_value,omitempty"`
	Message     string `json:"message"`
}

// RunSanityCheck performs a silent sanity check comparing reducer state to database state.
// If differences are found, it writes a warning and a detailed diff file to ~/.tk/
// Returns true if differences were found, false otherwise.
func RunSanityCheck(db *database.DB, config *config.Config) bool {
	comparison, err := compareState(db, config)
	if err != nil {
		// Silently ignore errors - this is a best-effort check
		return false
	}

	if len(comparison.Differences) == 0 {
		// States match, nothing to do
		return false
	}

	// Write diff to file
	if err := writeDiffFile(comparison); err != nil {
		// Silently ignore write errors
		return false
	}

	// Print warning to stderr
	fmt.Fprintf(os.Stderr, "⚠️  Warning: State mismatch detected between events and database. See %s for details.\n",
		getDiffFilePath())

	return true
}

// compareState builds state from events using the reducer and compares it to database projections
func compareState(db *database.DB, config *config.Config) (*StateComparison, error) {
	comparison := &StateComparison{
		Timestamp: time.Now(),
	}

	// Get all events
	events, err := db.GetEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	comparison.EventCount = len(events)

	// Build reducer state from events
	red, err := reducer.BuildFromEventsWithConfig(events, config)
	if err != nil {
		return nil, fmt.Errorf("failed to build reducer state: %w", err)
	}

	// Get tasks from reducer
	reducerTasks := red.Tasks()
	comparison.ReducerTaskCount = len(reducerTasks)

	// Get tasks from database
	dbTasks, err := getTasksFromDatabase(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks from database: %w", err)
	}
	comparison.DatabaseTaskCount = len(dbTasks)

	// Compare tasks
	comparison.Differences = compareTasks(reducerTasks, dbTasks)

	return comparison, nil
}

// getTasksFromDatabase retrieves all tasks from the database projection tables
func getTasksFromDatabase(db *database.DB) (map[string]*dbTask, error) {
	tasks := make(map[string]*dbTask)

	rows, err := db.Query(`
		SELECT task_uid, project_uid, title, created_at, created_by
		FROM tasks
		ORDER BY task_uid
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var task dbTask
		var createdAtUnix int64
		if err := rows.Scan(&task.TaskUID, &task.ProjectUID, &task.Title, &createdAtUnix, &task.CreatedBy); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		task.CreatedAt = time.Unix(createdAtUnix, 0)
		tasks[task.TaskUID] = &task
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tasks: %w", err)
	}

	return tasks, nil
}

// dbTask represents a task from the database projection table
type dbTask struct {
	TaskUID    string
	ProjectUID string
	Title      string
	CreatedAt  time.Time
	CreatedBy  string
}

// compareTasks compares reducer tasks with database tasks and returns differences
func compareTasks(reducerTasks map[string]*types.Task, dbTasks map[string]*dbTask) []Difference {
	var diffs []Difference

	// Check for tasks in reducer but not in database
	for taskUID := range reducerTasks {
		if _, exists := dbTasks[taskUID]; !exists {
			diffs = append(diffs, Difference{
				Type:    "missing_in_db",
				TaskUID: taskUID,
				Message: fmt.Sprintf("Task %s exists in reducer but not in database", taskUID),
			})
		}
	}

	// Check for tasks in database but not in reducer
	for taskUID := range dbTasks {
		if _, exists := reducerTasks[taskUID]; !exists {
			diffs = append(diffs, Difference{
				Type:    "missing_in_reducer",
				TaskUID: taskUID,
				Message: fmt.Sprintf("Task %s exists in database but not in reducer", taskUID),
			})
		}
	}

	// For tasks that exist in both, compare fields
	for taskUID, reducerTask := range reducerTasks {
		dbTask, exists := dbTasks[taskUID]
		if !exists {
			continue // Already reported as missing_in_db
		}

		// Compare title
		if reducerTask.Title != dbTask.Title {
			diffs = append(diffs, Difference{
				Type:        "field_mismatch",
				TaskUID:     taskUID,
				Field:       "title",
				ReducerVal:  reducerTask.Title,
				DatabaseVal: dbTask.Title,
				Message:     fmt.Sprintf("Title mismatch for task %s", taskUID),
			})
		}

		// Compare created_at (allowing for small time differences due to serialization)
		reducerCreatedAt := reducerTask.CreatedAt.Unix()
		dbCreatedAt := dbTask.CreatedAt.Unix()
		if reducerCreatedAt != dbCreatedAt {
			diffs = append(diffs, Difference{
				Type:        "field_mismatch",
				TaskUID:     taskUID,
				Field:       "created_at",
				ReducerVal:  reducerTask.CreatedAt.Format(time.RFC3339),
				DatabaseVal: dbTask.CreatedAt.Format(time.RFC3339),
				Message:     fmt.Sprintf("CreatedAt mismatch for task %s", taskUID),
			})
		}
	}

	return diffs
}

// getDiffFilePath returns the path to the diff file in ~/.tk/
func getDiffFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".tk", "state-diff.json")
}

// writeDiffFile writes the comparison results to a JSON file
func writeDiffFile(comparison *StateComparison) error {
	diffPath := getDiffFilePath()
	if diffPath == "" {
		return fmt.Errorf("could not determine diff file path")
	}

	// Ensure .tk directory exists
	dir := filepath.Dir(diffPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(comparison, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal comparison: %w", err)
	}

	// Write file
	if err := os.WriteFile(diffPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write diff file: %w", err)
	}

	return nil
}

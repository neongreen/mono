package debug

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/reducer"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

type DoctorCollision struct {
	ProjectUID     string
	ProjectAlias   string
	Number         int64
	TaskDisplayIDs []string
}

type DoctorReport struct {
	Issues        []string
	InvalidEvents []string
	Collisions    []DoctorCollision
}

func (r *DoctorReport) ProblemCount() int {
	return len(r.Issues) + len(r.InvalidEvents) + len(r.Collisions)
}

// DoctorCmd is the doctor subcommand for debug
var DoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Verify database health and report issues",
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		jsonOutput, _ := cobraCmd.Flags().GetBool("json")

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		report, err := RunDoctor(db)
		if err != nil {
			return err
		}

		if jsonOutput {
			output, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal report: %w", err)
			}
			fmt.Println(string(output))
		} else {
			PrintDoctorReport(os.Stdout, report)
		}

		if report.ProblemCount() > 0 {
			return fmt.Errorf("doctor found %d issue(s)", report.ProblemCount())
		}

		return nil
	},
}

func init() {
	DoctorCmd.Flags().Bool("json", false, "Output as JSON")
}

func RunDoctor(db *database.DB) (*DoctorReport, error) {
	report := &DoctorReport{}

	if err := checkOrphanTasks(db, report); err != nil {
		return nil, err
	}
	if err := checkMissingNumbers(db, report); err != nil {
		return nil, err
	}
	if err := checkBrokenAliases(db, report); err != nil {
		return nil, err
	}
	if err := checkEventPayloads(db, report); err != nil {
		return nil, err
	}
	if err := checkEventOrdering(db, report); err != nil {
		return nil, err
	}
	if err := checkEventProjectionConsistency(db, report); err != nil {
		return nil, err
	}
	if err := collectCollisions(db, report); err != nil {
		return nil, err
	}

	sort.Strings(report.Issues)
	sort.Strings(report.InvalidEvents)
	sort.Slice(report.Collisions, func(i, j int) bool {
		if report.Collisions[i].ProjectAlias == report.Collisions[j].ProjectAlias {
			return report.Collisions[i].Number < report.Collisions[j].Number
		}
		return report.Collisions[i].ProjectAlias < report.Collisions[j].ProjectAlias
	})

	return report, nil
}

func PrintDoctorReport(w io.Writer, report *DoctorReport) {
	if report.ProblemCount() == 0 {
		fmt.Fprintln(w, "✓ Doctor found no issues")
		return
	}

	if len(report.Issues) > 0 {
		fmt.Fprintln(w, "Issues:")
		for _, issue := range report.Issues {
			fmt.Fprintf(w, "  - %s\n", issue)
		}
	}

	if len(report.InvalidEvents) > 0 {
		fmt.Fprintln(w, "Invalid events:")
		for _, evt := range report.InvalidEvents {
			fmt.Fprintf(w, "  - %s\n", evt)
		}
	}

	if len(report.Collisions) > 0 {
		fmt.Fprintln(w, "Number collisions:")
		for _, col := range report.Collisions {
			fmt.Fprintf(w, "  - %s #%d\n", col.ProjectAlias, col.Number)
			for _, task := range col.TaskDisplayIDs {
				fmt.Fprintf(w, "      * %s\n", task)
			}
		}
	}
}

func checkOrphanTasks(db *database.DB, report *DoctorReport) error {
	rows, err := db.Db.Query(`
        SELECT task_uid, project_uid
        FROM tasks
        WHERE project_uid NOT IN (SELECT project_uid FROM projects)
    `)
	if err != nil {
		return fmt.Errorf("failed to query orphan tasks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var taskUID, projectUID string
		if err := rows.Scan(&taskUID, &projectUID); err != nil {
			return fmt.Errorf("failed to scan orphan task: %w", err)
		}
		display, err := database.RenderTaskDisplayID(db, taskUID)
		if err != nil {
			display = taskUID
		}
		report.Issues = append(report.Issues, fmt.Sprintf("task %s references missing project %s", display, projectUID))
	}
	return rows.Err()
}

func checkMissingNumbers(db *database.DB, report *DoctorReport) error {
	rows, err := db.Db.Query(`
        SELECT task_uid FROM tasks
        WHERE task_uid NOT IN (SELECT task_uid FROM task_numbers)
    `)
	if err != nil {
		return fmt.Errorf("failed to query tasks without numbers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var taskUID string
		if err := rows.Scan(&taskUID); err != nil {
			return fmt.Errorf("failed to scan task without number: %w", err)
		}
		display, err := database.RenderTaskDisplayID(db, taskUID)
		if err != nil {
			display = taskUID
		}
		report.Issues = append(report.Issues, fmt.Sprintf("task %s is missing a number assignment", display))
	}
	return rows.Err()
}

func checkBrokenAliases(db *database.DB, report *DoctorReport) error {
	rows, err := db.Db.Query(`
        SELECT alias, node, project_uid
        FROM project_aliases
        WHERE project_uid NOT IN (SELECT project_uid FROM projects)
    `)
	if err != nil {
		return fmt.Errorf("failed to query aliases: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var alias, node, projectUID string
		if err := rows.Scan(&alias, &node, &projectUID); err != nil {
			return fmt.Errorf("failed to scan alias: %w", err)
		}
		report.Issues = append(report.Issues, fmt.Sprintf("alias %s (node %s) points to missing project %s", alias, node, projectUID))
	}
	return rows.Err()
}

func checkEventPayloads(db *database.DB, report *DoctorReport) error {
	rows, err := db.Db.Query(`SELECT id, payload FROM events`)
	if err != nil {
		return fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var payload []byte
		if err := rows.Scan(&id, &payload); err != nil {
			return fmt.Errorf("failed to scan event payload: %w", err)
		}
		if len(payload) == 0 || !json.Valid(payload) {
			report.InvalidEvents = append(report.InvalidEvents, id)
		}
	}
	return rows.Err()
}

func checkEventProjectionConsistency(db *database.DB, report *DoctorReport) error {
	// Rebuild state from events (in memory, read-only)
	events, err := db.GetEvents()
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	// Try to rebuild, but collect errors instead of failing
	r := reducer.NewReducer()
	var replayErrors []string

	for i, e := range events {
		if err := r.Apply(e); err != nil {
			// Collect error instead of failing immediately
			replayErrors = append(replayErrors, fmt.Sprintf(
				"Event %d: %s (kind=%s, TS=%d) - %v",
				i+1, e.ID, e.Kind, e.TS, err))
		}
	}

	// Report replay errors
	if len(replayErrors) > 0 {
		report.Issues = append(report.Issues,
			fmt.Sprintf("Found %d events that failed to apply during replay:", len(replayErrors)))
		// Limit to first 20 errors to avoid overwhelming output
		for i, errMsg := range replayErrors {
			if i >= 20 {
				report.Issues = append(report.Issues,
					fmt.Sprintf("... and %d more replay errors", len(replayErrors)-20))
				break
			}
			report.Issues = append(report.Issues, "  "+errMsg)
		}
	}

	// Compare task count
	rows, err := db.Db.Query(`SELECT COUNT(*) FROM tasks`)
	if err != nil {
		return fmt.Errorf("failed to count tasks: %w", err)
	}
	defer rows.Close()

	var dbTaskCount int
	if rows.Next() {
		rows.Scan(&dbTaskCount)
	}
	rows.Close()

	reducerTaskCount := len(r.Tasks())
	if dbTaskCount != reducerTaskCount {
		report.Issues = append(report.Issues, fmt.Sprintf("task count mismatch: database has %d tasks, event replay produces %d tasks", dbTaskCount, reducerTaskCount))
	}

	// Check each task in database exists in reducer with same title
	taskRows, err := db.Db.Query(`SELECT task_uid, title FROM tasks`)
	if err != nil {
		return fmt.Errorf("failed to query tasks: %w", err)
	}
	defer taskRows.Close()

	for taskRows.Next() {
		var taskUID, dbTitle string
		if err := taskRows.Scan(&taskUID, &dbTitle); err != nil {
			return fmt.Errorf("failed to scan task: %w", err)
		}

		reducerTask, ok := r.GetTask(taskUID)
		if !ok {
			display, _ := database.RenderTaskDisplayID(db, taskUID)
			if display == "" {
				display = taskUID
			}
			report.Issues = append(report.Issues, fmt.Sprintf("task %s exists in database but not in event replay", display))
			continue
		}

		if reducerTask.Title != dbTitle {
			display, _ := database.RenderTaskDisplayID(db, taskUID)
			if display == "" {
				display = taskUID
			}
			report.Issues = append(report.Issues, fmt.Sprintf("task %s title mismatch: db=%q, events=%q", display, dbTitle, reducerTask.Title))
		}
	}

	return nil
}

func checkEventOrdering(db *database.DB, report *DoctorReport) error {
	events, err := db.GetEvents()
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	// Track when tasks are created (event index)
	taskCreated := make(map[string]int)

	for i, e := range events {
		if e.Kind == string(types.EventKindTaskCreated) {
			var payload types.TaskCreatedPayload
			if err := json.Unmarshal(e.Payload, &payload); err == nil {
				taskCreated[payload.TaskUID] = i
			}
		}
	}

	// Check events that reference tasks
	for i, e := range events {
		taskUUID := extractTaskUUIDFromEvent(e)
		if taskUUID == "" {
			continue
		}

		createdIndex, exists := taskCreated[taskUUID]
		if !exists {
			// Task never created
			report.Issues = append(report.Issues,
				fmt.Sprintf("Event %s (%s, TS=%d) references task %s that was never created",
					e.ID, e.Kind, e.TS, taskUUID))
		} else if i < createdIndex {
			// Event comes before task.created
			report.Issues = append(report.Issues,
				fmt.Sprintf("Event %s (%s, TS=%d, position=%d) comes before task.created (TS=%d, position=%d) for task %s",
					e.ID, e.Kind, e.TS, i+1, events[createdIndex].TS, createdIndex+1, taskUUID))
		}

		// Check for corrupted timestamps
		if e.CreatedAt.Unix() <= 0 {
			report.Issues = append(report.Issues,
				fmt.Sprintf("Event %s has corrupt created_at timestamp: %v", e.ID, e.CreatedAt))
		}
	}

	return nil
}

func extractTaskUUIDFromEvent(e types.Event) string {
	var payload map[string]interface{}
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return ""
	}

	// Try task_uuid first (v4 format)
	if uuid, ok := payload["task_uuid"].(string); ok {
		return uuid
	}
	// Fall back to task_uid (older format)
	if uuid, ok := payload["task_uid"].(string); ok {
		return uuid
	}
	return ""
}

func collectCollisions(db *database.DB, report *DoctorReport) error {
	collisions, err := GetNumberCollisions(db, "")
	if err != nil {
		return err
	}
	report.Collisions = append(report.Collisions, collisions...)
	return nil
}

// GetNumberCollisions finds task number collisions (exported for cmd package)
func GetNumberCollisions(db *database.DB, projectFilter string) ([]DoctorCollision, error) {
	baseQuery := `
        SELECT project_uid, number
        FROM task_numbers
    `
	var args []interface{}
	if projectFilter != "" {
		baseQuery += " WHERE project_uid = ?"
		args = append(args, projectFilter)
	}
	baseQuery += " GROUP BY project_uid, number HAVING COUNT(*) > 1"

	rows, err := db.Db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query collisions: %w", err)
	}
	defer rows.Close()

	var collisions []DoctorCollision
	for rows.Next() {
		var projectUID string
		var number int64
		if err := rows.Scan(&projectUID, &number); err != nil {
			return nil, fmt.Errorf("failed to scan collision row: %w", err)
		}

		tasksRows, err := db.Db.Query(`
            SELECT task_uid FROM task_numbers
            WHERE project_uid = ? AND number = ?
        `, projectUID, number)
		if err != nil {
			return nil, fmt.Errorf("failed to query collision tasks: %w", err)
		}

		var taskDisplayIDs []string
		for tasksRows.Next() {
			var taskUID string
			if err := tasksRows.Scan(&taskUID); err != nil {
				tasksRows.Close()
				return nil, fmt.Errorf("failed to scan collision task: %w", err)
			}
			display, err := database.RenderTaskDisplayID(db, taskUID)
			if err != nil {
				display = taskUID
			}
			taskDisplayIDs = append(taskDisplayIDs, display)
		}
		tasksRows.Close()

		alias, err := database.PreferredAliasForProject(db, types.ProjectUID(projectUID))
		if err != nil {
			alias = projectUID
		}
		if alias == "" {
			alias = projectUID
		}

		collisions = append(collisions, DoctorCollision{
			ProjectUID:     projectUID,
			ProjectAlias:   alias,
			Number:         number,
			TaskDisplayIDs: taskDisplayIDs,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return collisions, nil
}

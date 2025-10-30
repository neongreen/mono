package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

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

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Verify database health and report issues",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		
		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		report, err := runDoctor(db)
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
			printDoctorReport(os.Stdout, report)
		}

		if report.ProblemCount() > 0 {
			return fmt.Errorf("doctor found %d issue(s)", report.ProblemCount())
		}

		return nil
	},
}

func init() {
	doctorCmd.Flags().Bool("json", false, "Output as JSON")
}

func runDoctor(db *DB) (*DoctorReport, error) {
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

func printDoctorReport(w io.Writer, report *DoctorReport) {
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

func checkOrphanTasks(db *DB, report *DoctorReport) error {
	rows, err := db.db.Query(`
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
		display, err := RenderTaskDisplayID(db, taskUID)
		if err != nil {
			display = taskUID
		}
		report.Issues = append(report.Issues, fmt.Sprintf("task %s references missing project %s", display, projectUID))
	}
	return rows.Err()
}

func checkMissingNumbers(db *DB, report *DoctorReport) error {
	rows, err := db.db.Query(`
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
		display, err := RenderTaskDisplayID(db, taskUID)
		if err != nil {
			display = taskUID
		}
		report.Issues = append(report.Issues, fmt.Sprintf("task %s is missing a number assignment", display))
	}
	return rows.Err()
}

func checkBrokenAliases(db *DB, report *DoctorReport) error {
	rows, err := db.db.Query(`
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

func checkEventPayloads(db *DB, report *DoctorReport) error {
	rows, err := db.db.Query(`SELECT id, payload FROM events`)
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

func collectCollisions(db *DB, report *DoctorReport) error {
	collisions, err := getNumberCollisions(db, "")
	if err != nil {
		return err
	}
	report.Collisions = append(report.Collisions, collisions...)
	return nil
}

func getNumberCollisions(db *DB, projectFilter string) ([]DoctorCollision, error) {
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

	rows, err := db.db.Query(baseQuery, args...)
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

		tasksRows, err := db.db.Query(`
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
			display, err := RenderTaskDisplayID(db, taskUID)
			if err != nil {
				display = taskUID
			}
			taskDisplayIDs = append(taskDisplayIDs, display)
		}
		tasksRows.Close()

		alias, err := preferredAliasForProject(db, projectUID)
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

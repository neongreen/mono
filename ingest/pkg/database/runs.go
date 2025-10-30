package database

import (
	"database/sql"
	"fmt"
	"time"
)

// CreateRun creates a new ingestion run
func (d *Database) CreateRun(repoPath string, runType string) (int64, error) {
	result, err := d.db.Exec(
		"INSERT INTO runs (repo_path, run_type, start_time, status) VALUES (?, ?, ?, ?)",
		repoPath,
		runType,
		time.Now(),
		"in_progress",
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create run: %w", err)
	}

	return result.LastInsertId()
}

// FinishRun marks a run as completed
func (d *Database) FinishRun(runID int64, status string) error {
	_, err := d.db.Exec(
		"UPDATE runs SET end_time = ?, status = ? WHERE id = ?",
		time.Now(),
		status,
		runID,
	)
	if err != nil {
		return fmt.Errorf("failed to finish run: %w", err)
	}

	return nil
}

// UpdateRunItemCount updates item count for a run based on run type
func (d *Database) UpdateRunItemCount(runID int64) error {
	// Get run type first
	var runType string
	err := d.db.QueryRow("SELECT run_type FROM runs WHERE id = ?", runID).Scan(&runType)
	if err != nil {
		return fmt.Errorf("failed to get run type: %w", err)
	}

	var query string
	var args []interface{}
	switch runType {
	case "git":
		query = "UPDATE runs SET item_count = (SELECT COUNT(*) FROM commits WHERE run_id = ?) WHERE id = ?"
		args = []interface{}{runID, runID}
	case "fs":
		query = "UPDATE runs SET item_count = (SELECT COUNT(*) FROM fs_entries WHERE run_id = ?) WHERE id = ?"
		args = []interface{}{runID, runID}
	case "cmd":
		lineCount, err := d.totalCommandOutputLines(runID)
		if err != nil {
			return fmt.Errorf("failed to calculate command output lines: %w", err)
		}
		_, err = d.db.Exec("UPDATE runs SET item_count = ? WHERE id = ?", lineCount, runID)
		if err != nil {
			return fmt.Errorf("failed to update run item count: %w", err)
		}
		return nil
	case "github":
		query = "UPDATE runs SET item_count = (SELECT COUNT(*) FROM github_issues WHERE run_id = ?) + (SELECT COUNT(*) FROM github_prs WHERE run_id = ?) WHERE id = ?"
		args = []interface{}{runID, runID, runID}
	case "linear":
		query = "UPDATE runs SET item_count = (SELECT COUNT(*) FROM linear_issues WHERE run_id = ?) WHERE id = ?"
		args = []interface{}{runID, runID}
	default:
		query = "UPDATE runs SET item_count = 0 WHERE id = ?"
		args = []interface{}{runID}
	}

	_, err = d.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update run item count: %w", err)
	}

	return nil
}

func (d *Database) totalCommandOutputLines(runID int64) (int, error) {
	rows, err := d.db.Query("SELECT stdout, stderr FROM cmd_runs WHERE run_id = ?", runID)
	if err != nil {
		return 0, fmt.Errorf("failed to select command outputs: %w", err)
	}
	defer rows.Close()

	total := 0
	for rows.Next() {
		var stdout, stderr sql.NullString
		if err := rows.Scan(&stdout, &stderr); err != nil {
			return 0, fmt.Errorf("failed to scan command outputs: %w", err)
		}
		if stdout.Valid {
			total += countOutputLines(stdout.String)
		}
		if stderr.Valid {
			total += countOutputLines(stderr.String)
		}
	}

	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("error iterating command outputs: %w", err)
	}

	return total, nil
}

// UpdateRunCounts is deprecated, use UpdateRunItemCount instead
func (d *Database) UpdateRunCounts(runID int64) error {
	return d.UpdateRunItemCount(runID)
}

// GetAllRuns retrieves all ingestion runs
func (d *Database) GetAllRuns() ([]Run, error) {
	rows, err := d.db.Query(`
		SELECT id, repo_path, run_type, start_time, end_time, item_count, status
		FROM runs
		ORDER BY start_time DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query runs: %w", err)
	}
	defer rows.Close()

	var runs []Run
	for rows.Next() {
		var run Run
		var endTime sql.NullTime
		err := rows.Scan(
			&run.ID,
			&run.RepoPath,
			&run.RunType,
			&run.StartTime,
			&endTime,
			&run.ItemCount,
			&run.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan run: %w", err)
		}

		if endTime.Valid {
			run.EndTime = &endTime.Time
		}

		runs = append(runs, run)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating runs: %w", err)
	}

	return runs, nil
}

package database

import "fmt"

// CreateCmdRun creates a new command run record
func (d *Database) CreateCmdRun(runID int64, command string, exitCode int, stdout, stderr string, durationMs int64) error {
	_, err := d.db.Exec(
		"INSERT INTO cmd_runs (run_id, command, exit_code, stdout, stderr, duration_ms) VALUES (?, ?, ?, ?, ?, ?)",
		runID,
		command,
		exitCode,
		stdout,
		stderr,
		durationMs,
	)
	if err != nil {
		return fmt.Errorf("failed to create cmd run: %w", err)
	}

	return nil
}

// UpdateRunFileCount is deprecated, use UpdateRunItemCount instead
func (d *Database) UpdateRunFileCount(runID int64) error {
	return d.UpdateRunItemCount(runID)
}

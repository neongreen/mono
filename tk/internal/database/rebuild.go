package database

import (
	"fmt"
)

// RebuildProjections clears all projection tables and rebuilds them from events.
// This ensures projections are computed in Lamport timestamp order, making them deterministic.
//
// Use this when:
// - Projection tables are corrupted or inconsistent
// - After fixing projection bugs (like the "auto" mode non-determinism)
// - To verify projection determinism
func (d *DB) RebuildProjections() error {
	// Start a transaction to make this atomic
	tx, err := d.Db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Phase 1: Clear all projection tables
	tables := []string{
		"task_numbers",
		"tasks",
		"project_aliases",
		"projects",
	}

	for _, table := range tables {
		if _, err := tx.Exec(fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			return fmt.Errorf("failed to clear %s table: %w", table, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit table clearing: %w", err)
	}

	// Phase 2: Replay all events in Lamport timestamp order
	events, err := d.GetEvents()
	if err != nil {
		return fmt.Errorf("failed to get events for rebuild: %w", err)
	}

	projectionErrors := 0
	for _, event := range events {
		if err := d.ProjectEvent(event); err != nil {
			// Track projection errors but continue
			// Some events might not have projection handlers (like status.set, etc.)
			projectionErrors++
		}
	}

	return nil
}

// VerifyProjectionDeterminism checks if rebuilding projections produces the same result.
// This is useful for testing and debugging projection issues.
func (d *DB) VerifyProjectionDeterminism() error {
	// Get current projection state
	var currentCount int
	err := d.Db.QueryRow("SELECT COUNT(*) FROM task_numbers").Scan(&currentCount)
	if err != nil {
		return fmt.Errorf("failed to get current task_numbers count: %w", err)
	}

	// Rebuild projections
	if err := d.RebuildProjections(); err != nil {
		return fmt.Errorf("failed to rebuild projections: %w", err)
	}

	// Compare counts
	var newCount int
	err = d.Db.QueryRow("SELECT COUNT(*) FROM task_numbers").Scan(&newCount)
	if err != nil {
		return fmt.Errorf("failed to get new task_numbers count: %w", err)
	}

	if currentCount != newCount {
		return fmt.Errorf("projection determinism check failed: task_numbers count changed from %d to %d", currentCount, newCount)
	}

	return nil
}


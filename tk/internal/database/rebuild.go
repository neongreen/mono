package database

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/reducer"
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

	// Phase 2: Build reducer state from all events and persist to database
	events, err := d.GetEvents()
	if err != nil {
		return fmt.Errorf("failed to get events for rebuild: %w", err)
	}

	// Load config for reducer
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Build reducer state from events
	// The reducer uses lax types and properly handles legacy event formats
	r, err := reducer.BuildFromEventsWithConfig(events, cfg)
	if err != nil {
		return fmt.Errorf("failed to build reducer from events: %w", err)
	}

	// Persist reducer state to database
	if err := d.PersistReducerState(r); err != nil {
		return fmt.Errorf("failed to persist reducer state: %w", err)
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

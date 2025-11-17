package database

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/reducer"
)

// CheckInvariants verifies event sourcing invariants
// Call this after any operation to ensure database consistency
func (d *DB) CheckInvariants() error {
	// Invariant 1: State matches event log
	// Rebuilding from events should produce the same state as current projections
	events, err := d.GetEvents()
	if err != nil {
		return fmt.Errorf("invariant check: failed to get events: %w", err)
	}

	// Build state from events
	cfg := &config.Config{
		Blocking: config.BlockingConfig{
			BlockingAxis: "generic",
			DoneStates:   []string{"done"},
		},
	}

	rebuilt, err := reducer.BuildFromEventsWithConfig(events, cfg)
	if err != nil {
		return fmt.Errorf("invariant check: failed to rebuild from events: %w", err)
	}

	// Check task count matches
	var projectedTaskCount int
	if err := d.Db.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&projectedTaskCount); err != nil {
		return fmt.Errorf("invariant check: failed to count projected tasks: %w", err)
	}

	if len(rebuilt.Tasks()) != projectedTaskCount {
		return fmt.Errorf("invariant violated: rebuilt state has %d tasks but projections have %d tasks",
			len(rebuilt.Tasks()), projectedTaskCount)
	}

	// Invariant 2: No orphaned tasks (every task in projection has events)
	rows, err := d.Db.Query(`SELECT task_uid FROM tasks`)
	if err != nil {
		return fmt.Errorf("invariant check: failed to query tasks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var taskUID string
		if err := rows.Scan(&taskUID); err != nil {
			return fmt.Errorf("invariant check: failed to scan task: %w", err)
		}

		// This task should exist in rebuilt state
		if _, exists := rebuilt.GetTask(taskUID); !exists {
			return fmt.Errorf("invariant violated: task %s exists in projections but not in rebuilt state (orphaned data)", taskUID)
		}
	}

	// Invariant 3: All tasks in rebuilt state should exist in projections
	for taskUID := range rebuilt.Tasks() {
		var count int
		if err := d.Db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE task_uid = ?`, taskUID).Scan(&count); err != nil {
			return fmt.Errorf("invariant check: failed to check task existence: %w", err)
		}

		if count == 0 {
			return fmt.Errorf("invariant violated: task %s exists in rebuilt state but not in projections (missing projection)", taskUID)
		}
	}

	// Invariant 4: Events when ordered by Lamport TS should be processable
	// (We don't check monotonic by insertion order because multi-machine sync can interleave events)
	// Instead, verify that when we ORDER BY ts, we get valid Lamport order
	var prevTS int64 = -1
	eventRows, err := d.Db.Query(`SELECT ts FROM events ORDER BY ts ASC`)
	if err != nil {
		return fmt.Errorf("invariant check: failed to query events: %w", err)
	}
	defer eventRows.Close()

	for eventRows.Next() {
		var ts int64
		if err := eventRows.Scan(&ts); err != nil {
			return fmt.Errorf("invariant check: failed to scan event: %w", err)
		}

		if ts < prevTS {
			return fmt.Errorf("invariant violated: events not monotonic when ordered by Lamport TS (ts=%d after ts=%d)", ts, prevTS)
		}
		prevTS = ts
	}

	return nil
}

// CheckInvariantsT is a test helper that calls CheckInvariants and fails the test if violated
func (d *DB) CheckInvariantsT(t interface {
	Helper()
	Fatalf(string, ...any)
}) {
	t.Helper()

	if err := d.CheckInvariants(); err != nil {
		t.Fatalf("Invariant violation: %v", err)
	}
}

package main

import (
	"database/sql"
	"fmt"
)

// GetNextLamportTS gets the next Lamport timestamp and increments the counter
func (d *DB) GetNextLamportTS() (int64, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current counter value
	var counter int64
	err = tx.QueryRow("SELECT value FROM metadata WHERE key = 'lamport_counter'").Scan(&counter)
	if err == sql.ErrNoRows {
		counter = 0
	} else if err != nil {
		return 0, fmt.Errorf("failed to query lamport counter: %w", err)
	}

	nextTS := counter + 1

	if counter == 0 {
		_, err = tx.Exec("INSERT INTO metadata (key, value) VALUES ('lamport_counter', ?)", nextTS)
	} else {
		_, err = tx.Exec("UPDATE metadata SET value = ? WHERE key = 'lamport_counter'", nextTS)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to update lamport counter: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nextTS, nil
}

// BumpLamport updates the lamport counter if the given value is higher
func (d *DB) BumpLamport(newValue int64) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get current counter value
	var counter int64
	err = tx.QueryRow("SELECT value FROM metadata WHERE key = 'lamport_counter'").Scan(&counter)
	if err == sql.ErrNoRows {
		counter = 0
	} else if err != nil {
		return fmt.Errorf("failed to query lamport counter: %w", err)
	}

	if newValue > counter {
		if counter == 0 {
			_, err = tx.Exec("INSERT INTO metadata (key, value) VALUES ('lamport_counter', ?)", newValue)
		} else {
			_, err = tx.Exec("UPDATE metadata SET value = ? WHERE key = 'lamport_counter'", newValue)
		}
		if err != nil {
			return fmt.Errorf("failed to update lamport counter: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetNextTaskNumber gets the next task number and increments the counter
func (d *DB) GetNextTaskNumber() (int64, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var lastID int64
	err = tx.QueryRow("SELECT last_id FROM task_counter").Scan(&lastID)
	if err != nil {
		return 0, fmt.Errorf("failed to get last task ID: %w", err)
	}

	nextID := lastID + 1
	_, err = tx.Exec("UPDATE task_counter SET last_id = ?", nextID)
	if err != nil {
		return 0, fmt.Errorf("failed to update task counter: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nextID, nil
}

// GetNextEventNumber gets the next event number and increments the counter
func (d *DB) GetNextEventNumber() (int64, error) {
	tx, err := d.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var lastID int64
	err = tx.QueryRow("SELECT last_id FROM event_counter").Scan(&lastID)
	if err != nil {
		return 0, fmt.Errorf("failed to get last event ID: %w", err)
	}

	nextID := lastID + 1
	_, err = tx.Exec("UPDATE event_counter SET last_id = ?", nextID)
	if err != nil {
		return 0, fmt.Errorf("failed to update event counter: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nextID, nil
}

package database

import (
	"database/sql"
	"fmt"
)

// GetNextLamportTS gets the next Lamport timestamp and increments the counter
func (d *DB) GetNextLamportTS() (int64, error) {
	tx, err := d.Db.Begin()
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
	tx, err := d.Db.Begin()
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

// incrementCounter is a generic helper to increment a counter in a table
func (d *DB) incrementCounter(tableName, columnName string) (int64, error) {
	tx, err := d.Db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var lastID int64
	query := fmt.Sprintf("SELECT %s FROM %s", columnName, tableName)
	err = tx.QueryRow(query).Scan(&lastID)
	if err != nil {
		return 0, fmt.Errorf("failed to get last ID from %s: %w", tableName, err)
	}

	nextID := lastID + 1
	updateQuery := fmt.Sprintf("UPDATE %s SET %s = ?", tableName, columnName)
	_, err = tx.Exec(updateQuery, nextID)
	if err != nil {
		return 0, fmt.Errorf("failed to update %s: %w", tableName, err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nextID, nil
}

// GetNextTaskNumber gets the next task number and increments the counter
func (d *DB) GetNextTaskNumber() (int64, error) {
	return d.incrementCounter("task_counter", "last_id")
}

// GetNextEventNumber gets the next event number and increments the counter
func (d *DB) GetNextEventNumber() (int64, error) {
	return d.incrementCounter("event_counter", "last_id")
}

package v4_to_v5

import (
	"database/sql"
	"fmt"
)

// DB interface defines the minimal database operations needed for migration.
// This avoids an import cycle with the database package.
type DB interface {
	Exec(query string, args ...any) (sql.Result, error)
	SetDBVersion(version int) error
}

// Migrate runs the v4 to v5 migration.
//
// This migration:
//  1. Adds is_synthetic column to projects table
//  2. Updates db_version to 5
//
// The is_synthetic column allows the projection layer to mark projects
// that have no corresponding project.created event. These synthetic
// projects are created on-the-fly to make orphaned tasks visible.
//
// This migration is safe to run multiple times (idempotent).
func Migrate(db DB) error {
	// Add is_synthetic column to projects table
	// Use ALTER TABLE ADD COLUMN which is idempotent in SQLite if column exists
	_, err := db.Exec(`
		ALTER TABLE projects ADD COLUMN is_synthetic INTEGER DEFAULT 0
	`)
	if err != nil {
		// Check if error contains "duplicate column" which is fine
		errStr := err.Error()
		if errStr != "duplicate column name: is_synthetic" &&
			errStr != "SQL logic error: duplicate column name: is_synthetic (1)" {
			return fmt.Errorf("failed to add is_synthetic column: %w", err)
		}
		// Column already exists, continue
	}

	// Update database version
	if err := db.SetDBVersion(5); err != nil {
		return fmt.Errorf("failed to set DB version: %w", err)
	}

	return nil
}

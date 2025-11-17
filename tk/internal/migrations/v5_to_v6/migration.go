package v5_to_v6

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

// Migrate runs the v5 to v6 migration.
//
// This migration:
//  1. Creates container_kinds table
//  2. Creates containers table
//  3. Creates container_members table
//  4. Adds indexes for performance
//  5. Updates db_version to 6
//
// This adds support for primitive containers (queue, stack, group) with
// event-sourced kind definitions. See tk/specs/v6-event-defined-capabilities.md
//
// This migration is safe to run multiple times (idempotent).
func Migrate(db DB) error {
	// Create container_kinds table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS container_kinds (
			name TEXT PRIMARY KEY,
			primitive TEXT NOT NULL,
			description TEXT,
			deprecated INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			created_by TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create container_kinds table: %w", err)
	}

	// Create containers table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS containers (
			id TEXT PRIMARY KEY,
			primitive TEXT NOT NULL,
			kind TEXT NOT NULL,
			name TEXT NOT NULL,
			metadata TEXT,
			removed INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (kind) REFERENCES container_kinds(name)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create containers table: %w", err)
	}

	// Create container_members table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS container_members (
			container_id TEXT NOT NULL,
			item_id TEXT NOT NULL,
			position INTEGER,
			removed INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (container_id, item_id),
			FOREIGN KEY (container_id) REFERENCES containers(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create container_members table: %w", err)
	}

	// Create indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_container_members_container_pos
			ON container_members(container_id, position)`,
		`CREATE INDEX IF NOT EXISTS idx_containers_kind
			ON containers(kind)`,
		`CREATE INDEX IF NOT EXISTS idx_containers_removed
			ON containers(removed)`,
		`CREATE INDEX IF NOT EXISTS idx_container_members_removed
			ON container_members(removed)`,
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Update database version
	if err := db.SetDBVersion(6); err != nil {
		return fmt.Errorf("failed to set DB version: %w", err)
	}

	return nil
}

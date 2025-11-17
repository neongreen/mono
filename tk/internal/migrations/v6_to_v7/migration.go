package v6_to_v7

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

// Migrate runs the v6 to v7 migration.
//
// This migration:
//  1. Creates item_kinds table
//  2. Adds builtin item kinds (task, idea, wish, goal, requirement, constraint)
//  3. Adds item_kind column to tasks table (defaults to 'task')
//  4. Creates index on item_kind for performance
//  5. Updates db_version to 7
//
// This adds support for custom item kinds that can be defined at runtime via events.
//
// This migration is safe to run multiple times (idempotent).
func Migrate(db DB) error {
	// Create item_kinds table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS item_kinds (
			name TEXT PRIMARY KEY,
			description TEXT,
			llm_hint TEXT,
			builtin INTEGER NOT NULL DEFAULT 0,
			deprecated INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL,
			created_by TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create item_kinds table: %w", err)
	}

	// Insert builtin item kinds
	builtinKinds := []struct {
		name        string
		description string
		llmHint     string
	}{
		{"task", "Work items to be done", "Use for tracking work that needs to be completed"},
		{"idea", "Ideas to explore or consider", "Use for capturing ideas that may become tasks later"},
		{"wish", "Desired features or improvements", "Use for wishlist items or nice-to-have features"},
		{"goal", "High-level objectives", "Use for tracking goals and objectives"},
		{"requirement", "Requirements or specifications", "Use for documenting requirements and specifications"},
		{"constraint", "Constraints or limitations", "Use for documenting constraints and limitations"},
	}

	for _, kind := range builtinKinds {
		_, err = db.Exec(`
			INSERT OR IGNORE INTO item_kinds (name, description, llm_hint, builtin, deprecated, created_at, created_by)
			VALUES (?, ?, ?, 1, 0, 0, 'system')
		`, kind.name, kind.description, kind.llmHint)
		if err != nil {
			return fmt.Errorf("failed to insert builtin %s kind: %w", kind.name, err)
		}
	}

	// Add item_kind column to tasks table (defaults to 'task')
	// This will fail gracefully if column already exists (idempotent)
	_, err = db.Exec(`
		ALTER TABLE tasks ADD COLUMN item_kind TEXT NOT NULL DEFAULT 'task'
	`)
	if err != nil {
		// SQLite returns error if column already exists - this is OK for idempotent migration
		// We just continue - the important thing is the column exists after this runs
	}

	// Create index on item_kind for performance
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tasks_item_kind ON tasks(item_kind)
	`)
	if err != nil {
		return fmt.Errorf("failed to create item_kind index: %w", err)
	}

	// Update database version
	if err := db.SetDBVersion(7); err != nil {
		return fmt.Errorf("failed to set DB version: %w", err)
	}

	return nil
}

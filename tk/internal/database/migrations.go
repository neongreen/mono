package database

import (
	"fmt"

	v5_to_v6 "github.com/neongreen/mono/tk/internal/migrations/v5_to_v6"
	v6_to_v7 "github.com/neongreen/mono/tk/internal/migrations/v6_to_v7"
)

const (
	// MinSupportedDBVersion is the minimum database version this binary can work with
	MinSupportedDBVersion = 5

	// MaxSupportedDBVersion is the maximum database version this binary can work with
	MaxSupportedDBVersion = 8
)

// Migration represents a database migration
type Migration struct {
	FromVersion int
	ToVersion   int
	Name        string
	Run         func(*DB) error
}

// migrations is the list of all available migrations
var migrations = []Migration{
	{
		FromVersion: 5,
		ToVersion:   6,
		Name:        "Add container primitives (queue/stack/group)",
		Run: func(db *DB) error {
			return v5_to_v6.Migrate(db)
		},
	},
	{
		FromVersion: 6,
		ToVersion:   7,
		Name:        "Add item kinds (task/decision/resource/etc)",
		Run: func(db *DB) error {
			return v6_to_v7.Migrate(db)
		},
	},
	{
		FromVersion: 7,
		ToVersion:   8,
		Name:        "Add soft deletion (deleted columns)",
		Run: func(db *DB) error {
			// Drop and recreate projection tables with new schema
			// This is simpler than ALTER TABLE and guaranteed correct
			tx, err := db.Db.Begin()
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}
			defer tx.Rollback()

			// Drop projection tables (events table is preserved!)
			tables := []string{"task_numbers", "tasks", "projects"}
			for _, table := range tables {
				if _, err := tx.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
					return fmt.Errorf("failed to drop %s: %w", table, err)
				}
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit: %w", err)
			}

			// Recreate tables with new schema (includes deleted columns)
			if err := db.CreateProjectTables(); err != nil {
				return fmt.Errorf("failed to recreate tables: %w", err)
			}

			// Rebuild projections from events
			if err := db.RebuildProjections(); err != nil {
				return fmt.Errorf("failed to rebuild projections: %w", err)
			}

			return nil
		},
	},
}

// RunMigrationsIfNeeded checks the current database version and runs any pending migrations.
//
// This function is called automatically when opening a database with OpenExistingDB().
// Migrations run in sequence from the current version to the latest version.
//
// Following design principle from tk-283: "Upgrades never require immediate action"
// - Migrations run automatically without user prompts
// - Data becomes immediately visible and operable after migration
// - No forced resolution of issues
func (d *DB) RunMigrationsIfNeeded() error {
	currentVersion, err := d.GetDBVersion()
	if err != nil {
		return fmt.Errorf("failed to get current database version: %w", err)
	}

	// Check version compatibility
	if currentVersion > MaxSupportedDBVersion {
		return fmt.Errorf(
			"database version %d is too new for this version of tk (max supported: %d)\n"+
				"Please upgrade tk to a newer version that supports this database",
			currentVersion,
			MaxSupportedDBVersion,
		)
	}

	if currentVersion < MinSupportedDBVersion {
		return fmt.Errorf(
			"database version %d is too old for this version of tk (min supported: %d)\n"+
				"Please use an older version of tk to migrate this database first",
			currentVersion,
			MinSupportedDBVersion,
		)
	}

	// Find migrations that need to run
	var pendingMigrations []Migration
	for _, m := range migrations {
		if m.FromVersion >= currentVersion && m.ToVersion > currentVersion {
			pendingMigrations = append(pendingMigrations, m)
		}
	}

	// No migrations needed
	if len(pendingMigrations) == 0 {
		return nil
	}

	// Run migrations in sequence
	for _, m := range pendingMigrations {
		fmt.Printf("Migrating database v%d → v%d: %s\n", m.FromVersion, m.ToVersion, m.Name)
		if err := m.Run(d); err != nil {
			return fmt.Errorf("migration %d→%d failed: %w", m.FromVersion, m.ToVersion, err)
		}
	}

	fmt.Printf("✓ Database migration complete. Current version: %d\n", pendingMigrations[len(pendingMigrations)-1].ToVersion)
	return nil
}

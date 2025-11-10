package database

import (
	"fmt"

	v4_to_v5 "github.com/neongreen/mono/tk/internal/migrations/v4_to_v5"
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
		FromVersion: 4,
		ToVersion:   5,
		Name:        "Add synthetic projects support",
		Run: func(db *DB) error {
			return v4_to_v5.Migrate(db)
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

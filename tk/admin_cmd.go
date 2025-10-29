package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Administrative commands for tk",
	Long:  `Administrative commands for tk database management and maintenance.`,
}

var rollbackV4Cmd = &cobra.Command{
	Use:   "rollback-v4",
	Short: "Rollback v4 migration and restore v3 backup",
	Long: `Rollback the v4 migration by restoring the v3 backup.

This command:
1. Restores tk.db.v3.bak as tk.db
2. Resets meta.version_major = 3
3. Allows you to use v1/v2 binaries again

The v4 segments in v4/ remain untouched and can be ignored.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := GetDBPath()
		if err != nil {
			return err
		}

		// Perform rollback
		fmt.Println("Rolling back v4 migration...")
		fmt.Printf("Restoring backup from %s%s\n", path, v4BackupSuffix)

		if err := RollbackV4(path); err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}

		// Open the restored database and reset version
		db, err := OpenDB(path)
		if err != nil {
			return fmt.Errorf("failed to open restored database: %w", err)
		}
		defer db.Close()

		// Set version back to 3
		if err := db.SetDBVersion(v3SpecVersion); err != nil {
			return fmt.Errorf("failed to reset version: %w", err)
		}

		fmt.Println("Rollback complete!")
		fmt.Println("You can now use v1/v2 tk binaries with this database.")
		return nil
	},
}

func init() {
	adminCmd.AddCommand(rollbackV4Cmd)
	adminCmd.AddCommand(validateMigrationCmd)
	adminCmd.AddCommand(rebuildFromRemoteCmd)

	// Add flags to rebuild-from-remote command
	rebuildFromRemoteCmd.Flags().Bool("debug", false, "Enable debug output with detailed information about each step")
}

var validateMigrationCmd = &cobra.Command{
	Use:   "validate-migration",
	Short: "Validate database for v4 migration compatibility",
	Long: `Validate the database for v4 migration without performing the migration.

This command checks for common issues that might cause migration to fail:
- Tasks with empty or malformed task IDs
- Tasks referencing non-existent prefixes
- Events with missing required fields

Run this before attempting v4 migration to identify potential issues.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Check if already migrated
		version, err := db.GetDBVersion()
		if err != nil {
			return fmt.Errorf("failed to get database version: %w", err)
		}
		if version >= 4 {
			fmt.Println("Database is already at version 4 or higher")
			return nil
		}

		fmt.Println("Validating database for v4 migration...")
		fmt.Println()

		var issues []string

		// Check for task.created events with empty or malformed task IDs
		rows, err := db.db.Query(`
SELECT id, payload 
FROM events 
WHERE kind = 'task.created'
ORDER BY ts
`)
		if err != nil {
			return fmt.Errorf("failed to query events: %w", err)
		}
		defer rows.Close()

		taskCount := 0
		for rows.Next() {
			var eventID string
			var payload []byte
			if err := rows.Scan(&eventID, &payload); err != nil {
				return fmt.Errorf("failed to scan event: %w", err)
			}

			var taskPayload TaskCreatedPayload
			if err := json.Unmarshal(payload, &taskPayload); err != nil {
				issues = append(issues, fmt.Sprintf("Event %s: failed to parse task.created payload: %v", eventID, err))
				continue
			}

			taskCount++

			// Check if TaskID is empty
			if taskPayload.TaskID == "" {
				issues = append(issues, fmt.Sprintf("Event %s: task.created has empty task_id (uuid=%s, title=%q)",
					eventID, taskPayload.TaskUUID, taskPayload.Title))
				continue
			}

			// Try to parse TaskID
			parts := strings.Split(taskPayload.TaskID, "-")
			if len(parts) != 3 {
				issues = append(issues, fmt.Sprintf("Event %s: task.created has malformed task_id=%q (expected format: prefix-number-node)",
					eventID, taskPayload.TaskID))
				continue
			}

			// Check if number is valid
			if _, err := strconv.ParseInt(parts[1], 10, 64); err != nil {
				issues = append(issues, fmt.Sprintf("Event %s: task.created has invalid number in task_id=%q: %v",
					eventID, taskPayload.TaskID, err))
			}
		}

		if err := rows.Err(); err != nil {
			return fmt.Errorf("error iterating events: %w", err)
		}

		// Check for prefixes
		var prefixCount int
		err = db.db.QueryRow("SELECT COUNT(*) FROM prefixes WHERE removed = 0").Scan(&prefixCount)
		if err != nil {
			return fmt.Errorf("failed to count prefixes: %w", err)
		}

		fmt.Printf("Found %d task.created events\n", taskCount)
		fmt.Printf("Found %d active prefixes\n", prefixCount)
		fmt.Println()

		if len(issues) == 0 {
			fmt.Println("✓ No issues found - database appears ready for v4 migration")
			return nil
		}

		fmt.Printf("✗ Found %d issue(s):\n\n", len(issues))
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
		fmt.Println()
		fmt.Println("These issues may cause migration to fail. Please review and fix them before migrating.")
		return fmt.Errorf("validation found %d issue(s)", len(issues))
	},
}

var rebuildFromRemoteCmd = &cobra.Command{
	Use:   "rebuild-from-remote [remote-name]",
	Short: "Rebuild local database from remote segments",
	Long: `Rebuild the local database from scratch using segments from a remote.

This command:
1. Deletes the local database (creates a backup first)
2. Creates a fresh v3 database
3. Ingests all segments from the specified remote
4. Migrates the database to v4

This is useful when:
- The local database is corrupted
- You want to start fresh from remote data
- Migration from v3 failed and you want to retry

Examples:
  tk admin rebuild-from-remote icloud
  tk admin rebuild-from-remote icloud --debug
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		remoteName := args[0]
		debug, _ := cmd.Flags().GetBool("debug")

		if debug {
			fmt.Println("Debug mode enabled")
		}

		// Get database path
		dbPath, err := GetDBPath()
		if err != nil {
			return err
		}

		if debug {
			fmt.Printf("Database path: %s\n", dbPath)
		}

		// Load config to verify remote exists
		config, err := LoadConfig()
		if err != nil {
			return err
		}

		remote, exists := config.Remotes[remoteName]
		if !exists {
			return fmt.Errorf("remote '%s' not found in config", remoteName)
		}

		if debug {
			fmt.Printf("Remote config: %+v\n", remote)
		}

		// Create backup of current database
		backupPath := dbPath + ".pre-rebuild.bak"
		if _, err := os.Stat(dbPath); err == nil {
			fmt.Printf("Creating backup at %s\n", backupPath)
			if err := copyFile(dbPath, backupPath); err != nil {
				return fmt.Errorf("failed to create backup: %w", err)
			}
			if debug {
				info, _ := os.Stat(backupPath)
				fmt.Printf("Backup created: size=%d bytes\n", info.Size())
			}
		}

		// Delete current database
		if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove database: %w", err)
		}

		fmt.Println("Creating fresh v3 database...")

		// Create new database
		db, err := OpenDB(dbPath)
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}

		// Initialize with v3 schema (don't create v4 tables yet)
		if err := db.InitDB(); err != nil {
			db.Close()
			return fmt.Errorf("failed to initialize database: %w", err)
		}

		if debug {
			fmt.Println("Database initialized with v3 schema")
		}

		// Set version to 3 (pre-migration)
		if err := db.SetDBVersion(3); err != nil {
			db.Close()
			return fmt.Errorf("failed to set version: %w", err)
		}

		if debug {
			version, _ := db.GetDBVersion()
			fmt.Printf("Database version set to: %d\n", version)

			// Show database schema
			rows, err := db.db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
			if err == nil {
				var tables []string
				for rows.Next() {
					var name string
					if err := rows.Scan(&name); err == nil {
						tables = append(tables, name)
					}
				}
				rows.Close()
				fmt.Printf("Database tables: %v\n", tables)
			}
		}

		fmt.Printf("Ingesting segments from remote '%s'...\n", remoteName)

		// Ingest from remote (v3 events)
		if err := ingestRemote(db, remoteName, remote); err != nil {
			db.Close()
			return fmt.Errorf("failed to ingest from remote: %w", err)
		}

		if debug {
			// Count events
			var eventCount int
			db.db.QueryRow("SELECT COUNT(*) FROM events").Scan(&eventCount)
			fmt.Printf("Total events in database: %d\n", eventCount)
		}

		db.Close()

		fmt.Println()
		fmt.Println("✓ Events ingested successfully!")
		fmt.Println()

		// Reopen database and run migration to v4
		db, err = OpenDB(dbPath)
		if err != nil {
			return fmt.Errorf("failed to reopen database: %w", err)
		}

		if err := db.InitDB(); err != nil {
			db.Close()
			return fmt.Errorf("failed to initialize: %w", err)
		}

		// Check if v4 migration is needed (should be true since we set version to 3)
		needsMigration, err := db.NeedsMigrationToV4()
		if err != nil {
			db.Close()
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if debug {
			fmt.Printf("Needs migration to v4: %v\n", needsMigration)
		}

		if needsMigration {
			fmt.Println("Migrating database to v4...")
			fmt.Printf("Creating migration backup at %s%s\n", dbPath, v4BackupSuffix)

			if err := db.MigrateToV4(dbPath); err != nil {
				db.Close()
				return fmt.Errorf("failed to migrate to v4: %w", err)
			}

			fmt.Println("Migration to v4 complete!")

			if debug {
				version, _ := db.GetDBVersion()
				fmt.Printf("Database version after migration: %d\n", version)

				// Show database schema after migration
				rows, err := db.db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
				if err == nil {
					var tables []string
					for rows.Next() {
						var name string
						if err := rows.Scan(&name); err == nil {
							tables = append(tables, name)
						}
					}
					rows.Close()
					fmt.Printf("Database tables after migration: %v\n", tables)
				}
			}
		}

		fmt.Println()
		fmt.Println("Running health check...")

		report, err := runDoctor(db)
		if err != nil {
			db.Close()
			return fmt.Errorf("doctor check failed: %w", err)
		}

		db.Close()

		printDoctorReport(os.Stdout, report)

		if report.ProblemCount() > 0 {
			fmt.Println()
			fmt.Println("Note: You can resolve these issues and rerun 'tk doctor' at any time.")
		}

		return nil
	},
}

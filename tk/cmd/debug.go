package cmd

import (
	"fmt"
	"os"

	conflicts_pkg "github.com/neongreen/mono/tk/cmd/conflicts"
	debug_pkg "github.com/neongreen/mono/tk/cmd/debug"
	events_pkg "github.com/neongreen/mono/tk/cmd/debug/events"
	node_pkg "github.com/neongreen/mono/tk/cmd/debug/node"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Parent command; JSON only required for read-only data commands
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debugging and diagnostic commands",
	Long:  `Commands for debugging, diagnostics, and database management.`,
}

func init() {
	// Database management
	debugCmd.AddCommand(rebuildFromRemoteCmd)
	debugCmd.AddCommand(fixTimestampsCmd)
	debugCmd.AddCommand(debug_pkg.RebuildCmd)

	// Diagnostic commands
	debugCmd.AddCommand(idCmd)
	debugCmd.AddCommand(nodeCmd)
	debugCmd.AddCommand(eventsCmd)
	debugCmd.AddCommand(debug_pkg.DoctorCmd)
	debugCmd.AddCommand(debug_pkg.RepairCmd)
	debugCmd.AddCommand(conflicts_pkg.NumbersCmd)

	// Add flags to rebuild-from-remote command
	rebuildFromRemoteCmd.Flags().Bool("debug", false, "Enable debug output with detailed information about each step")
}

// nodeCmd is a parent command for node-related debug commands
// cobralint:exemptjson reason: Parent command; JSON only required for read-only data commands
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage node ID",
}

// eventsCmd is a parent command for event-related debug commands
// cobralint:exemptjson reason: Parent command; JSON only required for read-only data commands
var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Debug commands for inspecting events",
}

func init() {
	// Add node subcommands
	nodeCmd.AddCommand(node_pkg.ShowCmd)
	nodeCmd.AddCommand(node_pkg.RegenCmd)

	// Add events subcommands
	eventsCmd.AddCommand(events_pkg.ListCmd)
	eventsCmd.AddCommand(events_pkg.ShowCmd)
	eventsCmd.AddCommand(events_pkg.StatsCmd)
}

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var fixTimestampsCmd = &cobra.Command{
	Use:   "fix-timestamps",
	Short: "Reassign Lamport timestamps to events based on creation time",
	Long: `Reassign Lamport timestamps to all events based on their CreatedAt field.

Events will be assigned sequential timestamps (1, 2, 3, ...)
in order of their CreatedAt time.

This command is safe to run multiple times.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		fmt.Println("Reassigning Lamport timestamps based on creation time...")

		// Get all events ordered by CreatedAt, then by ID for deterministic ordering
		rows, err := db.Query(`
			SELECT id FROM events
			ORDER BY created_at ASC, id ASC
		`)
		if err != nil {
			return fmt.Errorf("failed to query events: %w", err)
		}

		var eventIDs []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return fmt.Errorf("failed to scan event ID: %w", err)
			}
			eventIDs = append(eventIDs, id)
		}
		rows.Close()

		if len(eventIDs) == 0 {
			fmt.Println("No events found")
			return nil
		}

		fmt.Printf("Updating %d events...\n", len(eventIDs))

		// Begin transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback()

		// Update each event with its new timestamp
		stmt, err := tx.Prepare(`UPDATE events SET ts = ? WHERE id = ?`)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		for i, eventID := range eventIDs {
			newTS := int64(i + 1) // Start from 1
			if _, err := stmt.Exec(newTS, eventID); err != nil {
				return fmt.Errorf("failed to update event %s: %w", eventID, err)
			}
		}

		// Update the metadata to track the latest timestamp
		_, err = tx.Exec(`
			INSERT OR REPLACE INTO metadata (key, value)
			VALUES ('lamport_ts', ?)
		`, int64(len(eventIDs)))
		if err != nil {
			return fmt.Errorf("failed to update lamport_ts metadata: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		fmt.Println("✓ Timestamps updated successfully!")
		fmt.Printf("Latest Lamport timestamp: %d\n", len(eventIDs))
		return nil
	},
}

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var rebuildFromRemoteCmd = &cobra.Command{
	Use:   "rebuild-from-remote [remote-name]",
	Short: "Rebuild local database from remote segments",
	Long: `Rebuild the local database from scratch using segments from a remote.

This command:
1. Deletes the local database (creates a backup first)
2. Creates a fresh database
3. Ingests all segments from the specified remote

This is useful when:
- The local database is corrupted
- You want to start fresh from remote data

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
		dbPath, err := database.GetDBPath()
		if err != nil {
			return err
		}

		if debug {
			fmt.Printf("Database path: %s\n", dbPath)
		}

		// Load config to verify remote exists
		config, err := config_pkg.LoadConfig()
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
			data, err := os.ReadFile(dbPath)
			if err != nil {
				return fmt.Errorf("failed to read database: %w", err)
			}
			if err := os.WriteFile(backupPath, data, 0o644); err != nil {
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

		fmt.Println("Creating fresh database...")

		// Create new database
		db, err := database.OpenDB(dbPath)
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}

		// Initialize schema
		if err := db.InitDB(); err != nil {
			db.Close()
			return fmt.Errorf("failed to initialize database: %w", err)
		}

		// Ensure DB version is set to 4
		if err := db.SetDBVersion(4); err != nil {
			db.Close()
			return fmt.Errorf("failed to set version: %w", err)
		}

		if debug {
			fmt.Println("Database initialized")

			// Show database schema
			rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
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

		// Ingest from remote
		if err := IngestRemote(db, remoteName, remote); err != nil {
			db.Close()
			return fmt.Errorf("failed to ingest from remote: %w", err)
		}

		// Run migrations to bring DB to latest version
		if err := db.RunMigrationsIfNeeded(); err != nil {
			db.Close()
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		if debug {
			// Count events
			var eventCount int
			db.QueryRow("SELECT COUNT(*) FROM events").Scan(&eventCount)
			fmt.Printf("Total events in database: %d\n", eventCount)
		}

		fmt.Println()
		fmt.Println("✓ Events ingested successfully!")
		fmt.Println()

		fmt.Println("Running health check...")

		report, err := debug_pkg.RunDoctor(db)
		if err != nil {
			db.Close()
			return fmt.Errorf("doctor check failed: %w", err)
		}

		db.Close()

		debug_pkg.PrintDoctorReport(os.Stdout, report)

		if report.ProblemCount() > 0 {
			fmt.Println()
			fmt.Println("Note: You can resolve these issues and rerun 'tk doctor' at any time.")
		}

		return nil
	},
}

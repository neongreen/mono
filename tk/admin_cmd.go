package main

import (
	"encoding/json"
	"fmt"
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

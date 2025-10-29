package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Administrative commands for tk",
	Long:  `Administrative commands for tk database management and maintenance.`,
}

func init() {
	adminCmd.AddCommand(fixTimestampsCmd)
}

var fixTimestampsCmd = &cobra.Command{
	Use:   "fix-timestamps",
	Short: "Reassign Lamport timestamps to events based on creation time",
	Long: `Reassign Lamport timestamps to all events based on their CreatedAt field.

Events will be assigned sequential timestamps (1, 2, 3, ...)
in order of their CreatedAt time.

This command is safe to run multiple times.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		fmt.Println("Reassigning Lamport timestamps based on creation time...")

		// Get all events ordered by CreatedAt, then by ID for deterministic ordering
		rows, err := db.db.Query(`
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
		tx, err := db.db.Begin()
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


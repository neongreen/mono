package debug

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

// RepairCmd repairs database inconsistencies
var RepairCmd = &cobra.Command{
	Use:   "unsafe-repair-timestamps",
	Short: "UNSAFE: Repair Lamport timestamp corruption",
	Long: `UNSAFE: Repair Lamport timestamp corruption detected by 'tk debug doctor'.

⚠️  DANGER - MULTI-MACHINE SYNC INCOMPATIBLE ⚠️

This command modifies event Lamport timestamps in place. If you use 'tk sync':

1. DO NOT run this on one machine and expect sync to propagate the fix
2. Events with same ID but different TS will NOT sync (duplicates are skipped)
3. Machines will become permanently out of sync

SAFE USAGE FOR MULTI-MACHINE SETUPS:

1. Run 'tk debug doctor' on all machines to confirm same corruption
2. DELETE ALL REMOTE SEGMENTS (clear the remote completely):
     rm -rf ~/Library/Mobile\ Documents/com~apple~CloudDocs/tk-events/personal/segments/*
     rm -f ~/Library/Mobile\ Documents/com~apple~CloudDocs/tk-events/personal/index.json
3. Run this repair on ONE machine (the "source of truth")
4. Run 'tk export' then 'tk push <remote>' to repopulate remote with fixed timestamps
5. On OTHER machines: DELETE local database, then 'tk pull <remote>' + 'tk ingest'
6. Verify: 'tk debug doctor' should be clean on all machines

SINGLE-MACHINE USAGE:
  If you don't use sync, this is safe to run.

Options:
  --dry-run   Show what would be fixed without making changes

Examples:
  tk debug doctor                                 # See corruption
  tk debug unsafe-repair-timestamps --dry-run     # Preview fix
  tk debug unsafe-repair-timestamps               # Apply fix`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		return repairLamportTimestamps(db, dryRun)
	},
}

func init() {
	RepairCmd.Flags().Bool("dry-run", false, "Show what would be fixed without making changes")
}

// repairLamportTimestamps recomputes all Lamport timestamps based on wall clock order
func repairLamportTimestamps(db *database.DB, dryRun bool) error {
	// Get events in wall-clock order (this is the "correct" chronological order)
	rows, err := db.Db.Query(`
		SELECT id, ts, created_at
		FROM events
		ORDER BY created_at, id
	`)
	if err != nil {
		return fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	type update struct {
		ID    string
		OldTS int64
		NewTS int64
	}
	var updates []update

	newTS := int64(1)
	for rows.Next() {
		var id string
		var oldTS int64
		var createdAtNano int64
		if err := rows.Scan(&id, &oldTS, &createdAtNano); err != nil {
			return fmt.Errorf("failed to scan event: %w", err)
		}

		if oldTS != newTS {
			updates = append(updates, update{
				ID:    id,
				OldTS: oldTS,
				NewTS: newTS,
			})
		}
		newTS++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating events: %w", err)
	}

	if len(updates) == 0 {
		fmt.Println("No Lamport timestamp repairs needed - all timestamps are correct")
		return nil
	}

	fmt.Printf("Found %d events with incorrect Lamport timestamps\n", len(updates))

	if dryRun {
		fmt.Println("\nPreview of changes (first 10 and last 5):")
		for i, u := range updates {
			if i < 10 || i >= len(updates)-5 {
				fmt.Printf("  %s: TS %d → %d\n", u.ID, u.OldTS, u.NewTS)
			} else if i == 10 {
				fmt.Printf("  ... (%d more) ...\n", len(updates)-15)
			}
		}
		fmt.Printf("\nWould update Lamport counter to: %d\n", newTS-1)
		fmt.Println("\nTo apply changes, run without --dry-run")
		return nil
	}

	// Apply updates in transaction
	tx, err := db.Db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`UPDATE events SET ts = ? WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, u := range updates {
		if _, err := stmt.Exec(u.NewTS, u.ID); err != nil {
			return fmt.Errorf("failed to update event %s: %w", u.ID, err)
		}
	}

	// Update Lamport counter
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO metadata (key, value)
		VALUES ('lamport_counter', ?)
	`, newTS-1)
	if err != nil {
		return fmt.Errorf("failed to update lamport counter: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("\n✓ Repaired %d Lamport timestamps\n", len(updates))
	fmt.Printf("✓ Updated Lamport counter to: %d\n", newTS-1)
	fmt.Println("\nRecommendation: Run 'tk debug doctor' to verify the repairs")

	return nil
}

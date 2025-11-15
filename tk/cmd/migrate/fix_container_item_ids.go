package migrate

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var FixContainerItemIDsCmd = &cobra.Command{
	Use:   "fix-container-item-ids",
	Short: "Fix display IDs in container_members (convert to task UIDs)",
	Long: `Scans container_members for display IDs (tk-123) and converts them to
task UIDs (tsk_01ABC...).

This fixes data created before task reference resolution was implemented
in container membership operations (tk-426).

The command will:
  1. Find all container members with display ID format (non-UID strings)
  2. Resolve each display ID to its task UID
  3. Update container_members.item_id in place
  4. Report conversion statistics

This is a direct table update, not event-based, since it's fixing
historical data corruption.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Check database version
		version, err := db.GetDBVersion()
		if err != nil {
			return err
		}
		if version < 6 {
			fmt.Println("✓ Database version < 6, no container data to fix")
			return nil
		}

		// Find all container members
		rows, err := db.Db.Query(`
			SELECT container_id, item_id
			FROM container_members
		`)
		if err != nil {
			return fmt.Errorf("failed to query container members: %w", err)
		}
		defer rows.Close()

		type Member struct {
			ContainerID string
			ItemID      string
		}

		var members []Member
		for rows.Next() {
			var containerID, itemID string
			if err := rows.Scan(&containerID, &itemID); err != nil {
				return fmt.Errorf("failed to scan member: %w", err)
			}
			members = append(members, Member{ContainerID: containerID, ItemID: itemID})
		}

		if len(members) == 0 {
			fmt.Println("✓ No container members found")
			return nil
		}

		fmt.Printf("Found %d container member(s)\n", len(members))

		// Process each member
		fixedCount := 0
		skippedCount := 0
		errorCount := 0

		for _, m := range members {
			// Check if item_id is already a task UID
			if types.NewTaskRef(m.ItemID).IsTaskUID() {
				skippedCount++
				continue
			}

			// Try to resolve display ID to task UID
			taskUID, err := database.ResolveTaskReference(db, types.NewTaskRef(m.ItemID))
			if err != nil {
				fmt.Printf("⚠ Failed to resolve %s in %s: %v\n", m.ItemID, m.ContainerID, err)
				errorCount++
				continue
			}

			// Update container_members in place
			_, err = db.Db.Exec(`
				UPDATE container_members
				SET item_id = ?
				WHERE container_id = ? AND item_id = ?
			`, taskUID, m.ContainerID, m.ItemID)
			if err != nil {
				fmt.Printf("⚠ Failed to update %s in %s: %v\n", m.ItemID, m.ContainerID, err)
				errorCount++
				continue
			}

			fmt.Printf("✓ Fixed %s → %s (in %s)\n", m.ItemID, taskUID, m.ContainerID)
			fixedCount++
		}

		// Summary
		fmt.Println()
		fmt.Printf("Summary:\n")
		fmt.Printf("  Fixed:   %d\n", fixedCount)
		fmt.Printf("  Skipped: %d (already task UIDs)\n", skippedCount)
		if errorCount > 0 {
			fmt.Printf("  Errors:  %d\n", errorCount)
		}

		return nil
	},
}

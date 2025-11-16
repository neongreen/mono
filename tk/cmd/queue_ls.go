package cmd

import (
	"database/sql"
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var queueLsCmd = &cobra.Command{
	Use:   "queue-ls",
	Short: "List queues or queue members",
	Long: `List all queues, or list members of a specific queue.

Examples:
  tk queue list       # List all queues
  tk queue list q-1   # List members of queue q-1`,
	Args: cobra.MaximumNArgs(1),
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
			return fmt.Errorf("containers require database v6 or higher, but current version is v%d", version)
		}

		// If queue ID provided, list members
		if len(args) == 1 {
			return listQueueMembers(db, args[0])
		}

		// Otherwise list all queues
		rows, err := db.Db.Query(`
			SELECT id, kind, name, metadata
			FROM containers
			WHERE primitive = ? AND removed = 0
			ORDER BY id
		`, types.PrimitiveQueue)
		if err != nil {
			return fmt.Errorf("failed to query queues: %w", err)
		}
		defer rows.Close()

		// Print header
		fmt.Printf("%-8s %-12s %-30s\n", "ID", "KIND", "NAME")
		fmt.Println("────────────────────────────────────────────────────────────────")

		count := 0
		for rows.Next() {
			var id string
			var kind string
			var name string
			var metadata sql.NullString

			if err := rows.Scan(&id, &kind, &name, &metadata); err != nil {
				return fmt.Errorf("failed to scan row: %w", err)
			}

			fmt.Printf("%-8s %-12s %-30s\n", id, kind, name)
			count++
		}

		if count == 0 {
			fmt.Println("No queues found.")
			fmt.Println("\nCreate one with: tk queue create <kind> <name>")
		}

		return nil
	},
}

func listQueueMembers(db *database.DB, queueID string) error {
	// Verify queue exists
	var primitive string
	var queueName string
	var removed int
	err := db.Db.QueryRow(`
		SELECT primitive, name, removed
		FROM containers
		WHERE id = ?
	`, queueID).Scan(&primitive, &queueName, &removed)
	if err != nil {
		return fmt.Errorf("queue %q not found", queueID)
	}

	if primitive != string(types.PrimitiveQueue) {
		return fmt.Errorf("%q is a %s, not a queue", queueID, primitive)
	}

	if removed == 1 {
		return fmt.Errorf("queue %q has been removed", queueID)
	}

	// Query members
	rows, err := db.Db.Query(`
		SELECT item_id, position
		FROM container_members
		WHERE container_id = ? AND removed = 0
		ORDER BY position ASC
	`, queueID)
	if err != nil {
		return fmt.Errorf("failed to query members: %w", err)
	}
	defer rows.Close()

	// Print header
	fmt.Printf("Queue %s: %s\n", queueID, queueName)
	fmt.Println()
	fmt.Printf("%-4s %s\n", "POS", "ITEM")
	fmt.Println("─────────────────────")

	count := 0
	for rows.Next() {
		var itemID string
		var position int64

		if err := rows.Scan(&itemID, &position); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Render task UID as display ID (tk-123) for user output
		displayID, err := database.RenderTaskDisplayID(db, itemID)
		if err != nil {
			// If rendering fails, show the raw UID (defensive)
			displayID = itemID
		}

		fmt.Printf("%-4d %s\n", position, displayID)
		count++
	}

	if count == 0 {
		fmt.Println("(empty)")
	} else {
		fmt.Printf("\n%d items in queue\n", count)
	}

	return nil
}

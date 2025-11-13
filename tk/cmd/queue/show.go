package queue

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var ShowCmd = &cobra.Command{
	Use:   "show <queue-id>",
	Short: "Show queue details",
	Long: `Display detailed information about a queue.

Example:
  tk queue show q-1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		queueID := args[0]

		// Query queue details
		var primitive string
		var kind string
		var name string
		var metadata sql.NullString
		var removed int
		err = db.Db.QueryRow(`
			SELECT primitive, kind, name, metadata, removed
			FROM containers
			WHERE id = ?
		`, queueID).Scan(&primitive, &kind, &name, &metadata, &removed)
		if err != nil {
			return fmt.Errorf("queue %q not found", queueID)
		}

		if primitive != string(types.PrimitiveQueue) {
			return fmt.Errorf("%q is a %s, not a queue", queueID, primitive)
		}

		// Count members
		var memberCount int
		err = db.Db.QueryRow(`
			SELECT COUNT(*)
			FROM container_members
			WHERE container_id = ? AND removed = 0
		`, queueID).Scan(&memberCount)
		if err != nil {
			return fmt.Errorf("failed to count members: %w", err)
		}

		// Display
		fmt.Printf("Queue: %s\n", queueID)
		fmt.Printf("Name: %s\n", name)
		fmt.Printf("Kind: %s\n", kind)
		fmt.Printf("Members: %d\n", memberCount)

		if removed == 1 {
			fmt.Println("Status: REMOVED")
		}

		if metadata.Valid && metadata.String != "" && metadata.String != "null" {
			fmt.Println("\nMetadata:")
			// Pretty-print JSON metadata
			var metaMap map[string]any
			if err := json.Unmarshal([]byte(metadata.String), &metaMap); err == nil {
				for k, v := range metaMap {
					fmt.Printf("  %s: %v\n", k, v)
				}
			} else {
				fmt.Printf("  %s\n", metadata.String)
			}
		}

		return nil
	},
}

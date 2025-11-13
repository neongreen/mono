package group

import (
	"database/sql"
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all groups",
	Long:  `List all group containers.`,
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

		// Query groups
		rows, err := db.Db.Query(`
			SELECT id, kind, name, metadata
			FROM containers
			WHERE primitive = ? AND removed = 0
			ORDER BY id
		`, types.PrimitiveGroup)
		if err != nil {
			return fmt.Errorf("failed to query groups: %w", err)
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
			fmt.Println("No groups found.")
			fmt.Println("\nCreate one with: tk group create <kind> <name>")
		}

		return nil
	},
}

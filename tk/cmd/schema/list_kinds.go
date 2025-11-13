package schema

import (
	"database/sql"
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

var (
	listKindsAll bool
)

var ListKindsCmd = &cobra.Command{
	Use:   "list-kinds",
	Short: "List defined container kinds",
	Long:  `List all defined container kinds (queue, stack, group).`,
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

		// Query container kinds
		query := `
			SELECT name, primitive, description, deprecated
			FROM container_kinds
		`
		if !listKindsAll {
			query += ` WHERE deprecated = 0`
		}
		query += ` ORDER BY primitive, name`

		rows, err := db.Db.Query(query)
		if err != nil {
			return fmt.Errorf("failed to query container kinds: %w", err)
		}
		defer rows.Close()

		// Print header
		fmt.Printf("%-12s %-10s %-40s %s\n", "NAME", "PRIMITIVE", "DESCRIPTION", "STATUS")
		fmt.Println("────────────────────────────────────────────────────────────────────────────")

		count := 0
		for rows.Next() {
			var name string
			var primitive string
			var description sql.NullString
			var deprecated int

			if err := rows.Scan(&name, &primitive, &description, &deprecated); err != nil {
				return fmt.Errorf("failed to scan row: %w", err)
			}

			desc := ""
			if description.Valid {
				desc = description.String
			}

			status := "active"
			if deprecated == 1 {
				status = "deprecated"
			}

			fmt.Printf("%-12s %-10s %-40s %s\n", name, primitive, desc, status)
			count++
		}

		if count == 0 {
			fmt.Println("No container kinds defined.")
			fmt.Println("\nDefine one with: tk schema add-kind <primitive> <name>")
		}

		return nil
	},
}

func init() {
	ListKindsCmd.Flags().BoolVar(&listKindsAll, "all", false, "Show deprecated kinds")
}

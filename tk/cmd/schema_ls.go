package cmd

import (
	"database/sql"
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

var (
	listAll bool
)

var schemaLsCmd = &cobra.Command{
	Use:   "schema-ls",
	Short: "List defined schema kinds",
	Long:  `List all defined container kinds (queue, stack, group) and item kinds (task, decision, resource, etc.).`,
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

		// Show item kinds (v7+)
		if version >= 7 {
			fmt.Println("Item Kinds:")
			fmt.Printf("  %-15s %-50s %s\n", "NAME", "DESCRIPTION", "STATUS")
			fmt.Println("  ───────────────────────────────────────────────────────────────────────────────")

			itemQuery := `
				SELECT name, description, deprecated
				FROM item_kinds
			`
			if !listAll {
				itemQuery += ` WHERE deprecated = 0`
			}
			itemQuery += ` ORDER BY builtin DESC, name`

			rows, err := db.Db.Query(itemQuery)
			if err != nil {
				return fmt.Errorf("failed to query item kinds: %w", err)
			}

			itemCount := 0
			for rows.Next() {
				var name string
				var description sql.NullString
				var deprecated int

				if err := rows.Scan(&name, &description, &deprecated); err != nil {
					rows.Close()
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

				fmt.Printf("  %-15s %-50s %s\n", name, desc, status)
				itemCount++
			}
			rows.Close()

			if itemCount == 0 {
				fmt.Println("  (no item kinds defined)")
			}
			fmt.Println()
		}

		// Show container kinds (v6+)
		if version < 6 {
			return fmt.Errorf("containers require database v6 or higher, but current version is v%d", version)
		}

		fmt.Println("Container Kinds:")
		fmt.Printf("  %-15s %-10s %-40s %s\n", "NAME", "TYPE", "DESCRIPTION", "STATUS")
		fmt.Println("  ──────────────────────────────────────────────────────────────────────────────────")

		// Query container kinds
		query := `
			SELECT name, primitive, description, deprecated
			FROM container_kinds
		`
		if !listAll {
			query += ` WHERE deprecated = 0`
		}
		query += ` ORDER BY primitive, name`

		rows, err := db.Db.Query(query)
		if err != nil {
			return fmt.Errorf("failed to query container kinds: %w", err)
		}
		defer rows.Close()

		containerCount := 0
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

			fmt.Printf("  %-15s %-10s %-40s %s\n", name, primitive, desc, status)
			containerCount++
		}

		if containerCount == 0 {
			fmt.Println("  (no container kinds defined)")
		}

		return nil
	},
}

func init() {
	schemaLsCmd.Flags().BoolVar(&listAll, "all", false, "Show deprecated kinds")
	schemaLsCmd.Flags().Bool("json", false, "Output as JSON")
}

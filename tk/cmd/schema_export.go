package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

var (
	exportJSON bool
)

// SchemaExport represents the exported schema structure
type SchemaExport struct {
	QueueKinds []ContainerKindExport `json:"queue_kinds"`
	StackKinds []ContainerKindExport `json:"stack_kinds"`
	GroupKinds []ContainerKindExport `json:"group_kinds"`
}

// ContainerKindExport represents a container kind in the export
type ContainerKindExport struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

var schemaExportCmd = &cobra.Command{
	Use:   "schema-export",
	Short: "Export schema definition",
	Long: `Export schema definitions (container kinds, etc.) for use by agents and tools.

Example:
  tk schema export --json`,
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
			return fmt.Errorf("schema export requires database v6 or higher, but current version is v%d", version)
		}

		// Query all active container kinds
		rows, err := db.Db.Query(`
			SELECT name, primitive, description
			FROM container_kinds
			WHERE deprecated = 0
			ORDER BY primitive, name
		`)
		if err != nil {
			return fmt.Errorf("failed to query container kinds: %w", err)
		}
		defer rows.Close()

		export := SchemaExport{
			QueueKinds: []ContainerKindExport{},
			StackKinds: []ContainerKindExport{},
			GroupKinds: []ContainerKindExport{},
		}

		for rows.Next() {
			var name string
			var primitive string
			var description sql.NullString

			if err := rows.Scan(&name, &primitive, &description); err != nil {
				return fmt.Errorf("failed to scan row: %w", err)
			}

			kind := ContainerKindExport{
				Name: name,
			}
			if description.Valid {
				kind.Description = &description.String
			}

			switch primitive {
			case "queue":
				export.QueueKinds = append(export.QueueKinds, kind)
			case "stack":
				export.StackKinds = append(export.StackKinds, kind)
			case "group":
				export.GroupKinds = append(export.GroupKinds, kind)
			}
		}

		if exportJSON {
			// Output as JSON
			output, err := json.MarshalIndent(export, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(output))
		} else {
			// Output as human-readable text
			fmt.Println("Container Kinds Schema")
			fmt.Println("======================")
			fmt.Println()

			if len(export.QueueKinds) > 0 {
				fmt.Println("Queue kinds:")
				for _, k := range export.QueueKinds {
					if k.Description != nil {
						fmt.Printf("  - %s: %s\n", k.Name, *k.Description)
					} else {
						fmt.Printf("  - %s\n", k.Name)
					}
				}
				fmt.Println()
			}

			if len(export.StackKinds) > 0 {
				fmt.Println("Stack kinds:")
				for _, k := range export.StackKinds {
					if k.Description != nil {
						fmt.Printf("  - %s: %s\n", k.Name, *k.Description)
					} else {
						fmt.Printf("  - %s\n", k.Name)
					}
				}
				fmt.Println()
			}

			if len(export.GroupKinds) > 0 {
				fmt.Println("Group kinds:")
				for _, k := range export.GroupKinds {
					if k.Description != nil {
						fmt.Printf("  - %s: %s\n", k.Name, *k.Description)
					} else {
						fmt.Printf("  - %s\n", k.Name)
					}
				}
				fmt.Println()
			}

			if len(export.QueueKinds) == 0 && len(export.StackKinds) == 0 && len(export.GroupKinds) == 0 {
				fmt.Println("No container kinds defined.")
			}
		}

		return nil
	},
}

func init() {
	ExportCmd.Flags().BoolVar(&exportJSON, "json", false, "Output as JSON")
}

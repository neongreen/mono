package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

var projectLsCmd = &cobra.Command{
	Use:   "project-ls",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		db, err := database.OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Query projects
		rows, err := db.Db.Query(`
			SELECT project_uid, type, name, description, created_by, created_at, COALESCE(is_synthetic, 0)
			FROM projects
			ORDER BY created_at
		`)
		if err != nil {
			return fmt.Errorf("failed to query projects: %w", err)
		}
		defer rows.Close()

		// Get node ID for filtering aliases
		nodeID, err := db.GetOrCreateNodeID()
		if err != nil {
			return err
		}

		type ProjectOutput struct {
			UID                 string   `json:"uid"`
			Name                string   `json:"name"`
			Type                string   `json:"type"`
			Aliases             []string `json:"aliases"`
			LocalPreferredAlias string   `json:"local_preferred_alias,omitempty"` // First alias for this node, used for display
			Description         string   `json:"description"`
			CreatedBy           string   `json:"created_by"`
			CreatedAt           int64    `json:"created_at"`
			IsSynthetic         bool     `json:"is_synthetic"`
		}

		var projects []ProjectOutput

		for rows.Next() {
			var projectUID, typ, name, description, createdBy string
			var createdAt int64
			var isSynthetic int

			if err := rows.Scan(&projectUID, &typ, &name, &description, &createdBy, &createdAt, &isSynthetic); err != nil {
				return err
			}

			// Get aliases for this project
			aliasRows, err := db.Db.Query(`
			SELECT alias FROM project_aliases
			WHERE project_uid = ? AND node = ?
		`, projectUID, nodeID)
			if err != nil {
				return err
			}

			aliases := []string{}
			for aliasRows.Next() {
				var alias string
				if err := aliasRows.Scan(&alias); err != nil {
					aliasRows.Close()
					return err
				}
				aliases = append(aliases, alias)
			}
			aliasRows.Close()

			// Set preferred alias to first local alias (if any exist)
			localPreferredAlias := ""
			if len(aliases) > 0 {
				localPreferredAlias = aliases[0]
			}

			projects = append(projects, ProjectOutput{
				UID:                 projectUID,
				Name:                name,
				Type:                typ,
				Aliases:             aliases,
				LocalPreferredAlias: localPreferredAlias,
				Description:         description,
				CreatedBy:           createdBy,
				CreatedAt:           createdAt,
				IsSynthetic:         isSynthetic == 1,
			})
		}

		if err := rows.Err(); err != nil {
			return err
		}

		if jsonOutput {
			output, err := json.MarshalIndent(projects, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal projects: %w", err)
			}
			fmt.Println(string(output))
		} else {
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"UID", "Name", "Type", "Aliases", "Description", "Created By"})

			t.SetStyle(table.StyleLight)
			t.Style().Options.SeparateRows = true
			t.Style().Options.DrawBorder = false

			for _, project := range projects {
				aliasStr := ""
				if len(project.Aliases) > 0 {
					aliasStr = strings.Join(project.Aliases, ", ")
				}

				// Add [synthetic] marker to name if it's a synthetic project
				displayName := project.Name
				if project.IsSynthetic {
					displayName = project.Name + " [synthetic]"
				}

				t.AppendRow(table.Row{project.UID, displayName, project.Type, aliasStr, project.Description, project.CreatedBy})
			}

			t.Render()
		}

		return nil
	},
}

func init() {
	projectLsCmd.Flags().Bool("json", false, "Output as JSON")
}

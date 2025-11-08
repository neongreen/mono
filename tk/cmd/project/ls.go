package project

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/spf13/cobra"
)

var LsCmd = &cobra.Command{
	Use:   "ls",
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
			SELECT project_uid, type, name, description, created_by, created_at
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
			UID         string   `json:"uid"`
			Name        string   `json:"name"`
			Type        string   `json:"type"`
			Aliases     []string `json:"aliases"`
			Description string   `json:"description"`
			CreatedBy   string   `json:"created_by"`
			CreatedAt   int64    `json:"created_at"`
		}

		var projects []ProjectOutput

		for rows.Next() {
			var projectUID, typ, name, description, createdBy string
			var createdAt int64

			if err := rows.Scan(&projectUID, &typ, &name, &description, &createdBy, &createdAt); err != nil {
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

			projects = append(projects, ProjectOutput{
				UID:         projectUID,
				Name:        name,
				Type:        typ,
				Aliases:     aliases,
				Description: description,
				CreatedBy:   createdBy,
				CreatedAt:   createdAt,
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
				t.AppendRow(table.Row{project.UID, project.Name, project.Type, aliasStr, project.Description, project.CreatedBy})
			}

			t.Render()
		}

		return nil
	},
}

func init() {
	LsCmd.Flags().Bool("json", false, "Output as JSON")
}

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
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

		// Load config for reducer
		config, err := config_pkg.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get reducer with all project state
		reducer, err := db.GetCachedReducerWithConfig(config)
		if err != nil {
			return fmt.Errorf("failed to build reducer: %w", err)
		}

		type ProjectOutput struct {
			UID         string `json:"uid"`
			Name        string `json:"name"`
			Type        string `json:"type"`
			Description string `json:"description"`
			CreatedBy   string `json:"created_by"`
			CreatedAt   int64  `json:"created_at"`
			IsSynthetic bool   `json:"is_synthetic"`
		}

		var projects []ProjectOutput

		// Get all projects from reducer
		for _, project := range reducer.GetAllProjects() {
			projects = append(projects, ProjectOutput{
				UID:         project.ProjectUID,
				Name:        project.Name,
				Type:        project.Type,
				Description: project.Description,
				CreatedBy:   project.CreatedBy,
				CreatedAt:   project.CreatedAt.Unix(),
				IsSynthetic: project.IsSynthetic,
			})
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
			t.AppendHeader(table.Row{"UID", "Name", "Type", "Description", "Created By"})

			t.SetStyle(table.StyleLight)
			t.Style().Options.SeparateRows = true
			t.Style().Options.DrawBorder = false

			for _, project := range projects {
				// Add [synthetic] marker to name if it's a synthetic project
				displayName := project.Name
				if project.IsSynthetic {
					displayName = project.Name + " [synthetic]"
				}

				t.AppendRow(table.Row{project.UID, displayName, project.Type, project.Description, project.CreatedBy})
			}

			t.Render()
		}

		return nil
	},
}

func init() {
	projectLsCmd.Flags().Bool("json", false, "Output as JSON")
}

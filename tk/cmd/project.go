package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long:  `Create and manage projects.`,
}

var projectCreateCmd = &cobra.Command{
	Use:   "create <name> [description]",
	Short: "Create a new project",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Check database version - projects are v4+ only
		version, err := db.GetDBVersion()
		if err != nil {
			return err
		}
		if version < 4 {
			return fmt.Errorf("projects require database v4 or higher, but current version is v%d", version)
		}

		name := args[0]
		description := ""
		if len(args) > 1 {
			description = args[1]
		}

		// Get current user and node
		actor, err := getCurrentUser()
		if err != nil {
			return err
		}

		nodeID, err := db.GetOrCreateNodeID()
		if err != nil {
			return err
		}

		// Generate new project UID
		projectUID := types.NewProjectUID()

		// Create project.created event
		payload := types.ProjectCreatedPayload{
			ProjectUID:  string(projectUID),
			Type:        "local",
			Name:        name,
			Description: description,
			CreatedBy:   actor,
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		event := types.Event{
			ID:        generateEventID(db),
			TS:        getNextLamportTimestamp(db),
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      "human",
			Kind:      string(types.EventKindProjectCreated),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event into projects table
		if err := db.ProjectProjectCreatedEvent(event); err != nil {
			return fmt.Errorf("failed to project event: %w", err)
		}

		fmt.Printf("Created project %s: %s\n", projectUID, name)

		// Optionally create an alias
		alias, _ := cmd.Flags().GetString("alias")
		if alias != "" {
			// Create project.alias.add event
			aliasPayload := types.ProjectAliasAddPayload{
				ProjectUID: string(projectUID),
				Alias:      alias,
				Node:       nodeID,
				AddedBy:    actor,
			}

			aliasPayloadJSON, err := json.Marshal(aliasPayload)
			if err != nil {
				return fmt.Errorf("failed to marshal alias payload: %w", err)
			}

			aliasEvent := types.Event{
				ID:        generateEventID(db),
				TS:        getNextLamportTimestamp(db),
				CreatedAt: time.Now(),
				Actor:     actor,
				Role:      "human",
				Kind:      string(types.EventKindProjectAliasAdd),
				Payload:   aliasPayloadJSON,
			}

			if err := db.InsertEvent(aliasEvent); err != nil {
				return fmt.Errorf("failed to insert alias event: %w", err)
			}

			// Project the alias
			if err := db.ProjectProjectAliasAddEvent(aliasEvent); err != nil {
				return fmt.Errorf("failed to project alias: %w", err)
			}

			fmt.Printf("Added alias '%s' for project %s\n", alias, projectUID)
		}

		return nil
	},
}

var projectLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		db, err := OpenExistingDB()
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

var projectAliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage project aliases",
	Long:  `Add or remove project aliases.`,
}

var projectAliasAddCmd = &cobra.Command{
	Use:   "add <project-ref> <alias>",
	Short: "Add an alias for a project (project-ref can be a UID, alias, or name)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Resolve the project reference to a ProjectUID
		projectRef := types.NewProjectRef(args[0])
		projectUID, err := database.ResolveProjectRef(db, projectRef)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}

		alias := args[1]

		// Get current user and node
		actor, err := getCurrentUser()
		if err != nil {
			return err
		}

		nodeID, err := db.GetOrCreateNodeID()
		if err != nil {
			return err
		}

		// Create project.alias.add event
		payload := types.ProjectAliasAddPayload{
			ProjectUID: projectUID.String(),
			Alias:      alias,
			Node:       nodeID,
			AddedBy:    actor,
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		event := types.Event{
			ID:        generateEventID(db),
			TS:        getNextLamportTimestamp(db),
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      "human",
			Kind:      string(types.EventKindProjectAliasAdd),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the alias
		if err := db.ProjectProjectAliasAddEvent(event); err != nil {
			return fmt.Errorf("failed to project alias: %w", err)
		}

		fmt.Printf("Added alias '%s' for project %s\n", alias, projectUID)
		return nil
	},
}

var projectAliasRemoveCmd = &cobra.Command{
	Use:   "remove <alias>",
	Short: "Remove an alias for a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		alias := args[0]

		// Get current node
		nodeID, err := db.GetOrCreateNodeID()
		if err != nil {
			return err
		}

		// Find the project UID for this alias
		var projectUID string
		err = db.Db.QueryRow(`
			SELECT project_uid FROM project_aliases 
			WHERE alias = ? AND node = ?
		`, alias, nodeID).Scan(&projectUID)
		if err != nil {
			return fmt.Errorf("alias not found: %w", err)
		}

		// Get current user
		actor, err := getCurrentUser()
		if err != nil {
			return err
		}

		// Create project.alias.remove event
		payload := types.ProjectAliasRemovePayload{
			ProjectUID: projectUID,
			Alias:      alias,
			Node:       nodeID,
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		event := types.Event{
			ID:        generateEventID(db),
			TS:        getNextLamportTimestamp(db),
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      "human",
			Kind:      string(types.EventKindProjectAliasRemove),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event (removes the alias from the table)
		if err := db.ProjectProjectAliasRemoveEvent(event); err != nil {
			return fmt.Errorf("failed to project alias removal: %w", err)
		}

		fmt.Printf("Removed alias '%s' for project %s\n", alias, projectUID)
		return nil
	},
}

var projectRmCmd = &cobra.Command{
	Use:   "rm <project-ref>",
	Short: "Delete a project (project-ref can be a UID, alias, or name)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := OpenExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Resolve the project reference to a ProjectUID
		ref := types.NewProjectRef(args[0])
		projectUID, err := database.ResolveProjectRef(db, ref)
		if err != nil {
			return fmt.Errorf("failed to resolve project: %w", err)
		}

		// Get current user
		actor, err := getCurrentUser()
		if err != nil {
			return err
		}

		// Create project.delete event
		payload := types.ProjectDeletePayload{
			ProjectUID: projectUID.String(),
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		event := types.Event{
			ID:        generateEventID(db),
			TS:        getNextLamportTimestamp(db),
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      "human",
			Kind:      string(types.EventKindProjectDelete),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event (removes the project from the table)
		if err := db.ProjectProjectDeleteEvent(event); err != nil {
			return fmt.Errorf("failed to project project deletion: %w", err)
		}

		fmt.Printf("Deleted project %s\n", projectUID)
		return nil
	},
}

// Helper functions

func generateEventID(db *database.DB) string {
	// Use the ULID-based event ID
	return string(types.NewEventID())
}

func getNextLamportTimestamp(db *database.DB) int64 {
	// Get the current max timestamp and increment
	var maxTS int64
	err := db.Db.QueryRow("SELECT COALESCE(MAX(ts), 0) FROM events").Scan(&maxTS)
	if err != nil {
		return 1
	}
	return maxTS + 1
}

func init() {
	projectCmd.AddCommand(projectCreateCmd)

	projectLsCmd.Flags().Bool("json", false, "Output as JSON")
	projectCmd.AddCommand(projectLsCmd)

	projectCmd.AddCommand(projectRmCmd)

	projectCmd.AddCommand(projectAliasCmd)

	projectAliasCmd.AddCommand(projectAliasAddCmd)
	projectAliasCmd.AddCommand(projectAliasRemoveCmd)

	// Flags for project create
	projectCreateCmd.Flags().String("alias", "", "Create an alias for this project on the current node")
}

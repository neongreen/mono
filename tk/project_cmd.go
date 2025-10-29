package main

import (
	"encoding/json"
	"fmt"
	"time"

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
		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

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
		projectUID := NewProjectUID()

		// Create project.created event
		payload := ProjectCreatedPayload{
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

		event := Event{
			ID:        generateEventID(db),
			TS:        getNextLamportTimestamp(db),
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      "human",
			Kind:      string(EventKindProjectCreated),
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
			aliasPayload := ProjectAliasAddPayload{
				ProjectUID: string(projectUID),
				Alias:      alias,
				Node:       nodeID,
				AddedBy:    actor,
			}

			aliasPayloadJSON, err := json.Marshal(aliasPayload)
			if err != nil {
				return fmt.Errorf("failed to marshal alias payload: %w", err)
			}

			aliasEvent := Event{
				ID:        generateEventID(db),
				TS:        getNextLamportTimestamp(db),
				CreatedAt: time.Now(),
				Actor:     actor,
				Role:      "human",
				Kind:      string(EventKindProjectAliasAdd),
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

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		// Query projects
		rows, err := db.db.Query(`
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

		t := table.NewWriter()
		t.AppendHeader(table.Row{"UID", "Name", "Type", "Aliases", "Description", "Created By"})

		for rows.Next() {
			var projectUID, typ, name, description, createdBy string
			var createdAt int64

			if err := rows.Scan(&projectUID, &typ, &name, &description, &createdBy, &createdAt); err != nil {
				return err
			}

			// Get aliases for this project
			aliasRows, err := db.db.Query(`
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

			aliasStr := ""
			if len(aliases) > 0 {
				aliasStr = fmt.Sprintf("%v", aliases)
			}

			t.AppendRow(table.Row{projectUID, name, typ, aliasStr, description, createdBy})
		}

		fmt.Println(t.Render())
		return rows.Err()
	},
}

var projectAliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage project aliases",
	Long:  `Add or remove project aliases.`,
}

var projectAliasAddCmd = &cobra.Command{
	Use:   "add <project-uid> <alias>",
	Short: "Add an alias for a project",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openExistingDB()
		if err != nil {
			return err
		}
		defer db.Close()

		projectUID := args[0]
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
		payload := ProjectAliasAddPayload{
			ProjectUID: projectUID,
			Alias:      alias,
			Node:       nodeID,
			AddedBy:    actor,
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		event := Event{
			ID:        generateEventID(db),
			TS:        getNextLamportTimestamp(db),
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      "human",
			Kind:      string(EventKindProjectAliasAdd),
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
		db, err := openExistingDB()
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
		err = db.db.QueryRow(`
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
		payload := ProjectAliasRemovePayload{
			ProjectUID: projectUID,
			Alias:      alias,
			Node:       nodeID,
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		event := Event{
			ID:        generateEventID(db),
			TS:        getNextLamportTimestamp(db),
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      "human",
			Kind:      string(EventKindProjectAliasRemove),
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

// Helper functions

func generateEventID(db *DB) string {
	// Use the ULID-based event ID
	return string(NewEventID())
}

func getNextLamportTimestamp(db *DB) int64 {
	// Get the current max timestamp and increment
	var maxTS int64
	err := db.db.QueryRow("SELECT COALESCE(MAX(ts), 0) FROM events").Scan(&maxTS)
	if err != nil {
		return 1
	}
	return maxTS + 1
}

func init() {
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectAliasCmd)

	projectAliasCmd.AddCommand(projectAliasAddCmd)
	projectAliasCmd.AddCommand(projectAliasRemoveCmd)

	// Flags for project create
	projectCreateCmd.Flags().String("alias", "", "Create an alias for this project on the current node")
}

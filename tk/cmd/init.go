package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
	"github.com/spf13/cobra"
)

// cobralint:exemptjson reason: Modifies state; JSON only required for read-only commands
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new tk database",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := database.GetDBPath()
		if err != nil {
			return err
		}

		if database.DBExists(path) {
			return fmt.Errorf("database already exists at %s", path)
		}

		db, err := database.OpenDB(path)
		if err != nil {
			return err
		}
		defer db.Close()

		if err := db.InitDB(); err != nil {
			return err
		}

		// Set initial DB version to 8 (latest)
		if err := db.SetDBVersion(8); err != nil {
			return fmt.Errorf("failed to set DB version: %w", err)
		}

		// Run migrations if needed (should be none for new DBs)
		if err := db.RunMigrationsIfNeeded(); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		fmt.Printf("Database initialized at %s\n", path)

		// Create default "me" project (used as default when -p flag is not specified)
		fmt.Printf("Creating default project \"me\"...\n")
		actor, err := utils.GetCurrentUser()
		if err != nil {
			return fmt.Errorf("failed to get current user: %w", err)
		}

		projectUID := types.NewProjectUID()
		payload := types.ProjectCreatedPayload{
			ProjectUID:  projectUID,
			Type:        types.ProjectTypeLocal,
			Name:        "me",
			Description: "Personal tasks",
			CreatedBy:   actor,
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		eventID, err := database.GenerateEventID(db)
		if err != nil {
			return fmt.Errorf("failed to generate event ID: %w", err)
		}

		ts, err := db.GetNextLamportTS()
		if err != nil {
			return fmt.Errorf("failed to get next lamport timestamp: %w", err)
		}

		event := types.Event{
			ID:        eventID,
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      "human",
			Kind:      string(types.EventKindProjectCreated),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}

		if err := db.RebuildProjections(); err != nil {
			return fmt.Errorf("failed to project event: %w", err)
		}

		fmt.Printf("✓ Created default project \"me\" (Personal tasks)\n")
		fmt.Printf("  This project will be used by default when you don't specify -p.\n")
		fmt.Printf("\nYou can now create tasks with:\n")
		fmt.Printf("  tk new \"Your task title\"\n")
		return nil
	},
}

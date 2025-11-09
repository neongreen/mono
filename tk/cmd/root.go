package cmd

import (
	"log/slog"
	"os"

	"github.com/golang-cz/devslog"
	"github.com/spf13/cobra"
)

var debugFlag bool

var RootCmd = &cobra.Command{
	Use:   "tk",
	Short: "tk - system-wide event-sourced task tracker",
	Long:  `tk is a command-line tool that tracks tasks system-wide using an append-only event log in a single SQLite database.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debugFlag {
			slog.SetDefault(slog.New(
				devslog.NewHandler(os.Stderr, &devslog.Options{
					HandlerOptions:  &slog.HandlerOptions{Level: slog.LevelDebug},
					NewLineAfterLog: true,
				}),
			))
			slog.Debug("debug logging enabled")
		}
	},
}

// TODO: move these abandoned comments!

// Always use project-based path

// When --unset is true, expect exactly 1 arg (task-id)

// Otherwise, expect exactly 2 args (task-id and state)

// Get next Lamport timestamp from DB

// Get next Lamport timestamp from DB

// Load config for relation processing

// Build reducer to get task and all its IDs (current + aliases)
// Use cached reducer for performance

// Always output JSON for now (backward compatibility)
// TODO: Add human-readable format if needed

// Respect color environment variables

// Load config for relation processing

// Use cached reducer for performance

// Filter by project if specified

// Filter tasks by project

// Filter by axis if specified

// Filter by blocked status if specified

// Sort tasks based on the --sort flag

// JSON output mode

// Get terminal width for wrapping

// default width if terminal size cannot be determined

// Group and render tasks based on groupBy flag

// Group tasks by project

// To maintain consistent order

// First, get all projects to ensure we include empty ones

// Initialize all projects in the grouped map

// Now add tasks to their respective groups

// Group by project alias

// Fallback to UID

// If this is a new group (shouldn't happen if we got all projects), add it

// Sort projects alphabetically

// Render a table for each project group

// Add blank line between tables

// Group tasks by status

// Render a table for each status group

// Add blank line between tables

// No grouping - render single table

func init() {
	RootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging")

	RootCmd.AddCommand(initCmd)

	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "Database commands",
		Long: `Database management commands.

Default database location: ~/.tk/tk.db

You can override the database location using the TK_DB_PATH environment variable:
  export TK_DB_PATH=/custom/path/tk.db
  tk ls  # Uses custom database

This is useful for:
- Testing with isolated databases
- Running multiple tk instances
- Custom database locations`,
	}
	dbPathCmd.Flags().Bool("json", false, "Output as JSON")
	dbCmd.AddCommand(dbPathCmd)
	RootCmd.AddCommand(dbCmd)

	newCmd.Flags().StringP("project", "p", "tk", "Project alias or UID to use")
	newCmd.Flags().String("parent", "", "Parent task (creates a subtask relation)")
	RootCmd.AddCommand(newCmd)

	RootCmd.AddCommand(markCmd)
	RootCmd.AddCommand(statusCmd)

	RootCmd.AddCommand(noteCmd)

	showCmd.Flags().Bool("json", false, "Output as JSON")
	RootCmd.AddCommand(showCmd)

	lsCmd.Flags().String("axis", "", "Filter by axis:state")
	lsCmd.Flags().String("sort", "created", "Sort order: created, id, or title (default: created)")
	lsCmd.Flags().StringSliceP("project", "p", []string{}, "Filter by project (alias, UID, or name; can be specified multiple times)")
	lsCmd.Flags().Bool("aliases", false, "Show task aliases")
	lsCmd.Flags().String("group", "project", "Group tasks by: project, status, or none (default: project)")
	lsCmd.Flags().Bool("blocked", false, "Show only blocked tasks")
	lsCmd.Flags().Bool("unblocked", false, "Show only unblocked tasks")
	lsCmd.Flags().Bool("json", false, "Output tasks as JSON")
	RootCmd.AddCommand(lsCmd)

	RootCmd.AddCommand(editCmd)
	RootCmd.AddCommand(describeCmd)
	RootCmd.AddCommand(rmCmd)
	RootCmd.AddCommand(mvCmd)
	RootCmd.AddCommand(relateCmd)
	RootCmd.AddCommand(dupCmd)
	RootCmd.AddCommand(blockersCmd)
	RootCmd.AddCommand(blockedCmd)
	RootCmd.AddCommand(graphCmd)
	RootCmd.AddCommand(conflictsCmd)
	RootCmd.AddCommand(remoteCmd)
	RootCmd.AddCommand(ingestCmd)
	RootCmd.AddCommand(importBeadsCmd)
	RootCmd.AddCommand(pushCmd)
	RootCmd.AddCommand(pullCmd)
	RootCmd.AddCommand(syncCmd)
	RootCmd.AddCommand(debugCmd)
	RootCmd.AddCommand(projectCmd)
	RootCmd.AddCommand(metaCmd)
	RootCmd.AddCommand(migrateCmd)
}

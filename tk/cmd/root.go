package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tk",
	Short: "tk - system-wide event-sourced task tracker",
	Long:  `tk is a command-line tool that tracks tasks system-wide using an append-only event log in a single SQLite database.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

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
	rootCmd.AddCommand(initCmd)

	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "Database commands",
	}
	dbPathCmd.Flags().Bool("json", false, "Output as JSON")
	dbCmd.AddCommand(dbPathCmd)
	rootCmd.AddCommand(dbCmd)

	newCmd.Flags().StringP("project", "p", "tk", "Project alias or UID to use")
	rootCmd.AddCommand(newCmd)

	markCmd.Flags().String("axis", "generic", "Status axis")
	markCmd.Flags().String("role", "human", "Actor role")
	markCmd.Flags().Bool("unset", false, "Unset the status (clear it)")
	rootCmd.AddCommand(markCmd)
	statusCmd.AddCommand(statusSyncCmd)
	rootCmd.AddCommand(statusCmd)

	rootCmd.AddCommand(noteCmd)

	showCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(showCmd)

	lsCmd.Flags().String("axis", "", "Filter by axis:state")
	lsCmd.Flags().String("sort", "created", "Sort order: created, id, or title (default: created)")
	lsCmd.Flags().StringSliceP("project", "p", []string{}, "Filter by project (alias, UID, or name; can be specified multiple times)")
	lsCmd.Flags().Bool("aliases", false, "Show task aliases")
	lsCmd.Flags().String("group", "project", "Group tasks by: project, status, or none (default: project)")
	lsCmd.Flags().Bool("blocked", false, "Show only blocked tasks")
	lsCmd.Flags().Bool("unblocked", false, "Show only unblocked tasks")
	lsCmd.Flags().Bool("json", false, "Output tasks as JSON")
	rootCmd.AddCommand(lsCmd)

	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(describeCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(mvCmd)
	rootCmd.AddCommand(relateCmd)
	rootCmd.AddCommand(dupCmd)
	rootCmd.AddCommand(blockersCmd)
	rootCmd.AddCommand(blockedCmd)
	rootCmd.AddCommand(graphCmd)
	rootCmd.AddCommand(conflictsCmd)
	rootCmd.AddCommand(remoteCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(ingestCmd)
	rootCmd.AddCommand(importBeadsCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(debugCmd)
	rootCmd.AddCommand(projectCmd)
}

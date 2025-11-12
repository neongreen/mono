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

	newCmd.Flags().StringP("project", "p", "me", "Project alias or UID to use")
	newCmd.Flags().String("parent", "", "Parent task (creates a subtask relation)")
	RootCmd.AddCommand(newCmd)

	RootCmd.AddCommand(markCmd)
	RootCmd.AddCommand(statusCmd)

	RootCmd.AddCommand(noteCmd)

	showCmd.Flags().Bool("json", false, "Output as JSON")
	RootCmd.AddCommand(showCmd)

	lsCmd.Flags().String("status", "", "Filter by status. Supports multiple values with comma: --status wip,next (next, wip, done, closed)")
	lsCmd.Flags().String("axis", "", "Filter by axis:state")
	lsCmd.Flags().String("sort", "created", "Sort order: created, id, or title (default: created)")
	lsCmd.Flags().StringSliceP("project", "p", []string{}, "Filter by project (alias, UID, or name; can be specified multiple times)")
	lsCmd.Flags().Bool("aliases", false, "Show task aliases")
	lsCmd.Flags().String("group", "project", "Group tasks by: project, status, or none (default: project)")
	lsCmd.Flags().Bool("blocked", false, "Show only blocked tasks")
	lsCmd.Flags().Bool("unblocked", false, "Show only unblocked tasks")
	lsCmd.Flags().Bool("json", false, "Output tasks as JSON")
	// Hide --axis flag from help but keep it functional
	lsCmd.Flags().MarkHidden("axis")
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
	RootCmd.AddCommand(logCmd)
}

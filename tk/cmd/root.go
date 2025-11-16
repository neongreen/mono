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
	// Allow arbitrary args so paths like /foo work
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If an argument starts with /, treat it as a path query
		if len(args) > 0 && len(args[0]) > 0 && args[0][0] == '/' {
			// Delegate to query command
			return queryCmd.RunE(cmd, args)
		}
		// Otherwise show help
		return cmd.Help()
	},
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging")
	// Add json flag for path queries
	RootCmd.Flags().Bool("json", false, "Output as JSON (for path queries)")

	//
	// Core task management commands
	//
	RootCmd.AddCommand(initCmd)

	newCmd.Flags().StringP("project", "p", "me", "Project alias or UID to use")
	newCmd.Flags().String("parent", "", "Parent task (creates a subtask relation)")
	newCmd.Flags().String("kind", "task", "Item kind (task, decision, resource, etc.)")
	RootCmd.AddCommand(newCmd)

	lsCmd.Flags().String("status", "", "Filter by status. Supports multiple values with comma: --status wip,next (next, wip, done, closed)")
	lsCmd.Flags().String("axis", "", "Filter by axis:state")
	lsCmd.Flags().String("kind", "", "Filter by item kind. Supports multiple values with comma: --kind decision,resource")
	lsCmd.Flags().String("query", "", "Taskset query expression (e.g., 'status(wip) & kind(decision)')")
	lsCmd.Flags().String("sort", "created", "Sort order: created, updated, id, title, or status. Add -desc for descending (e.g., updated-desc for most recently updated)")
	lsCmd.Flags().StringSliceP("project", "p", []string{}, "Filter by project (alias, UID, or name; can be specified multiple times)")
	lsCmd.Flags().Bool("aliases", false, "Show task aliases")
	lsCmd.Flags().String("group", "project", "Group tasks by: project, status, or none (default: project, or none when --sort is used)")
	lsCmd.Flags().Bool("blocked", false, "Show only blocked tasks")
	lsCmd.Flags().Bool("unblocked", false, "Show only unblocked tasks")
	lsCmd.Flags().Bool("json", false, "Output tasks as JSON")
	lsCmd.Flags().String("grep", "", "Filter by regex pattern (RE2 syntax; searches title and notes)")
	lsCmd.Flags().Int("limit", 0, "Limit number of tasks displayed (0 = no limit)")
	lsCmd.Flags().String("in", "", "Filter by container (show only tasks in this container)")
	lsCmd.Flags().MarkHidden("axis") // Hide --axis flag from help but keep it functional
	RootCmd.AddCommand(lsCmd)

	showCmd.Flags().Bool("json", false, "Output as JSON")
	RootCmd.AddCommand(showCmd)

	queryCmd.Flags().Bool("json", false, "Output as JSON")
	RootCmd.AddCommand(queryCmd)

	RootCmd.AddCommand(markCmd)
	RootCmd.AddCommand(editCmd)
	RootCmd.AddCommand(describeCmd)
	RootCmd.AddCommand(noteCmd)

	attachCmd.Flags().Bool("list", false, "List attachments for a task")
	attachCmd.Flags().String("get", "", "Get an attachment by ID")
	attachCmd.Flags().String("open", "", "Open an attachment by ID")
	attachCmd.Flags().StringP("description", "d", "", "Description for the attachment")
	RootCmd.AddCommand(attachCmd)

	RootCmd.AddCommand(rmCmd)
	RootCmd.AddCommand(mvCmd)
	RootCmd.AddCommand(historyCmd)

	//
	// Relations & dependencies
	//
	RootCmd.AddCommand(relateAddCmd)
	RootCmd.AddCommand(relateLsCmd)
	RootCmd.AddCommand(relateRmCmd)
	RootCmd.AddCommand(relateDupCmd)
	RootCmd.AddCommand(dupCmd) // Alias
	RootCmd.AddCommand(relateBlockersCmd)
	RootCmd.AddCommand(blockersCmd) // Alias
	RootCmd.AddCommand(relateBlockedCmd)
	RootCmd.AddCommand(blockedCmd) // Alias
	RootCmd.AddCommand(relateGraphCmd)
	RootCmd.AddCommand(graphCmd) // Alias
	RootCmd.AddCommand(relateConflictsCmd)
	RootCmd.AddCommand(conflictsCmd) // Alias
	RootCmd.AddCommand(taskConflictsCmd)

	//
	// Projects
	//
	RootCmd.AddCommand(projectCreateCmd)
	RootCmd.AddCommand(projectLsCmd)
	RootCmd.AddCommand(projectRenameCmd)
	RootCmd.AddCommand(projectRmCmd)

	//
	// Sync & remote
	//
	RootCmd.AddCommand(syncCmd)
	RootCmd.AddCommand(pushCmd)
	RootCmd.AddCommand(pullCmd)
	RootCmd.AddCommand(ingestCmd)
	RootCmd.AddCommand(importBeadsCmd)
	RootCmd.AddCommand(syncStatusCmd)

	RootCmd.AddCommand(remoteAddCmd)
	RootCmd.AddCommand(remoteLsCmd)
	RootCmd.AddCommand(remoteRmCmd)

	//
	// Containers - Queues
	//
	RootCmd.AddCommand(queueCreateCmd)
	RootCmd.AddCommand(queuePushCmd)
	RootCmd.AddCommand(queuePopCmd)
	RootCmd.AddCommand(queueLsCmd)
	RootCmd.AddCommand(queueShowCmd)
	RootCmd.AddCommand(queueRenameCmd)
	RootCmd.AddCommand(queueRmCmd)

	//
	// Containers - Stacks
	//
	RootCmd.AddCommand(stackCreateCmd)
	RootCmd.AddCommand(stackPushCmd)
	RootCmd.AddCommand(stackPopCmd)
	RootCmd.AddCommand(stackLsCmd)
	RootCmd.AddCommand(stackShowCmd)
	RootCmd.AddCommand(stackRenameCmd)
	RootCmd.AddCommand(stackRmCmd)

	//
	// Containers - Groups
	//
	RootCmd.AddCommand(groupCreateCmd)
	RootCmd.AddCommand(groupAddtaskCmd)
	RootCmd.AddCommand(groupRmtaskCmd)
	RootCmd.AddCommand(groupLsCmd)
	RootCmd.AddCommand(groupShowCmd)
	RootCmd.AddCommand(groupRenameCmd)
	RootCmd.AddCommand(groupDeleteCmd)

	//
	// Schema & metadata
	//
	RootCmd.AddCommand(schemaAddCmd)
	RootCmd.AddCommand(schemaLsCmd)
	RootCmd.AddCommand(schemaExportCmd)

	RootCmd.AddCommand(metaSetCmd)
	RootCmd.AddCommand(metaGetCmd)
	RootCmd.AddCommand(metaLsCmd)
	RootCmd.AddCommand(metaClaimsCmd)

	//
	// Debug
	//
	RootCmd.AddCommand(debugDoctorCmd)
	RootCmd.AddCommand(debugRepairCmd)
	RootCmd.AddCommand(debugRebuildCmd)
	RootCmd.AddCommand(debugEventsLsCmd)
	RootCmd.AddCommand(debugEventsShowCmd)
	RootCmd.AddCommand(debugEventsStatsCmd)
	RootCmd.AddCommand(debugNodeShowCmd)
	RootCmd.AddCommand(debugNodeRegenCmd)
	RootCmd.AddCommand(debugIdCmd)
	RootCmd.AddCommand(idCmd) // Alias

	//
	// Migration & logs
	//
	RootCmd.AddCommand(migrateFixContainerItemIdsCmd)
	RootCmd.AddCommand(migrateFixRelocateBugCmd)
	RootCmd.AddCommand(migrateScanDeprecatedCmd)

	RootCmd.AddCommand(logQueryCmd)
	RootCmd.AddCommand(logSearchCmd)

	//
	// Database & system
	//
	dbPathCmd.Flags().Bool("json", false, "Output as JSON")
	RootCmd.AddCommand(dbPathCmd)

	RootCmd.AddCommand(statuslineCmd)
	RootCmd.AddCommand(mcpCmd)

	// Apply "See Also" sections to all commands
	// This adds cross-references to help improve command discoverability
	ApplySeeAlso(RootCmd)
}

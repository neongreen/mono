package cmd

import (
	"github.com/spf13/cobra"
)

var dryRun bool

var RootCmd = &cobra.Command{
	Use:   "conf",
	Short: "Smart configuration manager with autocompletion",
	Long: `conf is a smart config manager that provides intelligent configuration 
management with autocomplete for tools like jj (Jujutsu) and mise. It understands 
tool schemas and provides surgical TOML editing while preserving formatting.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute runs the root command
//
//nolint:uselesswrapper // Provides stable public API
func Execute() error {
	return RootCmd.Execute()
}

func init() {
	// Add dry-run flag to root command
	RootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making any modifications")

	// Register all commands
	RootCmd.AddCommand(jjCmd)
	RootCmd.AddCommand(claudeCmd)
	RootCmd.AddCommand(miseCmd)
	RootCmd.AddCommand(starshipCmd)
	RootCmd.AddCommand(shimsCmd)
	RootCmd.AddCommand(applyCmd)
	// statusCmd is registered in status.go init()
	RootCmd.AddCommand(syncCmd)
	RootCmd.AddCommand(importCmd)
	RootCmd.AddCommand(completionCmd)
	RootCmd.AddCommand(versionCmd)
}

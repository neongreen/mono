package cmd

import (
	"github.com/spf13/cobra"
)

var dryRun bool

var rootCmd = &cobra.Command{
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
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add dry-run flag to root command
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making any modifications")

	// Register all commands
	rootCmd.AddCommand(jjCmd)
	rootCmd.AddCommand(claudeCmd)
	rootCmd.AddCommand(miseCmd)
	rootCmd.AddCommand(starshipCmd)
	rootCmd.AddCommand(shimsCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(versionCmd)
}

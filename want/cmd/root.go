package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/golang-cz/devslog"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	dryRun    bool
	planJSON  bool
	debugFlag bool
)

var RootCmd = &cobra.Command{
	Use:   "want",
	Short: "want - Interactive task fulfillment tool for macOS",
	Long: `want helps you get things you need on your system through CLI commands.
It's an interactive assistant that respects your preferences.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	// Accept any arguments
	Args: cobra.ArbitraryArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if debugFlag {
			slog.SetDefault(slog.New(
				devslog.NewHandler(os.Stderr, &devslog.Options{
					HandlerOptions:  &slog.HandlerOptions{Level: slog.LevelDebug},
					NewLineAfterLog: true,
				}),
			))
			slog.Debug("debug logging enabled")
		}
		return ensureConfigDirectory()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		// Default handler for requirements
		return handleRequirement(args, dryRun, planJSON)
	},
}

// Execute runs the root command
//nolint:uselesswrapper // Provides stable public API
func Execute() error {
	return RootCmd.Execute()
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without doing it")
	RootCmd.PersistentFlags().BoolVar(&planJSON, "plan-json", false, "Output the fulfillment plan as JSON")
	RootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug logging")
}

func ensureConfigDirectory() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "want")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create want config directory %s: %w", configDir, err)
	}

	return nil
}

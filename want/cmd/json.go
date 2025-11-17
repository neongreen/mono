package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var jsonCmd = &cobra.Command{
	Use:   "json <command>",
	Short: "Convert command output to JSON",
	Long: `Convert command output to JSON using jc.

This command uses jc (JSON CLI) to convert the output of common CLI tools
to JSON format, making it easier to parse and process programmatically.

Examples:
  want json ps                    # Get running processes as JSON
  want json df                    # Get disk usage as JSON
  want json netstat               # Get network statistics as JSON`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleJsonCommand(args, dryRun, planJSON)
	},
}

func init() {
	RootCmd.AddCommand(jsonCmd)
}

// Function reference that will be set by main package
var HandleJsonCommandFunc func([]string, bool, bool) error

func handleJsonCommand(args []string, dryRun, planJSON bool) error {
	if HandleJsonCommandFunc != nil {
		return HandleJsonCommandFunc(args, dryRun, planJSON)
	}
	return fmt.Errorf("json command handler not initialized")
}

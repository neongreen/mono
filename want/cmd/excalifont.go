package cmd

import (
	"github.com/spf13/cobra"
)

var excalifontCmd = &cobra.Command{
	Use:   "excalifont",
	Short: "Download and install Excalifont",
	Long: `Download and install the Excalifont font.

This command downloads the Excalifont Regular font from excalidraw.com,
converts it from woff2 to ttf format, and helps you install it on your system.

On macOS, it will automatically open Font Book for installation.
On other systems, it will save the font file for manual installation.

Example:
  want excalifont                  # Download and install Excalifont`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleExcalifontCommand(args, dryRun, planJSON)
	},
}

func init() {
	rootCmd.AddCommand(excalifontCmd)
}

// Function reference that will be set by main package
var HandleExcalifontCommandFunc func([]string, bool, bool) error

func handleExcalifontCommand(args []string, dryRun, planJSON bool) error {
	if HandleExcalifontCommandFunc != nil {
		return HandleExcalifontCommandFunc(args, dryRun, planJSON)
	}
	return nil
}

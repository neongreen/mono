package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mdCmd = &cobra.Command{
	Use:   "md <url>",
	Short: "Convert URL to markdown",
	Long: `Convert a webpage URL to markdown format.

This command uses markitdown to convert web pages to markdown,
making it easier to read and process web content.

Examples:
  want md https://example.com                # Convert webpage to markdown
  want md https://github.com/owner/repo      # Convert GitHub page to markdown`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleMarkdownCommand(args, dryRun, planJSON)
	},
}

func init() {
	RootCmd.AddCommand(mdCmd)
}

// Function reference that will be set by main package
var HandleMarkdownCommandFunc func([]string, bool, bool) error

func handleMarkdownCommand(args []string, dryRun, planJSON bool) error {
	if HandleMarkdownCommandFunc != nil {
		return HandleMarkdownCommandFunc(args, dryRun, planJSON)
	}
	return fmt.Errorf("markdown command handler not initialized")
}

package cmd

import (
	"fmt"
	"os"

	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/neongreen/mono/conf/pkg/diff"
	"github.com/neongreen/mono/conf/pkg/tools"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply [tool]",
	Short: "Apply desired state to target configuration files",
	Long: `Sync the desired state from conf config to target files.

Examples:
  conf apply          # Apply all tools
  conf apply jj       # Apply only jj config
  conf apply --dry-run jj  # Preview what would be applied`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to load conf config: %v\n", err)
			os.Exit(1)
		}

		var toolsToApply []string
		if len(args) == 1 {
			toolsToApply = []string{args[0]}
		} else {
			// Apply all tools
			for toolName := range conf.Tools {
				toolsToApply = append(toolsToApply, toolName)
			}
		}

		for _, toolName := range toolsToApply {
			if err := applyTool(conf, toolName, dryRun); err != nil {
				fmt.Fprintf(os.Stderr, "Error applying %s: %v\n", toolName, err)
				os.Exit(1)
			}
		}
	},
}

// applyTool applies desired state for a specific tool
func applyTool(conf *config.Config, toolName string, dryRun bool) error {
	tool, exists := conf.GetTool(toolName)
	if !exists {
		return fmt.Errorf("tool %s not configured", toolName)
	}

	if len(tool.Values) == 0 {
		fmt.Printf("%s: No values to apply\n", toolName)
		return nil
	}

	fmt.Printf("Applying %s configuration...\n", toolName)

	// Use the nested structure directly instead of flattening to strings
	// This avoids the need to parse quoted keys and is more efficient
	if dryRun {
		fmt.Printf("  Would apply %d top-level config sections\n", len(tool.Values))
	} else {
		// Capture the "before" state
		before, err := readFileContentSafe(tool.ConfigPath)
		if err != nil {
			return fmt.Errorf("failed to read config file before applying: %w", err)
		}

		// Apply the configuration
		if err := tools.ApplyAllToolValues(toolName, tool.Values); err != nil {
			return fmt.Errorf("failed to apply %s configuration: %w", toolName, err)
		}

		// Capture the "after" state
		after, err := readFileContentSafe(tool.ConfigPath)
		if err != nil {
			return fmt.Errorf("failed to read config file after applying: %w", err)
		}

		// Display diff if there were changes
		if diff.DisplayUnifiedDiff(before, after, tool.ConfigPath) {
			fmt.Printf("  ✓ Applied configuration successfully\n")
		} else {
			fmt.Printf("  ✓ No changes needed (already in sync)\n")
		}
	}

	return nil
}

// readFileContentSafe reads file content, returning empty string if file doesn't exist
func readFileContentSafe(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(content), nil
}

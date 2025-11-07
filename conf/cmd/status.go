package cmd

import (
	"fmt"
	"os"

	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [tool]",
	Short: "Show drift between desired and actual state",
	Long: `Compare desired state in conf config with actual values in target files.

Examples:
  conf status         # Show status for all tools
  conf status jj      # Show status for jj only`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to load conf config: %v\n", err)
			os.Exit(1)
		}

		var toolsToCheck []string
		if len(args) == 1 {
			toolsToCheck = []string{args[0]}
		} else {
			// Check all tools
			for toolName := range conf.Tools {
				toolsToCheck = append(toolsToCheck, toolName)
			}
		}

		for _, toolName := range toolsToCheck {
			if err := showToolStatus(conf, toolName); err != nil {
				fmt.Fprintf(os.Stderr, "Error checking %s status: %v\n", toolName, err)
				os.Exit(1)
			}
		}
	},
}

// showToolStatus shows drift status for a specific tool
func showToolStatus(conf *config.Config, toolName string) error {
	tool, exists := conf.GetTool(toolName)
	if !exists {
		return fmt.Errorf("tool %s not configured", toolName)
	}

	fmt.Printf("%s status:\n", toolName)

	flatValues := config.FlattenValues(tool.Values)
	if len(flatValues) == 0 {
		fmt.Printf("  No managed values\n")
		return nil
	}

	actualValues, err := getTargetConfigValues(toolName)
	if err != nil {
		return fmt.Errorf("failed to read current %s configuration: %w", toolName, err)
	}
	actualFlat := config.FlattenValues(actualValues)

	hasChanges := false
	for path, desiredValue := range flatValues {
		actualValue, exists := actualFlat[path]

		if !exists {
			fmt.Printf("  %s: MISSING (desired: %v)\n", path, desiredValue)
			hasChanges = true
		} else if fmt.Sprintf("%v", actualValue) != fmt.Sprintf("%v", desiredValue) {
			fmt.Printf("  %s: DRIFT (desired: %v, actual: %v)\n", path, desiredValue, actualValue)
			hasChanges = true
		} else {
			fmt.Printf("  %s: IN SYNC (%v)\n", path, actualValue)
		}
	}

	if !hasChanges {
		fmt.Printf("  All values in sync\n")
	}

	return nil
}

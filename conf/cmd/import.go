package cmd

import (
	"fmt"
	"os"

	"github.com/neongreen/mono/conf/pkg/config"
	claudetool "github.com/neongreen/mono/conf/pkg/tools/claude"
	jjtool "github.com/neongreen/mono/conf/pkg/tools/jj"
	misetool "github.com/neongreen/mono/conf/pkg/tools/mise"
	starshiptool "github.com/neongreen/mono/conf/pkg/tools/starship"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import [tool]",
	Short: "Import configuration values from target files into conf state",
	Long: `Read configuration values from target files (e.g., ~/.config/jj/config.toml) 
and import them into conf's state management (stored in ~/.config/conf/).

This is useful for:
- Migrating existing configurations to conf management
- Capturing manual changes made to config files
- Syncing local configurations into conf state

Examples:
  conf import           # Import all tools
  conf import jj        # Import only jj config
  conf import --dry-run # Preview what would be imported`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to load conf config: %v\n", err)
			os.Exit(1)
		}

		var toolsToImport []string
		if len(args) == 1 {
			toolsToImport = []string{args[0]}
		} else {
			// Import all tools
			for toolName := range conf.Tools {
				toolsToImport = append(toolsToImport, toolName)
			}
		}

		for _, toolName := range toolsToImport {
			if err := importTool(conf, toolName, dryRun); err != nil {
				fmt.Fprintf(os.Stderr, "Error importing %s: %v\n", toolName, err)
				os.Exit(1)
			}
		}

		if !dryRun {
			fmt.Println("\n✓ Import complete")
		}
	},
}

// importTool imports configuration values from a tool's target file into conf's state
func importTool(conf *config.Config, toolName string, dryRun bool) error {
	tool, exists := conf.GetTool(toolName)
	if !exists {
		return fmt.Errorf("tool %s not configured", toolName)
	}

	fmt.Printf("Importing %s configuration from %s...\n", toolName, tool.ConfigPath)

	// Get all values from the target config file
	values, err := getTargetConfigValues(toolName)
	if err != nil {
		return fmt.Errorf("failed to read target config: %w", err)
	}

	flatValues := config.FlattenValues(values)

	if len(flatValues) == 0 {
		fmt.Printf("  No values found in %s\n", tool.ConfigPath)
		return nil
	}

	fmt.Printf("  Found %d values\n", len(flatValues))

	if dryRun {
		// Preview what would be imported
		for path, value := range flatValues {
			fmt.Printf("  Would import: %s.%s = %v\n", toolName, path, value)
		}
	} else {
		// Import all values into conf state
		for path, value := range flatValues {
			fmt.Printf("  ✓ Imported %s.%s = %v\n", toolName, path, value)
		}

		conf.MergeToolValues(toolName, values)

		// Save conf state
		if err := conf.Save(); err != nil {
			return fmt.Errorf("failed to save conf state: %w", err)
		}

		fmt.Printf("  ✓ Saved to conf state\n")
	}

	return nil
}

// getTargetConfigValues reads all values from a tool's target config file
func getTargetConfigValues(toolName string) (map[string]any, error) {
	switch toolName {
	case "jj":
		jjTool, err := jjtool.NewJJTool()
		if err != nil {
			return nil, err
		}
		return jjTool.GetAllValues()
	case "claude":
		claudeTool, err := claudetool.NewClaudeTool()
		if err != nil {
			return nil, err
		}
		return claudeTool.GetAllValues()
	case "mise":
		miseTool, err := misetool.NewMiseTool()
		if err != nil {
			return nil, err
		}
		return miseTool.GetAllValues()
	case "starship":
		starshipTool, err := starshiptool.NewStarshipTool()
		if err != nil {
			return nil, err
		}
		return starshipTool.GetAllValues()
	default:
		return nil, fmt.Errorf("unsupported tool: %s", toolName)
	}
}

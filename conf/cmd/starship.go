package cmd

import (
	"fmt"
	"os"

	starshiptool "github.com/neongreen/mono/conf/pkg/tools/starship"
	"github.com/spf13/cobra"
)

var starshipCmd = &cobra.Command{
	Use:   "starship [config.path] [value]",
	Short: "Configure starship settings",
	Long: `Get or set configuration values in ~/.config/starship.toml using dotted path notation.

Examples:
  conf starship                          # List common settings
  conf starship add_newline              # Get current value
  conf starship add_newline true         # Set boolean value
  conf starship command_timeout 500      # Set timeout value`,
	Args: cobra.RangeArgs(0, 2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return starshipCompletion(args, toComplete)
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Create starship tool with dry-run mode
		starshipTool, err := starshiptool.NewStarshipToolWithDryRun(dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize starship tool: %v\n", err)
			os.Exit(1)
		}

		// Default to list when no arguments provided
		if len(args) == 0 {
			settings, err := starshipTool.ListAllSettings()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to list settings: %v\n", err)
				os.Exit(1)
			}

			renderSettingsTable(settings, starshipTool.GetConfigPath())
			return
		}

		configPath := args[0]

		// GET operation: only config path provided
		if len(args) == 1 {
			value, err := starshipTool.GetConfig(configPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
				os.Exit(1)
			}
			if value == nil {
				fmt.Printf("%s = (not set)\n", configPath)
			} else {
				fmt.Printf("%s = %v\n", configPath, value)
			}
			return
		}

		// SET operation: config path and value provided
		if configPath == "" {
			fmt.Fprintf(os.Stderr, "Error: invalid configuration path: %s\n", configPath)
			os.Exit(1)
		}
		value := args[1]
		parsedValue := parseValue(value)

		if dryRun {
			// Show preview of what would happen
			preview, err := starshipTool.PreviewSetConfig(configPath, parsedValue)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Print("DRY RUN: ")
			fmt.Print(preview)
		} else {
			// Set the configuration
			err = starshipTool.SetConfig(configPath, parsedValue)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✓ Set starship config: %s = %v\n", configPath, parsedValue)
		}

		fmt.Printf("Config file: %s\n", starshipTool.GetConfigPath())
	},
}

var starshipListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all starship configuration options",
	Long:  `Display a list of all starship configuration options with descriptions, types, and current values.`,
	Run: func(cmd *cobra.Command, args []string) {
		starshipTool, err := starshiptool.NewStarshipTool()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize starship tool: %v\n", err)
			os.Exit(1)
		}

		settings, err := starshipTool.ListAllSettings()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to list settings: %v\n", err)
			os.Exit(1)
		}

		renderSettingsTable(settings, starshipTool.GetConfigPath())
	},
}

func init() {
	starshipCmd.AddCommand(starshipListCmd)
}

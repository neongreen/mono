package cmd

import (
	"fmt"
	"os"

	misetool "github.com/neongreen/mono/conf/pkg/tools/mise"
	"github.com/spf13/cobra"
)

var miseCmd = &cobra.Command{
	Use:   "mise [config.path] [value]",
	Short: "Configure mise settings",
	Long: `Get or set configuration values in ~/.config/mise/config.toml using dotted path notation.

Examples:
  conf mise                               # List common settings
  conf mise settings.experimental         # Get current value
  conf mise settings.experimental true    # Set boolean value
  conf mise settings.jobs 4               # Set numeric value`,
	Args: cobra.RangeArgs(0, 2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return miseCompletion(args, toComplete)
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Create mise tool with dry-run mode
		miseTool, err := misetool.NewMiseToolWithDryRun(dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize mise tool: %v\n", err)
			os.Exit(1)
		}

		// Default to list when no arguments provided
		if len(args) == 0 {
			settings, err := miseTool.ListAllSettings()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to list settings: %v\n", err)
				os.Exit(1)
			}

			renderSettingsTable(settings, miseTool.GetConfigPath())
			return
		}

		configPath := args[0]

		// GET operation: only config path provided
		if len(args) == 1 {
			value, err := miseTool.GetConfig(configPath)
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
			preview, err := miseTool.PreviewSetConfig(configPath, parsedValue)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Print("DRY RUN: ")
			fmt.Print(preview)
		} else {
			// Set the configuration
			err = miseTool.SetConfig(configPath, parsedValue)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✓ Set mise config: %s = %v\n", configPath, parsedValue)
		}

		fmt.Printf("Config file: %s\n", miseTool.GetConfigPath())
	},
}

var miseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all mise configuration options from schema",
	Long:  `Display a list of all mise configuration options from the official schema.`,
	Run: func(cmd *cobra.Command, args []string) {
		miseTool, err := misetool.NewMiseTool()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize mise tool: %v\n", err)
			os.Exit(1)
		}

		settings, err := miseTool.ListAllSettings()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to list settings: %v\n", err)
			os.Exit(1)
		}

		renderSettingsTable(settings, miseTool.GetConfigPath())
	},
}

func init() {
	miseCmd.AddCommand(miseListCmd)
}

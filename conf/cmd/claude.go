package cmd

import (
	"fmt"
	"os"

	"github.com/neongreen/mono/conf/pkg/config"
	claudetool "github.com/neongreen/mono/conf/pkg/tools/claude"
	"github.com/spf13/cobra"
)

var claudeListFlag bool
var claudeForce bool

var claudeCmd = &cobra.Command{
	Use:   "claude [config.path] [value]",
	Short: "Configure Claude Code settings",
	Long: `Get or set configuration values in ~/.claude/settings.json using dotted path notation.

Examples:
  conf claude --list                          # List all available settings with current values
  conf claude model                           # Get current value
  conf claude model "sonnet"                  # Set value
  conf claude alwaysThinkingEnabled true      # Enable extended thinking`,
	Args: cobra.RangeArgs(0, 2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return claudeCompletion(args, toComplete)
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Create claude tool
		claudeTool, err := claudetool.NewClaudeToolWithDryRun(dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize claude tool: %v\n", err)
			os.Exit(1)
		}
		claudeTool.SetForce(claudeForce)

		// Handle --list flag or no arguments (default to list)
		if claudeListFlag || len(args) == 0 {
			settings, err := claudeTool.ListAllSettings()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to list settings: %v\n", err)
				os.Exit(1)
			}

			renderSettingsTable(settings, claudeTool.GetConfigPath())
			return
		}

		configPath := args[0]

		// Validate config path
		if configPath == "" {
			fmt.Fprintf(os.Stderr, "Error: invalid configuration path: %s\n", configPath)
			os.Exit(1)
		}

		// GET operation: only config path provided
		if len(args) == 1 {
			// Load conf's state config to show desired state
			conf, err := config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to load conf config: %v\n", err)
				os.Exit(1)
			}

			// Get actual value from target file
			actualValue, err := claudeTool.GetConfig(configPath)
			if err != nil && !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
				os.Exit(1)
			}

			// Get desired value from conf state
			desiredValue, hasDesired := conf.GetToolValue("claude", configPath)

			// Show state comparison
			if hasDesired {
				fmt.Printf("Desired: %s = %s\n", configPath, formatValueAsTOML(desiredValue))
				if actualValue == nil {
					fmt.Printf("Actual:  %s = (not set)\n", configPath)
					fmt.Printf("Status:  DRIFT - value not applied\n")
					fmt.Printf("\nTo apply the desired value:\n")
					fmt.Printf("  conf apply claude  # Applies ALL drifting claude values\n")
					fmt.Printf("  conf apply claude --dry-run  # Preview changes first\n")
				} else if fmt.Sprintf("%v", actualValue) == fmt.Sprintf("%v", desiredValue) {
					fmt.Printf("Actual:  %s = %s\n", configPath, formatValueAsTOML(actualValue))
					fmt.Printf("Status:  IN SYNC\n")
				} else {
					fmt.Printf("Actual:  %s = %s\n", configPath, formatValueAsTOML(actualValue))
					fmt.Printf("Status:  DRIFT - values differ\n")
					fmt.Printf("\nTo apply the desired value:\n")
					fmt.Printf("  conf apply claude  # Applies ALL drifting claude values\n")
					fmt.Printf("  conf apply claude --dry-run  # Preview changes first\n")
					fmt.Printf("\nTo update desired to match actual:\n")
					fmt.Printf("  conf claude %s %s\n", configPath, formatValueAsTOML(actualValue))
				}
			} else {
				if actualValue == nil {
					fmt.Printf("%s = (not set)\n", configPath)
				} else {
					fmt.Printf("Actual:  %s = %s\n", configPath, formatValueAsTOML(actualValue))
					fmt.Printf("Status:  UNMANAGED - not in conf state\n")
				}
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

		// Load conf's state config
		conf, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to load conf config: %v\n", err)
			os.Exit(1)
		}

		if dryRun {
			// Show preview of what would happen
			preview, err := claudeTool.PreviewSetConfig(configPath, parsedValue)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Print("DRY RUN: ")
			fmt.Print(preview)
			fmt.Printf("Would also record in conf state: claude.%s = %v\n", configPath, parsedValue)
		} else {
			// 1. Record desired state in conf config
			conf.SetToolValue("claude", configPath, parsedValue)
			if err := conf.Save(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to save conf state: %v\n", err)
				os.Exit(1)
			}

			// 2. Apply to target file
			err = claudeTool.SetConfig(configPath, parsedValue)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✓ Set claude config: %s = %v\n", configPath, parsedValue)
			fmt.Printf("✓ Recorded in conf state\n")
		}

		fmt.Printf("Config file: %s\n", claudeTool.GetConfigPath())
	},
}

func init() {
	claudeCmd.Flags().BoolVar(&claudeListFlag, "list", false, "List all available Claude Code configuration settings with current values")
	claudeCmd.Flags().BoolVar(&claudeForce, "force", false, "Bypass schema validation when setting values")
}

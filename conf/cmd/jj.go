package cmd

import (
	"fmt"
	"os"

	"github.com/neongreen/mono/conf/pkg/config"
	jjtool "github.com/neongreen/mono/conf/pkg/tools/jj"
	"github.com/neongreen/mono/lib/cli"
	"github.com/spf13/cobra"
)

var jjListFlag bool
var jjForce bool

var jjCmd = &cobra.Command{
	Use:   "jj [config.path] [value]",
	Short: "Configure jj (Jujutsu) settings",
	Long: `Get or set configuration values in ~/.config/jj/config.toml using dotted path notation.

Examples:
  conf jj --list                       # List all available settings with current values  
  conf jj user.name                    # Get current value
  conf jj user.name "John Doe"         # Set value
  conf jj user.email john@example.com  # Set email`,
	Args: cobra.RangeArgs(0, 2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return jjCompletion(args, toComplete)
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Create jj tool
		jjTool, err := jjtool.NewJJToolWithDryRun(dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize jj tool: %v\n", err)
			os.Exit(1)
		}
		jjTool.SetForce(jjForce)

		// Handle --list flag or no arguments (default to list)
		if jjListFlag || len(args) == 0 {
			settings, err := jjTool.ListAllSettings()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to list settings: %v\n", err)
				os.Exit(1)
			}

			renderSettingsTable(settings, jjTool.GetConfigPath())
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
			actualValue, err := jjTool.GetConfig(configPath)
			if err != nil && !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
				os.Exit(1)
			}

			// Get desired value from conf state
			desiredValue, hasDesired := conf.GetToolValue("jj", configPath)

			// Show state comparison
			if hasDesired {
				fmt.Printf("Desired: %s = %s\n", cli.Key(configPath), cli.Value(formatValueAsTOML(desiredValue)))
				if actualValue == nil {
					fmt.Printf("Actual:  %s = %s\n", cli.Key(configPath), cli.Muted("(not set)"))
					fmt.Printf("Status:  %s\n", cli.Warning("DRIFT - value not applied"))
					fmt.Printf("\nTo apply the desired value:\n")
					fmt.Printf("  conf apply jj  # Applies ALL drifting jj values\n")
					fmt.Printf("  conf apply jj --dry-run  # Preview changes first\n")
				} else if fmt.Sprintf("%v", actualValue) == fmt.Sprintf("%v", desiredValue) {
					fmt.Printf("Actual:  %s = %s\n", cli.Key(configPath), cli.Value(formatValueAsTOML(actualValue)))
					fmt.Printf("Status:  %s\n", cli.Success("IN SYNC"))
				} else {
					fmt.Printf("Actual:  %s = %s\n", cli.Key(configPath), cli.Value(formatValueAsTOML(actualValue)))
					fmt.Printf("Status:  %s\n", cli.Warning("DRIFT - values differ"))
					fmt.Printf("\nTo apply the desired value:\n")
					fmt.Printf("  conf apply jj  # Applies ALL drifting jj values\n")
					fmt.Printf("  conf apply jj --dry-run  # Preview changes first\n")
					fmt.Printf("\nTo update desired to match actual:\n")
					fmt.Printf("  conf jj %s %s\n", configPath, formatValueAsTOML(actualValue))
				}
			} else {
				if actualValue == nil {
					fmt.Printf("%s = %s\n", cli.Key(configPath), cli.Muted("(not set)"))
				} else {
					fmt.Printf("Actual:  %s = %s\n", cli.Key(configPath), cli.Value(formatValueAsTOML(actualValue)))
					fmt.Printf("Status:  %s\n", cli.Muted("UNMANAGED - not in conf state"))
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
			preview, err := jjTool.PreviewSetConfig(configPath, parsedValue)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Print("DRY RUN: ")
			fmt.Print(preview)
			fmt.Printf("Would also record in conf state: jj.%s = %v\n", configPath, parsedValue)
		} else {
			// 1. Record desired state in conf config
			conf.SetToolValue("jj", configPath, parsedValue)
			if err := conf.Save(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to save conf state: %v\n", err)
				os.Exit(1)
			}

			// 2. Apply to target file
			err = jjTool.SetConfig(configPath, parsedValue)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✓ Set jj config: %s = %v\n", configPath, parsedValue)
			fmt.Printf("✓ Recorded in conf state\n")
		}

		fmt.Printf("Config file: %s\n", jjTool.GetConfigPath())
	},
}

func init() {
	jjCmd.Flags().BoolVar(&jjListFlag, "list", false, "List all available jj configuration settings with current values")
	jjCmd.Flags().BoolVar(&jjForce, "force", false, "Bypass schema validation when setting values")
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"conf/pkg/config"
	"conf/pkg/schemas"
	"conf/pkg/sync"
	"conf/pkg/tools"
	jjtool "conf/pkg/tools/jj"
	misetool "conf/pkg/tools/mise"
	shimstool "conf/pkg/tools/shims"
	starshiptool "conf/pkg/tools/starship"
	"github.com/spf13/cobra"
)

var dryRun bool

var rootCmd = &cobra.Command{
	Use:   "conf",
	Short: "Smart configuration manager with autocompletion",
	Long: `conf is a smart config manager that provides intelligent configuration 
management with autocomplete for tools like jj (Jujutsu) and mise. It understands 
tool schemas and provides surgical TOML editing while preserving formatting.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var jjListFlag bool

// jjCompletion provides schema-aware completion for jj commands
func jjCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Create jj tool to access schema
	jjTool, err := jjtool.NewJJTool()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	switch len(args) {
	case 0:
		// First argument: complete config paths from schema
		settings, err := jjTool.ListAllSettings()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		var completions []string
		for _, setting := range settings {
			// Filter by what user has typed so far
			if strings.HasPrefix(setting.Path, toComplete) {
				description := setting.Description
				if description == "" {
					description = fmt.Sprintf("Type: %s", setting.Type)
				}

				// Add current value info if set
				valueInfo := ""
				if setting.IsSet {
					valueInfo = fmt.Sprintf(" (current: %v)", setting.CurrentValue)
				}

				// Format: path<tab>description + value info
				completion := fmt.Sprintf("%s\t%s%s", setting.Path, description, valueInfo)
				completions = append(completions, completion)
			}
		}

		return completions, cobra.ShellCompDirectiveDefault

	case 1:
		// Second argument: complete values based on schema info for the given path
		configPath := args[0]

		// Get property info for this path
		settings, err := jjTool.ListAllSettings()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		var targetSetting *schemas.SettingInfo
		for _, setting := range settings {
			if setting.Path == configPath {
				targetSetting = &setting
				break
			}
		}

		if targetSetting == nil {
			return nil, cobra.ShellCompDirectiveDefault
		}

		var completions []string

		// If setting has enum values, complete with those
		if len(targetSetting.Enum) > 0 {
			for _, enumVal := range targetSetting.Enum {
				if strings.HasPrefix(enumVal, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tValid enum value", enumVal))
				}
			}
			return completions, cobra.ShellCompDirectiveDefault
		}

		// Provide type-based suggestions
		switch targetSetting.Type {
		case "boolean":
			for _, val := range []string{"true", "false"} {
				if strings.HasPrefix(val, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tBoolean value", val))
				}
			}
		case "string":
			// Show current value as suggestion if set
			if targetSetting.IsSet && targetSetting.CurrentValue != nil {
				currentVal := fmt.Sprintf("%v", targetSetting.CurrentValue)
				if strings.HasPrefix(currentVal, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tCurrent value", currentVal))
				}
			}
			// Show default value as suggestion if available
			if targetSetting.Default != nil {
				defaultVal := fmt.Sprintf("%v", targetSetting.Default)
				if strings.HasPrefix(defaultVal, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tDefault value", defaultVal))
				}
			}
		case "integer":
			// Show current and default values for integers
			if targetSetting.IsSet && targetSetting.CurrentValue != nil {
				currentVal := fmt.Sprintf("%v", targetSetting.CurrentValue)
				if strings.HasPrefix(currentVal, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tCurrent value", currentVal))
				}
			}
			if targetSetting.Default != nil {
				defaultVal := fmt.Sprintf("%v", targetSetting.Default)
				if strings.HasPrefix(defaultVal, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tDefault value", defaultVal))
				}
			}
		}

		return completions, cobra.ShellCompDirectiveDefault

	default:
		// No completion for additional arguments
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

var jjCmd = &cobra.Command{
	Use:   "jj [config.path] [value]",
	Short: "Configure jj (Jujutsu) settings",
	Long: `Get or set configuration values in ~/.config/jj/config.toml using dotted path notation.

Examples:
  conf jj --list                       # List all available settings with current values  
  conf jj user.name                    # Get current value
  conf jj user.name "John Doe"         # Set value
  conf jj user.email john@example.com  # Set email`,
	Args:              cobra.RangeArgs(0, 2),
	ValidArgsFunction: jjCompletion,
	Run: func(cmd *cobra.Command, args []string) {
		// Create jj tool
		jjTool, err := jjtool.NewJJToolWithDryRun(dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize jj tool: %v\n", err)
			os.Exit(1)
		}

		// Handle --list flag
		if jjListFlag {
			settings, err := jjTool.ListAllSettings()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to list settings: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("All jj configuration settings:")
			fmt.Println()

			for _, setting := range settings {
				fmt.Printf("  %s\n", setting.Path)
				fmt.Printf("    Type: %s\n", setting.Type)
				if setting.Description != "" {
					fmt.Printf("    Description: %s\n", setting.Description)
				}
				if setting.Default != nil {
					fmt.Printf("    Default: %v\n", setting.Default)
				}
				if len(setting.Enum) > 0 {
					fmt.Printf("    Valid values: %v\n", setting.Enum)
				}

				if setting.IsSet {
					fmt.Printf("    Current value: %v ✓\n", setting.CurrentValue)
				} else {
					fmt.Printf("    Current value: (not set)\n")
				}
				fmt.Println()
			}

			fmt.Printf("Config file: %s\n", jjTool.GetConfigPath())
			return
		}

		// Require arguments when not listing
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "Error: config path required (use --list to see available options)\n")
			os.Exit(1)
		}

		configPath := args[0]

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
				fmt.Printf("Desired: %s = %v\n", configPath, desiredValue)
				if actualValue == nil {
					fmt.Printf("Actual:  %s = (not set)\n", configPath)
					fmt.Printf("Status:  DRIFT - value not applied\n")
				} else if fmt.Sprintf("%v", actualValue) == fmt.Sprintf("%v", desiredValue) {
					fmt.Printf("Actual:  %s = %v\n", configPath, actualValue)
					fmt.Printf("Status:  IN SYNC\n")
				} else {
					fmt.Printf("Actual:  %s = %v\n", configPath, actualValue)
					fmt.Printf("Status:  DRIFT - values differ\n")
				}
			} else {
				if actualValue == nil {
					fmt.Printf("%s = (not set)\n", configPath)
				} else {
					fmt.Printf("Actual:  %s = %v\n", configPath, actualValue)
					fmt.Printf("Status:  UNMANAGED - not in conf state\n")
				}
			}
			return
		}

		// SET operation: config path and value provided
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

// miseCompletion provides completion for mise commands
func miseCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Create mise tool
	miseTool, err := misetool.NewMiseTool()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	switch len(args) {
	case 0:
		// First argument: complete config paths from common settings
		settings := miseTool.ListCommonSettings()
		var completions []string
		for _, setting := range settings {
			if strings.HasPrefix(setting.Path, toComplete) {
				completion := fmt.Sprintf("%s\t%s", setting.Path, setting.Description)
				completions = append(completions, completion)
			}
		}
		return completions, cobra.ShellCompDirectiveDefault

	case 1:
		// Second argument: provide type-based suggestions
		configPath := args[0]
		settings := miseTool.ListCommonSettings()

		for _, setting := range settings {
			if setting.Path == configPath {
				switch setting.Type {
				case "boolean":
					var completions []string
					for _, val := range []string{"true", "false"} {
						if strings.HasPrefix(val, toComplete) {
							completions = append(completions, fmt.Sprintf("%s\tBoolean value", val))
						}
					}
					return completions, cobra.ShellCompDirectiveDefault
				case "integer":
					// Show example as suggestion
					if strings.HasPrefix(setting.Example, toComplete) {
						return []string{fmt.Sprintf("%s\tExample value", setting.Example)}, cobra.ShellCompDirectiveDefault
					}
				}
				break
			}
		}
		return nil, cobra.ShellCompDirectiveDefault

	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

var miseCmd = &cobra.Command{
	Use:   "mise [config.path] [value]",
	Short: "Configure mise settings",
	Long: `Get or set configuration values in ~/.config/mise/config.toml using dotted path notation.

Examples:
  conf mise settings.experimental         # Get current value
  conf mise settings.experimental true    # Set boolean value
  conf mise settings.jobs 4               # Set numeric value`,
	Args:              cobra.RangeArgs(1, 2),
	ValidArgsFunction: miseCompletion,
	Run: func(cmd *cobra.Command, args []string) {
		configPath := args[0]

		// Create mise tool with dry-run mode
		miseTool, err := misetool.NewMiseToolWithDryRun(dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize mise tool: %v\n", err)
			os.Exit(1)
		}

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
		value := args[1]
		parsedValue := parseValue(value)

		if dryRun {
			// Show preview of what would happen
			preview, err := miseTool.PreviewSetConfig(configPath, parsedValue)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
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
	Short: "List common mise configuration options",
	Long:  `Display a list of commonly used mise configuration options with descriptions and examples.`,
	Run: func(cmd *cobra.Command, args []string) {
		miseTool, err := misetool.NewMiseTool()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize mise tool: %v\n", err)
			os.Exit(1)
		}

		settings := miseTool.ListCommonSettings()

		fmt.Println("Common mise configuration settings:")
		fmt.Println()

		for _, setting := range settings {
			fmt.Printf("  %s\n", setting.Path)
			fmt.Printf("    Type: %s\n", setting.Type)
			fmt.Printf("    Description: %s\n", setting.Description)
			fmt.Printf("    Example: %s\n", setting.Example)
			fmt.Println()
		}

		fmt.Printf("Config file: %s\n", miseTool.GetConfigPath())
	},
}

// starshipCompletion provides completion for starship commands
func starshipCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Create starship tool
	starshipTool, err := starshiptool.NewStarshipTool()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	switch len(args) {
	case 0:
		// First argument: complete config paths from common settings
		settings := starshipTool.ListCommonSettings()
		var completions []string
		for _, setting := range settings {
			if strings.HasPrefix(setting.Path, toComplete) {
				completion := fmt.Sprintf("%s\t%s", setting.Path, setting.Description)
				completions = append(completions, completion)
			}
		}
		return completions, cobra.ShellCompDirectiveDefault

	case 1:
		// Second argument: provide type-based suggestions
		configPath := args[0]
		settings := starshipTool.ListCommonSettings()

		for _, setting := range settings {
			if setting.Path == configPath {
				switch setting.Type {
				case "boolean":
					var completions []string
					for _, val := range []string{"true", "false"} {
						if strings.HasPrefix(val, toComplete) {
							completions = append(completions, fmt.Sprintf("%s\tBoolean value", val))
						}
					}
					return completions, cobra.ShellCompDirectiveDefault
				case "integer":
					// Show example as suggestion
					if strings.HasPrefix(setting.Example, toComplete) {
						return []string{fmt.Sprintf("%s\tExample value", setting.Example)}, cobra.ShellCompDirectiveDefault
					}
				case "string":
					// Show example as suggestion
					if strings.HasPrefix(setting.Example, toComplete) {
						return []string{fmt.Sprintf("%s\tExample value", setting.Example)}, cobra.ShellCompDirectiveDefault
					}
				}
				break
			}
		}
		return nil, cobra.ShellCompDirectiveDefault

	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

var starshipCmd = &cobra.Command{
	Use:   "starship [config.path] [value]",
	Short: "Configure starship settings",
	Long: `Get or set configuration values in ~/.config/starship.toml using dotted path notation.

Examples:
  conf starship add_newline              # Get current value
  conf starship add_newline true         # Set boolean value
  conf starship command_timeout 500      # Set timeout value`,
	Args:              cobra.RangeArgs(1, 2),
	ValidArgsFunction: starshipCompletion,
	Run: func(cmd *cobra.Command, args []string) {
		configPath := args[0]

		// Create starship tool with dry-run mode
		starshipTool, err := starshiptool.NewStarshipToolWithDryRun(dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize starship tool: %v\n", err)
			os.Exit(1)
		}

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
		value := args[1]
		parsedValue := parseValue(value)

		if dryRun {
			// Show preview of what would happen
			preview, err := starshipTool.PreviewSetConfig(configPath, parsedValue)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
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

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate schema-aware shell completion scripts",
	Long: `Generate shell completion scripts with intelligent schema-aware suggestions.

The completion system provides:
- Configuration path completion from actual schemas
- Type-aware value suggestions (boolean, enum, etc.)
- Current value display for existing settings
- Descriptions for all configuration options

Quick setup (recommended):
  # Bash - add to ~/.bashrc
  eval "$(conf completion bash)"

  # Zsh - add to ~/.zshrc
  eval "$(conf completion zsh)"

  # Fish - add to ~/.config/fish/config.fish
  conf completion fish | source

Persistent installation (alternative):
  # Bash
  conf completion bash > ~/.local/share/bash-completion/completions/conf

  # Zsh
  conf completion zsh > ~/.oh-my-zsh/completions/_conf

  # Fish
  conf completion fish > ~/.config/fish/completions/conf.fish

After adding to your shell config, restart your shell or source the file.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		shell := args[0]
		switch shell {
		case "bash":
			rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			rootCmd.GenFishCompletion(os.Stdout, true)
		default:
			fmt.Fprintf(os.Stderr, "Unsupported shell: %s\n", shell)
			fmt.Fprintln(os.Stderr, "\nSupported shells: bash, zsh, fish")
			os.Exit(1)
		}
	},
}

// parseValue attempts to parse a string value into the appropriate type
func parseValue(value string) interface{} {
	// Try boolean first
	if value == "true" || value == "false" {
		return value == "true"
	}

	// Try integer
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}

	// Try float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// Default to string
	return value
}

func init() {
	// Add dry-run flag to root command
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making any modifications")

	// Add --list flag to jj command
	jjCmd.Flags().BoolVar(&jjListFlag, "list", false, "List all available jj configuration settings with current values")

	miseCmd.AddCommand(miseListCmd)

	// Add shims subcommands
	shimsCmd.AddCommand(shimsCreateCmd)
	shimsCmd.AddCommand(shimsRemoveCmd)
	shimsCmd.AddCommand(shimsListCmd)

	rootCmd.AddCommand(jjCmd)
	rootCmd.AddCommand(miseCmd)
	rootCmd.AddCommand(starshipCmd)
	rootCmd.AddCommand(shimsCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(completionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var shimsCmd = &cobra.Command{
	Use:   "shims",
	Short: "Manage command shims",
	Long: `Create and manage executable command shims in ~/.local/bin/conf-shims/.
	
Shims are executable scripts that act as aliases for longer commands.
They work across all shells (bash, zsh, fish) without modifying shell config files.

Add ~/.local/bin/conf-shims to your PATH to use the shims.`,
}

var shimsCreateCmd = &cobra.Command{
	Use:   "create [name] [command]",
	Short: "Create a new command shim",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		command := args[1]

		shimsTool, err := shimstool.NewShimsToolWithDryRun(dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		err = shimsTool.CreateShim(name, command)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Created shim: %s -> %s\n", name, command)
	},
}

var shimsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all managed command shims",
	Run: func(cmd *cobra.Command, args []string) {
		shimsTool, err := shimstool.NewShimsTool()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		shims, err := shimsTool.ListShims()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(shims) == 0 {
			fmt.Println("No shims found.")
			return
		}

		for _, shim := range shims {
			fmt.Printf("%-12s -> %s\n", shim.Name, shim.Command)
		}
	},
}

var shimsRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove an existing command shim",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		shimsTool, err := shimstool.NewShimsToolWithDryRun(dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		err = shimsTool.RemoveShim(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Removed shim: %s\n", name)
	},
}

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

var migrateCmd = &cobra.Command{
	Use:   "migrate [tool]",
	Short: "Migrate config from centralized to per-tool files",
	Long: `Convert configuration from centralized format (all values in config.toml)
to per-tool files format (metadata in config.toml, values in {tool}.toml).

This creates separate TOML files for each tool that uses the tool's native schema.
For example, jj values will be moved from config.toml to ~/.config/conf/jj.toml.

Examples:
  conf migrate          # Migrate all tools
  conf migrate jj       # Migrate only jj config
  conf migrate --dry-run jj  # Preview what would be migrated`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to load conf config: %v\n", err)
			os.Exit(1)
		}

		var toolsToMigrate []string
		if len(args) == 1 {
			toolsToMigrate = []string{args[0]}
		} else {
			// Migrate all tools
			for toolName := range conf.Tools {
				toolsToMigrate = append(toolsToMigrate, toolName)
			}
		}

		for _, toolName := range toolsToMigrate {
			if err := migrateTool(conf, toolName, dryRun); err != nil {
				fmt.Fprintf(os.Stderr, "Error migrating %s: %v\n", toolName, err)
				os.Exit(1)
			}
		}

		// Save the updated config (removes values from config.toml)
		if !dryRun {
			if err := conf.Save(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to save config: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("\n✓ Migration complete")
		}
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync [tool]",
	Short: "Sync configuration with iCloud Drive",
	Long: `Merge local configuration with iCloud Drive using simple Last-Write-Wins strategy.

This command downloads config from iCloud, merges with local config, and uploads
the merged result back to iCloud. Conflicts are resolved using Last-Write-Wins
based on file modification times.

iCloud Drive location:
  ~/Library/Mobile Documents/com~apple~CloudDocs/conf/

  Tool configs are stored as:
    - jj.toml (for jj config)
    - mise.toml (for mise config)
    - etc.

Examples:
  conf sync           # Sync all tools
  conf sync jj        # Sync only jj config
  conf sync --dry-run # Preview what would be synced`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		conf, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to load conf config: %v\n", err)
			os.Exit(1)
		}

		configDir, err := config.ConfigDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to get config directory: %v\n", err)
			os.Exit(1)
		}

		var toolsToSync []string
		if len(args) == 1 {
			toolsToSync = []string{args[0]}
		} else {
			// Sync all tools
			for toolName := range conf.Tools {
				toolsToSync = append(toolsToSync, toolName)
			}
		}

		for _, toolName := range toolsToSync {
			if err := syncTool(conf, configDir, toolName, dryRun); err != nil {
				fmt.Fprintf(os.Stderr, "Error syncing %s: %v\n", toolName, err)
				os.Exit(1)
			}
		}

		if !dryRun {
			fmt.Println("\n✓ Sync complete")
		}
	},
}

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

// syncTool syncs a tool's config with iCloud Drive
func syncTool(conf *config.Config, configDir, toolName string, dryRun bool) error {
	tool, exists := conf.GetTool(toolName)
	if !exists {
		return fmt.Errorf("tool %s not configured", toolName)
	}

	fmt.Printf("Syncing %s configuration...\n", toolName)

	// Load sync metadata
	metadata, err := sync.LoadSyncMetadata(configDir)
	if err != nil {
		return fmt.Errorf("failed to load sync metadata: %w", err)
	}

	// Get local values
	localValues, err := sync.GetLocalValues(conf, toolName)
	if err != nil {
		return fmt.Errorf("failed to get local values: %w", err)
	}

	// Download from iCloud
	icloudValues, err := sync.DownloadFromICloud(toolName)
	if err != nil {
		return fmt.Errorf("failed to download from iCloud: %w", err)
	}

	if icloudValues == nil && len(localValues) == 0 {
		fmt.Printf("  %s: No configuration to sync\n", toolName)
		return nil
	}

	// Handle first-time sync cases
	if icloudValues == nil {
		// No iCloud data, upload local
		fmt.Printf("  No iCloud data found, uploading local config (%d values)\n", len(localValues))
		if !dryRun {
			if err := sync.UploadToICloud(toolName, localValues); err != nil {
				return fmt.Errorf("failed to upload to iCloud: %w", err)
			}

			// Update metadata
			perToolPath := filepath.Join(configDir, toolName+".toml")
			if localHash, err := sync.ComputeFileHash(perToolPath); err == nil {
				metadata.UpdateToolState(toolName, localHash, localHash)
				metadata.Save(configDir)
			}

			fmt.Printf("  ✓ Uploaded to iCloud\n")
		}
		return nil
	}

	if len(localValues) == 0 {
		// No local data, download from iCloud
		fmt.Printf("  No local data found, downloading from iCloud (%d values)\n", len(icloudValues))
		if !dryRun {
			// Set values in config
			for k, v := range icloudValues {
				conf.SetToolValue(toolName, k, v)
			}

			if err := conf.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			// Update metadata
			icloudPath, _ := sync.ICloudDrivePath()
			icloudFilePath := filepath.Join(icloudPath, toolName+".toml")
			if icloudHash, err := sync.ComputeFileHash(icloudFilePath); err == nil {
				metadata.UpdateToolState(toolName, icloudHash, icloudHash)
				metadata.Save(configDir)
			}

			fmt.Printf("  ✓ Downloaded from iCloud\n")
		}
		return nil
	}

	// Both have data - need to merge
	fmt.Printf("  Local: %d values, iCloud: %d values\n", len(localValues), len(icloudValues))

	// Get file modification times for LWW
	perToolPath := filepath.Join(configDir, toolName+".toml")
	icloudPath, _ := sync.ICloudDrivePath()
	icloudFilePath := filepath.Join(icloudPath, toolName+".toml")

	localStat, _ := os.Stat(perToolPath)
	icloudStat, _ := os.Stat(icloudFilePath)

	localMtime := localStat.ModTime().Unix()
	icloudMtime := icloudStat.ModTime().Unix()

	// Merge configs
	merged := sync.MergeConfigs(localValues, icloudValues, localMtime, icloudMtime)

	fmt.Printf("  Merged: %d values\n", len(merged))

	if dryRun {
		fmt.Printf("  Would upload merged config to iCloud\n")
		fmt.Printf("  Would update local config\n")
	} else {
		// Upload merged result to iCloud
		if err := sync.UploadToICloud(toolName, merged); err != nil {
			return fmt.Errorf("failed to upload merged config: %w", err)
		}

		// Update local config
		tool.Values = merged
		conf.Tools[toolName] = tool
		if err := conf.Save(); err != nil {
			return fmt.Errorf("failed to save local config: %w", err)
		}

		// Update metadata
		if icloudHash, err := sync.ComputeFileHash(icloudFilePath); err == nil {
			if localHash, err := sync.ComputeFileHash(perToolPath); err == nil {
				metadata.UpdateToolState(toolName, icloudHash, localHash)
				metadata.Save(configDir)
			}
		}

		fmt.Printf("  ✓ Synced with iCloud\n")
	}

	return nil
}

// migrateTool migrates a tool's config from centralized to per-tool file
func migrateTool(conf *config.Config, toolName string, dryRun bool) error {
	tool, exists := conf.GetTool(toolName)
	if !exists {
		return fmt.Errorf("tool %s not configured", toolName)
	}

	// Check if tool has values to migrate
	if tool.Values == nil || len(tool.Values) == 0 {
		fmt.Printf("%s: No values to migrate (already using per-tool file or no values set)\n", toolName)
		return nil
	}

	configDir, err := config.ConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	perToolPath := fmt.Sprintf("%s/%s.toml", configDir, toolName)

	// Check if per-tool file already exists
	if _, err := os.Stat(perToolPath); err == nil {
		fmt.Printf("%s: Per-tool file already exists at %s, skipping\n", toolName, perToolPath)
		return nil
	}

	fmt.Printf("Migrating %s configuration to %s\n", toolName, perToolPath)
	fmt.Printf("  %d values to migrate\n", len(tool.Values))

	if dryRun {
		fmt.Printf("  Would create: %s\n", perToolPath)
		for path, value := range tool.Values {
			fmt.Printf("    %s = %v\n", path, value)
		}
	} else {
		// Create the per-tool file
		// The savePerToolConfigs will handle writing values to it
		// We just need to create an empty file so Save() knows to write values there
		if err := os.WriteFile(perToolPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create per-tool file: %w", err)
		}
		fmt.Printf("  ✓ Created %s with %d values\n", perToolPath, len(tool.Values))
	}

	return nil
}

// applyTool applies desired state for a specific tool
func applyTool(conf *config.Config, toolName string, dryRun bool) error {
	tool, exists := conf.GetTool(toolName)
	if !exists {
		return fmt.Errorf("tool %s not configured", toolName)
	}

	if tool.Values == nil || len(tool.Values) == 0 {
		fmt.Printf("%s: No values to apply\n", toolName)
		return nil
	}

	fmt.Printf("Applying %s configuration...\n", toolName)

	for path, value := range tool.Values {
		if dryRun {
			fmt.Printf("  Would set %s.%s = %v\n", toolName, path, value)
		} else {
			if err := applyToolValue(toolName, path, value); err != nil {
				return fmt.Errorf("failed to apply %s.%s: %w", toolName, path, err)
			}
			fmt.Printf("  ✓ Set %s.%s = %v\n", toolName, path, value)
		}
	}

	return nil
}

// applyToolValue applies a single configuration value to a tool
func applyToolValue(toolName, path string, value interface{}) error {
	return tools.ApplyToolValue(toolName, path, value)
}

// showToolStatus shows drift status for a specific tool
func showToolStatus(conf *config.Config, toolName string) error {
	tool, exists := conf.GetTool(toolName)
	if !exists {
		return fmt.Errorf("tool %s not configured", toolName)
	}

	fmt.Printf("%s status:\n", toolName)

	if tool.Values == nil || len(tool.Values) == 0 {
		fmt.Printf("  No managed values\n")
		return nil
	}

	hasChanges := false
	for path, desiredValue := range tool.Values {
		actualValue, err := getActualValue(toolName, path)
		if err != nil {
			fmt.Printf("  %s: ERROR - %v\n", path, err)
			hasChanges = true
			continue
		}

		if actualValue == nil {
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

// getActualValue gets the actual value from a tool config file
func getActualValue(toolName, path string) (interface{}, error) {
	return tools.GetActualValue(toolName, path)
}

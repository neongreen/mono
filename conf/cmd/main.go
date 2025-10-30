package main

import (
	"fmt"
	"os"

	"github.com/neongreen/mono/conf/pkg/config"

	claudetool "github.com/neongreen/mono/conf/pkg/tools/claude"
	jjtool "github.com/neongreen/mono/conf/pkg/tools/jj"
	misetool "github.com/neongreen/mono/conf/pkg/tools/mise"
	shimstool "github.com/neongreen/mono/conf/pkg/tools/shims"
	starshiptool "github.com/neongreen/mono/conf/pkg/tools/starship"
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
var claudeListFlag bool

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

var claudeCmd = &cobra.Command{
	Use:   "claude [config.path] [value]",
	Short: "Configure Claude Code settings",
	Long: `Get or set configuration values in ~/.config/claude/config.json using dotted path notation.

Examples:
  conf claude --list                          # List all available settings with current values
  conf claude model                           # Get current value
  conf claude model "sonnet"                  # Set value
  conf claude alwaysThinkingEnabled true      # Enable extended thinking`,
	Args:              cobra.RangeArgs(0, 2),
	ValidArgsFunction: claudeCompletion,
	Run: func(cmd *cobra.Command, args []string) {
		// Create claude tool
		claudeTool, err := claudetool.NewClaudeToolWithDryRun(dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize claude tool: %v\n", err)
			os.Exit(1)
		}

		// Handle --list flag
		if claudeListFlag {
			settings, err := claudeTool.ListAllSettings()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to list settings: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("All Claude Code configuration settings:")
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

			fmt.Printf("Config file: %s\n", claudeTool.GetConfigPath())
			return
		}

		// Require arguments when not listing
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "Error: config path required (use --list to see available options)\n")
			os.Exit(1)
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

var starshipListCmd = &cobra.Command{
	Use:   "list",
	Short: "List common starship configuration options",
	Long:  `Display a list of commonly used starship configuration options with descriptions and examples.`,
	Run: func(cmd *cobra.Command, args []string) {
		starshipTool, err := starshiptool.NewStarshipTool()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize starship tool: %v\n", err)
			os.Exit(1)
		}

		settings := starshipTool.ListCommonSettings()

		fmt.Println("Common starship configuration settings:")
		fmt.Println()

		for _, setting := range settings {
			fmt.Printf("  %s\n", setting.Path)
			fmt.Printf("    Type: %s\n", setting.Type)
			fmt.Printf("    Description: %s\n", setting.Description)
			fmt.Printf("    Example: %s\n", setting.Example)
			fmt.Println()
		}

		fmt.Printf("Config file: %s\n", starshipTool.GetConfigPath())
	},
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

func init() {
	// Add dry-run flag to root command
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making any modifications")

	// Add --list flags to commands
	jjCmd.Flags().BoolVar(&jjListFlag, "list", false, "List all available jj configuration settings with current values")
	claudeCmd.Flags().BoolVar(&claudeListFlag, "list", false, "List all available Claude Code configuration settings with current values")

	miseCmd.AddCommand(miseListCmd)
	starshipCmd.AddCommand(starshipListCmd)

	// Add shims subcommands
	shimsCmd.AddCommand(shimsCreateCmd)
	shimsCmd.AddCommand(shimsRemoveCmd)
	shimsCmd.AddCommand(shimsListCmd)

	rootCmd.AddCommand(jjCmd)
	rootCmd.AddCommand(claudeCmd)
	rootCmd.AddCommand(miseCmd)
	rootCmd.AddCommand(starshipCmd)
	rootCmd.AddCommand(shimsCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(importCmd)
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

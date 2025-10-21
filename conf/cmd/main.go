package main

import (
	"fmt"
	"os"
	"strconv"

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

var jjCmd = &cobra.Command{
	Use:   "jj [config.path] [value]",
	Short: "Configure jj (Jujutsu) settings",
	Long:  `Get or set configuration values in ~/.config/jj/config.toml using dotted path notation.

Examples:
  conf jj user.name                    # Get current value
  conf jj user.name "John Doe"         # Set value
  conf jj user.email john@example.com  # Set email`,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		configPath := args[0]

		// Create jj tool
		jjTool, err := jjtool.NewJJToolWithDryRun(dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize jj tool: %v\n", err)
			os.Exit(1)
		}

		// GET operation: only config path provided
		if len(args) == 1 {
			value, err := jjTool.GetConfig(configPath)
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
			preview, err := jjTool.PreviewSetConfig(configPath, parsedValue)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Print(preview)
		} else {
			// Set the configuration
			err = jjTool.SetConfig(configPath, parsedValue)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✓ Set jj config: %s = %v\n", configPath, parsedValue)
		}

		fmt.Printf("Config file: %s\n", jjTool.GetConfigPath())
	},
}

var jjListCmd = &cobra.Command{
	Use:   "list",
	Short: "List common jj configuration options",
	Long:  `Display a list of commonly used jj configuration options with descriptions and examples.`,
	Run: func(cmd *cobra.Command, args []string) {
		jjTool, err := jjtool.NewJJTool()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize jj tool: %v\n", err)
			os.Exit(1)
		}

		settings := jjTool.ListCommonSettings()

		fmt.Println("Common jj configuration settings:")
		fmt.Println()

		for _, setting := range settings {
			fmt.Printf("  %s\n", setting.Path)
			fmt.Printf("    Type: %s\n", setting.Type)
			fmt.Printf("    Description: %s\n", setting.Description)
			fmt.Printf("    Example: %s\n", setting.Example)
			fmt.Println()
		}

		fmt.Printf("Config file: %s\n", jjTool.GetConfigPath())
	},
}

var miseCmd = &cobra.Command{
	Use:   "mise [config.path] [value]",
	Short: "Configure mise settings",
	Long:  `Get or set configuration values in ~/.config/mise/config.toml using dotted path notation.

Examples:
  conf mise settings.experimental         # Get current value
  conf mise settings.experimental true    # Set boolean value
  conf mise settings.jobs 4               # Set numeric value`,
	Args:  cobra.RangeArgs(1, 2),
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

var starshipCmd = &cobra.Command{
	Use:   "starship [config.path] [value]",
	Short: "Configure starship settings",
	Long:  `Get or set configuration values in ~/.config/starship.toml using dotted path notation.

Examples:
  conf starship add_newline              # Get current value
  conf starship add_newline true         # Set boolean value
  conf starship command_timeout 500      # Set timeout value`,
	Args:  cobra.RangeArgs(1, 2),
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
	Short: "Generate shell completion scripts",
	Args:  cobra.ExactArgs(1),
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
			fmt.Printf("Unsupported shell: %s\n", shell)
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

	jjCmd.AddCommand(jjListCmd)
	miseCmd.AddCommand(miseListCmd)
	
	// Add shims subcommands
	shimsCmd.AddCommand(shimsCreateCmd)
	shimsCmd.AddCommand(shimsRemoveCmd)
	shimsCmd.AddCommand(shimsListCmd)

	rootCmd.AddCommand(jjCmd)
	rootCmd.AddCommand(miseCmd)
	rootCmd.AddCommand(starshipCmd)
	rootCmd.AddCommand(shimsCmd)
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
	Long:  `Create and manage executable command shims in ~/.local/bin/conf-shims/.
	
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


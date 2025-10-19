package main

import (
	"fmt"
	"os"
	"strconv"

	jjtool "conf/pkg/tools/jj"
	misetool "conf/pkg/tools/mise"
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
	Long:  `Set configuration values in ~/.config/jj/config.toml using dotted path notation.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		configPath := args[0]
		value := args[1]

		// Create jj tool with dry-run mode
		jjTool, err := jjtool.NewJJToolWithDryRun(dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize jj tool: %v\n", err)
			os.Exit(1)
		}

		// Parse value with type detection
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
	Long:  `Set configuration values in ~/.config/mise/config.toml using dotted path notation.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		configPath := args[0]
		value := args[1]

		// Create mise tool with dry-run mode
		miseTool, err := misetool.NewMiseToolWithDryRun(dryRun)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize mise tool: %v\n", err)
			os.Exit(1)
		}

		// Parse value with type detection
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

	rootCmd.AddCommand(jjCmd)
	rootCmd.AddCommand(miseCmd)
	rootCmd.AddCommand(completionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

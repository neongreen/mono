package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	jjtool "github.com/neongreen/monorepo/conf/pkg/tools/jj"
)

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
	Long:  `Set configuration values in ~/.jjconfig.toml using dotted path notation.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		configPath := args[0]
		value := args[1]
		
		// Create jj tool
		jjTool, err := jjtool.NewJJTool()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to initialize jj tool: %v\n", err)
			os.Exit(1)
		}
		
		// Try to parse value as different types
		var parsedValue interface{}
		
		// Try boolean first
		if value == "true" || value == "false" {
			parsedValue = value == "true"
		} else if intVal, err := strconv.Atoi(value); err == nil {
			// Try integer
			parsedValue = intVal
		} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			// Try float
			parsedValue = floatVal
		} else {
			// Default to string
			parsedValue = value
		}
		
		// Set the configuration
		err = jjTool.SetConfig(configPath, parsedValue)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("✓ Set jj config: %s = %v\n", configPath, parsedValue)
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

func init() {
	jjCmd.AddCommand(jjListCmd)
	rootCmd.AddCommand(jjCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

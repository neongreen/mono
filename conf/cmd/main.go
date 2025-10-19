package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
		fmt.Printf("Setting jj config: %s = %s\n", configPath, value)
		fmt.Println("(Implementation pending)")
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
		fmt.Printf("Setting mise config: %s = %s\n", configPath, value)
		fmt.Println("(Implementation pending)")
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for conf.

To load completions:

Bash:
  source <(conf completion bash)

Zsh:
  conf completion zsh > _conf
  # Move _conf to somewhere in your $fpath

Fish:
  conf completion fish | source
`,
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
			fmt.Printf("Unsupported shell: %s\n", shell)
			os.Exit(1)
		}
	},
}

func init() {
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

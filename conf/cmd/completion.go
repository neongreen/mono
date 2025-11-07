package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/neongreen/mono/conf/pkg/schemas"
	claudetool "github.com/neongreen/mono/conf/pkg/tools/claude"
	jjtool "github.com/neongreen/mono/conf/pkg/tools/jj"
	misetool "github.com/neongreen/mono/conf/pkg/tools/mise"
	starshiptool "github.com/neongreen/mono/conf/pkg/tools/starship"
	"github.com/spf13/cobra"
)

// completionOptions controls behavior of generic completion function
type completionOptions struct {
	showCurrentValuesInList bool // Show current values in case 0 (list all settings)
	handleStringType        bool // Handle string type completions in case 1
	showIntegerCurrentValue bool // Show current value for integer/number types in case 1
}

// genericCompletion provides schema-aware completion using a settings provider function
func genericCompletion(
	args []string,
	toComplete string,
	getSettings func() ([]schemas.SettingInfo, error),
	opts completionOptions,
) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		// List all settings
		settings, err := getSettings()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		var completions []string
		for _, setting := range settings {
			if strings.HasPrefix(setting.Path, toComplete) {
				description := setting.Description
				if description == "" {
					description = fmt.Sprintf("Type: %s", setting.Type)
				}

				valueInfo := ""
				if opts.showCurrentValuesInList && setting.IsSet {
					valueInfo = fmt.Sprintf(" (current: %v)", setting.CurrentValue)
				}

				completion := fmt.Sprintf("%s\t%s%s", setting.Path, description, valueInfo)
				completions = append(completions, completion)
			}
		}

		return completions, cobra.ShellCompDirectiveDefault

	case 1:
		// Value completion for a specific setting
		configPath := args[0]

		settings, err := getSettings()
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

		// Handle enum values
		if len(targetSetting.Enum) > 0 {
			for _, enumVal := range targetSetting.Enum {
				if strings.HasPrefix(enumVal, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tValid enum value", enumVal))
				}
			}
			return completions, cobra.ShellCompDirectiveDefault
		}

		// Handle different types
		switch targetSetting.Type {
		case "boolean":
			for _, val := range []string{"true", "false"} {
				if strings.HasPrefix(val, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tBoolean value", val))
				}
			}
		case "string":
			if opts.handleStringType {
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
		case "integer":
			if opts.showIntegerCurrentValue && targetSetting.IsSet && targetSetting.CurrentValue != nil {
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
		case "number":
			// Handle number type (used by starship and mise)
			if targetSetting.Default != nil {
				defaultVal := fmt.Sprintf("%v", targetSetting.Default)
				if strings.HasPrefix(defaultVal, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tDefault value", defaultVal))
				}
			}
		}

		return completions, cobra.ShellCompDirectiveDefault

	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// jjCompletion provides schema-aware completion for jj commands
func jjCompletion(args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	jjTool, err := jjtool.NewJJTool()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return genericCompletion(args, toComplete, jjTool.ListAllSettings, completionOptions{
		showCurrentValuesInList: true,
		handleStringType:        true,
		showIntegerCurrentValue: true,
	})
}

// claudeCompletion provides schema-aware completion for claude commands
func claudeCompletion(args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	claudeTool, err := claudetool.NewClaudeTool()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return genericCompletion(args, toComplete, claudeTool.ListAllSettings, completionOptions{
		showCurrentValuesInList: true,
		handleStringType:        true,
		showIntegerCurrentValue: true,
	})
}

// miseCompletion provides completion for mise commands
func miseCompletion(args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	miseTool, err := misetool.NewMiseTool()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return genericCompletion(args, toComplete, miseTool.ListAllSettings, completionOptions{
		showCurrentValuesInList: false,
		handleStringType:        false,
		showIntegerCurrentValue: false,
	})
}

// starshipCompletion provides completion for starship commands
func starshipCompletion(args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	starshipTool, err := starshiptool.NewStarshipTool()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return genericCompletion(args, toComplete, starshipTool.ListAllSettings, completionOptions{
		showCurrentValuesInList: true,
		handleStringType:        true,
		showIntegerCurrentValue: false,
	})
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

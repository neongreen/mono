package main

import (
	"fmt"
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
	cmd *cobra.Command,
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
func jjCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	jjTool, err := jjtool.NewJJTool()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return genericCompletion(cmd, args, toComplete, jjTool.ListAllSettings, completionOptions{
		showCurrentValuesInList: true,
		handleStringType:        true,
		showIntegerCurrentValue: true,
	})
}

// claudeCompletion provides schema-aware completion for claude commands
func claudeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	claudeTool, err := claudetool.NewClaudeTool()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return genericCompletion(cmd, args, toComplete, claudeTool.ListAllSettings, completionOptions{
		showCurrentValuesInList: true,
		handleStringType:        true,
		showIntegerCurrentValue: true,
	})
}

// miseCompletion provides completion for mise commands
func miseCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	miseTool, err := misetool.NewMiseTool()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return genericCompletion(cmd, args, toComplete, miseTool.ListAllSettings, completionOptions{
		showCurrentValuesInList: false,
		handleStringType:        false,
		showIntegerCurrentValue: false,
	})
}

// starshipCompletion provides completion for starship commands
func starshipCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	starshipTool, err := starshiptool.NewStarshipTool()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return genericCompletion(cmd, args, toComplete, starshipTool.ListAllSettings, completionOptions{
		showCurrentValuesInList: true,
		handleStringType:        true,
		showIntegerCurrentValue: false,
	})
}

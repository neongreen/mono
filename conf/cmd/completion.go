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

// jjCompletion provides schema-aware completion for jj commands
func jjCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

	jjTool, err := jjtool.NewJJTool()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	switch len(args) {
	case 0:

		settings, err := jjTool.ListAllSettings()
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
				if setting.IsSet {
					valueInfo = fmt.Sprintf(" (current: %v)", setting.CurrentValue)
				}

				completion := fmt.Sprintf("%s\t%s%s", setting.Path, description, valueInfo)
				completions = append(completions, completion)
			}
		}

		return completions, cobra.ShellCompDirectiveDefault

	case 1:

		configPath := args[0]

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

		if len(targetSetting.Enum) > 0 {
			for _, enumVal := range targetSetting.Enum {
				if strings.HasPrefix(enumVal, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tValid enum value", enumVal))
				}
			}
			return completions, cobra.ShellCompDirectiveDefault
		}

		switch targetSetting.Type {
		case "boolean":
			for _, val := range []string{"true", "false"} {
				if strings.HasPrefix(val, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tBoolean value", val))
				}
			}
		case "string":

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
		case "integer":

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

		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// claudeCompletion provides schema-aware completion for claude commands
func claudeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

	claudeTool, err := claudetool.NewClaudeTool()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	switch len(args) {
	case 0:

		settings, err := claudeTool.ListAllSettings()
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
				if setting.IsSet {
					valueInfo = fmt.Sprintf(" (current: %v)", setting.CurrentValue)
				}

				completion := fmt.Sprintf("%s\t%s%s", setting.Path, description, valueInfo)
				completions = append(completions, completion)
			}
		}

		return completions, cobra.ShellCompDirectiveDefault

	case 1:

		configPath := args[0]

		settings, err := claudeTool.ListAllSettings()
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

		if len(targetSetting.Enum) > 0 {
			for _, enumVal := range targetSetting.Enum {
				if strings.HasPrefix(enumVal, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tValid enum value", enumVal))
				}
			}
			return completions, cobra.ShellCompDirectiveDefault
		}

		switch targetSetting.Type {
		case "boolean":
			for _, val := range []string{"true", "false"} {
				if strings.HasPrefix(val, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tBoolean value", val))
				}
			}
		case "string":

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
		case "integer":

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

		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// miseCompletion provides completion for mise commands
func miseCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

	miseTool, err := misetool.NewMiseTool()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	switch len(args) {
	case 0:

		settings, err := miseTool.ListAllSettings()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		var completions []string
		for _, setting := range settings {
			if strings.HasPrefix(setting.Path, toComplete) {
				completion := fmt.Sprintf("%s\t%s", setting.Path, setting.Description)
				completions = append(completions, completion)
			}
		}
		return completions, cobra.ShellCompDirectiveDefault

	case 1:

		configPath := args[0]
		settings, err := miseTool.ListAllSettings()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

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
				case "integer", "number":

					if setting.Default != nil {
						defaultStr := fmt.Sprintf("%v", setting.Default)
						if strings.HasPrefix(defaultStr, toComplete) {
							return []string{fmt.Sprintf("%s\tDefault value", defaultStr)}, cobra.ShellCompDirectiveDefault
						}
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

// starshipCompletion provides completion for starship commands
func starshipCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

	starshipTool, err := starshiptool.NewStarshipTool()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	switch len(args) {
	case 0:

		settings, err := starshipTool.ListAllSettings()
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
				if setting.IsSet {
					valueInfo = fmt.Sprintf(" (current: %v)", setting.CurrentValue)
				}

				completion := fmt.Sprintf("%s\t%s%s", setting.Path, description, valueInfo)
				completions = append(completions, completion)
			}
		}
		return completions, cobra.ShellCompDirectiveDefault

	case 1:

		configPath := args[0]

		settings, err := starshipTool.ListAllSettings()
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

		if len(targetSetting.Enum) > 0 {
			for _, enumVal := range targetSetting.Enum {
				if strings.HasPrefix(enumVal, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tValid enum value", enumVal))
				}
			}
			return completions, cobra.ShellCompDirectiveDefault
		}

		switch targetSetting.Type {
		case "boolean":
			for _, val := range []string{"true", "false"} {
				if strings.HasPrefix(val, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tBoolean value", val))
				}
			}
		case "string":

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
		case "integer", "number":

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

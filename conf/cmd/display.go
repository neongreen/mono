package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/neongreen/mono/conf/pkg/schemas"
	"github.com/neongreen/mono/lib/cli"
)

// renderSettingsTable renders a table of settings with colors and proper formatting
func renderSettingsTable(settings []schemas.SettingInfo, configPath string) {
	if len(settings) == 0 {
		fmt.Println("No settings available")
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Setting", "Type", "Value", "Description"})

	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.Style().Options.DrawBorder = false

	// Configure column widths for better display
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: false, WidthMax: 40},                                  // Setting path
		{Number: 2, AutoMerge: false, WidthMax: 15},                                  // Type
		{Number: 3, AutoMerge: false, WidthMax: 30, WidthMaxEnforcer: text.WrapSoft}, // Value
		{Number: 4, AutoMerge: false, WidthMax: 50, WidthMaxEnforcer: text.WrapSoft}, // Description
	})

	for _, setting := range settings {
		// Format the setting path with color
		pathStr := cli.Key(setting.Path)

		// Format type with color
		typeStr := cli.Type(setting.Type)

		// Format value with color and status
		var valueStr string
		if setting.IsSet {
			valueStr = cli.Successf("✓ %v", formatValue(setting.CurrentValue))
		} else if setting.Default != nil {
			valueStr = cli.Secondaryf("(default: %v)", formatValue(setting.Default))
		} else {
			valueStr = cli.Muted("(not set)")
		}

		// Add enum values if present
		if len(setting.Enum) > 0 {
			valueStr += "\n" + cli.Mutedf("values: %s", strings.Join(setting.Enum, ", "))
		}

		// Format description
		descStr := setting.Description
		if descStr == "" {
			descStr = cli.Muted("-")
		}

		t.AppendRow(table.Row{pathStr, typeStr, valueStr, descStr})
	}

	t.Render()

	// Show config file path at the bottom
	fmt.Println()
	fmt.Printf("Config file: %s\n", configPath)
}

// formatValue formats an interface{} value for display
func formatValue(v any) string {
	if v == nil {
		return "nil"
	}

	switch val := v.(type) {
	case string:
		// Quote strings for clarity
		return fmt.Sprintf("%q", val)
	case []any:
		// Format arrays
		if len(val) == 0 {
			return "[]"
		}
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = formatValue(item)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]any:
		// Format objects
		if len(val) == 0 {
			return "{}"
		}
		return "{...}"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// CommonSetting represents a common configuration setting
type CommonSetting struct {
	Path        string
	Type        string
	Description string
	Example     string
}

// renderCommonSettingsTable renders a table of common settings (for mise/starship)
func renderCommonSettingsTable(settings []CommonSetting, configPath string) {
	if len(settings) == 0 {
		fmt.Println("No settings available")
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Setting", "Type", "Description", "Example"})

	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.Style().Options.DrawBorder = false

	// Configure column widths for better display
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: false, WidthMax: 30},                                  // Setting path
		{Number: 2, AutoMerge: false, WidthMax: 12},                                  // Type
		{Number: 3, AutoMerge: false, WidthMax: 45, WidthMaxEnforcer: text.WrapSoft}, // Description
		{Number: 4, AutoMerge: false, WidthMax: 25, WidthMaxEnforcer: text.WrapSoft}, // Example
	})

	for _, setting := range settings {
		// Format with colors
		pathStr := cli.Key(setting.Path)
		typeStr := cli.Type(setting.Type)
		descStr := setting.Description
		exampleStr := cli.Value(setting.Example)

		t.AppendRow(table.Row{pathStr, typeStr, descStr, exampleStr})
	}

	t.Render()

	// Show config file path at the bottom
	fmt.Println()
	fmt.Printf("Config file: %s\n", configPath)
}

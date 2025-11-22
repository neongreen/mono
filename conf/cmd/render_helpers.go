package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/neongreen/mono/lib/cli"
)

// formatValue formats an interface{} value for display in tables.
func formatValue(v any) string {
	if v == nil {
		return "nil"
	}

	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case []any:
		if len(val) == 0 {
			return "[]"
		}
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = formatValue(item)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]any:
		if len(val) == 0 {
			return "{}"
		}
		return "{...}"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// formatValueShort renders a value as TOML-compatible text and truncates long strings for readability.
func formatValueShort(v any) string {
	if v == nil {
		return "(not set)"
	}

	str := formatValueAsTOML(v)
	str = strings.ReplaceAll(str, "\n", "\\n")

	const maxLen = 80
	if len(str) > maxLen {
		extra := len(str) - maxLen
		str = str[:maxLen] + cli.Mutedf("… (+%d chars)", extra)
	}

	return str
}

func formatLabeledValue(label string, v any) string {
	labelStr := color.New(color.FgHiBlack, color.Italic).Sprint(label)
	if v == nil {
		return fmt.Sprintf("%s %s", cli.Muted("(not set)"), labelStr)
	}
	return fmt.Sprintf("%s %s", formatValueShort(v), labelStr)
}

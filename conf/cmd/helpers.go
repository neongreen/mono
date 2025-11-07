package main

import (
	"strconv"

	tomlcp "github.com/neongreen/mono/lib/toml"
)

// parseValue attempts to parse a string value into the appropriate type
func parseValue(value string) any {

	if value == "true" || value == "false" {
		return value == "true"
	}

	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}

	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	return value
}

// formatValueAsTOML formats a value as TOML-compatible string representation
func formatValueAsTOML(value any) string {
	if value == nil {
		return "(not set)"
	}

	// Use lib/toml's formatValueToString which handles all TOML types properly
	formatted, err := tomlcp.FormatValueToString(value)
	if err != nil {
		// Fallback to basic formatting if it fails (though it shouldn't)
		// This handles edge cases that formatValueToString doesn't support
		return "(unsupported type)"
	}

	return formatted
}

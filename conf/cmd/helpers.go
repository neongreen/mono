package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// parseValue attempts to parse a string value into the appropriate type
func parseValue(value string) interface{} {

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
func formatValueAsTOML(value interface{}) string {
	if value == nil {
		return "(not set)"
	}

	// Marshal to TOML and extract just the value part
	// We create a temporary map to marshal
	temp := map[string]interface{}{"x": value}
	bytes, err := toml.Marshal(temp)
	if err != nil {
		// Fallback to basic formatting if marshaling fails
		return fmt.Sprintf("%v", value)
	}

	// Extract the value part (after "x = ")
	str := strings.TrimSpace(string(bytes))
	if strings.HasPrefix(str, "x = ") {
		return strings.TrimPrefix(str, "x = ")
	}

	// If we have a multiline value, it starts with "x = ["
	if strings.HasPrefix(str, "x = [") {
		return strings.TrimPrefix(str, "x = ")
	}

	// Fallback
	return fmt.Sprintf("%v", value)
}

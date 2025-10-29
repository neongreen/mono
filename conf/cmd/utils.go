package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/neongreen/mono/conf/pkg/tools"
	claudetool "github.com/neongreen/mono/conf/pkg/tools/claude"
	jjtool "github.com/neongreen/mono/conf/pkg/tools/jj"
	misetool "github.com/neongreen/mono/conf/pkg/tools/mise"
	starshiptool "github.com/neongreen/mono/conf/pkg/tools/starship"
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

// getTargetConfigValues reads all values from a tool's target config file
func getTargetConfigValues(toolName string) (map[string]interface{}, error) {
	switch toolName {
	case "jj":
		jjTool, err := jjtool.NewJJTool()
		if err != nil {
			return nil, err
		}
		return jjTool.GetAllValues()
	case "claude":
		claudeTool, err := claudetool.NewClaudeTool()
		if err != nil {
			return nil, err
		}
		return claudeTool.GetAllValues()
	case "mise":
		miseTool, err := misetool.NewMiseTool()
		if err != nil {
			return nil, err
		}
		return miseTool.GetAllValues()
	case "starship":
		starshipTool, err := starshiptool.NewStarshipTool()
		if err != nil {
			return nil, err
		}
		return starshipTool.GetAllValues()
	default:
		return nil, fmt.Errorf("unsupported tool: %s", toolName)
	}
}

// readFileContentSafe reads file content, returning empty string if file doesn't exist
func readFileContentSafe(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(content), nil
}

// applyToolValue applies a single configuration value to a tool
func applyToolValue(toolName, path string, value interface{}) error {
	return tools.ApplyToolValue(toolName, path, value)
}

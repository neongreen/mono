package tools

import (
	"fmt"

	claudetool "github.com/neongreen/mono/conf/pkg/tools/claude"
	jjtool "github.com/neongreen/mono/conf/pkg/tools/jj"
	misetool "github.com/neongreen/mono/conf/pkg/tools/mise"
	starshiptool "github.com/neongreen/mono/conf/pkg/tools/starship"
)

// Tool interface represents a configuration tool
type Tool interface {
	SetConfig(path string, value any) error
	GetConfig(path string) (any, error)
	// SetAllValues sets multiple configuration values from a nested map structure
	// This is more efficient than flattening and setting individual paths
	SetAllValues(values map[string]any) error
}

// ToolFactory is a function that creates a new tool instance
type ToolFactory func() (Tool, error)

// toolRegistry maps tool names to their factory functions
var toolRegistry = map[string]ToolFactory{
	"jj": func() (Tool, error) {
		return jjtool.NewJJTool()
	},
	"mise": func() (Tool, error) {
		return misetool.NewMiseTool()
	},
	"starship": func() (Tool, error) {
		return starshiptool.NewStarshipTool()
	},
	"claude": func() (Tool, error) {
		return claudetool.NewClaudeTool()
	},
}

// GetTool returns a tool instance by name
func GetTool(toolName string) (Tool, error) {
	factory, exists := toolRegistry[toolName]
	if !exists {
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}

	return factory()
}

// GetSupportedTools returns a list of all supported tool names
func GetSupportedTools() []string {
	var tools []string
	for name := range toolRegistry {
		tools = append(tools, name)
	}
	return tools
}

// ApplyToolValue applies a single configuration value to a tool using the registry
func ApplyToolValue(toolName, path string, value any) error {
	tool, err := GetTool(toolName)
	if err != nil {
		return err
	}

	return tool.SetConfig(path, value)
}

// GetActualValue gets the actual value from a tool config file using the registry
func GetActualValue(toolName, path string) (any, error) {
	tool, err := GetTool(toolName)
	if err != nil {
		return nil, err
	}

	return tool.GetConfig(path)
}

// ApplyAllToolValues applies a nested map of configuration values to a tool
// This is more efficient than flattening and applying individual paths
func ApplyAllToolValues(toolName string, values map[string]any) error {
	tool, err := GetTool(toolName)
	if err != nil {
		return err
	}

	return tool.SetAllValues(values)
}

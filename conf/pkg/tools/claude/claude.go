package claude

import (
	"fmt"
	"strings"

	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/neongreen/mono/conf/pkg/editors"
	"github.com/neongreen/mono/conf/pkg/schemas"
)

// ClaudeTool implements Claude Code configuration management
type ClaudeTool struct {
	configPath string
	editor     *editors.JSONEditor
	parser     *schemas.ClaudeSchemaParser
	dryRun     bool
}

// NewClaudeTool creates a new Claude tool instance
func NewClaudeTool() (*ClaudeTool, error) {
	return NewClaudeToolWithDryRun(false)
}

// NewClaudeToolWithDryRun creates a new Claude tool instance with dry-run mode
func NewClaudeToolWithDryRun(dryRun bool) (*ClaudeTool, error) {
	// Load conf configuration to get Claude config path
	conf, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load conf configuration: %w", err)
	}

	claudeConfig, exists := conf.GetTool("claude")
	if !exists {
		return nil, fmt.Errorf("claude tool not configured in conf")
	}

	// Create JSON editor for Claude config file
	editor := editors.NewJSONEditorWithDryRun(claudeConfig.ConfigPath, dryRun)

	// Create Claude schema parser
	parser, err := schemas.NewClaudeSchemaParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create Claude schema parser: %w", err)
	}

	return &ClaudeTool{
		configPath: claudeConfig.ConfigPath,
		editor:     editor,
		parser:     parser,
		dryRun:     dryRun,
	}, nil
}

// SetDryRun enables or disables dry-run mode
func (c *ClaudeTool) SetDryRun(dryRun bool) {
	c.dryRun = dryRun
	c.editor.SetDryRun(dryRun)
}

// SetConfig sets a configuration value using dotted path notation
func (c *ClaudeTool) SetConfig(path string, value any) error {
	// Validate the path exists in schema
	if !c.parser.ValidatePath(path) {
		return c.createInvalidPathError(path)
	}

	// Set the value using the JSON editor
	if err := c.editor.SetValue(path, value); err != nil {
		return fmt.Errorf("failed to set Claude config %s: %w", path, err)
	}

	return nil
}

// GetConfig retrieves a configuration value using dotted path notation
func (c *ClaudeTool) GetConfig(path string) (any, error) {
	// Validate the path exists in schema
	if !c.parser.ValidatePath(path) {
		return nil, c.createInvalidPathError(path)
	}

	// Get the value using the JSON editor
	value, err := c.editor.GetValue(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get Claude config %s: %w", path, err)
	}

	return value, nil
}

// UnsetConfig removes a configuration value using dotted path notation
func (c *ClaudeTool) UnsetConfig(path string) error {
	// Validate the path exists in schema
	if !c.parser.ValidatePath(path) {
		return fmt.Errorf("invalid configuration path: %s", path)
	}

	// Unset the value using the JSON editor
	if err := c.editor.UnsetValue(path); err != nil {
		return fmt.Errorf("failed to unset Claude config %s: %w", path, err)
	}

	return nil
}

// PreviewSetConfig shows what setting a config value would do without doing it
func (c *ClaudeTool) PreviewSetConfig(path string, value any) (string, error) {
	// Validate the path exists in schema
	if !c.parser.ValidatePath(path) {
		return "", fmt.Errorf("invalid configuration path: %s", path)
	}

	return c.editor.PreviewSetValue(path, value)
}

// PreviewUnsetConfig shows what unsetting a config value would do without doing it
func (c *ClaudeTool) PreviewUnsetConfig(path string) (string, error) {
	// Validate the path exists in schema
	if !c.parser.ValidatePath(path) {
		return "", fmt.Errorf("invalid configuration path: %s", path)
	}

	return c.editor.PreviewUnsetValue(path)
}

// GetConfigPath returns the path to the Claude configuration file
func (c *ClaudeTool) GetConfigPath() string {
	return c.configPath
}

// IsDryRun returns whether dry-run mode is enabled
func (c *ClaudeTool) IsDryRun() bool {
	return c.dryRun
}

// GetAllValues returns all configuration values from the Claude config file as a nested map
func (c *ClaudeTool) GetAllValues() (map[string]any, error) {
	return c.editor.GetAllValues()
}

// SetAllValues sets multiple configuration values from a nested map structure
func (c *ClaudeTool) SetAllValues(values map[string]any) error {
	if c.dryRun {
		fmt.Println("DRY RUN: Would set all values")
		return nil
	}

	return c.editor.SetAllValues(values)
}

// createInvalidPathError creates a helpful error message for invalid configuration paths
func (c *ClaudeTool) createInvalidPathError(path string) error {
	// Get all valid paths from schema
	allPaths := c.parser.GetAllPaths()

	// Find similar paths (simple string matching for now)
	var suggestions []string
	for _, validPath := range allPaths {
		if containsSubstring(validPath, path) || containsSubstring(path, validPath) {
			suggestions = append(suggestions, validPath)
			if len(suggestions) >= 3 { // Limit suggestions
				break
			}
		}
	}

	var errorMsg strings.Builder
	errorMsg.WriteString(fmt.Sprintf("invalid configuration path: %s", path))

	if len(suggestions) > 0 {
		errorMsg.WriteString("\n\nDid you mean one of these?")
		for _, suggestion := range suggestions {
			errorMsg.WriteString(fmt.Sprintf("\n  - %s", suggestion))
		}
	} else {
		errorMsg.WriteString("\n\nUse 'conf claude list' to see available configuration options")
	}

	return fmt.Errorf("%s", errorMsg.String())
}

// containsSubstring checks if s contains substr (case-insensitive)
func containsSubstring(s, substr string) bool {
	if len(substr) < 3 { // Avoid too short matches
		return false
	}
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) && s[:len(substr)] == substr) ||
			(len(s) > len(substr) && s[len(s)-len(substr):] == substr))
}

// ListAllSettings returns comprehensive information about all Claude settings from schema
func (c *ClaudeTool) ListAllSettings() ([]schemas.SettingInfo, error) {
	// Get all settings from schema
	schemaSettings := c.parser.GetAllSettingsWithInfo()

	// Enhance with current values
	for i := range schemaSettings {
		currentValue, err := c.editor.GetValue(schemaSettings[i].Path)
		if err == nil && currentValue != nil {
			schemaSettings[i].CurrentValue = currentValue
			schemaSettings[i].IsSet = true
		} else {
			schemaSettings[i].IsSet = false
		}
	}

	return schemaSettings, nil
}

// ListCommonSettings returns a list of commonly used Claude Code settings with descriptions
func (c *ClaudeTool) ListCommonSettings() []CommonSetting {
	return []CommonSetting{
		{
			Path:        "model",
			Description: "Claude model to use (deprecated: use env.ANTHROPIC_MODEL instead)",
			Type:        "string",
			Example:     "sonnet",
		},
		{
			Path:        "alwaysThinkingEnabled",
			Description: "Always use extended thinking mode",
			Type:        "boolean",
			Example:     "true",
		},
		{
			Path:        "outputStyle",
			Description: "Output formatting style",
			Type:        "string",
			Example:     "markdown",
		},
		{
			Path:        "apiKeyHelper",
			Description: "Command to retrieve API key dynamically",
			Type:        "string",
			Example:     "/path/to/get-api-key.sh",
		},
		{
			Path:        "spinnerTipsEnabled",
			Description: "Show helpful tips in the spinner",
			Type:        "boolean",
			Example:     "true",
		},
	}
}

// CommonSetting represents a commonly used configuration setting
type CommonSetting struct {
	Path        string
	Description string
	Type        string
	Example     string
}

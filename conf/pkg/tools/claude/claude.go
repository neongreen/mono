package claude

import (
	"fmt"

	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/neongreen/mono/conf/pkg/editors"
)

// ClaudeTool implements Claude Code configuration management
type ClaudeTool struct {
	configPath string
	editor     *editors.JSONEditor
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

	return &ClaudeTool{
		configPath: claudeConfig.ConfigPath,
		editor:     editor,
		dryRun:     dryRun,
	}, nil
}

// SetDryRun enables or disables dry-run mode
func (c *ClaudeTool) SetDryRun(dryRun bool) {
	c.dryRun = dryRun
	c.editor.SetDryRun(dryRun)
}

// SetConfig sets a configuration value using dotted path notation
func (c *ClaudeTool) SetConfig(path string, value interface{}) error {
	// Set the value using the JSON editor
	if err := c.editor.SetValue(path, value); err != nil {
		return fmt.Errorf("failed to set Claude config %s: %w", path, err)
	}

	return nil
}

// GetConfig retrieves a configuration value using dotted path notation
func (c *ClaudeTool) GetConfig(path string) (interface{}, error) {
	// Get the value using the JSON editor
	value, err := c.editor.GetValue(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get Claude config %s: %w", path, err)
	}

	return value, nil
}

// UnsetConfig removes a configuration value using dotted path notation
func (c *ClaudeTool) UnsetConfig(path string) error {
	// Unset the value using the JSON editor
	if err := c.editor.UnsetValue(path); err != nil {
		return fmt.Errorf("failed to unset Claude config %s: %w", path, err)
	}

	return nil
}

// PreviewSetConfig shows what setting a config value would do without doing it
func (c *ClaudeTool) PreviewSetConfig(path string, value interface{}) (string, error) {
	return c.editor.PreviewSetValue(path, value)
}

// PreviewUnsetConfig shows what unsetting a config value would do without doing it
func (c *ClaudeTool) PreviewUnsetConfig(path string) (string, error) {
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
func (c *ClaudeTool) GetAllValues() (map[string]interface{}, error) {
	return c.editor.GetAllValues()
}

// SetAllValues sets multiple configuration values from a nested map structure
func (c *ClaudeTool) SetAllValues(values map[string]interface{}) error {
	if c.dryRun {
		fmt.Println("DRY RUN: Would set all values")
		return nil
	}

	return c.editor.SetAllValues(values)
}

// ListCommonSettings returns a list of commonly used Claude Code settings with descriptions
func (c *ClaudeTool) ListCommonSettings() []CommonSetting {
	return []CommonSetting{
		{
			Path:        "api.key",
			Description: "API key for Claude Code",
			Type:        "string",
			Example:     "sk-ant-...",
		},
		{
			Path:        "api.url",
			Description: "API endpoint URL",
			Type:        "string",
			Example:     "https://api.anthropic.com",
		},
		{
			Path:        "model",
			Description: "Default model to use",
			Type:        "string",
			Example:     "claude-3-5-sonnet-20241022",
		},
		{
			Path:        "max_tokens",
			Description: "Maximum number of tokens in response",
			Type:        "number",
			Example:     "4096",
		},
		{
			Path:        "temperature",
			Description: "Sampling temperature (0-1)",
			Type:        "number",
			Example:     "0.7",
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

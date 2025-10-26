package jj

import (
	"fmt"

	"conf/pkg/config"
	"conf/pkg/editors"
	"conf/pkg/schemas"
)

// JJTool implements jj configuration management
type JJTool struct {
	configPath string
	editor     *editors.TOMLEditor
	parser     *schemas.JJSchemaParser
	dryRun     bool
}

// NewJJTool creates a new jj tool instance
func NewJJTool() (*JJTool, error) {
	return NewJJToolWithDryRun(false)
}

// NewJJToolWithDryRun creates a new jj tool instance with dry-run mode
func NewJJToolWithDryRun(dryRun bool) (*JJTool, error) {
	// Load conf configuration to get jj config path
	conf, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load conf configuration: %w", err)
	}

	jjConfig, exists := conf.GetTool("jj")
	if !exists {
		return nil, fmt.Errorf("jj tool not configured in conf")
	}

	// Create TOML editor for jj config file
	editor := editors.NewTOMLEditorWithDryRun(jjConfig.ConfigPath, dryRun)

	// Create jj schema parser
	parser, err := schemas.NewJJSchemaParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create jj schema parser: %w", err)
	}

	return &JJTool{
		configPath: jjConfig.ConfigPath,
		editor:     editor,
		parser:     parser,
		dryRun:     dryRun,
	}, nil
}

// SetDryRun enables or disables dry-run mode
func (j *JJTool) SetDryRun(dryRun bool) {
	j.dryRun = dryRun
	j.editor.SetDryRun(dryRun)
}

// SetConfig sets a configuration value using dotted path notation
func (j *JJTool) SetConfig(path string, value interface{}) error {
	// Validate the path exists in schema
	if !j.parser.ValidatePath(path) {
		return j.createInvalidPathError(path)
	}

	// Set the value using the TOML editor
	if err := j.editor.SetValue(path, value); err != nil {
		return fmt.Errorf("failed to set jj config %s: %w", path, err)
	}

	return nil
}

// GetConfig retrieves a configuration value using dotted path notation
func (j *JJTool) GetConfig(path string) (interface{}, error) {
	// Validate the path exists in schema
	if !j.parser.ValidatePath(path) {
		return nil, j.createInvalidPathError(path)
	}

	// Get the value using the TOML editor
	value, err := j.editor.GetValue(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get jj config %s: %w", path, err)
	}

	return value, nil
}

// UnsetConfig removes a configuration value using dotted path notation
func (j *JJTool) UnsetConfig(path string) error {
	// Validate the path exists in schema
	if !j.parser.ValidatePath(path) {
		return fmt.Errorf("invalid configuration path: %s", path)
	}

	// Unset the value using the TOML editor
	if err := j.editor.UnsetValue(path); err != nil {
		return fmt.Errorf("failed to unset jj config %s: %w", path, err)
	}

	return nil
}

// PreviewSetConfig shows what setting a config value would do without doing it
func (j *JJTool) PreviewSetConfig(path string, value interface{}) (string, error) {
	// Validate the path exists in schema
	if !j.parser.ValidatePath(path) {
		return "", fmt.Errorf("invalid configuration path: %s", path)
	}

	return j.editor.PreviewSetValue(path, value)
}

// PreviewUnsetConfig shows what unsetting a config value would do without doing it
func (j *JJTool) PreviewUnsetConfig(path string) (string, error) {
	// Validate the path exists in schema
	if !j.parser.ValidatePath(path) {
		return "", fmt.Errorf("invalid configuration path: %s", path)
	}

	return j.editor.PreviewUnsetValue(path)
}

// GetConfigPath returns the path to the jj configuration file
func (j *JJTool) GetConfigPath() string {
	return j.configPath
}

// IsDryRun returns whether dry-run mode is enabled
func (j *JJTool) IsDryRun() bool {
	return j.dryRun
}

// ListCommonSettings returns a list of commonly used jj settings with descriptions
func (j *JJTool) ListCommonSettings() []CommonSetting {
	return []CommonSetting{
		{
			Path:        "user.name",
			Description: "Full name of the user, used in commits",
			Type:        "string",
			Example:     "Alice Smith",
		},
		{
			Path:        "user.email",
			Description: "User's email address, used in commits",
			Type:        "string",
			Example:     "alice@example.com",
		},
		{
			Path:        "ui.default-command",
			Description: "Default command to run when no command is specified",
			Type:        "string",
			Example:     "log",
		},
		{
			Path:        "snapshot.max-new-file-size",
			Description: "Maximum size of new files to automatically track",
			Type:        "integer",
			Example:     "1048576",
		},
	}
}

// ListAllSettings returns comprehensive information about all jj settings from schema
func (j *JJTool) ListAllSettings() ([]schemas.SettingInfo, error) {
	// Get all settings from schema
	schemaSettings := j.parser.GetAllSettingsWithInfo()

	// Enhance with current values
	for i := range schemaSettings {
		currentValue, err := j.editor.GetValue(schemaSettings[i].Path)
		if err == nil && currentValue != nil {
			schemaSettings[i].CurrentValue = currentValue
			schemaSettings[i].IsSet = true
		} else {
			schemaSettings[i].IsSet = false
		}
	}

	return schemaSettings, nil
}

// createInvalidPathError creates a helpful error message for invalid configuration paths
func (j *JJTool) createInvalidPathError(path string) error {
	// Get all valid paths from schema
	allPaths := j.parser.GetAllPaths()

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

	errorMsg := fmt.Sprintf("invalid configuration path: %s", path)

	if len(suggestions) > 0 {
		errorMsg += "\n\nDid you mean one of these?"
		for _, suggestion := range suggestions {
			errorMsg += fmt.Sprintf("\n  - %s", suggestion)
		}
	} else {
		errorMsg += "\n\nUse 'conf jj list' to see available configuration options"
	}

	return fmt.Errorf("%s", errorMsg)
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

// CommonSetting represents a commonly used configuration setting
type CommonSetting struct {
	Path        string
	Description string
	Type        string
	Example     string
}

// GetAllValues returns all configuration values from the jj config file as a nested map
func (j *JJTool) GetAllValues() (map[string]interface{}, error) {
	return j.editor.GetAllValues()
}

// SetAllValues sets multiple configuration values from a nested map structure
// This is more efficient than setting individual paths as it avoids the need
// to flatten/unflatten the structure and parse quoted keys
func (j *JJTool) SetAllValues(values map[string]interface{}) error {
	if j.dryRun {
		fmt.Println("DRY RUN: Would set all values")
		return nil
	}

	// We don't validate individual paths here because we're working with
	// the native nested structure. The TOML library will handle the writing.
	// Schema validation would need to be done at a different level if needed.

	return j.editor.SetAllValues(values)
}

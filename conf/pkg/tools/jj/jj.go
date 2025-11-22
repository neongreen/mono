package jj

import (
	"fmt"
	"strings"

	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/neongreen/mono/conf/pkg/editors"
	"github.com/neongreen/mono/conf/pkg/schemas"
	"github.com/neongreen/mono/lib/configschema"
)

// JJTool implements jj configuration management
type JJTool struct {
	configPath string
	editor     *editors.TOMLEditor
	parser     *schemas.JJSchemaParser
	dryRun     bool
	force      bool
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
		force:      false,
	}, nil
}

// SetDryRun enables or disables dry-run mode
func (j *JJTool) SetDryRun(dryRun bool) {
	j.dryRun = dryRun
	j.editor.SetDryRun(dryRun)
}

// SetForce enables or disables schema validation bypass
func (j *JJTool) SetForce(force bool) {
	j.force = force
}

// SetConfig sets a configuration value using dotted path notation
func (j *JJTool) SetConfig(path string, value any) error {
	// Validate the path exists in schema
	if !j.parser.ValidatePath(path) {
		return j.createInvalidPathError(path)
	}

	if !j.force {
		if err := j.parser.ValidateValue(path, value); err != nil {
			return fmt.Errorf("invalid value for %s: %w", path, err)
		}
	}

	// Set the value using the TOML editor
	if err := j.editor.SetValue(path, value); err != nil {
		return fmt.Errorf("failed to set jj config %s: %w", path, err)
	}

	return nil
}

// GetConfig retrieves a configuration value using dotted path notation
func (j *JJTool) GetConfig(path string) (any, error) {
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
func (j *JJTool) PreviewSetConfig(path string, value any) (string, error) {
	// Validate the path exists in schema
	if !j.parser.ValidatePath(path) {
		return "", fmt.Errorf("invalid configuration path: %s", path)
	}

	if !j.force {
		if err := j.parser.ValidateValue(path, value); err != nil {
			return "", fmt.Errorf("invalid value for %s: %w", path, err)
		}
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

// IsForce returns whether schema validation is bypassed
func (j *JJTool) IsForce() bool {
	return j.force
}

// ListAllSettings returns comprehensive information about all jj settings from schema
func (j *JJTool) ListAllSettings() ([]configschema.SettingInfo, error) {
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

	var errorMsg strings.Builder
	errorMsg.WriteString(fmt.Sprintf("invalid configuration path: %s", path))

	if len(suggestions) > 0 {
		errorMsg.WriteString("\n\nDid you mean one of these?")
		for _, suggestion := range suggestions {
			errorMsg.WriteString(fmt.Sprintf("\n  - %s", suggestion))
		}
	} else {
		errorMsg.WriteString("\n\nUse 'conf jj list' to see available configuration options")
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

// GetAllValues returns all configuration values from the jj config file as a nested map
func (j *JJTool) GetAllValues() (map[string]any, error) {
	return j.editor.GetAllValues()
}

// SetAllValues sets multiple configuration values from a nested map structure
// This is more efficient than setting individual paths as it avoids the need
// to flatten/unflatten the structure and parse quoted keys
func (j *JJTool) SetAllValues(values map[string]any) error {
	if !j.force {
		if err := j.parser.ValidateDocument(values); err != nil {
			return fmt.Errorf("invalid jj configuration: %w", err)
		}
	}

	if j.dryRun {
		fmt.Println("DRY RUN: Would set all values")
		return nil
	}

	// We don't validate individual paths here because we're working with
	// the native nested structure. The TOML library will handle the writing.
	// Schema validation would need to be done at a different level if needed.

	return j.editor.SetAllValues(values)
}

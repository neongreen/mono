package starship

import (
	"fmt"

	"conf/pkg/config"
	"conf/pkg/editors"
)

// StarshipTool implements starship configuration management
type StarshipTool struct {
	configPath string
	editor     *editors.TOMLEditor
	dryRun     bool
}

// NewStarshipTool creates a new starship tool instance
func NewStarshipTool() (*StarshipTool, error) {
	return NewStarshipToolWithDryRun(false)
}

// NewStarshipToolWithDryRun creates a new starship tool instance with dry-run mode
func NewStarshipToolWithDryRun(dryRun bool) (*StarshipTool, error) {
	// Load conf configuration to get starship config path
	conf, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load conf configuration: %w", err)
	}

	starshipConfig, exists := conf.GetTool("starship")
	if !exists {
		return nil, fmt.Errorf("starship tool not configured in conf")
	}

	// Create TOML editor for starship config file
	editor := editors.NewTOMLEditorWithDryRun(starshipConfig.ConfigPath, dryRun)

	return &StarshipTool{
		configPath: starshipConfig.ConfigPath,
		editor:     editor,
		dryRun:     dryRun,
	}, nil
}

// SetDryRun enables or disables dry-run mode
func (s *StarshipTool) SetDryRun(dryRun bool) {
	s.dryRun = dryRun
	s.editor.SetDryRun(dryRun)
}

// SetConfig sets a configuration value using dotted path notation
func (s *StarshipTool) SetConfig(path string, value interface{}) error {
	// Validate the path is reasonable (basic validation)
	if !s.isValidPath(path) {
		return s.createInvalidPathError(path)
	}

	// Set the value using the TOML editor
	if err := s.editor.SetValue(path, value); err != nil {
		return fmt.Errorf("failed to set starship config %s: %w", path, err)
	}

	return nil
}

// GetConfig retrieves a configuration value using dotted path notation
func (s *StarshipTool) GetConfig(path string) (interface{}, error) {
	// Validate the path is reasonable (basic validation)
	if !s.isValidPath(path) {
		return nil, s.createInvalidPathError(path)
	}

	// Get the value using the TOML editor
	value, err := s.editor.GetValue(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get starship config %s: %w", path, err)
	}

	return value, nil
}

// UnsetConfig removes a configuration value using dotted path notation
func (s *StarshipTool) UnsetConfig(path string) error {
	// Validate the path is reasonable (basic validation)
	if !s.isValidPath(path) {
		return s.createInvalidPathError(path)
	}

	// Unset the value using the TOML editor
	if err := s.editor.UnsetValue(path); err != nil {
		return fmt.Errorf("failed to unset starship config %s: %w", path, err)
	}

	return nil
}

// PreviewSetConfig shows what setting a config value would do without doing it
func (s *StarshipTool) PreviewSetConfig(path string, value interface{}) (string, error) {
	// Validate the path is reasonable (basic validation)
	if !s.isValidPath(path) {
		return "", s.createInvalidPathError(path)
	}

	return s.editor.PreviewSetValue(path, value)
}

// GetConfigPath returns the path to the starship configuration file
func (s *StarshipTool) GetConfigPath() string {
	return s.configPath
}

// IsDryRun returns whether dry-run mode is enabled
func (s *StarshipTool) IsDryRun() bool {
	return s.dryRun
}

// SetAllValues sets multiple configuration values from a nested map structure
// This is more efficient than setting individual paths as it avoids the need
// to flatten/unflatten the structure and parse quoted keys
func (s *StarshipTool) SetAllValues(values map[string]interface{}) error {
	if s.dryRun {
		fmt.Println("DRY RUN: Would set all values")
		return nil
	}

	// We don't validate individual paths here because we're working with
	// the native nested structure. The TOML library will handle the writing.
	return s.editor.SetAllValues(values)
}

// isValidPath performs basic validation on starship config paths
func (s *StarshipTool) isValidPath(path string) bool {
	// Basic validation - just check it's not empty and has reasonable format
	if path == "" {
		return false
	}

	// Allow any dotted path for starship (it's very flexible)
	return true
}

// createInvalidPathError creates a helpful error message for invalid configuration paths
func (s *StarshipTool) createInvalidPathError(path string) error {
	errorMsg := fmt.Sprintf("invalid configuration path: %s", path)
	errorMsg += "\n\nUse 'conf starship list' to see available configuration options"
	return fmt.Errorf("%s", errorMsg)
}

// ListCommonSettings returns a list of commonly used starship settings with descriptions
func (s *StarshipTool) ListCommonSettings() []CommonSetting {
	return []CommonSetting{
		{
			Path:        "format",
			Description: "Custom format string for the prompt",
			Type:        "string",
			Example:     "$all$character",
		},
		{
			Path:        "add_newline",
			Description: "Add a blank line before the prompt",
			Type:        "boolean",
			Example:     "true",
		},
		{
			Path:        "command_timeout",
			Description: "Timeout for commands run by starship (in milliseconds)",
			Type:        "integer",
			Example:     "500",
		},
		{
			Path:        "scan_timeout",
			Description: "Timeout for scanning files and directories (in milliseconds)",
			Type:        "integer",
			Example:     "30",
		},
		{
			Path:        "character.success_symbol",
			Description: "Symbol shown when the last command succeeded",
			Type:        "string",
			Example:     "[➜](bold green)",
		},
		{
			Path:        "character.error_symbol",
			Description: "Symbol shown when the last command failed",
			Type:        "string",
			Example:     "[➜](bold red)",
		},
		{
			Path:        "directory.truncation_length",
			Description: "Number of parent directories to show",
			Type:        "integer",
			Example:     "3",
		},
		{
			Path:        "git_branch.format",
			Description: "Format string for git branch display",
			Type:        "string",
			Example:     "on [$symbol$branch]($style) ",
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

// GetAllValues returns all configuration values from the starship config file as a nested map
func (s *StarshipTool) GetAllValues() (map[string]interface{}, error) {
	return s.editor.GetAllValues()
}

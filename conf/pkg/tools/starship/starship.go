package starship

import (
	"fmt"

	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/neongreen/mono/conf/pkg/editors"
	"github.com/neongreen/mono/conf/pkg/schemas"
)

// StarshipTool implements starship configuration management
type StarshipTool struct {
	configPath string
	editor     *editors.TOMLEditor
	parser     *schemas.StarshipSchemaParser
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

	// Create starship schema parser
	parser, err := schemas.NewStarshipSchemaParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create starship schema parser: %w", err)
	}

	return &StarshipTool{
		configPath: starshipConfig.ConfigPath,
		editor:     editor,
		parser:     parser,
		dryRun:     dryRun,
	}, nil
}

// SetDryRun enables or disables dry-run mode
func (s *StarshipTool) SetDryRun(dryRun bool) {
	s.dryRun = dryRun
	s.editor.SetDryRun(dryRun)
}

// SetConfig sets a configuration value using dotted path notation
func (s *StarshipTool) SetConfig(path string, value any) error {
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
func (s *StarshipTool) GetConfig(path string) (any, error) {
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
func (s *StarshipTool) PreviewSetConfig(path string, value any) (string, error) {
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
func (s *StarshipTool) SetAllValues(values map[string]any) error {
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

// GetAllValues returns all configuration values from the starship config file as a nested map
func (s *StarshipTool) GetAllValues() (map[string]any, error) {
	return s.editor.GetAllValues()
}

// ListAllSettings returns comprehensive information about all starship settings from schema
func (s *StarshipTool) ListAllSettings() ([]schemas.SettingInfo, error) {
	// Get all settings from schema
	schemaSettings := s.parser.GetAllSettingsWithInfo()

	// Read config file once to avoid re-parsing for every setting
	allValues, err := s.editor.GetAllValues()
	if err != nil {
		// If we can't read the config, return settings without current values
		return schemaSettings, nil
	}

	// Enhance with current values by looking up in the in-memory map
	for i := range schemaSettings {
		currentValue := lookupValueByPath(allValues, schemaSettings[i].Path)
		if currentValue != nil {
			schemaSettings[i].CurrentValue = currentValue
			schemaSettings[i].IsSet = true
		} else {
			schemaSettings[i].IsSet = false
		}
	}

	return schemaSettings, nil
}

// lookupValueByPath traverses a nested map using a dotted path to find a value
func lookupValueByPath(data map[string]any, path string) any {
	if path == "" {
		return nil
	}

	parts := splitPath(path)
	current := data

	for i, part := range parts {
		value, exists := current[part]
		if !exists {
			return nil
		}

		// If this is the last part, return the value
		if i == len(parts)-1 {
			return value
		}

		// Otherwise, expect a nested map and continue traversing
		nestedMap, ok := value.(map[string]any)
		if !ok {
			return nil
		}
		current = nestedMap
	}

	return nil
}

// splitPath splits a dotted path into parts, handling quoted segments
func splitPath(path string) []string {
	if path == "" {
		return nil
	}

	var parts []string
	var current []rune
	inQuotes := false

	for _, ch := range path {
		switch ch {
		case '"':
			inQuotes = !inQuotes
		case '.':
			if !inQuotes {
				if len(current) > 0 {
					parts = append(parts, string(current))
					current = nil
				}
			} else {
				current = append(current, ch)
			}
		default:
			current = append(current, ch)
		}
	}

	if len(current) > 0 {
		parts = append(parts, string(current))
	}

	return parts
}

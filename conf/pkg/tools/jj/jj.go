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
}

// NewJJTool creates a new jj tool instance
func NewJJTool() (*JJTool, error) {
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
	editor := editors.NewTOMLEditor(jjConfig.ConfigPath)

	// Create jj schema parser
	parser, err := schemas.NewJJSchemaParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create jj schema parser: %w", err)
	}

	return &JJTool{
		configPath: jjConfig.ConfigPath,
		editor:     editor,
		parser:     parser,
	}, nil
}

// SetConfig sets a configuration value using dotted path notation
func (j *JJTool) SetConfig(path string, value interface{}) error {
	// Validate the path exists in schema
	if !j.parser.ValidatePath(path) {
		return fmt.Errorf("invalid configuration path: %s", path)
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
		return nil, fmt.Errorf("invalid configuration path: %s", path)
	}

	// Get the value using the TOML editor
	value, err := j.editor.GetValue(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get jj config %s: %w", path, err)
	}

	return value, nil
}

// GetConfigPath returns the path to the jj configuration file
func (j *JJTool) GetConfigPath() string {
	return j.configPath
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

// CommonSetting represents a commonly used configuration setting
type CommonSetting struct {
	Path        string
	Description string
	Type        string
	Example     string
}

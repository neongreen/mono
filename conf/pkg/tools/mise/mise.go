package mise

import (
	"fmt"
	"slices"
	"strings"

	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/neongreen/mono/conf/pkg/editors"
	"github.com/neongreen/mono/conf/pkg/schemas"
)

// MiseTool implements mise configuration management
type MiseTool struct {
	configPath string
	editor     *editors.TOMLEditor
	schema     *schemas.MiseSchema
	dryRun     bool
}

// NewMiseTool creates a new mise tool instance
func NewMiseTool() (*MiseTool, error) {
	return NewMiseToolWithDryRun(false)
}

// NewMiseToolWithDryRun creates a new mise tool instance with dry-run mode
func NewMiseToolWithDryRun(dryRun bool) (*MiseTool, error) {
	// Load conf configuration to get mise config path
	conf, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load conf configuration: %w", err)
	}

	miseConfig, exists := conf.GetTool("mise")
	if !exists {
		return nil, fmt.Errorf("mise tool not configured in conf")
	}

	// Create TOML editor for mise config file
	editor := editors.NewTOMLEditorWithDryRun(miseConfig.ConfigPath, dryRun)

	// Create mise schema parser
	schema, err := schemas.LoadMiseSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to load mise schema: %w", err)
	}

	return &MiseTool{
		configPath: miseConfig.ConfigPath,
		editor:     editor,
		schema:     schema,
		dryRun:     dryRun,
	}, nil
}

// SetDryRun enables or disables dry-run mode
func (m *MiseTool) SetDryRun(dryRun bool) {
	m.dryRun = dryRun
	m.editor.SetDryRun(dryRun)
}

// SetConfig sets a configuration value using dotted path notation
func (m *MiseTool) SetConfig(path string, value any) error {
	// Validate the path exists in schema
	if !m.ValidatePath(path) {
		return fmt.Errorf("invalid configuration path: %s", path)
	}

	// Set the value using the TOML editor
	if err := m.editor.SetValue(path, value); err != nil {
		return fmt.Errorf("failed to set mise config %s: %w", path, err)
	}

	return nil
}

// GetConfig retrieves a configuration value using dotted path notation
func (m *MiseTool) GetConfig(path string) (any, error) {
	// Validate the path exists in schema
	if !m.ValidatePath(path) {
		return nil, fmt.Errorf("invalid configuration path: %s", path)
	}

	// Get the value using the TOML editor
	value, err := m.editor.GetValue(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get mise config %s: %w", path, err)
	}

	return value, nil
}

// UnsetConfig removes a configuration value using dotted path notation
func (m *MiseTool) UnsetConfig(path string) error {
	// Validate the path exists in schema
	if !m.ValidatePath(path) {
		return fmt.Errorf("invalid configuration path: %s", path)
	}

	// Unset the value using the TOML editor
	if err := m.editor.UnsetValue(path); err != nil {
		return fmt.Errorf("failed to unset mise config %s: %w", path, err)
	}

	return nil
}

// PreviewSetConfig shows what setting a config value would do without doing it
func (m *MiseTool) PreviewSetConfig(path string, value any) (string, error) {
	// Validate the path exists in schema
	if !m.ValidatePath(path) {
		return "", fmt.Errorf("invalid configuration path: %s", path)
	}

	return m.editor.PreviewSetValue(path, value)
}

// PreviewUnsetConfig shows what unsetting a config value would do without doing it
func (m *MiseTool) PreviewUnsetConfig(path string) (string, error) {
	// Validate the path exists in schema
	if !m.ValidatePath(path) {
		return "", fmt.Errorf("invalid configuration path: %s", path)
	}

	return m.editor.PreviewUnsetValue(path)
}

// GetCompletionOptions returns completion options for a given path
func (m *MiseTool) GetCompletionOptions(path string) []schemas.CompletionOption {
	return m.schema.GetCompletionOptions(path)
}

// ValidatePath checks if a configuration path is valid
func (m *MiseTool) ValidatePath(path string) bool {
	// For mise, we'll be more permissive since the schema is custom
	// Check if it's in our known schema fields, or allow any path for flexibility
	options := m.schema.GetCompletionOptions("")
	if path == "" {
		return true
	}

	// Check top-level paths
	parts := splitPath(path)
	if len(parts) == 0 {
		return true
	}

	topLevel := parts[0]
	for _, option := range options {
		if option.Name == topLevel {
			return true
		}
	}

	// Allow common mise patterns even if not in schema
	commonPrefixes := []string{"tools", "env", "tasks", "settings", "alias", "python", "ruby", "node"}
	return slices.Contains(commonPrefixes, topLevel)
}

// GetConfigPath returns the path to the mise configuration file
func (m *MiseTool) GetConfigPath() string {
	return m.configPath
}

// IsDryRun returns whether dry-run mode is enabled
func (m *MiseTool) IsDryRun() bool {
	return m.dryRun
}

// SetAllValues sets multiple configuration values from a nested map structure
// This is more efficient than setting individual paths as it avoids the need
// to flatten/unflatten the structure and parse quoted keys
func (m *MiseTool) SetAllValues(values map[string]any) error {
	if m.dryRun {
		fmt.Println("DRY RUN: Would set all values")
		return nil
	}

	// We don't validate individual paths here because we're working with
	// the native nested structure. The TOML library will handle the writing.
	return m.editor.SetAllValues(values)
}

// ListCommonSettings returns a list of commonly used mise settings with descriptions
func (m *MiseTool) ListCommonSettings() []CommonSetting {
	return []CommonSetting{
		{
			Path:        "settings.experimental",
			Description: "Enable experimental features in mise",
			Type:        "boolean",
			Example:     "true",
		},
		{
			Path:        "settings.verbose",
			Description: "Enable verbose output",
			Type:        "boolean",
			Example:     "false",
		},
		{
			Path:        "settings.jobs",
			Description: "Number of parallel jobs for installation",
			Type:        "integer",
			Example:     "4",
		},
		{
			Path:        "settings.legacy_version_file",
			Description: "Enable support for legacy version files",
			Type:        "boolean",
			Example:     "true",
		},
		{
			Path:        "env.NODE_ENV",
			Description: "Set Node.js environment",
			Type:        "string",
			Example:     "development",
		},
		{
			Path:        "tools.node",
			Description: "Node.js version to use",
			Type:        "string",
			Example:     "20",
		},
		{
			Path:        "tools.python",
			Description: "Python version to use",
			Type:        "string",
			Example:     "3.11",
		},
		{
			Path:        "tasks.dev.run",
			Description: "Development task command",
			Type:        "string",
			Example:     "npm run dev",
		},
		{
			Path:        "python.venv_auto_create",
			Description: "Automatically create Python virtual environments",
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

// Helper function to split dotted path
func splitPath(path string) []string {
	if path == "" {
		return []string{}
	}
	return strings.Split(path, ".")
}

// GetAllValues returns all configuration values from the mise config file as a nested map
func (m *MiseTool) GetAllValues() (map[string]any, error) {
	return m.editor.GetAllValues()
}

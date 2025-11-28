package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/creachadair/tomledit/parser"
)

// Config represents the main configuration for conf
type Config struct {
	Tools   map[string]ToolConfig   `toml:"tools"`
	Folders map[string]FolderConfig `toml:"folders,omitempty"`
	Shims   map[string]string       `toml:"shims,omitempty"`
}

// ToolConfig represents configuration for a specific tool
type ToolConfig struct {
	Name       string         `toml:"name"`
	ConfigPath string         `toml:"path"`
	SchemaPath string         `toml:"schema,omitempty"`
	Values     map[string]any `toml:"values,omitempty"`
}

// FolderConfig represents configuration for a tracked folder
type FolderConfig struct {
	Name         string   `toml:"name"`
	SourcePath   string   `toml:"source_path"`
	TrackedSince string   `toml:"tracked_since,omitempty"`
	Exclude      []string `toml:"exclude,omitempty"`
}

// DefaultConfig returns a new Config with default settings
func DefaultConfig() *Config {
	return &Config{
		Tools: map[string]ToolConfig{
			"jj": {
				Name:       "jj",
				ConfigPath: "~/.config/jj/config.toml",
				SchemaPath: "embedded://jj.json",
			},
			"mise": {
				Name:       "mise",
				ConfigPath: "~/.config/mise/config.toml",
				SchemaPath: "embedded://mise.toml",
			},
			"starship": {
				Name:       "starship",
				ConfigPath: "~/.config/starship.toml",
			},
			"claude": {
				Name:       "claude",
				ConfigPath: "~/.claude/settings.json",
				SchemaPath: "embedded://claude.json",
			},
		},
	}
}

// ConfigDir returns the directory where conf stores its configuration
func ConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "conf")
	return configDir, nil
}

// ConfigPath returns the path to conf's configuration file
func ConfigPath() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.toml"), nil
}

// ExpandPath expands tilde (~) in paths to the user's home directory
func ExpandPath(path string) (string, error) {
	if path == "" {
		return path, nil
	}

	// Only expand if path starts with ~
	if path[0] != '~' {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Handle ~/ or just ~
	if len(path) == 1 || path[1] == '/' {
		return filepath.Join(homeDir, path[1:]), nil
	}

	// Don't support ~user syntax for now
	return path, nil
}

// GetTool returns the configuration for a specific tool with expanded paths
func (c *Config) GetTool(name string) (ToolConfig, bool) {
	tool, exists := c.Tools[name]
	if !exists {
		return tool, false
	}

	// Expand tilde in config path
	if expandedPath, err := ExpandPath(tool.ConfigPath); err == nil {
		tool.ConfigPath = expandedPath
	}

	return tool, true
}

// SetTool sets the configuration for a specific tool
func (c *Config) SetTool(name string, tool ToolConfig) {
	if c.Tools == nil {
		c.Tools = make(map[string]ToolConfig)
	}
	c.Tools[name] = tool
}

// SetToolValue sets a specific configuration value for a tool
// and ensures the per-tool config file will be created on Save()
func (c *Config) SetToolValue(toolName, path string, value any) {
	if c.Tools == nil {
		c.Tools = make(map[string]ToolConfig)
	}

	tool := c.Tools[toolName]
	key, err := parser.ParseKey(path)
	if err != nil || len(key) == 0 {
		return
	}

	tool.Values = ensureMap(tool.Values)
	setNestedValue(tool.Values, key, value)
	c.Tools[toolName] = tool

	// Ensure per-tool file will be created on Save()
	configDir, err := ConfigDir()
	if err != nil {
		return
	}
	perToolPath := filepath.Join(configDir, toolName+".toml")

	// Create empty file if it doesn't exist, so Save() will write to it
	if _, err := os.Stat(perToolPath); os.IsNotExist(err) {
		os.MkdirAll(configDir, 0o755)
		os.WriteFile(perToolPath, []byte{}, 0o644)
	}
}

// MergeToolValues merges a map of values into a tool's configuration.
// Nested maps are merged recursively.
func (c *Config) MergeToolValues(toolName string, values map[string]any) {
	if c.Tools == nil {
		c.Tools = make(map[string]ToolConfig)
	}

	tool := c.Tools[toolName]
	tool.Values = mergeNestedValues(tool.Values, values)
	c.Tools[toolName] = tool

	// Ensure per-tool file will be created on Save()
	configDir, err := ConfigDir()
	if err != nil {
		return
	}
	perToolPath := filepath.Join(configDir, toolName+".toml")

	if _, err := os.Stat(perToolPath); os.IsNotExist(err) {
		os.MkdirAll(configDir, 0o755)
		os.WriteFile(perToolPath, []byte{}, 0o644)
	}
}

// GetToolValue gets a specific configuration value for a tool
func (c *Config) GetToolValue(toolName, path string) (any, bool) {
	tool, exists := c.Tools[toolName]
	if !exists || tool.Values == nil {
		return nil, false
	}

	key, err := parser.ParseKey(path)
	if err != nil || len(key) == 0 {
		return nil, false
	}

	return getNestedValue(tool.Values, key)
}

// UnsetToolValue removes a specific configuration value for a tool
func (c *Config) UnsetToolValue(toolName, path string) {
	tool, exists := c.Tools[toolName]
	if !exists || tool.Values == nil {
		return
	}

	key, err := parser.ParseKey(path)
	if err != nil || len(key) == 0 {
		return
	}

	if unsetNestedValue(tool.Values, key) {
		c.Tools[toolName] = tool
	}
}

// SetShim sets a shim command
func (c *Config) SetShim(name, command string) {
	if c.Shims == nil {
		c.Shims = make(map[string]string)
	}
	c.Shims[name] = command
}

// GetShim gets a shim command
func (c *Config) GetShim(name string) (string, bool) {
	if c.Shims == nil {
		return "", false
	}
	command, exists := c.Shims[name]
	return command, exists
}

// UnsetShim removes a shim
func (c *Config) UnsetShim(name string) {
	if c.Shims == nil {
		return
	}
	delete(c.Shims, name)
}

// GetFolder returns the configuration for a specific folder with expanded paths
func (c *Config) GetFolder(name string) (FolderConfig, bool) {
	folder, exists := c.Folders[name]
	if !exists {
		return folder, false
	}

	// Expand tilde in source path
	if expandedPath, err := ExpandPath(folder.SourcePath); err == nil {
		folder.SourcePath = expandedPath
	}

	return folder, true
}

// SetFolder sets the configuration for a specific folder
func (c *Config) SetFolder(name string, folder FolderConfig) {
	if c.Folders == nil {
		c.Folders = make(map[string]FolderConfig)
	}
	c.Folders[name] = folder
}

// RemoveFolder removes a folder from the configuration
func (c *Config) RemoveFolder(name string) {
	if c.Folders == nil {
		return
	}
	delete(c.Folders, name)
}

// FolderManifestPath returns the path to a folder's manifest file
func FolderManifestPath(configDir, folderName string) string {
	return filepath.Join(configDir, folderName+".toml")
}

// FolderCopyPath returns the path to a folder's copy in conf directory
func FolderCopyPath(configDir, folderName string) string {
	return filepath.Join(configDir, folderName)
}

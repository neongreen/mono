package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the main configuration for conf
type Config struct {
	Tools map[string]ToolConfig `toml:"tools"`
	Shims map[string]string     `toml:"shims,omitempty"`
}

// ToolConfig represents configuration for a specific tool
type ToolConfig struct {
	Name       string                 `toml:"name"`
	ConfigPath string                 `toml:"path"`
	SchemaPath string                 `toml:"schema,omitempty"`
	Values     map[string]interface{} `toml:"values,omitempty"`
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

// Load loads the configuration including per-tool config files
// Loads tool definitions and values from main config.toml, then augments with per-tool files
// Paths are kept in tilde notation (~/) for portability
func Load() (*Config, error) {
	// Start with default tool metadata as base
	defaultConfig := DefaultConfig()

	// Try to load main config.toml to get all stored tools and values
	configPath, err := ConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	var config *Config

	if _, err := os.Stat(configPath); err == nil {
		// Main config.toml exists, load it
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read main config: %w", err)
		}

		var loadedConfig Config
		if err := toml.Unmarshal(data, &loadedConfig); err != nil {
			return nil, fmt.Errorf("failed to parse main config: %w", err)
		}

		config = &loadedConfig
	} else {
		// Main config.toml doesn't exist, start with default config
		config = defaultConfig
	}

	// Ensure default tools are available (preserve their metadata if not already in config)
	for name, defaultTool := range defaultConfig.Tools {
		if _, exists := config.Tools[name]; !exists {
			config.Tools[name] = defaultTool
		}
	}

	// Load values from per-tool config files if they exist
	// This augments the values loaded from main config.toml
	// Note: paths remain in tilde notation; GetTool() expands them when needed
	if err := config.loadPerToolConfigs(); err != nil {
		return nil, fmt.Errorf("failed to load per-tool configs: %w", err)
	}

	return config, nil
}

// convertNestedToFlat converts a nested map structure to flat map with dotted keys
// Example: {"user": {"name": "John"}} -> {"user.name": "John"}
func convertNestedToFlat(nested map[string]interface{}, prefix string) map[string]interface{} {
	flat := make(map[string]interface{})

	for key, value := range nested {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		// If value is a map, recurse
		if nestedMap, ok := value.(map[string]interface{}); ok {
			// Merge the flattened results
			for k, v := range convertNestedToFlat(nestedMap, fullKey) {
				flat[k] = v
			}
		} else {
			// Leaf value, add to flat map
			flat[fullKey] = value
		}
	}

	return flat
}

// loadPerToolConfigs loads values from per-tool config files (e.g., ~/.config/conf/jj.toml)
// and merges them with the tool metadata from config.toml
func (c *Config) loadPerToolConfigs() error {
	configDir, err := ConfigDir()
	if err != nil {
		return err
	}

	for toolName, tool := range c.Tools {
		// Check if per-tool config file exists
		perToolPath := filepath.Join(configDir, toolName+".toml")
		if _, err := os.Stat(perToolPath); err != nil {
			// Per-tool file doesn't exist, keep existing values from config.toml
			continue
		}

		// Read per-tool config file
		data, err := os.ReadFile(perToolPath)
		if err != nil {
			return fmt.Errorf("failed to read per-tool config %s: %w", perToolPath, err)
		}

		// Parse per-tool config into a nested map
		var perToolNested map[string]interface{}
		if err := toml.Unmarshal(data, &perToolNested); err != nil {
			return fmt.Errorf("failed to parse per-tool config %s: %w", perToolPath, err)
		}

		// Convert nested structure to flat dotted keys
		perToolFlat := convertNestedToFlat(perToolNested, "")

		// Merge per-tool values with tool metadata
		// Per-tool file values override config.toml values
		if tool.Values == nil {
			tool.Values = make(map[string]interface{})
		}
		for k, v := range perToolFlat {
			tool.Values[k] = v
		}
		c.Tools[toolName] = tool
	}

	return nil
}

// convertDottedToNested converts a flat map with dotted keys to nested map structure
// Example: {"user.name": "John"} -> {"user": {"name": "John"}}
func convertDottedToNested(flat map[string]interface{}) map[string]interface{} {
	nested := make(map[string]interface{})

	for key, value := range flat {
		parts := strings.Split(key, ".")
		current := nested

		// Navigate/create nested structure for all parts except the last
		for i := 0; i < len(parts)-1; i++ {
			part := parts[i]
			if _, exists := current[part]; !exists {
				current[part] = make(map[string]interface{})
			}
			// Type assert to map for next iteration
			if nextMap, ok := current[part].(map[string]interface{}); ok {
				current = nextMap
			}
		}

		// Set the final value
		current[parts[len(parts)-1]] = value
	}

	return nested
}

// Save saves the configuration to per-tool files
// Only creates/updates files for tools that already have per-tool files
// Also saves the main config.toml to preserve tool metadata
func (c *Config) Save() error {
	configDir, err := ConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config dir: %w", err)
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save tools to per-tool files if they exist
	for toolName, tool := range c.Tools {
		perToolPath := filepath.Join(configDir, toolName+".toml")

		// Save to per-tool file if it exists
		if _, err := os.Stat(perToolPath); err == nil {
			// Save values if they exist
			if tool.Values != nil && len(tool.Values) > 0 {
				// Convert dotted keys to nested structure
				nested := convertDottedToNested(tool.Values)

				data, err := toml.Marshal(nested)
				if err != nil {
					return fmt.Errorf("failed to marshal %s config: %w", toolName, err)
				}

				if err := os.WriteFile(perToolPath, data, 0644); err != nil {
					return fmt.Errorf("failed to write %s config: %w", toolName, err)
				}
			}
		}
	}

	// Save main config.toml to preserve tool metadata (paths, schema paths, etc.)
	// and values for tools that don't have per-tool files
	mainConfig := &Config{
		Tools: make(map[string]ToolConfig),
		Shims: c.Shims, // Always save shims in main config
	}

	for toolName, tool := range c.Tools {
		perToolPath := filepath.Join(configDir, toolName+".toml")

		// Only preserve tool metadata in main config if per-tool file exists
		// (the values go to the per-tool file, the metadata stays in main config)
		if _, err := os.Stat(perToolPath); err == nil {
			// Copy tool config but without Values (they go to per-tool file)
			mainConfig.Tools[toolName] = ToolConfig{
				Name:       tool.Name,
				ConfigPath: tool.ConfigPath,
				SchemaPath: tool.SchemaPath,
				// Values are stored in per-tool file, not main config
			}
		} else {
			// If per-tool file doesn't exist, save everything in main config
			mainConfig.Tools[toolName] = tool
		}
	}

	// Write main config.toml
	configPath, err := ConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	data, err := toml.Marshal(mainConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal main config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write main config: %w", err)
	}

	return nil
}

// savePerToolConfigs saves values to per-tool config files if they exist
func (c *Config) savePerToolConfigs() error {
	configDir, err := ConfigDir()
	if err != nil {
		return err
	}

	for toolName, tool := range c.Tools {
		perToolPath := filepath.Join(configDir, toolName+".toml")

		// Check if per-tool file exists
		if _, err := os.Stat(perToolPath); err != nil {
			// Per-tool file doesn't exist, keep values in config.toml
			continue
		}

		// Per-tool file exists, save values there
		if tool.Values != nil && len(tool.Values) > 0 {
			data, err := toml.Marshal(tool.Values)
			if err != nil {
				return fmt.Errorf("failed to marshal per-tool config for %s: %w", toolName, err)
			}

			if err := os.WriteFile(perToolPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write per-tool config %s: %w", perToolPath, err)
			}

			// Remove values from main config since they're in per-tool file
			tool.Values = nil
			c.Tools[toolName] = tool
		}
	}

	return nil
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
func (c *Config) SetToolValue(toolName, path string, value interface{}) {
	if c.Tools == nil {
		c.Tools = make(map[string]ToolConfig)
	}

	tool := c.Tools[toolName]
	if tool.Values == nil {
		tool.Values = make(map[string]interface{})
	}

	tool.Values[path] = value
	c.Tools[toolName] = tool

	// Ensure per-tool file will be created on Save()
	configDir, err := ConfigDir()
	if err != nil {
		return
	}
	perToolPath := filepath.Join(configDir, toolName+".toml")

	// Create empty file if it doesn't exist, so Save() will write to it
	if _, err := os.Stat(perToolPath); os.IsNotExist(err) {
		os.MkdirAll(configDir, 0755)
		os.WriteFile(perToolPath, []byte{}, 0644)
	}
}

// GetToolValue gets a specific configuration value for a tool
func (c *Config) GetToolValue(toolName, path string) (interface{}, bool) {
	tool, exists := c.Tools[toolName]
	if !exists || tool.Values == nil {
		return nil, false
	}

	value, exists := tool.Values[path]
	return value, exists
}

// UnsetToolValue removes a specific configuration value for a tool
func (c *Config) UnsetToolValue(toolName, path string) {
	tool, exists := c.Tools[toolName]
	if !exists || tool.Values == nil {
		return
	}

	delete(tool.Values, path)
	c.Tools[toolName] = tool
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

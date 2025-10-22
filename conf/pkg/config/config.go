package config

import (
	"fmt"
	"os"
	"path/filepath"

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
	ConfigPath string                 `toml:"config_path"`
	SchemaPath string                 `toml:"schema_path,omitempty"`
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

// Load loads the configuration from the config file, creating it if it doesn't exist
func Load() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// If config doesn't exist, create default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultConfig()
		if err := config.Save(); err != nil {
			return nil, fmt.Errorf("failed to save default config: %w", err)
		}
		return config, nil
	}

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Expand tilde in config paths
	for name, tool := range config.Tools {
		if expandedPath, err := ExpandPath(tool.ConfigPath); err == nil {
			tool.ConfigPath = expandedPath
			config.Tools[name] = tool
		}
	}

	// Load per-tool config files if they exist
	if err := config.loadPerToolConfigs(); err != nil {
		return nil, fmt.Errorf("failed to load per-tool configs: %w", err)
	}

	return &config, nil
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

		// Parse per-tool config into a map
		var perToolValues map[string]interface{}
		if err := toml.Unmarshal(data, &perToolValues); err != nil {
			return fmt.Errorf("failed to parse per-tool config %s: %w", perToolPath, err)
		}

		// Merge per-tool values with tool metadata
		// Per-tool file values override config.toml values
		if tool.Values == nil {
			tool.Values = make(map[string]interface{})
		}
		for k, v := range perToolValues {
			tool.Values[k] = v
		}
		c.Tools[toolName] = tool
	}

	return nil
}

// Save saves the configuration to the config file
func (c *Config) Save() error {
	configPath, err := ConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save per-tool configs if they exist
	if err := c.savePerToolConfigs(); err != nil {
		return fmt.Errorf("failed to save per-tool configs: %w", err)
	}

	// Marshal config to TOML (metadata only if per-tool files exist)
	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
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

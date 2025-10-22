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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "~"
	}

	return &Config{
		Tools: map[string]ToolConfig{
			"jj": {
				Name:       "jj",
				ConfigPath: filepath.Join(homeDir, ".config/jj/config.toml"),
				SchemaPath: "embedded://jj.json",
			},
			"mise": {
				Name:       "mise",
				ConfigPath: filepath.Join(homeDir, ".config", "mise", "config.toml"),
				SchemaPath: "embedded://mise.toml",
			},
			"starship": {
				Name:       "starship",
				ConfigPath: filepath.Join(homeDir, ".config", "starship.toml"),
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

	return &config, nil
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

	// Marshal config to TOML
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

// GetTool returns the configuration for a specific tool
func (c *Config) GetTool(name string) (ToolConfig, bool) {
	tool, exists := c.Tools[name]
	return tool, exists
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

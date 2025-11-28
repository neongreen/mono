package config

import (
	"fmt"
	"os"
	"path/filepath"

	tomlcp "github.com/neongreen/mono/lib/toml"
	tomlv2 "github.com/pelletier/go-toml/v2"
)

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
		if err := tomlv2.Unmarshal(data, &loadedConfig); err != nil {
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

	// Normalize tool values to nested maps (handles legacy dotted keys)
	for name, tool := range config.Tools {
		tool.Values = normalizeValues(tool.Values)
		config.Tools[name] = tool
	}

	// Load values from per-tool config files if they exist
	// This augments the values loaded from main config.toml
	// Note: paths remain in tilde notation; GetTool() expands them when needed
	if err := config.loadPerToolConfigs(); err != nil {
		return nil, fmt.Errorf("failed to load per-tool configs: %w", err)
	}

	return config, nil
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
		var perToolNested map[string]any
		if err := tomlv2.Unmarshal(data, &perToolNested); err != nil {
			return fmt.Errorf("failed to parse per-tool config %s: %w", perToolPath, err)
		}

		perToolValues := normalizeValues(perToolNested)
		tool.Values = mergeNestedValues(tool.Values, perToolValues)
		c.Tools[toolName] = tool
	}

	return nil
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
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Save tools to per-tool files if they exist
	for toolName, tool := range c.Tools {
		perToolPath := filepath.Join(configDir, toolName+".toml")

		// Save to per-tool file if it exists
		if _, err := os.Stat(perToolPath); err == nil {
			// Save values if they exist
			if len(tool.Values) > 0 {
				if err := tomlcp.WriteFile(perToolPath, tool.Values); err != nil {
					return fmt.Errorf("failed to write %s config: %w", toolName, err)
				}
			}
		}
	}

	// Save main config.toml to preserve tool metadata (paths, schema paths, etc.)
	// and values for tools that don't have per-tool files
	mainConfig := &Config{
		Tools:   make(map[string]ToolConfig),
		Folders: c.Folders, // Always save folder metadata in main config
		Shims:   c.Shims,   // Always save shims in main config
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

	mainConfigMap, err := marshalToMap(mainConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal main config: %w", err)
	}

	if err := tomlcp.WriteFile(configPath, mainConfigMap); err != nil {
		return fmt.Errorf("failed to write main config: %w", err)
	}

	return nil
}

func marshalToMap(v any) (map[string]any, error) {
	data, err := tomlv2.Marshal(v)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := tomlv2.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

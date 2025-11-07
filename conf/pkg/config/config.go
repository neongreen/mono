package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/creachadair/tomledit/parser"
	tomlcp "github.com/neongreen/mono/lib/toml"
	tomlv2 "github.com/pelletier/go-toml/v2"
)

// Config represents the main configuration for conf
type Config struct {
	Tools map[string]ToolConfig `toml:"tools"`
	Shims map[string]string     `toml:"shims,omitempty"`
}

// ToolConfig represents configuration for a specific tool
type ToolConfig struct {
	Name       string         `toml:"name"`
	ConfigPath string         `toml:"path"`
	SchemaPath string         `toml:"schema,omitempty"`
	Values     map[string]any `toml:"values,omitempty"`
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
				ConfigPath: "~/.config/claude/config.json",
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

	mainConfigMap, err := marshalToMap(mainConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal main config: %w", err)
	}

	if err := tomlcp.WriteFile(configPath, mainConfigMap); err != nil {
		return fmt.Errorf("failed to write main config: %w", err)
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

// FlattenValues converts nested configuration maps to dotted-path representation.
// Keys are quoted when necessary to produce TOML-compliant paths (e.g., aliases.".")
func FlattenValues(values map[string]any) map[string]any {
	result := make(map[string]any)
	normalized := normalizeValues(values)
	flattenRecursive(normalized, "", result)
	return result
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

// ExpandValues converts a map of dotted paths back into a nested map structure.
func ExpandValues(flat map[string]any) map[string]any {
	if flat == nil {
		return nil
	}

	result := make(map[string]any)
	for path, val := range flat {
		key, err := parser.ParseKey(path)
		if err != nil || len(key) == 0 {
			result[path] = normalizeValue(val)
			continue
		}
		setNestedValue(result, key, val)
	}
	return result
}

func flattenRecursive(values map[string]any, prefix string, result map[string]any) {
	if values == nil {
		return
	}

	for key, val := range values {
		formattedKey := formatKeySegment(key)

		fullKey := formattedKey
		if prefix != "" {
			fullKey = prefix + "." + formattedKey
		}

		if nested, ok := val.(map[string]any); ok {
			flattenRecursive(nested, fullKey, result)
			continue
		}

		result[fullKey] = val
	}
}

func formatKeySegment(key string) string {
	if isBareKey(key) {
		return key
	}
	escaped := strings.ReplaceAll(key, `"`, `\"`)
	return `"` + escaped + `"`
}

func isBareKey(key string) bool {
	if len(key) == 0 {
		return false
	}
	for _, r := range key {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

func ensureMap(m map[string]any) map[string]any {
	if m == nil {
		return make(map[string]any)
	}
	return m
}

func normalizeValues(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}

	normalized := make(map[string]any, len(values))
	for rawKey, rawVal := range values {
		normalizedVal := normalizeValue(rawVal)
		key, err := parser.ParseKey(rawKey)
		if err == nil && len(key) > 1 {
			setNestedValue(normalized, key, normalizedVal)
			continue
		}
		if err == nil && len(key) == 1 {
			normalized[key[0]] = normalizedVal
			continue
		}
		normalized[rawKey] = normalizedVal
	}
	return normalized
}

func normalizeValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		return normalizeValues(v)
	case map[any]any:
		converted := make(map[string]any, len(v))
		for k, elem := range v {
			strKey, ok := k.(string)
			if !ok {
				continue
			}
			converted[strKey] = normalizeValue(elem)
		}
		return normalizeValues(converted)
	case []any:
		out := make([]any, len(v))
		for i, elem := range v {
			out[i] = normalizeValue(elem)
		}
		return out
	default:
		return value
	}
}

func mergeNestedValues(dst map[string]any, src map[string]any) map[string]any {
	dst = normalizeValues(dst)
	if dst == nil {
		dst = make(map[string]any)
	}

	srcNormalized := normalizeValues(src)
	if srcNormalized == nil {
		return dst
	}

	for key, val := range srcNormalized {
		if existing, ok := dst[key]; ok {
			existingMap, existingIsMap := existing.(map[string]any)
			newMap, newIsMap := val.(map[string]any)
			if existingIsMap && newIsMap {
				dst[key] = mergeNestedValues(existingMap, newMap)
				continue
			}
		}
		dst[key] = val
	}

	return dst
}

func setNestedValue(m map[string]any, key parser.Key, value any) {
	if len(key) == 0 {
		return
	}

	current := ensureMap(m)
	for i := 0; i < len(key)-1; i++ {
		part := key[i]
		next, ok := current[part]
		if !ok {
			nextMap := make(map[string]any)
			current[part] = nextMap
			current = nextMap
			continue
		}

		nextMap, ok := next.(map[string]any)
		if !ok {
			nextMap = make(map[string]any)
			current[part] = nextMap
		}
		current = nextMap
	}

	current[key[len(key)-1]] = normalizeValue(value)
}

func getNestedValue(m map[string]any, key parser.Key) (any, bool) {
	if len(key) == 0 {
		return nil, false
	}

	current := m
	for i := range key {
		part := key[i]
		val, ok := current[part]
		if !ok {
			return nil, false
		}

		if i == len(key)-1 {
			return val, true
		}

		nextMap, ok := val.(map[string]any)
		if !ok {
			return nil, false
		}
		current = nextMap
	}

	return nil, false
}

func unsetNestedValue(m map[string]any, key parser.Key) bool {
	if len(key) == 0 {
		return false
	}

	current := m
	stack := make([]map[string]any, 0, len(key))
	stack = append(stack, current)

	for i := 0; i < len(key)-1; i++ {
		part := key[i]
		next, ok := current[part].(map[string]any)
		if !ok {
			return false
		}
		stack = append(stack, next)
		current = next
	}

	lastPart := key[len(key)-1]
	if _, exists := current[lastPart]; !exists {
		return false
	}
	delete(current, lastPart)

	// Cleanup empty maps from the bottom up
	for i := len(stack) - 1; i > 0; i-- {
		parent := stack[i-1]
		childKey := key[i-1]
		child, ok := parent[childKey].(map[string]any)
		if !ok {
			break
		}
		if len(child) == 0 {
			delete(parent, childKey)
		} else {
			break
		}
	}

	return true
}

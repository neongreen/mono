package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// GetConfigPath returns the path to the tk config file
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "tk")
	return filepath.Join(configDir, "config.toml"), nil
}

// LoadConfig loads the tk configuration file
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Return default config if file doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			Remotes:  make(map[string]RemoteConfig),
			Sync:     DefaultSyncConfig(),
			Blocking: DefaultBlockingConfig(),
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults if not specified
	if config.Sync.SegmentMaxBytes == 0 {
		config.Sync = DefaultSyncConfig()
	}
	if config.Blocking.BlockingAxis == "" {
		config.Blocking = DefaultBlockingConfig()
	}

	return &config, nil
}

// SaveConfig saves the tk configuration file
func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetStateDir returns the path to the tk state directory
func GetStateDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	stateDir := filepath.Join(home, ".tk")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create state directory: %w", err)
	}

	return stateDir, nil
}

package remote

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AddRemote adds a new remote to the configuration
func AddRemote(cfg *Config, name, remoteType, path string) error {
	// Validate remote type
	if err := ValidateRemoteType(remoteType); err != nil {
		return err
	}

	// Expand home directory
	expandedPath, err := ExpandPath(path)
	if err != nil {
		return err
	}

	// Check if remote already exists
	if _, exists := cfg.Remotes[name]; exists {
		return fmt.Errorf("remote '%s' already exists", name)
	}

	// Add remote with default configuration
	cfg.Remotes[name] = RemoteConfig{
		Type:   remoteType,
		Path:   expandedPath,
		Spaces: []string{"personal"},
		Push:   true,
		Pull:   true,
	}

	return nil
}

// RemoveRemote removes a remote from the configuration
func RemoveRemote(cfg *Config, name string) error {
	// Check if remote exists
	if _, exists := cfg.Remotes[name]; !exists {
		return fmt.Errorf("remote '%s' not found", name)
	}

	// Remove remote
	delete(cfg.Remotes, name)
	return nil
}

// GetRemote retrieves a specific remote from the configuration
func GetRemote(cfg *Config, name string) (*RemoteConfig, error) {
	remote, exists := cfg.Remotes[name]
	if !exists {
		return nil, fmt.Errorf("remote '%s' not found", name)
	}
	return &remote, nil
}

// ListRemotes returns all configured remotes
func ListRemotes(cfg *Config) map[string]RemoteConfig {
	return cfg.Remotes
}

// ValidateRemoteType checks if a remote type is supported
func ValidateRemoteType(remoteType string) error {
	if remoteType != "folder" {
		return fmt.Errorf("unsupported remote type: %s (only 'folder' is supported in v1)", remoteType)
	}
	return nil
}

// ExpandPath expands the home directory (~/) in a path
func ExpandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

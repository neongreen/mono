package sync

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/neongreen/mono/conf/pkg/config"
	"github.com/pelletier/go-toml/v2"
)

// ICloudDrivePath returns the path to iCloud Drive
func ICloudDrivePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	icloudPath := filepath.Join(homeDir, "Library", "Mobile Documents", "com~apple~CloudDocs")

	// Check if iCloud Drive path exists
	if _, err := os.Stat(icloudPath); os.IsNotExist(err) {
		return "", fmt.Errorf("iCloud Drive not found at %s (is iCloud Drive enabled?)", icloudPath)
	}

	// Create conf directory in iCloud Drive if it doesn't exist
	confPath := filepath.Join(icloudPath, "conf")
	if err := os.MkdirAll(confPath, 0o755); err != nil {
		return "", fmt.Errorf("failed to create conf directory in iCloud Drive: %w", err)
	}

	return confPath, nil
}

// DownloadFromICloud downloads a tool's config from iCloud Drive
func DownloadFromICloud(toolName string) (map[string]any, error) {
	icloudPath, err := ICloudDrivePath()
	if err != nil {
		return nil, err
	}

	toolPath := filepath.Join(icloudPath, toolName+".toml")

	// Check if file exists in iCloud
	if _, err := os.Stat(toolPath); os.IsNotExist(err) {
		// No iCloud data for this tool yet
		return nil, nil
	}

	// Read iCloud config
	data, err := os.ReadFile(toolPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read iCloud config: %w", err)
	}

	var values map[string]any
	if err := toml.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("failed to parse iCloud config: %w", err)
	}

	return values, nil
}

// UploadToICloud uploads a tool's config to iCloud Drive
func UploadToICloud(toolName string, values map[string]any) error {
	icloudPath, err := ICloudDrivePath()
	if err != nil {
		return err
	}

	toolPath := filepath.Join(icloudPath, toolName+".toml")

	// Marshal values to TOML
	data, err := toml.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to iCloud
	if err := os.WriteFile(toolPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write to iCloud: %w", err)
	}

	return nil
}

// MergeConfigs merges local and iCloud configs
// For conflicting keys, uses Last-Write-Wins based on file modification times
func MergeConfigs(local, icloud map[string]any, localMtime, icloudMtime int64) map[string]any {
	localFlat := config.FlattenValues(local)
	icloudFlat := config.FlattenValues(icloud)

	resultFlat := make(map[string]any, len(localFlat))
	maps.Copy(resultFlat, localFlat)

	// Add/override with iCloud keys
	for k, v := range icloudFlat {
		if _, exists := resultFlat[k]; !exists || icloudMtime > localMtime {
			resultFlat[k] = v
		}
	}

	return config.ExpandValues(resultFlat)
}

// GetLocalValues returns the values for a tool from local config
func GetLocalValues(conf *config.Config, toolName string) (map[string]any, error) {
	tool, exists := conf.GetTool(toolName)
	if !exists {
		return nil, fmt.Errorf("tool %s not configured", toolName)
	}

	if tool.Values == nil {
		return make(map[string]any), nil
	}

	return tool.Values, nil
}

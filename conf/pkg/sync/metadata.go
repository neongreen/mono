package sync

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tomlcp "github.com/neongreen/mono/lib/toml"
	tomlv2 "github.com/pelletier/go-toml/v2"
)

// SyncMetadata represents the sync state for all tools
type SyncMetadata struct {
	Tools map[string]ToolSyncState `toml:"tools"`
}

// ToolSyncState represents the sync state for a single tool
type ToolSyncState struct {
	LastICloudHash string    `toml:"last_icloud_hash"`
	LastSyncTime   time.Time `toml:"last_sync_time"`
	LocalHash      string    `toml:"local_hash,omitempty"`
}

// LoadSyncMetadata loads the sync metadata from ~/.config/conf/.sync-state
func LoadSyncMetadata(configDir string) (*SyncMetadata, error) {
	metadataPath := filepath.Join(configDir, ".sync-state")

	// If metadata file doesn't exist, return empty metadata
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return &SyncMetadata{
			Tools: make(map[string]ToolSyncState),
		}, nil
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sync metadata: %w", err)
	}

	var metadata SyncMetadata
	if err := tomlv2.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse sync metadata: %w", err)
	}

	if metadata.Tools == nil {
		metadata.Tools = make(map[string]ToolSyncState)
	}

	return &metadata, nil
}

// SaveSyncMetadata saves the sync metadata to ~/.config/conf/.sync-state
func (m *SyncMetadata) Save(configDir string) error {
	metadataPath := filepath.Join(configDir, ".sync-state")

	metadataMap, err := metadataToMap(m)
	if err != nil {
		return fmt.Errorf("failed to marshal sync metadata: %w", err)
	}

	if err := tomlcp.WriteFile(metadataPath, metadataMap); err != nil {
		return fmt.Errorf("failed to write sync metadata: %w", err)
	}

	return nil
}

// UpdateToolState updates the sync state for a specific tool
func (m *SyncMetadata) UpdateToolState(toolName, icloudHash, localHash string) {
	m.Tools[toolName] = ToolSyncState{
		LastICloudHash: icloudHash,
		LocalHash:      localHash,
		LastSyncTime:   time.Now(),
	}
}

// GetToolState returns the sync state for a specific tool
func (m *SyncMetadata) GetToolState(toolName string) (ToolSyncState, bool) {
	state, exists := m.Tools[toolName]
	return state, exists
}

// ComputeFileHash computes SHA256 hash of a file
func ComputeFileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file for hashing: %w", err)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func metadataToMap(m *SyncMetadata) (map[string]any, error) {
	data, err := tomlv2.Marshal(m)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := tomlv2.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

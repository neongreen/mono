package remote

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadJSON loads a JSON file into a struct
func LoadJSON[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &result, nil
}

// SaveJSON saves a struct to a JSON file
func SaveJSON[T any](path string, data *T) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// CollectSegmentFiles walks a directory and collects all .zst segment files
func CollectSegmentFiles(dir string) ([]string, error) {
	var segmentFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".zst" {
			segmentFiles = append(segmentFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return segmentFiles, nil
}

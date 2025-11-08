package remote

import (
	"encoding/json"
	"fmt"
	"io"
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

// copyFile copies a file from src to dst, creating parent directories as needed.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", dst, err)
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", dst, err)
	}
	defer func() {
		out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("failed to copy %s to %s: %w", src, dst, err)
	}

	if err := out.Sync(); err != nil {
		return fmt.Errorf("failed to sync %s: %w", dst, err)
	}

	if err := out.Chmod(0o644); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", dst, err)
	}

	return nil
}

// CacheSegmentFile copies a remote segment file into the local state cache.
func CacheSegmentFile(stateDir, remoteName, remoteBasePath string, seg SegmentInfo) error {
	src := filepath.Join(remoteBasePath, seg.Rel)
	dst := filepath.Join(stateDir, "remotes", remoteName, seg.Rel)
	if err := copyFile(src, dst); err != nil {
		return fmt.Errorf("failed to cache segment %s: %w", seg.Rel, err)
	}
	return nil
}

// RestoreSegmentFromCache ensures the remote segment file exists by copying from cache if missing.
// Returns (wasRestored=true, err) if file was copied from cache
// Returns (wasRestored=false, err=nil) if file already exists or cache doesn't exist
func RestoreSegmentFromCache(stateDir, remoteName, remoteBasePath string, seg SegmentInfo) (wasRestored bool, err error) {
	dst := filepath.Join(remoteBasePath, seg.Rel)
	if _, err := os.Stat(dst); err == nil {
		return false, nil // already present, no restoration needed
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to stat segment %s: %w", seg.Rel, err)
	}

	// Segment missing from remote, try to restore from cache
	cachePath := filepath.Join(stateDir, "remotes", remoteName, seg.Rel)
	if _, err := os.Stat(cachePath); err != nil {
		if os.IsNotExist(err) {
			// Cache doesn't exist - segment can't be restored
			return false, nil
		}
		return false, fmt.Errorf("failed to stat cache for segment %s: %w", seg.Rel, err)
	}

	// Copy from cache to remote
	if err := copyFile(cachePath, dst); err != nil {
		return false, fmt.Errorf("failed to restore segment %s from cache: %w", seg.Rel, err)
	}
	return true, nil
}

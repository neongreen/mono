package fs

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type FSEntry struct {
	Path       string
	IsDir      bool
	Size       int64
	Mode       string
	ModTime    time.Time
	Content    []byte
	SHA256Hash string
}

// WalkFilesystem walks through a filesystem path recursively
// The progressCallback is called periodically with the number of entries found so far
func WalkFilesystem(rootPath string, progressCallback func(int)) ([]FSEntry, error) {
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path does not exist: %s", absPath)
		}
		return nil, fmt.Errorf("failed to stat %s: %w", absPath, err)
	}

	var entries []FSEntry
	count := 0

	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip files/dirs we can't access
			return nil
		}

		relPath, err := filepath.Rel(absPath, path)
		if err != nil {
			relPath = path
		}
		if relPath == "." {
			relPath = "/"
		}

		entry := FSEntry{
			Path:    relPath,
			IsDir:   info.IsDir(),
			Size:    info.Size(),
			Mode:    info.Mode().String(),
			ModTime: info.ModTime(),
		}

		// Read file content if it's a regular file and not too large (limit to 10MB)
		if !info.IsDir() && info.Mode().IsRegular() && info.Size() <= 10*1024*1024 {
			content, err := os.ReadFile(path)
			if err == nil {
				entry.Content = content
				// Calculate SHA256 hash
				hash := sha256.Sum256(content)
				entry.SHA256Hash = fmt.Sprintf("%x", hash)
			}
		}

		entries = append(entries, entry)
		count++

		// Call progress callback every 100 entries or on first entry
		if progressCallback != nil && (count%100 == 0 || count == 1) {
			progressCallback(count)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk filesystem: %w", err)
	}

	return entries, nil
}

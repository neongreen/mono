package remote

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/neongreen/mono/tk/internal/segment"
)

// LoadIndexFile loads an index file from the given path
//
//nolint:uselesswrapper // Type-safe wrapper for LoadJSON
func LoadIndexFile(path string) (*IndexFile, error) {
	return LoadJSON[IndexFile](path)
}

// SaveIndexFile saves an index file to the given path
//
//nolint:uselesswrapper // Type-safe wrapper for SaveJSON
func SaveIndexFile(path string, index *IndexFile) error {
	return SaveJSON(path, index)
}

// ReconstructIndex scans a segments directory and reconstructs the index from disk
func ReconstructIndex(remotePath, space string) (*IndexFile, error) {
	segmentsDir := filepath.Join(remotePath, space, "segments")
	if _, err := os.Stat(segmentsDir); os.IsNotExist(err) {
		return &IndexFile{
			Schema:   "tk.index.v1",
			Space:    space,
			Segments: []SegmentInfo{},
		}, nil
	}

	var segments []SegmentInfo
	err := filepath.Walk(segmentsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".zst" {
			// Calculate relative path
			relPath, err := filepath.Rel(remotePath, path)
			if err != nil {
				return err
			}

			// Calculate SHA256
			sha, err := segment.CalculateSHA256(path)
			if err != nil {
				return fmt.Errorf("failed to calculate SHA256 for %s: %w", path, err)
			}

			segments = append(segments, SegmentInfo{
				Rel:    relPath,
				SHA256: sha,
				Size:   info.Size(),
				MTime:  info.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &IndexFile{
		Schema:   "tk.index.v1",
		Space:    space,
		Segments: segments,
	}, nil
}

// UpdateLocalIndex updates the local index mirror with new segments
func UpdateLocalIndex(indexPath string, newSegments []SegmentInfo) error {
	// Ensure directory exists
	dir := filepath.Dir(indexPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create index directory: %w", err)
	}

	// Load existing index or create new one
	var index IndexFile
	existingIndex, err := LoadIndexFile(indexPath)
	if err == nil && existingIndex != nil {
		index = *existingIndex
	} else {
		// Extract space from path
		parts := filepath.SplitList(indexPath)
		space := "personal" // default
		if len(parts) >= 2 {
			space = parts[len(parts)-2]
		}
		index = IndexFile{
			Schema:   "tk.index.v1",
			Space:    space,
			Segments: []SegmentInfo{},
		}
	}

	// Add new segments
	index.Segments = append(index.Segments, newSegments...)

	// Save index
	return SaveIndexFile(indexPath, &index)
}

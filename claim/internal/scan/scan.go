package scan

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/neongreen/mono/claim/internal/logger"
	"github.com/neongreen/mono/claim/internal/parse"
)

// ScannedFile represents a file that was scanned
type ScannedFile struct {
	Path    string
	Content []byte
}

// ignoredDirs are directories to skip during scanning
var ignoredDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	"dist":         true,
	"build":        true,
	"target":       true,
	".next":        true,
	".vscode":      true,
	".idea":        true,
}

// ScanFiles walks the directory tree and reads all non-ignored files
func ScanFiles(root string) ([]ScannedFile, error) {
	var files []ScannedFile

	logger.Debug("scanning directory", "root", root)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip ignored directories
		if d.IsDir() {
			if ignoredDirs[d.Name()] {
				logger.Debug("skipping ignored directory", "path", path)
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files (except at root level)
		if strings.HasPrefix(d.Name(), ".") && path != root {
			logger.Debug("skipping hidden file", "path", path)
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			// Skip files we can't read (permissions, etc)
			logger.Debug("skipping unreadable file", "path", path, "error", err)
			return nil
		}

		logger.Debug("scanned file", "path", path, "size", len(content))
		files = append(files, ScannedFile{
			Path:    path,
			Content: content,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	logger.Debug("scan complete", "files", len(files))
	return files, nil
}

// LoadLensFile loads additional lenses from a markdown file
func LoadLensFile(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read lens file: %w", err)
	}

	// Parse lenses from the file
	return parse.ParseLenses(string(content), path)
}

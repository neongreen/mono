package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// MoveFile moves a file from source to destination, creating destination directories if they don't exist.
func MoveFile(srcPath, destPath string) error {
	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("error creating destination directory %s: %w", destDir, err)
	}

	err := os.Rename(srcPath, destPath)
	if err != nil {
		return fmt.Errorf("error moving file from %s to %s: %w", srcPath, destPath, err)
	}
	return nil
}

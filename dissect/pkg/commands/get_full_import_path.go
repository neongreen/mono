package commands

import (
	"path/filepath"
)

// GetFullImportPath returns the full import path for a given file path.
func GetFullImportPath(filePath string) (string, error) {
	return RunGoListNoArgs(filepath.Dir(filePath))
}

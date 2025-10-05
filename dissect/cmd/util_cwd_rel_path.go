package main

import (
	"log/slog"
	"os"
	"path/filepath"
)

// cwdRelPath returns the relative path of the given file from the current working directory.
// We use it to log the file paths in a more readable way.
func cwdRelPath(filePath string) string {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		slog.Error("Error getting current working directory", "error", err)
		return filePath // Fallback to absolute path if error occurs
	}

	// Get the relative path from the current working directory
	rel, err := filepath.Rel(cwd, filePath)
	if err != nil {
		slog.Error("Error getting relative path", "file", filePath, "error", err)
		return filePath // Fallback to absolute path if error occurs
	}
	return rel
}

package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// DiscoverTraceLocations finds potential Claude Code trace directories
func DiscoverTraceLocations() ([]string, error) {
	var paths []string
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Common locations for Claude Code traces
	// Note: These are best guesses based on typical application data storage patterns
	// The actual location may vary depending on the Claude Code version and platform
	possibleLocations := []string{
		// Actual Claude Code storage locations
		filepath.Join(homeDir, ".claude", "projects"),
		filepath.Join(homeDir, ".claude", "debug"),
		filepath.Join(homeDir, ".claude", "traces"),
		// Legacy locations (fallback)
		// Linux/XDG locations
		filepath.Join(homeDir, ".config", "Claude", "traces"),
		filepath.Join(homeDir, ".local", "share", "Claude", "traces"),
		filepath.Join(homeDir, ".local", "share", "claude-code", "traces"),
		// macOS locations
		filepath.Join(homeDir, "Library", "Application Support", "Claude", "traces"),
		filepath.Join(homeDir, "Library", "Application Support", "claude-code", "traces"),
		// Windows locations (if on Windows)
		filepath.Join(homeDir, "AppData", "Roaming", "Claude", "traces"),
		filepath.Join(homeDir, "AppData", "Local", "Claude", "traces"),
		// Current directory fallback for testing
		"./traces",
	}

	for _, location := range possibleLocations {
		if info, err := os.Stat(location); err == nil && info.IsDir() {
			paths = append(paths, location)
		}
	}

	return paths, nil
}

// GetAllSearchedLocations returns all possible trace locations that are checked,
// regardless of whether they exist or not. This is useful for displaying where
// the tool searched for traces.
func GetAllSearchedLocations() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If we can't get home directory, return just the current directory fallback
		return []string{"./traces"}
	}

	// Same locations as DiscoverTraceLocations but always return all of them
	return []string{
		// Actual Claude Code storage locations
		filepath.Join(homeDir, ".claude", "projects"),
		filepath.Join(homeDir, ".claude", "debug"),
		filepath.Join(homeDir, ".claude", "traces"),
		// Legacy locations (fallback)
		// Linux/XDG locations
		filepath.Join(homeDir, ".config", "Claude", "traces"),
		filepath.Join(homeDir, ".local", "share", "Claude", "traces"),
		filepath.Join(homeDir, ".local", "share", "claude-code", "traces"),
		// macOS locations
		filepath.Join(homeDir, "Library", "Application Support", "Claude", "traces"),
		filepath.Join(homeDir, "Library", "Application Support", "claude-code", "traces"),
		// Windows locations (if on Windows)
		filepath.Join(homeDir, "AppData", "Roaming", "Claude", "traces"),
		filepath.Join(homeDir, "AppData", "Local", "Claude", "traces"),
		// Current directory fallback for testing
		"./traces",
	}
}

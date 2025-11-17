package main

import (
	"os"
	"path/filepath"
	"strings"
)

// findMonoRoot walks up the directory tree to find the mono repository root.
// Returns the root directory path and any error encountered.
func findMonoRoot(startDir string) (string, error) {
	dir := startDir
	for {
		// Check for markers that indicate mono repo root
		// 1. mise.toml file (most reliable)
		if _, err := os.Stat(filepath.Join(dir, "mise.toml")); err == nil {
			return dir, nil
		}

		// 2. go.mod with the right module path
		goModPath := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(goModPath); err == nil {
			if strings.Contains(string(data), "github.com/neongreen/mono") {
				return dir, nil
			}
		}

		// 3. .git directory with common project subdirectories
		if stat, err := os.Stat(filepath.Join(dir, ".git")); err == nil && stat.IsDir() {
			// Check for at least 2 common projects
			commonProjects := []string{"tk", "want", "dissect", "printpdf", "conf"}
			found := 0
			for _, proj := range commonProjects {
				if stat, err := os.Stat(filepath.Join(dir, proj)); err == nil && stat.IsDir() {
					found++
					if found >= 2 {
						return dir, nil
					}
				}
			}
		}

		// Move up one level
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}

package remote

import (
	"log/slog"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// isICloudPath checks if a path is inside iCloud Drive.
// Returns true if the path contains the iCloud Drive folder structure.
func isICloudPath(path string) bool {
	// Normalize path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// Check for iCloud Drive path patterns
	// Standard: ~/Library/Mobile Documents/com~apple~CloudDocs/
	// App-specific: ~/Library/Mobile Documents/iCloud~...
	return strings.Contains(absPath, "/Library/Mobile Documents/com~apple~CloudDocs/") ||
		strings.Contains(absPath, "/Library/Mobile Documents/iCloud~")
}

// forceDownloadICloudFiles attempts to download iCloud files using brctl.
// This is macOS-only and silently does nothing on other platforms or if brctl is not available.
// Returns true if download was initiated, false otherwise.
func forceDownloadICloudFiles(path string) bool {
	// Only run on macOS
	if runtime.GOOS != "darwin" {
		return false
	}

	// Check if this is an iCloud path
	if !isICloudPath(path) {
		return false
	}

	// Check if brctl is available
	if _, err := exec.LookPath("brctl"); err != nil {
		slog.Debug("brctl not found, skipping iCloud download", "path", path)
		return false
	}

	// Run brctl download
	slog.Info("downloading iCloud files (this may take a moment)...", "path", path)
	cmd := exec.Command("brctl", "download", path)

	// Run asynchronously - we don't wait for completion
	// brctl download returns immediately and files download in background
	if err := cmd.Start(); err != nil {
		slog.Warn("failed to start iCloud download", "path", path, "error", err)
		return false
	}

	// Don't wait for completion - let it download in background
	go func() {
		if err := cmd.Wait(); err != nil {
			slog.Debug("brctl download completed with error", "path", path, "error", err)
		} else {
			slog.Debug("brctl download initiated", "path", path)
		}
	}()

	return true
}

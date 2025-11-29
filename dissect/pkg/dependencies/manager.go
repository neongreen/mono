package dependencies

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Manager handles finding and installing gopls and goimports
type Manager struct {
	projectDir    string
	goplsPath     string
	goimportsPath string
}

// NewManager creates a dependency manager for a project
func NewManager(projectDir string) *Manager {
	return &Manager{
		projectDir: projectDir,
	}
}

// EnsureGopls finds or installs gopls, returns path
func (m *Manager) EnsureGopls() (string, error) {
	if m.goplsPath != "" {
		return m.goplsPath, nil
	}

	path, err := m.findOrInstall("gopls", "golang.org/x/tools/gopls")
	if err != nil {
		return "", err
	}

	m.goplsPath = path
	return path, nil
}

// EnsureGoimports finds or installs goimports, returns path
func (m *Manager) EnsureGoimports() (string, error) {
	if m.goimportsPath != "" {
		return m.goimportsPath, nil
	}

	path, err := m.findOrInstall("goimports", "golang.org/x/tools/cmd/goimports")
	if err != nil {
		return "", err
	}

	m.goimportsPath = path
	return path, nil
}

// findOrInstall is the core logic for finding or installing a Go tool
func (m *Manager) findOrInstall(tool string, importPath string) (string, error) {
	slog.Debug("Finding or installing tool", "tool", tool, "projectDir", m.projectDir)

	// Get project-specific GOBIN
	gobin, err := m.getGOBIN()
	if err != nil {
		return "", fmt.Errorf("failed to get GOBIN: %w", err)
	}

	toolPath := filepath.Join(gobin, tool)

	// Check if tool already exists
	if _, err := os.Stat(toolPath); err == nil {
		slog.Debug("Tool found", "tool", tool, "path", toolPath)
		return toolPath, nil
	}

	// Tool not found, install it
	slog.Info("Installing tool (one-time setup)", "tool", tool)

	installCmd := exec.Command("go", "install", importPath+"@latest") //nolint:gosec // G204: intentional execution of go install for tool management
	installCmd.Dir = m.projectDir
	installCmd.Env = append(os.Environ(), fmt.Sprintf("GOBIN=%s", gobin))

	// Set timeout for installation (2 minutes for gopls which can be slow)
	done := make(chan error, 1)
	go func() {
		done <- installCmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			return "", fmt.Errorf("failed to install %s: %w", tool, err)
		}
	case <-time.After(2 * time.Minute):
		return "", fmt.Errorf("installation of %s timed out after 2 minutes", tool)
	}

	// Verify installation and wait for the binary to be executable.
	// After go install completes, there can be a brief window where the file exists
	// but is still being written (causing "text file busy" errors on execution).
	if err := waitForExecutable(toolPath, 5*time.Second); err != nil {
		return "", fmt.Errorf("%s installation verification failed: %w", tool, err)
	}

	slog.Info("Tool installed successfully", "tool", tool, "path", toolPath)
	return toolPath, nil
}

// waitForExecutable waits until a binary file exists and can be opened for reading.
// This handles the race condition where go install has written the file
// but the OS hasn't fully released the write handle yet.
func waitForExecutable(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		// Check if file exists
		info, err := os.Stat(path)
		if err != nil {
			lastErr = err
			time.Sleep(50 * time.Millisecond)
			continue
		}

		// Check if it's a regular file
		if !info.Mode().IsRegular() {
			lastErr = fmt.Errorf("%s is not a regular file", path)
			time.Sleep(50 * time.Millisecond)
			continue
		}

		// Try to open the file for reading to verify it's not locked
		f, err := os.Open(path)
		if err != nil {
			lastErr = err
			time.Sleep(50 * time.Millisecond)
			continue
		}
		f.Close()

		// File is ready
		return nil
	}

	if lastErr != nil {
		return fmt.Errorf("timed out waiting for %s to be ready: %w", path, lastErr)
	}
	return fmt.Errorf("timed out waiting for %s to be ready", path)
}

// getGOBIN returns the GOBIN directory for the project
func (m *Manager) getGOBIN() (string, error) {
	// First try to get GOBIN from go env
	cmd := exec.Command("go", "env", "GOBIN")
	cmd.Dir = m.projectDir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run 'go env GOBIN': %w", err)
	}

	gobin := strings.TrimSpace(string(output))

	// If GOBIN is empty, fall back to GOPATH/bin
	if gobin == "" {
		cmd = exec.Command("go", "env", "GOPATH")
		cmd.Dir = m.projectDir
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to run 'go env GOPATH': %w", err)
		}

		gopath := strings.TrimSpace(string(output))
		if gopath == "" {
			return "", fmt.Errorf("both GOBIN and GOPATH are empty")
		}

		gobin = filepath.Join(gopath, "bin")
	}

	return gobin, nil
}

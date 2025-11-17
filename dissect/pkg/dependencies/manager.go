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

	installCmd := exec.Command("go", "install", importPath+"@latest")
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

	// Verify installation
	if _, err := os.Stat(toolPath); err != nil {
		return "", fmt.Errorf("%s was not installed at expected path %s", tool, toolPath)
	}

	slog.Info("Tool installed successfully", "tool", tool, "path", toolPath)
	return toolPath, nil
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

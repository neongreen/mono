package shims

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ShimsTool implements command shim management
type ShimsTool struct {
	shimsDir string
	dryRun   bool
}

// NewShimsTool creates a new shims tool instance
func NewShimsTool() (*ShimsTool, error) {
	return NewShimsToolWithDryRun(false)
}

// NewShimsToolWithDryRun creates a new shims tool instance with dry-run mode
func NewShimsToolWithDryRun(dryRun bool) (*ShimsTool, error) {
	// Default shims directory: ~/.local/bin/conf-shims
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	shimsDir := filepath.Join(homeDir, ".local", "bin", "conf-shims")

	return &ShimsTool{
		shimsDir: shimsDir,
		dryRun:   dryRun,
	}, nil
}

// SetDryRun enables or disables dry-run mode
func (s *ShimsTool) SetDryRun(dryRun bool) {
	s.dryRun = dryRun
}

// CreateShim creates a new command shim
func (s *ShimsTool) CreateShim(name, command string) error {
	if err := s.validateShimName(name); err != nil {
		return err
	}

	// Ensure shims directory exists
	if !s.dryRun {
		if err := os.MkdirAll(s.shimsDir, 0o755); err != nil {
			return fmt.Errorf("failed to create shims directory %s: %w", s.shimsDir, err)
		}
	}

	shimPath := filepath.Join(s.shimsDir, name)

	// Check if shim already exists
	if !s.dryRun {
		if _, err := os.Stat(shimPath); err == nil {
			return fmt.Errorf("shim '%s' already exists", name)
		}
	}

	// Generate shim content
	shimContent, err := s.generateShimContent(name, command)
	if err != nil {
		return fmt.Errorf("failed to generate shim content: %w", err)
	}

	if s.dryRun {
		fmt.Printf("Would create shim: %s\n", shimPath)
		fmt.Printf("Content:\n%s\n", shimContent)
		return nil
	}

	// Write shim file
	if err := os.WriteFile(shimPath, []byte(shimContent), 0o755); err != nil {
		return fmt.Errorf("failed to write shim file %s: %w", shimPath, err)
	}

	return nil
}

// RemoveShim removes an existing command shim
func (s *ShimsTool) RemoveShim(name string) error {
	if err := s.validateShimName(name); err != nil {
		return err
	}

	shimPath := filepath.Join(s.shimsDir, name)

	if s.dryRun {
		fmt.Printf("Would remove shim: %s\n", shimPath)
		return nil
	}

	// Check if shim exists
	if _, err := os.Stat(shimPath); os.IsNotExist(err) {
		return fmt.Errorf("shim '%s' does not exist", name)
	}

	// Remove shim file
	if err := os.Remove(shimPath); err != nil {
		return fmt.Errorf("failed to remove shim file %s: %w", shimPath, err)
	}

	return nil
}

// ListShims returns a list of all managed shims
func (s *ShimsTool) ListShims() ([]ShimInfo, error) {
	var shims []ShimInfo

	// Check if shims directory exists
	if _, err := os.Stat(s.shimsDir); os.IsNotExist(err) {
		return shims, nil // Return empty list if directory doesn't exist
	}

	// Read directory contents
	entries, err := os.ReadDir(s.shimsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read shims directory %s: %w", s.shimsDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		shimPath := filepath.Join(s.shimsDir, entry.Name())
		command, err := s.extractCommandFromShim(shimPath)
		if err != nil {
			// Skip files that aren't valid shims
			continue
		}

		shims = append(shims, ShimInfo{
			Name:    entry.Name(),
			Command: command,
			Path:    shimPath,
		})
	}

	return shims, nil
}

// GetShimsDir returns the shims directory path
func (s *ShimsTool) GetShimsDir() string {
	return s.shimsDir
}

// IsDryRun returns whether dry-run mode is enabled
func (s *ShimsTool) IsDryRun() bool {
	return s.dryRun
}

// validateShimName validates that a shim name is acceptable
func (s *ShimsTool) validateShimName(name string) error {
	if name == "" {
		return fmt.Errorf("shim name cannot be empty")
	}

	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("shim name cannot contain path separators")
	}

	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("shim name cannot start with a dot")
	}

	return nil
}

// generateShimContent generates the content for a shim file
func (s *ShimsTool) generateShimContent(name, command string) (string, error) {
	tmpl := `#!/bin/bash
# Managed by conf
# Shim: {{.Name}}
# Command: {{.Command}}

exec {{.Command}} "$@"
`

	t, err := template.New("shim").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := t.Execute(&buf, struct {
		Name    string
		Command string
	}{
		Name:    name,
		Command: command,
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// extractCommandFromShim extracts the original command from a shim file
func (s *ShimsTool) extractCommandFromShim(shimPath string) (string, error) {
	content, err := os.ReadFile(shimPath)
	if err != nil {
		return "", err
	}

	lines := strings.SplitSeq(string(content), "\n")
	for line := range lines {
		if after, ok := strings.CutPrefix(line, "# Command: "); ok {
			return after, nil
		}
	}

	return "", fmt.Errorf("not a valid conf-managed shim")
}

// ShimInfo represents information about a shim
type ShimInfo struct {
	Name    string
	Command string
	Path    string
}

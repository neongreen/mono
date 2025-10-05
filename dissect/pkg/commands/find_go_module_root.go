package commands

import (
	"fmt"
	"path/filepath"
	"strings"
)

// FindGoModuleRoot finds the Go module root by running 'go env GOMOD'.
func FindGoModuleRoot(filePath string) (string, error) {
	stdout, stderr, err := RunGoCommand("env", []string{"GOMOD"}, filepath.Dir(filePath), nil)
	if err != nil {
		return "", fmt.Errorf("failed to execute 'go env GOMOD': %w\nStderr: %s", err, stderr)
	}

	goModPath := strings.TrimSpace(stdout)
	if goModPath == "" {
		return "", fmt.Errorf("go env GOMOD returned empty output")
	}

	// The module root is the directory containing the go.mod file
	moduleRoot := filepath.Dir(goModPath)
	return moduleRoot, nil
}

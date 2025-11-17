package commands

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetModuleName returns the module name from go.mod
func GetModuleName(moduleRoot string) (string, error) {
	cmd := exec.Command("go", "list", "-m")
	cmd.Dir = moduleRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get module name: %w (output: %s)", err, string(output))
	}

	moduleName := strings.TrimSpace(string(output))
	if moduleName == "" {
		return "", fmt.Errorf("empty module name")
	}

	return moduleName, nil
}

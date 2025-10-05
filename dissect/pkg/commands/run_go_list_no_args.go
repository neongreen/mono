package commands

import (
	"fmt"
	"strings"
)

// RunGoListNoArgs runs 'go list' in the specified working directory.
func RunGoListNoArgs(workingDir string) (string, error) {
	stdout, stderr, err := RunGoCommand("list", nil, workingDir, nil)
	if err != nil {
		return "", fmt.Errorf("error running 'go list': %w\nStderr: %s", err, stderr)
	}
	return strings.TrimSpace(stdout), nil
}

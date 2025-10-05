package commands

import (
	"fmt"
)

// RunGoCommand executes a 'go' subcommand and captures stdout/stderr.
func RunGoCommand(subcommand string, args []string, workingDir string, env []string) (stdout, stderr string, err error) {
	fullArgs := []string{subcommand}
	fullArgs = append(fullArgs, args...)
	stdout, stderr, err = RunCommand("go", fullArgs, workingDir, env)
	if err != nil {
		return stdout, stderr, fmt.Errorf("error running 'go %s': %w\nStderr: %s", subcommand, err, stderr)
	}
	return stdout, stderr, nil
}

package command

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"
)

type CmdResult struct {
	Command    string
	ExitCode   int
	Stdout     string
	Stderr     string
	DurationMs int64
}

// RunCommand executes a shell command and captures its output
func RunCommand(command string) (*CmdResult, error) {
	startTime := time.Now()

	cmd := exec.Command("sh", "-c", command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(startTime)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to run command: %w", err)
		}
	}

	result := &CmdResult{
		Command:    command,
		ExitCode:   exitCode,
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		DurationMs: duration.Milliseconds(),
	}

	return result, nil
}

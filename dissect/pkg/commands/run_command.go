package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCommand executes a shell command and captures stdout/stderr.
func RunCommand(cmdName string, args []string, workingDir string, env []string) (stdout, stderr string, err error) {
	cmd := exec.Command(cmdName, args...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()
	if err != nil {
		return stdout, stderr, fmt.Errorf("command %s %s failed: %w\nStdout: %s\nStderr: %s", cmdName, strings.Join(args, " "), err, stdout, stderr)
	}
	return stdout, stderr, nil
}

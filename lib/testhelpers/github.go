package testhelpers

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

// SetupTestLogger sets up a test logger that captures output to a buffer.
// Returns the buffer containing the log output.
// The logger is automatically cleaned up after the test completes.
func SetupTestLogger(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	prev := slog.Default()
	slog.SetDefault(slog.New(handler))
	t.Cleanup(func() {
		slog.SetDefault(prev)
	})
	return &buf
}

// InstallGhStub installs a fake 'gh' CLI executable for testing.
// The stub outputs the specified stdout and exits with the given exit code.
// The fake executable is automatically cleaned up after the test completes.
func InstallGhStub(t *testing.T, stdout string, exitCode int) {
	t.Helper()
	dir := t.TempDir()

	var scriptName string
	var content string
	if runtime.GOOS == "windows" {
		scriptName = "gh.bat"
		content = "@echo off\r\n"
		if stdout != "" {
			content += "echo " + stdout + "\r\n"
		}
		content += "exit /b " + strconv.Itoa(exitCode) + "\r\n"
	} else {
		scriptName = "gh"
		content = "#!/bin/sh\n"
		if stdout != "" {
			content += "printf '%s\\n' '" + stdout + "'\n"
		}
		content += "exit " + strconv.Itoa(exitCode) + "\n"
	}

	mode := os.FileMode(0o755)
	if runtime.GOOS == "windows" {
		mode = 0o666
	}

	scriptPath := filepath.Join(dir, scriptName)
	if err := os.WriteFile(scriptPath, []byte(content), mode); err != nil {
		t.Fatalf("failed to write gh stub: %v", err)
	}

	originalPath := os.Getenv("PATH")
	if originalPath == "" {
		t.Setenv("PATH", dir)
		return
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+originalPath)
}

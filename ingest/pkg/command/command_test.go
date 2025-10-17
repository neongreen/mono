package command

import (
	"strings"
	"testing"
)

func TestRunCommandSuccess(t *testing.T) {
	result, err := RunCommand(`printf 'hello'`)
	if err != nil {
		t.Fatalf("RunCommand returned unexpected error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", result.ExitCode)
	}
	if result.Stdout != "hello" {
		t.Fatalf("expected stdout 'hello', got %q", result.Stdout)
	}
	if result.Stderr != "" {
		t.Fatalf("expected empty stderr, got %q", result.Stderr)
	}
	if result.DurationMs < 0 {
		t.Fatalf("expected non-negative duration, got %d", result.DurationMs)
	}
	if result.Command != `printf 'hello'` {
		t.Fatalf("expected command to be preserved, got %q", result.Command)
	}
}

func TestRunCommandNonZeroExit(t *testing.T) {
	result, err := RunCommand(`{ >&2 printf 'failure'; exit 42; }`)
	if err != nil {
		t.Fatalf("RunCommand returned unexpected error: %v", err)
	}

	if result.ExitCode != 42 {
		t.Fatalf("expected exit code 42, got %d", result.ExitCode)
	}
	if result.Stdout != "" {
		t.Fatalf("expected empty stdout for failing command, got %q", result.Stdout)
	}
	if strings.TrimSpace(result.Stderr) != "failure" {
		t.Fatalf("expected stderr 'failure', got %q", result.Stderr)
	}
}

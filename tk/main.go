package main

import (
	"bytes"
	"io"
	"os"
	"time"

	"github.com/neongreen/mono/tk/cmd"
	"github.com/neongreen/mono/tk/internal/invlog"
)

func main() {
	startTime := time.Now()

	// Capture stdout and stderr
	// Note: This approach replaces os.Stdout/os.Stderr with pipes, which means
	// TTY detection (isatty checks) will return false even when running interactively.
	// This is a known limitation. A proper fix would require platform-specific code
	// using syscall.Dup2 to preserve file descriptor numbers, which is complex and
	// error-prone. For tk's use case (debugging AI agent invocations), this limitation
	// is acceptable as colors/interactive features are less critical than capturing output.
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}

	rOut, wOut, err := os.Pipe()
	if err != nil {
		// If pipe creation fails, run without logging
		if err := cmd.Execute(); err != nil {
			os.Exit(1)
		}
		return
	}
	rErr, wErr, err2 := os.Pipe()
	if err2 != nil {
		// Clean up first pipe and run without logging
		rOut.Close()
		wOut.Close()
		if err := cmd.Execute(); err != nil {
			os.Exit(1)
		}
		return
	}

	os.Stdout = wOut
	os.Stderr = wErr

	// Copy output to both buffer and original stdout/stderr
	doneOut := make(chan bool)
	doneErr := make(chan bool)

	go func() {
		defer func() {
			rOut.Close()
			doneOut <- true
		}()
		io.Copy(io.MultiWriter(oldStdout, stdoutBuf), rOut)
	}()

	go func() {
		defer func() {
			rErr.Close()
			doneErr <- true
		}()
		io.Copy(io.MultiWriter(oldStderr, stderrBuf), rErr)
	}()

	// Execute the command
	exitCode := 0
	err = cmd.Execute()
	if err != nil {
		exitCode = 1
	}

	// Restore stdout and stderr
	wOut.Close()
	wErr.Close()
	<-doneOut
	<-doneErr
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Log the invocation
	duration := time.Since(startTime)
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}

	log := invlog.InvocationLog{
		Timestamp:  startTime,
		Command:    "tk",
		Args:       os.Args[1:],
		PID:        os.Getpid(),
		PPID:       os.Getppid(),
		User:       user,
		Success:    exitCode == 0,
		ExitCode:   exitCode,
		Stdout:     stdoutBuf.String(),
		Stderr:     stderrBuf.String(),
		DurationMs: duration.Milliseconds(),
	}

	// Write log entry (ignore errors to avoid disrupting the main command)
	_ = invlog.WriteLog(log)

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

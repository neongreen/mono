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
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	// Copy output to both buffer and original stdout/stderr
	doneOut := make(chan bool)
	doneErr := make(chan bool)

	go func() {
		io.Copy(io.MultiWriter(oldStdout, stdoutBuf), rOut)
		doneOut <- true
	}()

	go func() {
		io.Copy(io.MultiWriter(oldStderr, stderrBuf), rErr)
		doneErr <- true
	}()

	// Execute the command
	exitCode := 0
	err := cmd.Execute()
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

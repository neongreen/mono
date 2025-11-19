package main

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/golang-cz/devslog"
	"github.com/neongreen/mono/lib/version"
	"github.com/neongreen/mono/tk/cmd"
	"github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/invlog"
	"github.com/neongreen/mono/tk/internal/sanitycheck"
)

func main() {
	// Check if debug logging should be enabled by looking at command-line args
	// This needs to happen before running sanitycheck so debug logs are visible
	debugEnabled := false
	for _, arg := range os.Args {
		if arg == "--debug" {
			debugEnabled = true
			break
		}
	}

	if debugEnabled {
		slog.SetDefault(slog.New(
			devslog.NewHandler(os.Stderr, &devslog.Options{
				HandlerOptions:  &slog.HandlerOptions{Level: slog.LevelDebug},
				NewLineAfterLog: true,
			}),
		))
	}

	// Check if invocation logging should be skipped
	// Used by tkvscode extension to avoid log accumulation
	skipInvLog := os.Getenv("TK_SKIP_INVLOG") != ""

	// Run sanity check before executing any command
	// This is a silent, read-only check that compares reducer state to database state
	runSanityCheckIfPossible()

	if skipInvLog {
		// Run command directly without logging
		if err := cmd.RootCmd.Execute(); err != nil {
			os.Exit(1)
		}
		return
	}

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
		if err := cmd.RootCmd.Execute(); err != nil {
			os.Exit(1)
		}
		return
	}
	rErr, wErr, err2 := os.Pipe()
	if err2 != nil {
		// Clean up first pipe and run without logging
		rOut.Close()
		wOut.Close()
		if err := cmd.RootCmd.Execute(); err != nil {
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
	err = cmd.RootCmd.Execute()
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

	// Get binary path (os.Args[0] may be relative, try to get absolute path)
	binaryPath := os.Args[0]
	if absPath, err := os.Executable(); err == nil {
		binaryPath = absPath
	}

	log := invlog.InvocationLog{
		SchemaVersion: invlog.SchemaVersion,
		TkVersion:     version.Version,
		TkCommit:      version.GitCommit,
		Timestamp:     startTime.UnixNano(),
		Command:       "tk",
		Args:          os.Args[1:],
		PID:           os.Getpid(),
		PPID:          os.Getppid(),
		User:          user,
		Success:       exitCode == 0,
		ExitCode:      exitCode,
		Stdout:        stdoutBuf.String(),
		Stderr:        stderrBuf.String(),
		DurationMs:    duration.Milliseconds(),

		// Schema v2 fields (tk-81)
		TkCommitFull:  version.GitCommitFull,
		TkCommitTime:  version.CommitTime,
		TkGitModified: version.GitModified,
		TkBinaryPath:  binaryPath,
	}

	// Write log entry to rotating JSONL log (ignore errors to avoid disrupting the main command)
	_ = invlog.WriteLog(log)

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

// runSanityCheckIfPossible attempts to run the sanity check if the database exists.
// This function silently fails if there are any errors (database doesn't exist, etc.)
// to avoid disrupting normal program operation.
func runSanityCheckIfPossible() {
	// Skip sanity check for certain commands that don't need it
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		// Skip for commands that don't access the database
		switch cmd {
		case "init", "version", "help", "--help", "-h", "--version", "-v":
			return
		}
	}

	// Try to open the database
	db, err := database.OpenExistingDB()
	if err != nil {
		// Database doesn't exist or can't be opened - silently skip
		return
	}
	defer db.Close()

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		// Can't load config - silently skip
		return
	}

	// Run the sanity check
	// If differences are found, it will print a warning and write a diff file
	_ = sanitycheck.RunSanityCheck(db, cfg)
}

package jobs

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/neongreen/mono/ingest/pkg/command"
	"github.com/neongreen/mono/ingest/pkg/database"
)

// RunCommand executes a shell command and stores its output.
func RunCommand(ctx context.Context, out io.Writer, opts CommandOptions) (Result, error) {
	if out == nil {
		out = os.Stdout
	}
	_ = ctx

	if opts.Command == "" {
		return Result{}, fmt.Errorf("command is empty")
	}

	fmt.Fprintf(out, "Running command: %s\n", opts.Command)

	db, err := database.Open()
	if err != nil {
		return Result{}, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	runID, err := db.CreateRun(opts.Command, "cmd")
	if err != nil {
		return Result{}, fmt.Errorf("failed to create run: %w", err)
	}

	result := Result{RunID: runID}
	runStatus := "failed"
	defer func() {
		_ = db.FinishRun(runID, runStatus)
	}()

	fmt.Fprintf(out, "Started ingestion run #%d\n", runID)

	res, err := command.RunCommand(opts.Command)
	if err != nil {
		return Result{}, err
	}

	fmt.Fprintf(out, "Command completed with exit code: %d (took %dms)\n", res.ExitCode, res.DurationMs)
	if len(res.Stdout) > 0 {
		fmt.Fprintf(out, "stdout:\n%s\n", res.Stdout)
	}
	if len(res.Stderr) > 0 {
		fmt.Fprintf(out, "stderr:\n%s\n", res.Stderr)
	}

	if err := db.CreateCmdRun(runID, opts.Command, res.ExitCode, res.Stdout, res.Stderr, res.DurationMs); err != nil {
		return Result{}, fmt.Errorf("failed to store command run: %w", err)
	}

	if err := db.UpdateRunItemCount(runID); err != nil {
		return Result{}, fmt.Errorf("failed to update run item count: %w", err)
	}

	runStatus = "completed"
	lineCount := countLines(res.Stdout) + countLines(res.Stderr)
	result.ItemCount = lineCount
	result.Details = map[string]int{
		"lines": lineCount,
	}

	fmt.Fprintln(out, "Ingestion completed successfully!")
	return result, nil
}

func countLines(s string) int {
	if s == "" {
		return 0
	}
	count := 0
	for _, ch := range s {
		if ch == '\n' {
			count++
		}
	}
	if s[len(s)-1] != '\n' {
		count++
	}
	return count
}

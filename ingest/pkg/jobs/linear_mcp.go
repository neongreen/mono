package jobs

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/neongreen/mono/ingest/pkg/database"
	"github.com/neongreen/mono/ingest/pkg/linear"
	mcppkg "github.com/neongreen/mono/ingest/pkg/mcp"
)

// RunLinearMCP ingests issues from the Linear MCP server.
func RunLinearMCP(ctx context.Context, out io.Writer, cfg mcppkg.Config) (Result, error) {
	if out == nil {
		out = os.Stdout
	}

	client, err := mcppkg.NewClient(cfg)
	if err != nil {
		return Result{}, fmt.Errorf("failed to create MCP client: %w", err)
	}

	connectCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	session, err := client.Connect(connectCtx)
	if err != nil {
		return Result{}, fmt.Errorf("failed to connect to MCP server: %w", err)
	}
	defer session.Close()

	db, err := database.Open()
	if err != nil {
		return Result{}, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	runPath := "linear"
	if cfg.Endpoint != "" {
		runPath = fmt.Sprintf("linear:%s", cfg.Endpoint)
	}

	runID, err := db.CreateRun(runPath, "linear")
	if err != nil {
		return Result{}, fmt.Errorf("failed to create run: %w", err)
	}

	result := Result{RunID: runID}
	runStatus := "failed"
	defer func() {
		_ = db.FinishRun(runID, runStatus)
	}()

	fmt.Fprintf(out, "Started ingestion run #%d\n", runID)

	count, err := linear.IngestIssues(ctx, db, runID, session)
	if err != nil {
		return Result{}, err
	}

	if err := db.UpdateRunItemCount(runID); err != nil {
		return Result{}, fmt.Errorf("failed to update run item count: %w", err)
	}

	runStatus = "completed"
	result.ItemCount = count
	result.Details = map[string]int{"issues": count}
	return result, nil
}

package jobs

import (
	"context"
	"fmt"
	"io"
	"os"

	"ingest/pkg/database"
	"ingest/pkg/githubmcp"
	mcppkg "ingest/pkg/mcp"
)

// RunGitHubMCP ingests issues and pull requests from the GitHub MCP server.
func RunGitHubMCP(ctx context.Context, out io.Writer, cfg mcppkg.Config, opts GitHubOptions) (Result, error) {
	if out == nil {
		out = os.Stdout
	}

	if opts.Owner == "" || opts.Repo == "" {
		return Result{}, fmt.Errorf("owner and repo must be provided")
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

	repoSpec := fmt.Sprintf("%s/%s", opts.Owner, opts.Repo)
	runID, err := db.CreateRun(repoSpec, "github")
	if err != nil {
		return Result{}, fmt.Errorf("failed to create run: %w", err)
	}

	result := Result{RunID: runID}
	runStatus := "failed"
	defer func() {
		_ = db.FinishRun(runID, runStatus)
	}()

	fmt.Fprintf(out, "Started ingestion run #%d\n", runID)

	summary, err := githubmcp.IngestRepository(ctx, db, runID, session, opts.Owner, opts.Repo)
	if err != nil {
		return Result{}, err
	}

	if err := db.UpdateRunItemCount(runID); err != nil {
		return Result{}, fmt.Errorf("failed to update run item count: %w", err)
	}

	runStatus = "completed"
	result.ItemCount = summary.Issues + summary.PullRequests
	result.Details = map[string]int{
		"issues":        summary.Issues,
		"issueComments": summary.IssueComments,
		"pullRequests":  summary.PullRequests,
		"prComments":    summary.PullRequestComments,
	}

	return result, nil
}

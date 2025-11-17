package integration

import (
	"context"
	"os"
	"testing"

	"github.com/neongreen/mono/ingest/pkg/database"
	"github.com/neongreen/mono/ingest/pkg/jobs"
	"github.com/neongreen/mono/ingest/pkg/testutil"
)

// TestGitHubLiveIngestion exercises the GitHub REST ingestion against
// a real repository. It is skipped unless INGEST_GITHUB_LIVE is set to
// an owner/repo string (e.g. "octocat/Hello-World").
func TestGitHubLiveIngestion(t *testing.T) {
	repoSpec := os.Getenv("INGEST_GITHUB_LIVE")
	if repoSpec == "" {
		t.Skip("set INGEST_GITHUB_LIVE=owner/repo to run live GitHub ingestion test")
	}
	owner, repo, ok := splitRepo(repoSpec)
	if !ok {
		t.Fatalf("INGEST_GITHUB_LIVE must be owner/repo, got %q", repoSpec)
	}

	testutil.WithTempHome(t)

	result, err := jobs.RunGitHub(context.Background(), nil, jobs.GitHubOptions{Owner: owner, Repo: repo})
	if err != nil {
		t.Fatalf("RunGitHub failed: %v", err)
	}
	if result.ItemCount == 0 {
		t.Fatalf("expected at least one item ingested")
	}

	db, err := database.Open()
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT node_id, html_url, comment_count, reaction_total, participants_count FROM github_issues ORDER BY number LIMIT 1")
	if err != nil {
		t.Fatalf("query issues: %v", err)
	}
	if len(rows) == 0 {
		t.Fatalf("expected at least one issue row")
	}
	row := rows[0]
	if row["node_id"] == nil || row["html_url"] == nil {
		t.Fatalf("expected node_id/html_url to be populated: %+v", row)
	}
	if _, ok := row["comment_count"]; !ok {
		t.Fatalf("expected comment_count column in result")
	}
	if _, ok := row["reaction_total"]; !ok {
		t.Fatalf("expected reaction_total column in result")
	}
	if _, ok := row["participants_count"]; !ok {
		t.Fatalf("expected participants_count column in result")
	}

	reactions, err := db.Query("SELECT COUNT(*) AS total FROM github_comment_reactions")
	if err != nil {
		t.Fatalf("query reactions: %v", err)
	}
	if len(reactions) == 0 {
		t.Fatalf("expected reaction count row")
	}
}

func splitRepo(spec string) (string, string, bool) {
	for i := 0; i < len(spec); i++ {
		if spec[i] == '/' {
			if i == 0 || i == len(spec)-1 {
				return "", "", false
			}
			return spec[:i], spec[i+1:], true
		}
	}
	return "", "", false
}

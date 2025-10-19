package githubmcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"ingest/pkg/database"

	gh "github.com/google/go-github/v61/github"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type fakeSession struct {
	responses map[string][]*sdkmcp.CallToolResult
	errors    map[string]error
	calls     []string
}

func newFakeSession() *fakeSession {
	return &fakeSession{
		responses: make(map[string][]*sdkmcp.CallToolResult),
		errors:    make(map[string]error),
	}
}

func (f *fakeSession) enqueue(tool string, args map[string]any, payload any) {
	key := responseKey(tool, args)
	data, _ := json.Marshal(payload)
	f.responses[key] = append(f.responses[key], textResult(string(data)))
}

func (f *fakeSession) enqueueError(tool string, args map[string]any, err error) {
	key := responseKey(tool, args)
	f.errors[key] = err
}

func (f *fakeSession) CallTool(_ context.Context, tool string, args map[string]any) (*sdkmcp.CallToolResult, error) {
	key := responseKey(tool, args)
	f.calls = append(f.calls, key)
	if err, ok := f.errors[key]; ok {
		return nil, err
	}
	queue := f.responses[key]
	if len(queue) == 0 {
		return nil, fmt.Errorf("no fake response for %s", key)
	}
	res := queue[0]
	f.responses[key] = queue[1:]
	return res, nil
}

func responseKey(tool string, args map[string]any) string {
	type pair struct {
		k string
		v any
	}
	var parts []pair
	for k, v := range args {
		parts = append(parts, pair{k: k, v: v})
	}
	// deterministic order
	for i := 0; i < len(parts); i++ {
		for j := i + 1; j < len(parts); j++ {
			if parts[j].k < parts[i].k {
				parts[i], parts[j] = parts[j], parts[i]
			}
		}
	}
	result := tool
	for _, p := range parts {
		result += fmt.Sprintf("|%s=%v", p.k, p.v)
	}
	return result
}

func textResult(text string) *sdkmcp.CallToolResult {
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{
			&sdkmcp.TextContent{Text: text},
		},
	}
}

func TestIngestRepository(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	defer os.Unsetenv("HOME")

	db, err := database.Open()
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	runID, err := db.CreateRun("owner/repo", "github")
	if err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	session := newFakeSession()

	issue := &gh.Issue{
		Number:    githubInt(1),
		Title:     githubString("Demo issue"),
		Body:      githubString("issue body"),
		State:     githubString("open"),
		User:      &gh.User{Login: githubString("octocat")},
		CreatedAt: &gh.Timestamp{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		UpdatedAt: &gh.Timestamp{Time: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)},
		Labels: []*gh.Label{
			{Name: githubString("bug")},
		},
	}

	session.enqueue("list_issues", map[string]any{"owner": "owner", "repo": "repo", "perPage": 100}, map[string]any{
		"issues":   []*gh.Issue{issue},
		"pageInfo": map[string]any{"hasNextPage": false, "endCursor": ""},
	})

	issueComment := &gh.IssueComment{
		ID:        githubInt64(11),
		Body:      githubString("looks good"),
		User:      &gh.User{Login: githubString("reviewer")},
		CreatedAt: &gh.Timestamp{Time: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)},
		UpdatedAt: &gh.Timestamp{Time: time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)},
	}
	session.enqueue("get_issue_comments", map[string]any{
		"owner": "owner", "repo": "repo", "issue_number": 1, "page": 1, "perPage": 100,
	}, []*gh.IssueComment{issueComment})
	session.enqueue("get_issue_comments", map[string]any{
		"owner": "owner", "repo": "repo", "issue_number": 1, "page": 2, "perPage": 100,
	}, []*gh.IssueComment{})

	pr := &gh.PullRequest{
		Number:    githubInt(2),
		Title:     githubString("Demo PR"),
		Body:      githubString("pr body"),
		State:     githubString("open"),
		User:      &gh.User{Login: githubString("octocat")},
		CreatedAt: &gh.Timestamp{Time: time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)},
		UpdatedAt: &gh.Timestamp{Time: time.Date(2024, 1, 6, 0, 0, 0, 0, time.UTC)},
		Base:      &gh.PullRequestBranch{Ref: githubString("main")},
		Head:      &gh.PullRequestBranch{Ref: githubString("feature")},
	}
	session.enqueue("list_pull_requests", map[string]any{"owner": "owner", "repo": "repo", "page": 1, "perPage": 100}, []*gh.PullRequest{pr})
	session.enqueue("list_pull_requests", map[string]any{"owner": "owner", "repo": "repo", "page": 2, "perPage": 100}, []*gh.PullRequest{})

	session.enqueue("get_issue_comments", map[string]any{
		"owner": "owner", "repo": "repo", "issue_number": 2, "page": 1, "perPage": 100,
	}, []*gh.IssueComment{})

	reviewComment := &gh.PullRequestComment{
		ID:        githubInt64(21),
		Body:      githubString("nit"),
		User:      &gh.User{Login: githubString("maintainer")},
		CreatedAt: &gh.Timestamp{Time: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)},
		UpdatedAt: &gh.Timestamp{Time: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)},
	}
	session.enqueue("pull_request_read", map[string]any{
		"method": "get_review_comments", "owner": "owner", "repo": "repo", "pullNumber": 2, "page": 1, "perPage": 100,
	}, []*gh.PullRequestComment{reviewComment})
	session.enqueue("pull_request_read", map[string]any{
		"method": "get_review_comments", "owner": "owner", "repo": "repo", "pullNumber": 2, "page": 2, "perPage": 100,
	}, []*gh.PullRequestComment{})

	summary, err := IngestRepository(context.Background(), db, runID, session, "owner", "repo")
	if err != nil {
		t.Fatalf("ingest failed: %v", err)
	}

	if summary.Issues != 1 || summary.IssueComments != 1 || summary.PullRequests != 1 || summary.PullRequestComments != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}

	rows, err := db.Query("SELECT number, title FROM github_issues")
	if err != nil {
		t.Fatalf("query issues failed: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 issue row, got %d", len(rows))
	}

	rows, err = db.Query("SELECT number, title FROM github_prs")
	if err != nil {
		t.Fatalf("query prs failed: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 pr row, got %d", len(rows))
	}

	rows, err = db.Query("SELECT COUNT(*) AS count FROM github_comments")
	if err != nil {
		t.Fatalf("query comments failed: %v", err)
	}
	if rows[0]["count"].(int64) != 2 {
		t.Fatalf("expected 2 comments, got %v", rows[0]["count"])
	}

	runs, err := db.Query(fmt.Sprintf("SELECT item_count FROM runs WHERE id = %d", runID))
	if err != nil {
		t.Fatalf("query runs failed: %v", err)
	}
	if runs[0]["item_count"].(int64) != 2 {
		t.Fatalf("expected item_count 2, got %v", runs[0]["item_count"])
	}
}

func githubString(v string) *string {
	return &v
}

func githubInt(v int) *int {
	return &v
}

func githubInt64(v int64) *int64 {
	return &v
}

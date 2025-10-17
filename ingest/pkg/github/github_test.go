package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestIssueStructure(t *testing.T) {
	// Test that we can create and access Issue struct fields
	now := time.Now()
	issue := Issue{
		Number:    1,
		Title:     "Test Issue",
		Body:      "Test body",
		State:     "open",
		User:      User{Login: "testuser"},
		CreatedAt: now,
		UpdatedAt: now,
		ClosedAt:  nil,
		Labels:    []Label{{Name: "bug"}},
		Assignees: []User{{Login: "assignee1"}},
		Milestone: &Milestone{Title: "v1.0"},
	}

	if issue.Number != 1 {
		t.Errorf("Expected issue number 1, got %d", issue.Number)
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got '%s'", issue.Title)
	}
	if issue.User.Login != "testuser" {
		t.Errorf("Expected user 'testuser', got '%s'", issue.User.Login)
	}
	if len(issue.Labels) != 1 {
		t.Errorf("Expected 1 label, got %d", len(issue.Labels))
	}
}

func TestPullRequestStructure(t *testing.T) {
	// Test that we can create and access PullRequest struct fields
	now := time.Now()
	pr := PullRequest{
		Number:             2,
		Title:              "Test PR",
		Body:               "Test PR body",
		State:              "open",
		User:               User{Login: "prauthor"},
		CreatedAt:          now,
		UpdatedAt:          now,
		ClosedAt:           nil,
		MergedAt:           nil,
		Merged:             false,
		Draft:              false,
		Base:               Branch{Ref: "main"},
		Head:               Branch{Ref: "feature"},
		Labels:             []Label{{Name: "enhancement"}},
		Assignees:          []User{{Login: "reviewer1"}},
		RequestedReviewers: []User{{Login: "reviewer2"}},
		Milestone:          nil,
	}

	if pr.Number != 2 {
		t.Errorf("Expected PR number 2, got %d", pr.Number)
	}
	if pr.Title != "Test PR" {
		t.Errorf("Expected title 'Test PR', got '%s'", pr.Title)
	}
	if pr.Base.Ref != "main" {
		t.Errorf("Expected base ref 'main', got '%s'", pr.Base.Ref)
	}
	if pr.Head.Ref != "feature" {
		t.Errorf("Expected head ref 'feature', got '%s'", pr.Head.Ref)
	}
}

func TestCommentStructure(t *testing.T) {
	// Test that we can create and access Comment struct fields
	now := time.Now()
	comment := Comment{
		ID:        12345,
		Body:      "Test comment",
		User:      User{Login: "commenter"},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if comment.ID != 12345 {
		t.Errorf("Expected comment ID 12345, got %d", comment.ID)
	}
	if comment.Body != "Test comment" {
		t.Errorf("Expected body 'Test comment', got '%s'", comment.Body)
	}
	if comment.User.Login != "commenter" {
		t.Errorf("Expected user 'commenter', got '%s'", comment.User.Login)
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
	if client.ctx == nil {
		t.Fatal("expected client context to be non-nil")
	}
	if client.gh == nil {
		t.Fatal("expected go-github client to be non-nil")
	}
}

func TestNewClientWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := NewClientWithContext(ctx)
	if client.ctx != ctx {
		t.Fatal("expected client to use provided context")
	}
	if client.gh == nil {
		t.Fatal("expected go-github client to be non-nil")
	}
}

func TestClientFetchIssuesFiltersPRsAndAggregatesPages(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/repos/test/repo/issues" {
			http.NotFound(w, r)
			return
		}

		page := r.URL.Query().Get("page")
		switch page {
		case "", "1":
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Link", fmt.Sprintf("<http://%s/repos/test/repo/issues?page=2>; rel=\"next\"", r.Host))
			payload := []map[string]any{
				newIssuePayload(1, "Issue 1", "body 1", "open", "alice", "bug"),
				{
					"number": 2,
					"title":  "PR disguised",
					"state":  "open",
					"user":   map[string]any{"login": "robot"},
					"pull_request": map[string]any{
						"url": "http://example/pr",
					},
					"created_at": time.Now().UTC().Format(time.RFC3339),
					"updated_at": time.Now().UTC().Format(time.RFC3339),
				},
			}
			_ = json.NewEncoder(w).Encode(payload)
		case "2":
			w.Header().Set("Content-Type", "application/json")
			payload := []map[string]any{
				newIssuePayload(3, "Issue 2", "body 2", "closed", "bob", "enhancement"),
			}
			_ = json.NewEncoder(w).Encode(payload)
		default:
			http.NotFound(w, r)
		}
	}

	client := newTestGitHubClient(t, handler)

	var progress []int
	issues, err := client.FetchIssues("test", "repo", "all", func(count int) {
		progress = append(progress, count)
	})
	if err != nil {
		t.Fatalf("FetchIssues returned error: %v", err)
	}

	if len(issues) != 2 {
		t.Fatalf("expected 2 issues (excluding PR items), got %d", len(issues))
	}
	if issues[0].Number != 1 || issues[1].Number != 3 {
		t.Fatalf("unexpected issue numbers: %+v", issues)
	}

	if len(progress) == 0 || progress[len(progress)-1] != 2 {
		t.Fatalf("expected progress callback to end at 2, got %v", progress)
	}
}

func TestClientFetchPullRequestsAggregatesPages(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/repos/test/repo/pulls" {
			http.NotFound(w, r)
			return
		}

		page := r.URL.Query().Get("page")
		switch page {
		case "", "1":
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Link", fmt.Sprintf("<http://%s/repos/test/repo/pulls?page=2>; rel=\"next\"", r.Host))
			payload := []map[string]any{
				newPRPayload(10, "PR 10", "charlie", "main", "feature-1"),
			}
			_ = json.NewEncoder(w).Encode(payload)
		case "2":
			w.Header().Set("Content-Type", "application/json")
			payload := []map[string]any{
				newPRPayload(11, "PR 11", "dana", "main", "feature-2"),
				newPRPayload(12, "PR 12", "erin", "main", "feature-3"),
			}
			_ = json.NewEncoder(w).Encode(payload)
		default:
			http.NotFound(w, r)
		}
	}

	client := newTestGitHubClient(t, handler)

	var progress []int
	prs, err := client.FetchPullRequests("test", "repo", "all", func(count int) {
		progress = append(progress, count)
	})
	if err != nil {
		t.Fatalf("FetchPullRequests returned error: %v", err)
	}

	if len(prs) != 3 {
		t.Fatalf("expected 3 pull requests, got %d", len(prs))
	}
	if prs[0].Number != 10 || prs[2].Number != 12 {
		t.Fatalf("unexpected PR numbers: %+v", prs)
	}
	if len(progress) == 0 || progress[len(progress)-1] != 3 {
		t.Fatalf("expected progress callback to end at 3, got %v", progress)
	}
}

func TestClientFetchPRCommentsCombinesIssueAndReview(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/issues/5/comments"):
			page := r.URL.Query().Get("page")
			switch page {
			case "", "1":
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Link", fmt.Sprintf("<http://%s/repos/test/repo/issues/5/comments?page=2>; rel=\"next\"", r.Host))
				payload := []map[string]any{
					newCommentPayload(1000, "issue comment 1", "frank"),
				}
				_ = json.NewEncoder(w).Encode(payload)
			case "2":
				w.Header().Set("Content-Type", "application/json")
				payload := []map[string]any{
					newCommentPayload(1001, "issue comment 2", "grace"),
				}
				_ = json.NewEncoder(w).Encode(payload)
			default:
				http.NotFound(w, r)
			}
		case strings.HasSuffix(r.URL.Path, "/pulls/5/comments"):
			w.Header().Set("Content-Type", "application/json")
			payload := []map[string]any{
				newCommentPayload(2000, "review comment", "heidi"),
			}
			_ = json.NewEncoder(w).Encode(payload)
		default:
			http.NotFound(w, r)
		}
	}

	client := newTestGitHubClient(t, handler)

	comments, err := client.FetchPRComments("test", "repo", 5)
	if err != nil {
		t.Fatalf("FetchPRComments returned error: %v", err)
	}

	if len(comments) != 3 {
		t.Fatalf("expected 3 comments (2 issue + 1 review), got %d", len(comments))
	}
	var ids []int64
	for _, c := range comments {
		ids = append(ids, c.ID)
	}
	expected := []int64{1000, 1001, 2000}
	if fmt.Sprintf("%v", ids) != fmt.Sprintf("%v", expected) {
		t.Fatalf("expected comment IDs %v, got %v", expected, ids)
	}
}

func TestClientFetchIssuesError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/repos/test/repo/issues" {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		http.NotFound(w, r)
	}

	client := newTestGitHubClient(t, handler)

	_, err := client.FetchIssues("test", "repo", "all", nil)
	if err == nil {
		t.Fatal("expected error from FetchIssues")
	}
	if !strings.Contains(err.Error(), "failed to list issues") {
		t.Fatalf("expected contextual error, got %v", err)
	}
}

func newIssuePayload(number int, title, body, state, user, label string) map[string]any {
	now := time.Now().UTC().Format(time.RFC3339)
	return map[string]any{
		"number":     number,
		"title":      title,
		"body":       body,
		"state":      state,
		"user":       map[string]any{"login": user},
		"created_at": now,
		"updated_at": now,
		"labels": []map[string]any{
			{"name": label},
		},
		"assignees": []map[string]any{
			{"login": user},
		},
	}
}

func newPRPayload(number int, title, user, base, head string) map[string]any {
	now := time.Now().UTC().Format(time.RFC3339)
	return map[string]any{
		"number":     number,
		"title":      title,
		"body":       title + " body",
		"state":      "open",
		"user":       map[string]any{"login": user},
		"created_at": now,
		"updated_at": now,
		"base":       map[string]any{"ref": base},
		"head":       map[string]any{"ref": head},
		"labels": []map[string]any{
			{"name": "label"},
		},
		"assignees": []map[string]any{
			{"login": user},
		},
		"requested_reviewers": []map[string]any{
			{"login": "reviewer"},
		},
	}
}

func newCommentPayload(id int64, body, user string) map[string]any {
	now := time.Now().UTC().Format(time.RFC3339)
	return map[string]any{
		"id":         id,
		"body":       body,
		"user":       map[string]any{"login": user},
		"created_at": now,
		"updated_at": now,
	}
}

func newTestGitHubClient(t *testing.T, handler http.HandlerFunc) *Client {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client := NewClient()
	baseURL, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatalf("parse baseURL: %v", err)
	}
	client.gh.BaseURL = baseURL
	client.gh.UploadURL = baseURL

	return client
}

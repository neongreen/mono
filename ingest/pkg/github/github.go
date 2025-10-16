package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Issue represents a GitHub issue
type Issue struct {
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	State     string     `json:"state"`
	User      User       `json:"user"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at"`
	Labels    []Label    `json:"labels"`
	Assignees []User     `json:"assignees"`
	Milestone *Milestone `json:"milestone"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number             int        `json:"number"`
	Title              string     `json:"title"`
	Body               string     `json:"body"`
	State              string     `json:"state"`
	User               User       `json:"user"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	ClosedAt           *time.Time `json:"closed_at"`
	MergedAt           *time.Time `json:"merged_at"`
	Merged             bool       `json:"merged"`
	Draft              bool       `json:"draft"`
	Base               Branch     `json:"base"`
	Head               Branch     `json:"head"`
	Labels             []Label    `json:"labels"`
	Assignees          []User     `json:"assignees"`
	RequestedReviewers []User     `json:"requested_reviewers"`
	Milestone          *Milestone `json:"milestone"`
}

// Comment represents a GitHub comment
type Comment struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	User      User      `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User represents a GitHub user
type User struct {
	Login string `json:"login"`
}

// Label represents a GitHub label
type Label struct {
	Name string `json:"name"`
}

// Milestone represents a GitHub milestone
type Milestone struct {
	Title string `json:"title"`
}

// Branch represents a GitHub branch reference
type Branch struct {
	Ref string `json:"ref"`
}

// Client is a GitHub API client
type Client struct {
	baseURL string
	token   string
	ctx     context.Context
}

// NewClient creates a new GitHub API client
func NewClient() *Client {
	token := getGitHubToken()
	return &Client{
		baseURL: "https://api.github.com",
		token:   token,
		ctx:     context.Background(),
	}
}

// getGitHubToken retrieves GitHub token from environment or gh CLI
func getGitHubToken() string {
	// Check GITHUB_TOKEN first
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token
	}
	// Try gh CLI
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// FetchIssues fetches all issues for a repository with pagination
func (c *Client) FetchIssues(owner, repo string, state string, progressCallback func(count int)) ([]Issue, error) {
	var allIssues []Issue
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("%s/repos/%s/%s/issues?state=%s&page=%d&per_page=%d", c.baseURL, owner, repo, state, page, perPage)
		req, err := http.NewRequestWithContext(c.ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issues from %s: %w", url, err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("GitHub API returned status %d for %s/%s (URL: %s): %s", resp.StatusCode, owner, repo, url, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var issues []Issue
		if err := json.Unmarshal(body, &issues); err != nil {
			return nil, fmt.Errorf("failed to parse issues response: %w", err)
		}

		// Filter out pull requests (they appear in issues endpoint but have pull_request field)
		var actualIssues []Issue
		for _, issue := range issues {
			// Check if it's a pull request by trying to unmarshal the pull_request field
			var raw map[string]interface{}
			if err := json.Unmarshal(body, &raw); err == nil {
				// Re-parse to check for pull_request field
				issueBytes, _ := json.Marshal(issue)
				var issueMap map[string]interface{}
				json.Unmarshal(issueBytes, &issueMap)
				if _, hasPR := issueMap["pull_request"]; !hasPR {
					actualIssues = append(actualIssues, issue)
				}
			}
		}

		// Re-parse properly to filter out PRs
		actualIssues = []Issue{}
		var rawIssues []map[string]interface{}
		if err := json.Unmarshal(body, &rawIssues); err == nil {
			for _, rawIssue := range rawIssues {
				if _, hasPR := rawIssue["pull_request"]; !hasPR {
					issueBytes, _ := json.Marshal(rawIssue)
					var issue Issue
					if json.Unmarshal(issueBytes, &issue) == nil {
						actualIssues = append(actualIssues, issue)
					}
				}
			}
		}

		allIssues = append(allIssues, actualIssues...)

		if progressCallback != nil {
			progressCallback(len(allIssues))
		}

		// Check if there are more pages
		if len(issues) < perPage {
			break
		}
		page++
	}

	return allIssues, nil
}

// FetchPullRequests fetches all pull requests for a repository with pagination
func (c *Client) FetchPullRequests(owner, repo string, state string, progressCallback func(count int)) ([]PullRequest, error) {
	var allPRs []PullRequest
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("%s/repos/%s/%s/pulls?state=%s&page=%d&per_page=%d", c.baseURL, owner, repo, state, page, perPage)
		req, err := http.NewRequestWithContext(c.ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch pull requests from %s: %w", url, err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("GitHub API returned status %d for %s/%s (URL: %s): %s", resp.StatusCode, owner, repo, url, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var prs []PullRequest
		if err := json.Unmarshal(body, &prs); err != nil {
			return nil, fmt.Errorf("failed to parse pull requests response: %w", err)
		}

		allPRs = append(allPRs, prs...)

		if progressCallback != nil {
			progressCallback(len(allPRs))
		}

		// Check if there are more pages
		if len(prs) < perPage {
			break
		}
		page++
	}

	return allPRs, nil
}

// FetchIssueComments fetches all comments for an issue
func (c *Client) FetchIssueComments(owner, repo string, issueNumber int) ([]Comment, error) {
	var allComments []Comment
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments?page=%d&per_page=%d", c.baseURL, owner, repo, issueNumber, page, perPage)
		req, err := http.NewRequestWithContext(c.ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch comments from %s: %w", url, err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("GitHub API returned status %d for issue %d (URL: %s): %s", resp.StatusCode, issueNumber, url, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var comments []Comment
		if err := json.Unmarshal(body, &comments); err != nil {
			return nil, fmt.Errorf("failed to parse comments response: %w", err)
		}

		allComments = append(allComments, comments...)

		// Check if there are more pages
		if len(comments) < perPage {
			break
		}
		page++
	}

	return allComments, nil
}

// FetchPRComments fetches all comments for a pull request (includes review comments)
func (c *Client) FetchPRComments(owner, repo string, prNumber int) ([]Comment, error) {
	// Fetch issue comments (general PR comments)
	issueComments, err := c.FetchIssueComments(owner, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR issue comments: %w", err)
	}

	// Fetch review comments (code review comments)
	reviewComments, err := c.fetchPRReviewComments(owner, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR review comments: %w", err)
	}

	return append(issueComments, reviewComments...), nil
}

// fetchPRReviewComments fetches review comments for a pull request
func (c *Client) fetchPRReviewComments(owner, repo string, prNumber int) ([]Comment, error) {
	var allComments []Comment
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/comments?page=%d&per_page=%d", c.baseURL, owner, repo, prNumber, page, perPage)
		req, err := http.NewRequestWithContext(c.ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch review comments from %s: %w", url, err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("GitHub API returned status %d for PR %d (URL: %s): %s", resp.StatusCode, prNumber, url, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var comments []Comment
		if err := json.Unmarshal(body, &comments); err != nil {
			return nil, fmt.Errorf("failed to parse review comments response: %w", err)
		}

		allComments = append(allComments, comments...)

		// Check if there are more pages
		if len(comments) < perPage {
			break
		}
		page++
	}

	return allComments, nil
}

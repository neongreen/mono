package github

import (
	"context"
	"fmt"
	"time"

	api "github.com/google/go-github/v61/github"
	"github.com/neongreen/mono/lib/ghclient"
)

const defaultPerPage = 100

// Issue represents a GitHub issue
type Issue struct {
	Number           int        `json:"number"`
	Title            string     `json:"title"`
	Body             string     `json:"body"`
	State            string     `json:"state"`
	User             User       `json:"user"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	ClosedAt         *time.Time `json:"closed_at"`
	Labels           []Label    `json:"labels"`
	Assignees        []User     `json:"assignees"`
	Milestone        *Milestone `json:"milestone"`
	NodeID           string     `json:"node_id"`
	ID               int64      `json:"id"`
	HTMLURL          string     `json:"html_url"`
	APIURL           string     `json:"api_url"`
	CommentsURL      string     `json:"comments_url"`
	EventsURL        string     `json:"events_url"`
	StateReason      string     `json:"state_reason"`
	Locked           bool       `json:"locked"`
	ActiveLockReason string     `json:"active_lock_reason"`
	Draft            bool       `json:"draft"`
	ClosedBy         string     `json:"closed_by"`
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

// Client wraps the shared GitHub client helpers for ingest.
type Client struct {
	ctx context.Context
	gh  *api.Client
}

// NewClient creates a GitHub API client with shared HTTP configuration.
func NewClient() *Client {
	return NewClientWithContext(context.Background())
}

// NewClientWithContext creates a GitHub API client using the provided context.
func NewClientWithContext(ctx context.Context) *Client {
	if ctx == nil {
		ctx = context.Background()
	}

	return &Client{
		ctx: ctx,
		gh:  ghclient.NewClient(ctx),
	}
}

// FetchIssues fetches all issues for a repository with pagination.
func (c *Client) FetchIssues(owner, repo, state string, progressCallback func(count int)) ([]Issue, error) {
	opts := &api.IssueListByRepoOptions{
		State: state,
		ListOptions: api.ListOptions{
			PerPage: defaultPerPage,
		},
	}

	var allIssues []Issue
	for {
		items, resp, err := c.gh.Issues.ListByRepo(c.ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list issues for %s/%s: %w", owner, repo, err)
		}

		for _, item := range items {
			if item == nil || item.IsPullRequest() {
				continue
			}
			allIssues = append(allIssues, convertIssue(item))
		}

		if progressCallback != nil {
			progressCallback(len(allIssues))
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allIssues, nil
}

// FetchPullRequests fetches all pull requests for a repository with pagination.
func (c *Client) FetchPullRequests(owner, repo, state string, progressCallback func(count int)) ([]PullRequest, error) {
	opts := &api.PullRequestListOptions{
		State: state,
		ListOptions: api.ListOptions{
			PerPage: defaultPerPage,
		},
	}

	var allPRs []PullRequest
	for {
		items, resp, err := c.gh.PullRequests.List(c.ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list pull requests for %s/%s: %w", owner, repo, err)
		}

		for _, item := range items {
			if item == nil {
				continue
			}
			allPRs = append(allPRs, convertPullRequest(item))
		}

		if progressCallback != nil {
			progressCallback(len(allPRs))
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPRs, nil
}

// FetchIssueComments fetches all comments for an issue.
func (c *Client) FetchIssueComments(owner, repo string, issueNumber int) ([]Comment, error) {
	opts := &api.IssueListCommentsOptions{
		ListOptions: api.ListOptions{
			PerPage: defaultPerPage,
		},
	}

	var allComments []Comment
	for {
		items, resp, err := c.gh.Issues.ListComments(c.ctx, owner, repo, issueNumber, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list comments for issue %d in %s/%s: %w", issueNumber, owner, repo, err)
		}

		for _, item := range items {
			if item == nil {
				continue
			}
			allComments = append(allComments, convertIssueComment(item))
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allComments, nil
}

// FetchPRComments fetches all comments for a pull request (includes review comments).
func (c *Client) FetchPRComments(owner, repo string, prNumber int) ([]Comment, error) {
	issueComments, err := c.FetchIssueComments(owner, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR issue comments for %s/%s#%d: %w", owner, repo, prNumber, err)
	}

	reviewComments, err := c.fetchPRReviewComments(owner, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR review comments for %s/%s#%d: %w", owner, repo, prNumber, err)
	}

	return append(issueComments, reviewComments...), nil
}

func (c *Client) fetchPRReviewComments(owner, repo string, prNumber int) ([]Comment, error) {
	opts := &api.PullRequestListCommentsOptions{
		ListOptions: api.ListOptions{
			PerPage: defaultPerPage,
		},
	}

	var allComments []Comment
	for {
		items, resp, err := c.gh.PullRequests.ListComments(c.ctx, owner, repo, prNumber, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list PR review comments for %s/%s#%d: %w", owner, repo, prNumber, err)
		}

		for _, item := range items {
			if item == nil {
				continue
			}
			allComments = append(allComments, convertReviewComment(item))
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allComments, nil
}

func convertIssue(src *api.Issue) Issue {
	var closedBy string
	if user := src.GetClosedBy(); user != nil {
		closedBy = user.GetLogin()
	}

	return Issue{
		Number:           src.GetNumber(),
		Title:            src.GetTitle(),
		Body:             src.GetBody(),
		State:            src.GetState(),
		User:             convertUser(src.User),
		CreatedAt:        convertTimestamp(src.GetCreatedAt()),
		UpdatedAt:        convertTimestamp(src.GetUpdatedAt()),
		ClosedAt:         cloneTimestampPtr(src.ClosedAt),
		Labels:           convertLabels(src.Labels),
		Assignees:        convertUsers(src.Assignees),
		Milestone:        convertMilestone(src.Milestone),
		NodeID:           src.GetNodeID(),
		ID:               src.GetID(),
		HTMLURL:          src.GetHTMLURL(),
		APIURL:           src.GetURL(),
		CommentsURL:      src.GetCommentsURL(),
		EventsURL:        src.GetEventsURL(),
		StateReason:      src.GetStateReason(),
		Locked:           src.GetLocked(),
		ActiveLockReason: src.GetActiveLockReason(),
		Draft:            src.GetDraft(),
		ClosedBy:         closedBy,
	}
}

func convertPullRequest(src *api.PullRequest) PullRequest {
	return PullRequest{
		Number:             src.GetNumber(),
		Title:              src.GetTitle(),
		Body:               src.GetBody(),
		State:              src.GetState(),
		User:               convertUser(src.User),
		CreatedAt:          convertTimestamp(src.GetCreatedAt()),
		UpdatedAt:          convertTimestamp(src.GetUpdatedAt()),
		ClosedAt:           cloneTimestampPtr(src.ClosedAt),
		MergedAt:           cloneTimestampPtr(src.MergedAt),
		Merged:             src.GetMerged(),
		Draft:              src.GetDraft(),
		Base:               convertBranch(src.Base),
		Head:               convertBranch(src.Head),
		Labels:             convertLabels(src.Labels),
		Assignees:          convertUsers(src.Assignees),
		RequestedReviewers: convertUsers(src.RequestedReviewers),
		Milestone:          convertMilestone(src.Milestone),
	}
}

func convertIssueComment(src *api.IssueComment) Comment {
	return Comment{
		ID:        src.GetID(),
		Body:      src.GetBody(),
		User:      convertUser(src.User),
		CreatedAt: convertTimestamp(src.GetCreatedAt()),
		UpdatedAt: convertTimestamp(src.GetUpdatedAt()),
	}
}

func convertReviewComment(src *api.PullRequestComment) Comment {
	return Comment{
		ID:        src.GetID(),
		Body:      src.GetBody(),
		User:      convertUser(src.User),
		CreatedAt: convertTimestamp(src.GetCreatedAt()),
		UpdatedAt: convertTimestamp(src.GetUpdatedAt()),
	}
}

func convertUser(src *api.User) User {
	if src == nil {
		return User{}
	}
	return User{Login: src.GetLogin()}
}

func convertUsers(items []*api.User) []User {
	out := make([]User, 0, len(items))
	for _, item := range items {
		out = append(out, convertUser(item))
	}
	return out
}

func convertLabels(items []*api.Label) []Label {
	out := make([]Label, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, Label{Name: item.GetName()})
	}
	return out
}

func convertMilestone(src *api.Milestone) *Milestone {
	if src == nil {
		return nil
	}
	result := &Milestone{Title: src.GetTitle()}
	return result
}

func convertBranch(src *api.PullRequestBranch) Branch {
	if src == nil {
		return Branch{}
	}
	return Branch{Ref: src.GetRef()}
}

func convertTimestamp(value api.Timestamp) time.Time {
	return value.Time
}

func cloneTimestampPtr(t *api.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	clone := t.Time
	return &clone
}

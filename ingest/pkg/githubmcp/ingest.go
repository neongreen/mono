package githubmcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	gh "github.com/google/go-github/v61/github"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/neongreen/mono/ingest/pkg/database"
)

// ToolCaller represents the MCP session subset required for GitHub ingestion.
type ToolCaller interface {
	CallTool(ctx context.Context, name string, args map[string]any) (*sdkmcp.CallToolResult, error)
}

// Summary captures counts produced during ingestion.
type Summary struct {
	Issues              int
	IssueComments       int
	PullRequests        int
	PullRequestComments int
}

// IngestRepository fetches issues, pull requests, and their comments from the GitHub MCP server
// and stores them in the ingest database.
func IngestRepository(ctx context.Context, db *database.Database, runID int64, session ToolCaller, owner, repo string) (Summary, error) {
	issues, err := fetchIssues(ctx, session, owner, repo)
	if err != nil {
		return Summary{}, err
	}

	summary := Summary{Issues: len(issues)}

	for _, issue := range issues {
		if issue == nil || issue.IsPullRequest() {
			continue
		}

		labels := joinLabelNames(issue.Labels)
		assignees := joinUsers(issue.Assignees)
		milestone := ""
		if issue.Milestone != nil {
			milestone = issue.Milestone.GetTitle()
		}

		comments, err := fetchIssueComments(ctx, session, owner, repo, issue.GetNumber())
		if err != nil {
			return Summary{}, fmt.Errorf("failed to fetch comments for issue #%d: %w", issue.GetNumber(), err)
		}

		reactionsTotal := 0
		if reactions := issue.GetReactions(); reactions != nil {
			reactionsTotal += reactions.GetTotalCount()
		}

		participants := make(map[string]struct{})
		addParticipant := func(login string) {
			login = strings.TrimSpace(login)
			if login == "" {
				return
			}
			participants[login] = struct{}{}
		}
		addParticipant(userLogin(issue.User))
		addParticipant(userLogin(issue.GetClosedBy()))
		for _, user := range issue.Assignees {
			addParticipant(userLogin(user))
		}

		for _, comment := range comments {
			if comment == nil {
				continue
			}
			if err := db.CreateGitHubComment(
				runID,
				"issue",
				issue.GetNumber(),
				comment.GetID(),
				userLogin(comment.User),
				comment.GetBody(),
				toTime(comment.GetCreatedAt()),
				toTime(comment.GetUpdatedAt()),
			); err != nil {
				return Summary{}, fmt.Errorf("failed to store comment %d for issue #%d: %w", comment.GetID(), issue.GetNumber(), err)
			}
			addParticipant(userLogin(comment.User))
			if comment.GetReactions() != nil {
				reactionsTotal += comment.GetReactions().GetTotalCount()
			}
		}
		summary.IssueComments += len(comments)

		if err := db.CreateGitHubIssue(database.GitHubIssueRecord{
			RunID:             runID,
			Number:            issue.GetNumber(),
			Title:             issue.GetTitle(),
			Body:              issue.GetBody(),
			State:             issue.GetState(),
			Author:            userLogin(issue.User),
			CreatedAt:         toTime(issue.GetCreatedAt()),
			UpdatedAt:         toTime(issue.GetUpdatedAt()),
			ClosedAt:          toTimePtr(issue.ClosedAt),
			Labels:            labels,
			Assignees:         assignees,
			Milestone:         milestone,
			NodeID:            issue.GetNodeID(),
			IssueID:           issue.GetID(),
			HTMLURL:           issue.GetHTMLURL(),
			APIURL:            issue.GetURL(),
			CommentsURL:       issue.GetCommentsURL(),
			EventsURL:         issue.GetEventsURL(),
			StateReason:       issue.GetStateReason(),
			Locked:            issue.GetLocked(),
			ActiveLockReason:  issue.GetActiveLockReason(),
			Draft:             issue.GetDraft(),
			ClosedBy:          userLogin(issue.GetClosedBy()),
			CommentCount:      len(comments),
			ReactionsTotal:    reactionsTotal,
			ParticipantsCount: len(participants),
		}); err != nil {
			return Summary{}, fmt.Errorf("failed to store issue #%d: %w", issue.GetNumber(), err)
		}
	}

	prs, err := fetchPullRequests(ctx, session, owner, repo)
	if err != nil {
		return Summary{}, err
	}
	summary.PullRequests = len(prs)

	for _, pr := range prs {
		if pr == nil {
			continue
		}

		labels := joinLabelNames(pr.Labels)
		assignees := joinUsers(pr.Assignees)
		reviewers := joinUsers(pr.RequestedReviewers)
		milestone := ""
		if pr.Milestone != nil {
			milestone = pr.Milestone.GetTitle()
		}

		if err := db.CreateGitHubPR(
			runID,
			pr.GetNumber(),
			pr.GetTitle(),
			pr.GetBody(),
			pr.GetState(),
			userLogin(pr.User),
			toTime(pr.GetCreatedAt()),
			toTime(pr.GetUpdatedAt()),
			toTimePtr(pr.ClosedAt),
			toTimePtr(pr.MergedAt),
			pr.GetMerged(),
			pr.GetDraft(),
			branchRef(pr.Base),
			branchRef(pr.Head),
			labels,
			assignees,
			reviewers,
			milestone,
		); err != nil {
			return Summary{}, fmt.Errorf("failed to store pull request #%d: %w", pr.GetNumber(), err)
		}

		issueComments, err := fetchIssueComments(ctx, session, owner, repo, pr.GetNumber())
		if err != nil {
			return Summary{}, fmt.Errorf("failed to fetch issue comments for pull request #%d: %w", pr.GetNumber(), err)
		}
		for _, comment := range issueComments {
			if comment == nil {
				continue
			}
			if err := db.CreateGitHubComment(
				runID,
				"pr",
				pr.GetNumber(),
				comment.GetID(),
				userLogin(comment.User),
				comment.GetBody(),
				toTime(comment.GetCreatedAt()),
				toTime(comment.GetUpdatedAt()),
			); err != nil {
				return Summary{}, fmt.Errorf("failed to store issue comment %d for pull request #%d: %w", comment.GetID(), pr.GetNumber(), err)
			}
		}
		summary.PullRequestComments += len(issueComments)

		reviewComments, err := fetchReviewComments(ctx, session, owner, repo, pr.GetNumber())
		if err != nil {
			return Summary{}, fmt.Errorf("failed to fetch review comments for pull request #%d: %w", pr.GetNumber(), err)
		}
		for _, comment := range reviewComments {
			if comment == nil {
				continue
			}
			if err := db.CreateGitHubComment(
				runID,
				"pr",
				pr.GetNumber(),
				comment.GetID(),
				userLogin(comment.User),
				comment.GetBody(),
				toTime(comment.GetCreatedAt()),
				toTime(comment.GetUpdatedAt()),
			); err != nil {
				return Summary{}, fmt.Errorf("failed to store review comment %d for pull request #%d: %w", comment.GetID(), pr.GetNumber(), err)
			}
		}
		summary.PullRequestComments += len(reviewComments)
	}

	if err := db.UpdateRunItemCount(runID); err != nil {
		return summary, fmt.Errorf("failed to update run item count: %w", err)
	}

	return summary, nil
}

func fetchIssues(ctx context.Context, session ToolCaller, owner, repo string) ([]*gh.Issue, error) {
	var issues []*gh.Issue
	cursor := ""

	for {
		args := map[string]any{
			"owner":   owner,
			"repo":    repo,
			"perPage": 100,
		}
		if cursor != "" {
			args["after"] = cursor
		}

		result, err := session.CallTool(ctx, "list_issues", args)
		if err != nil {
			return nil, fmt.Errorf("list_issues failed: %w", err)
		}

		text, err := extractText(result)
		if err != nil {
			return nil, fmt.Errorf("list_issues returned invalid content: %w", err)
		}

		var payload struct {
			Issues   []*gh.Issue `json:"issues"`
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
		}
		if err := json.Unmarshal([]byte(text), &payload); err != nil {
			return nil, fmt.Errorf("failed to decode list_issues payload: %w", err)
		}

		issues = append(issues, payload.Issues...)

		if !payload.PageInfo.HasNextPage || payload.PageInfo.EndCursor == "" {
			break
		}
		cursor = payload.PageInfo.EndCursor
	}

	return issues, nil
}

func fetchPullRequests(ctx context.Context, session ToolCaller, owner, repo string) ([]*gh.PullRequest, error) {
	const perPage = 100
	page := 1
	var prs []*gh.PullRequest

	for {
		args := map[string]any{
			"owner":   owner,
			"repo":    repo,
			"perPage": perPage,
			"page":    page,
		}

		result, err := session.CallTool(ctx, "list_pull_requests", args)
		if err != nil {
			return nil, fmt.Errorf("list_pull_requests failed: %w", err)
		}

		text, err := extractText(result)
		if err != nil {
			return nil, fmt.Errorf("list_pull_requests returned invalid content: %w", err)
		}

		var batch []*gh.PullRequest
		if err := json.Unmarshal([]byte(text), &batch); err != nil {
			return nil, fmt.Errorf("failed to decode pull request list: %w", err)
		}

		if len(batch) == 0 {
			break
		}
		prs = append(prs, batch...)

		if len(batch) < perPage {
			break
		}
		page++
	}

	return prs, nil
}

func fetchIssueComments(ctx context.Context, session ToolCaller, owner, repo string, issueNumber int) ([]*gh.IssueComment, error) {
	const perPage = 100
	page := 1
	var comments []*gh.IssueComment

	for {
		args := map[string]any{
			"owner":        owner,
			"repo":         repo,
			"issue_number": issueNumber,
			"page":         page,
			"perPage":      perPage,
		}

		result, err := session.CallTool(ctx, "get_issue_comments", args)
		if err != nil {
			return nil, fmt.Errorf("get_issue_comments failed: %w", err)
		}

		text, err := extractText(result)
		if err != nil {
			return nil, fmt.Errorf("get_issue_comments returned invalid content: %w", err)
		}

		var batch []*gh.IssueComment
		if err := json.Unmarshal([]byte(text), &batch); err != nil {
			return nil, fmt.Errorf("failed to decode issue comments: %w", err)
		}

		if len(batch) == 0 {
			break
		}
		comments = append(comments, batch...)

		if len(batch) < perPage {
			break
		}
		page++
	}

	return comments, nil
}

func fetchReviewComments(ctx context.Context, session ToolCaller, owner, repo string, prNumber int) ([]*gh.PullRequestComment, error) {
	const perPage = 100
	page := 1
	var comments []*gh.PullRequestComment

	for {
		args := map[string]any{
			"method":     "get_review_comments",
			"owner":      owner,
			"repo":       repo,
			"pullNumber": prNumber,
			"page":       page,
			"perPage":    perPage,
		}

		result, err := session.CallTool(ctx, "pull_request_read", args)
		if err != nil {
			return nil, fmt.Errorf("pull_request_read failed: %w", err)
		}

		text, err := extractText(result)
		if err != nil {
			return nil, fmt.Errorf("pull_request_read returned invalid content: %w", err)
		}

		var batch []*gh.PullRequestComment
		if err := json.Unmarshal([]byte(text), &batch); err != nil {
			return nil, fmt.Errorf("failed to decode review comments: %w", err)
		}

		if len(batch) == 0 {
			break
		}
		comments = append(comments, batch...)

		if len(batch) < perPage {
			break
		}
		page++
	}

	return comments, nil
}

func extractText(result *sdkmcp.CallToolResult) (string, error) {
	if result == nil {
		return "", errors.New("empty tool result")
	}
	if result.IsError {
		return "", errors.New("tool reported failure")
	}
	if len(result.Content) == 0 {
		return "", errors.New("tool response missing content")
	}

	if text, ok := result.Content[0].(*sdkmcp.TextContent); ok {
		return text.Text, nil
	}
	return "", fmt.Errorf("unexpected content type %T", result.Content[0])
}

func joinLabelNames(labels []*gh.Label) string {
	names := make([]string, 0, len(labels))
	for _, label := range labels {
		if label == nil {
			continue
		}
		if name := strings.TrimSpace(label.GetName()); name != "" {
			names = append(names, name)
		}
	}
	return strings.Join(names, ",")
}

func joinUsers(users []*gh.User) string {
	names := make([]string, 0, len(users))
	for _, user := range users {
		if user == nil {
			continue
		}
		if login := strings.TrimSpace(user.GetLogin()); login != "" {
			names = append(names, login)
		}
	}
	return strings.Join(names, ",")
}

func userLogin(user *gh.User) string {
	if user == nil {
		return ""
	}
	return user.GetLogin()
}

func branchRef(branch *gh.PullRequestBranch) string {
	if branch == nil {
		return ""
	}
	return branch.GetRef()
}

func toTimePtr(ts *gh.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.Time
	return &t
}

func toTime(ts gh.Timestamp) time.Time {
	return ts.Time
}

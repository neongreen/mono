package jobs

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"ingest/pkg/database"
	"ingest/pkg/github"
)

// RunGitHub ingests issues and pull requests using the GitHub REST API.
func RunGitHub(ctx context.Context, out io.Writer, opts GitHubOptions) (Result, error) {
	if out == nil {
		out = os.Stdout
	}

	if opts.Owner == "" || opts.Repo == "" {
		return Result{}, fmt.Errorf("owner and repo must be provided")
	}

	db, err := database.Open()
	if err != nil {
		return Result{}, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	repoSpec := fmt.Sprintf("%s/%s", opts.Owner, opts.Repo)
	fmt.Fprintf(out, "Ingesting GitHub repository: %s\n", repoSpec)

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

	client := github.NewClientWithContext(ctx)

	fmt.Fprintln(out, "Fetching issues...")
	issues, err := client.FetchIssues(opts.Owner, opts.Repo, "all", func(count int) {
		fmt.Fprintf(out, "Found %d issues so far...\r", count)
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to fetch issues: %w", err)
	}
	fmt.Fprintf(out, "\nFound %d issues total\n", len(issues))

	issueCommentsCount := 0
	for i, issue := range issues {
		if (i+1)%10 == 0 || i+1 == len(issues) {
			fmt.Fprintf(out, "Processing issue %d/%d...\r", i+1, len(issues))
		}

		var labelNames []string
		for _, label := range issue.Labels {
			labelNames = append(labelNames, label.Name)
		}

		var assigneeNames []string
		for _, assignee := range issue.Assignees {
			assigneeNames = append(assigneeNames, assignee.Login)
		}

		milestone := ""
		if issue.Milestone != nil {
			milestone = issue.Milestone.Title
		}

		if err := db.CreateGitHubIssue(database.GitHubIssueRecord{
			RunID:            runID,
			Number:           issue.Number,
			Title:            issue.Title,
			Body:             issue.Body,
			State:            issue.State,
			Author:           issue.User.Login,
			CreatedAt:        issue.CreatedAt,
			UpdatedAt:        issue.UpdatedAt,
			ClosedAt:         issue.ClosedAt,
			Labels:           strings.Join(labelNames, ","),
			Assignees:        strings.Join(assigneeNames, ","),
			Milestone:        milestone,
			NodeID:           issue.NodeID,
			IssueID:          issue.ID,
			HTMLURL:          issue.HTMLURL,
			APIURL:           issue.APIURL,
			CommentsURL:      issue.CommentsURL,
			EventsURL:        issue.EventsURL,
			StateReason:      issue.StateReason,
			Locked:           issue.Locked,
			ActiveLockReason: issue.ActiveLockReason,
			Draft:            issue.Draft,
			ClosedBy:         issue.ClosedBy,
		}); err != nil {
			return Result{}, fmt.Errorf("failed to create issue #%d: %w", issue.Number, err)
		}

		comments, err := client.FetchIssueComments(opts.Owner, opts.Repo, issue.Number)
		if err != nil {
			return Result{}, fmt.Errorf("failed to fetch comments for issue #%d: %w", issue.Number, err)
		}
		issueCommentsCount += len(comments)

		for _, comment := range comments {
			if err := db.CreateGitHubComment(
				runID,
				"issue",
				issue.Number,
				comment.ID,
				comment.User.Login,
				comment.Body,
				comment.CreatedAt,
				comment.UpdatedAt,
			); err != nil {
				return Result{}, fmt.Errorf("failed to create comment %d for issue #%d: %w", comment.ID, issue.Number, err)
			}
		}
	}
	if len(issues) > 0 {
		fmt.Fprintf(out, "\nProcessed %d issues with their comments\n", len(issues))
	}

	fmt.Fprintln(out, "Fetching pull requests...")
	prs, err := client.FetchPullRequests(opts.Owner, opts.Repo, "all", func(count int) {
		fmt.Fprintf(out, "Found %d pull requests so far...\r", count)
	})
	if err != nil {
		return Result{}, fmt.Errorf("failed to fetch pull requests: %w", err)
	}
	fmt.Fprintf(out, "\nFound %d pull requests total\n", len(prs))

	prCommentCount := 0
	for i, pr := range prs {
		if (i+1)%10 == 0 || i+1 == len(prs) {
			fmt.Fprintf(out, "Processing PR %d/%d...\r", i+1, len(prs))
		}

		var labelNames []string
		for _, label := range pr.Labels {
			labelNames = append(labelNames, label.Name)
		}

		var assigneeNames []string
		for _, assignee := range pr.Assignees {
			assigneeNames = append(assigneeNames, assignee.Login)
		}

		var reviewerNames []string
		for _, reviewer := range pr.RequestedReviewers {
			reviewerNames = append(reviewerNames, reviewer.Login)
		}

		milestone := ""
		if pr.Milestone != nil {
			milestone = pr.Milestone.Title
		}

		if err := db.CreateGitHubPR(
			runID,
			pr.Number,
			pr.Title,
			pr.Body,
			pr.State,
			pr.User.Login,
			pr.CreatedAt,
			pr.UpdatedAt,
			pr.ClosedAt,
			pr.MergedAt,
			pr.Merged,
			pr.Draft,
			pr.Base.Ref,
			pr.Head.Ref,
			strings.Join(labelNames, ","),
			strings.Join(assigneeNames, ","),
			strings.Join(reviewerNames, ","),
			milestone,
		); err != nil {
			return Result{}, fmt.Errorf("failed to create pull request #%d: %w", pr.Number, err)
		}

		comments, err := client.FetchPRComments(opts.Owner, opts.Repo, pr.Number)
		if err != nil {
			return Result{}, fmt.Errorf("failed to fetch comments for PR #%d: %w", pr.Number, err)
		}
		prCommentCount += len(comments)

		for _, comment := range comments {
			if err := db.CreateGitHubComment(
				runID,
				"pr",
				pr.Number,
				comment.ID,
				comment.User.Login,
				comment.Body,
				comment.CreatedAt,
				comment.UpdatedAt,
			); err != nil {
				return Result{}, fmt.Errorf("failed to create comment %d for PR #%d: %w", comment.ID, pr.Number, err)
			}
		}
	}
	if len(prs) > 0 {
		fmt.Fprintf(out, "\nProcessed %d pull requests with their comments\n", len(prs))
	}

	if err := db.UpdateRunItemCount(runID); err != nil {
		return Result{}, fmt.Errorf("failed to update run item count: %w", err)
	}

	runStatus = "completed"
	result.ItemCount = len(issues) + len(prs)
	result.Details = map[string]int{
		"issues":        len(issues),
		"issueComments": issueCommentsCount,
		"pullRequests":  len(prs),
		"prComments":    prCommentCount,
	}

	fmt.Fprintln(out, "Ingestion completed successfully!")
	return result, nil
}

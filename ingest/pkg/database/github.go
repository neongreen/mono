package database

import (
	"fmt"
	"time"
)

// CreateGitHubIssue creates a new GitHub issue record
func (d *Database) CreateGitHubIssue(record GitHubIssueRecord) error {
	_, err := d.db.Exec(
		`INSERT INTO github_issues (
			run_id,
			number,
			title,
			body,
			state,
			author,
			created_at,
			updated_at,
			closed_at,
			labels,
			assignees,
			milestone,
			node_id,
			issue_id,
			html_url,
			api_url,
			comments_url,
			events_url,
			state_reason,
			locked,
			active_lock_reason,
			draft,
			closed_by,
			comment_count,
			reaction_total,
			participants_count
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.RunID,
		record.Number,
		record.Title,
		record.Body,
		record.State,
		record.Author,
		record.CreatedAt,
		record.UpdatedAt,
		record.ClosedAt,
		record.Labels,
		record.Assignees,
		record.Milestone,
		record.NodeID,
		record.IssueID,
		record.HTMLURL,
		record.APIURL,
		record.CommentsURL,
		record.EventsURL,
		record.StateReason,
		record.Locked,
		record.ActiveLockReason,
		record.Draft,
		record.ClosedBy,
		record.CommentCount,
		record.ReactionsTotal,
		record.ParticipantsCount,
	)
	if err != nil {
		return fmt.Errorf("failed to create github issue: %w", err)
	}

	return nil
}

// CreateGitHubCommentReaction stores a reaction for a GitHub comment.
func (d *Database) CreateGitHubCommentReaction(record GitHubCommentReaction) error {
	_, err := d.db.Exec(
		`INSERT INTO github_comment_reactions (
			run_id,
			item_type,
			item_number,
			comment_id,
			reactor,
			content
		) VALUES (?, ?, ?, ?, ?, ?)`,
		record.RunID,
		record.ItemType,
		record.ItemNumber,
		record.CommentID,
		record.Reactor,
		record.Content,
	)
	if err != nil {
		return fmt.Errorf("failed to create github comment reaction: %w", err)
	}
	return nil
}

// CreateGitHubPR creates a new GitHub pull request record
func (d *Database) CreateGitHubPR(runID int64, number int, title, body, state, author string, createdAt, updatedAt time.Time, closedAt, mergedAt *time.Time, merged, draft bool, baseBranch, headBranch, labels, assignees, reviewers, milestone string) error {
	_, err := d.db.Exec(
		"INSERT INTO github_prs (run_id, number, title, body, state, author, created_at, updated_at, closed_at, merged_at, merged, draft, base_branch, head_branch, labels, assignees, reviewers, milestone) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		runID,
		number,
		title,
		body,
		state,
		author,
		createdAt,
		updatedAt,
		closedAt,
		mergedAt,
		merged,
		draft,
		baseBranch,
		headBranch,
		labels,
		assignees,
		reviewers,
		milestone,
	)
	if err != nil {
		return fmt.Errorf("failed to create github pull request: %w", err)
	}

	return nil
}

// CreateGitHubComment creates a new GitHub comment record
func (d *Database) CreateGitHubComment(runID int64, itemType string, itemNumber int, commentID int64, author, body string, createdAt, updatedAt time.Time) error {
	_, err := d.db.Exec(
		"INSERT INTO github_comments (run_id, item_type, item_number, comment_id, author, body, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		runID,
		itemType,
		itemNumber,
		commentID,
		author,
		body,
		createdAt,
		updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create github comment: %w", err)
	}

	return nil
}

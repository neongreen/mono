package github

import (
	"context"
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

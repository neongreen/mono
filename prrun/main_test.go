package main

import (
	"testing"
)

func TestExtractProjectFromTag(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want string
	}{
		{
			name: "PR release with double dash",
			tag:  "dissect--pr-123.1",
			want: "dissect",
		},
		{
			name: "main release",
			tag:  "markdown-format--main.5",
			want: "markdown-format",
		},
		{
			name: "project with hyphen",
			tag:  "my-project--pr-1.1",
			want: "my-project",
		},
		{
			name: "single part tag",
			tag:  "project",
			want: "project",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractProjectFromTag(tt.tag)
			if got != tt.want {
				t.Errorf("extractProjectFromTag(%q) = %q, want %q", tt.tag, got, tt.want)
			}
		})
	}
}

func TestFindAllPRReleases(t *testing.T) {
	// This test requires network access and GitHub API
	// We'll just verify the function signature and basic error handling
	_, err := findAllPRReleases("", "", 0)
	if err == nil {
		t.Error("findAllPRReleases should fail with empty owner/repo")
	}
}

func TestExtractUniqueProjects(t *testing.T) {
	tests := []struct {
		name     string
		releases []GitHubRelease
		want     []string
	}{
		{
			name: "single project with multiple versions",
			releases: []GitHubRelease{
				{TagName: "printpdf--pr-54.1"},
				{TagName: "printpdf--pr-54.2"},
				{TagName: "printpdf--pr-54.3"},
			},
			want: []string{"printpdf"},
		},
		{
			name: "multiple different projects",
			releases: []GitHubRelease{
				{TagName: "dissect--pr-123.1"},
				{TagName: "markdown-format--pr-123.1"},
			},
			want: []string{"dissect", "markdown-format"},
		},
		{
			name: "multiple projects with multiple versions each",
			releases: []GitHubRelease{
				{TagName: "dissect--pr-123.1"},
				{TagName: "dissect--pr-123.2"},
				{TagName: "markdown-format--pr-123.1"},
				{TagName: "markdown-format--pr-123.2"},
			},
			want: []string{"dissect", "markdown-format"},
		},
		{
			name:     "empty releases list",
			releases: []GitHubRelease{},
			want:     []string{},
		},
		{
			name: "single release",
			releases: []GitHubRelease{
				{TagName: "dissect--pr-1.1"},
			},
			want: []string{"dissect"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractUniqueProjects(tt.releases)
			if len(got) != len(tt.want) {
				t.Errorf("extractUniqueProjects() = %v, want %v", got, tt.want)
				return
			}
			// Check each element is present (order doesn't matter for this test)
			gotMap := make(map[string]bool)
			for _, p := range got {
				gotMap[p] = true
			}
			for _, p := range tt.want {
				if !gotMap[p] {
					t.Errorf("extractUniqueProjects() missing project %q, got %v, want %v", p, got, tt.want)
				}
			}
		})
	}
}

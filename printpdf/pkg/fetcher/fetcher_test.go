package fetcher

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFetchLocalFile(t *testing.T) {
	// Create a temporary markdown file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	content := []byte("# Test\n\nThis is a test.")

	if err := os.WriteFile(tmpFile, content, 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test fetching
	result, contentType, err := Fetch(tmpFile, nil)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if contentType != ContentTypeMarkdown {
		t.Errorf("Expected content type %s, got %s", ContentTypeMarkdown, contentType)
	}

	if string(result) != string(content) {
		t.Errorf("Content mismatch. Expected %s, got %s", content, result)
	}
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		path     string
		content  []byte
		expected string
	}{
		{"test.md", []byte("# Test"), ContentTypeMarkdown},
		{"test.markdown", []byte("# Test"), ContentTypeMarkdown},
		{"test.html", []byte("<html>"), ContentTypeHTML},
		{"test.htm", []byte("<html>"), ContentTypeHTML},
		{"test.txt", []byte("<!DOCTYPE html>"), ContentTypeHTML},
		{"test.txt", []byte("plain text"), ContentTypeMarkdown},
	}

	for _, tt := range tests {
		result := detectContentType(tt.path, tt.content)
		if result != tt.expected {
			t.Errorf("detectContentType(%s) = %s, want %s", tt.path, result, tt.expected)
		}
	}
}

func TestIsGitHubURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://github.com/owner/repo/blob/main/file.md", true},
		{"https://github.com/owner/repo/blob/d0c77b78977c00723fa5e2f58a9c8e683cf0714c/path/to/file.md", true},
		{"https://github.com/owner/repo/raw/main/file.md", true},
		{"https://github.com/owner/repo/pull/123/files", true},
		{"https://github.com/owner/repo/files/d0c77b78977c00723fa5e2f58a9c8e683cf0714c/path/to/file.md", true},
		{"https://example.com/file.md", false},
		{"https://gitlab.com/owner/repo/blob/main/file.md", false},
	}

	for _, tt := range tests {
		result := isGitHubURL(tt.url)
		if result != tt.expected {
			t.Errorf("isGitHubURL(%s) = %v, want %v", tt.url, result, tt.expected)
		}
	}
}

func TestGitHubBlobRegex(t *testing.T) {
	tests := []struct {
		url         string
		shouldMatch bool
		owner       string
		repo        string
		ref         string
		path        string
	}{
		{
			url:         "https://github.com/neongreen/mono/blob/d0c77b78977c00723fa5e2f58a9c8e683cf0714c/mdbook-comments/DOCKER_DEMO.md",
			shouldMatch: true,
			owner:       "neongreen",
			repo:        "mono",
			ref:         "d0c77b78977c00723fa5e2f58a9c8e683cf0714c",
			path:        "mdbook-comments/DOCKER_DEMO.md",
		},
		{
			url:         "https://github.com/owner/repo/blob/main/path/to/file.md",
			shouldMatch: true,
			owner:       "owner",
			repo:        "repo",
			ref:         "main",
			path:        "path/to/file.md",
		},
		{
			url:         "https://github.com/owner/repo/files/abc123/path/to/file.md",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		matches := githubBlobRegex.FindStringSubmatch(tt.url)
		if tt.shouldMatch {
			if matches == nil {
				t.Errorf("githubBlobRegex did not match %s", tt.url)
				continue
			}
			if matches[1] != tt.owner {
				t.Errorf("Expected owner %s, got %s", tt.owner, matches[1])
			}
			if matches[2] != tt.repo {
				t.Errorf("Expected repo %s, got %s", tt.repo, matches[2])
			}
			if matches[3] != tt.ref {
				t.Errorf("Expected ref %s, got %s", tt.ref, matches[3])
			}
			if matches[4] != tt.path {
				t.Errorf("Expected path %s, got %s", tt.path, matches[4])
			}
		} else {
			if matches != nil {
				t.Errorf("githubBlobRegex should not match %s", tt.url)
			}
		}
	}
}

func TestGitHubFilesRegex(t *testing.T) {
	tests := []struct {
		url         string
		shouldMatch bool
		owner       string
		repo        string
		commitSha   string
		path        string
	}{
		{
			url:         "https://github.com/neongreen/mono/files/d0c77b78977c00723fa5e2f58a9c8e683cf0714c/mdbook-comments/DOCKER_DEMO.md",
			shouldMatch: true,
			owner:       "neongreen",
			repo:        "mono",
			commitSha:   "d0c77b78977c00723fa5e2f58a9c8e683cf0714c",
			path:        "mdbook-comments/DOCKER_DEMO.md",
		},
		{
			url:         "https://github.com/owner/repo/files/abc123/path/to/file.md",
			shouldMatch: true,
			owner:       "owner",
			repo:        "repo",
			commitSha:   "abc123",
			path:        "path/to/file.md",
		},
		{
			url:         "https://github.com/owner/repo/blob/main/file.md",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		matches := githubFilesRegex.FindStringSubmatch(tt.url)
		if tt.shouldMatch {
			if matches == nil {
				t.Errorf("githubFilesRegex did not match %s", tt.url)
				continue
			}
			if matches[1] != tt.owner {
				t.Errorf("Expected owner %s, got %s", tt.owner, matches[1])
			}
			if matches[2] != tt.repo {
				t.Errorf("Expected repo %s, got %s", tt.repo, matches[2])
			}
			if matches[3] != tt.commitSha {
				t.Errorf("Expected commitSha %s, got %s", tt.commitSha, matches[3])
			}
			if matches[4] != tt.path {
				t.Errorf("Expected path %s, got %s", tt.path, matches[4])
			}
		} else {
			if matches != nil {
				t.Errorf("githubFilesRegex should not match %s", tt.url)
			}
		}
	}
}

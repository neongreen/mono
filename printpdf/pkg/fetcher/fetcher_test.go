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
	
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test fetching
	result, contentType, err := Fetch(tmpFile)
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
		{"https://github.com/owner/repo/raw/main/file.md", true},
		{"https://github.com/owner/repo/pull/123/files", true},
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

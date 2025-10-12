package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	ContentTypeMarkdown = "markdown"
	ContentTypeHTML     = "html"
)

// Fetch retrieves content from various sources and returns the content, content type, and any error
func Fetch(input string) ([]byte, string, error) {
	// Check if it's a local file
	if _, err := os.Stat(input); err == nil {
		return fetchLocalFile(input)
	}

	// Check if it's a URL
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		return fetchURL(input)
	}

	// Try as a local file again with error
	return fetchLocalFile(input)
}

func fetchLocalFile(path string) ([]byte, string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	contentType := detectContentType(path, content)
	return content, contentType, nil
}

func fetchURL(urlStr string) ([]byte, string, error) {
	// Check if it's a GitHub URL
	if isGitHubURL(urlStr) {
		return fetchGitHubFile(urlStr)
	}

	// Regular HTTP fetch
	return fetchHTTP(urlStr)
}

func fetchHTTP(urlStr string) ([]byte, string, error) {
	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP error: %s", resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	// Determine content type from URL or content
	contentType := detectContentType(urlStr, content)
	
	// If it looks like HTML but not markdown, use Readability
	if contentType == ContentTypeHTML && !strings.HasSuffix(urlStr, ".md") {
		content, err = extractReadableContent(content)
		if err != nil {
			return nil, "", fmt.Errorf("failed to extract readable content: %w", err)
		}
	}

	return content, contentType, nil
}

func detectContentType(path string, content []byte) string {
	ext := strings.ToLower(filepath.Ext(path))
	
	switch ext {
	case ".md", ".markdown":
		return ContentTypeMarkdown
	case ".html", ".htm":
		return ContentTypeHTML
	}

	// Check content
	contentStr := string(content)
	if strings.Contains(contentStr, "<!DOCTYPE html") || 
	   strings.Contains(contentStr, "<html") ||
	   strings.Contains(contentStr, "<HTML") {
		return ContentTypeHTML
	}

	// Default to markdown
	return ContentTypeMarkdown
}

// isGitHubURL checks if the URL is a GitHub file URL
func isGitHubURL(urlStr string) bool {
	return strings.Contains(urlStr, "github.com") && 
		   (strings.Contains(urlStr, "/blob/") || 
		    strings.Contains(urlStr, "/raw/") ||
			strings.Contains(urlStr, "/pull/"))
}

// GitHub URL patterns:
// - https://github.com/owner/repo/blob/branch/path/to/file.md
// - https://github.com/owner/repo/blob/commit-sha/path/to/file.md
// - https://github.com/owner/repo/pull/123/files#diff-abc123
var githubBlobRegex = regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/blob/([^/]+)/(.+)`)
var githubRawRegex = regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/raw/([^/]+)/(.+)`)
var githubPullFileRegex = regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/pull/(\d+)/files`)

func fetchGitHubFile(urlStr string) ([]byte, string, error) {
	// Try to convert blob URL to raw URL
	if matches := githubBlobRegex.FindStringSubmatch(urlStr); matches != nil {
		owner, repo, ref, path := matches[1], matches[2], matches[3], matches[4]
		rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, ref, path)
		return fetchGitHubRaw(rawURL)
	}

	// Already a raw URL
	if matches := githubRawRegex.FindStringSubmatch(urlStr); matches != nil {
		return fetchGitHubRaw(urlStr)
	}

	// Pull request URL - need to use GitHub API
	if matches := githubPullFileRegex.FindStringSubmatch(urlStr); matches != nil {
		return nil, "", fmt.Errorf("pull request file URLs not yet supported - please use the raw file URL")
	}

	return nil, "", fmt.Errorf("unsupported GitHub URL format")
}

func fetchGitHubRaw(rawURL string) ([]byte, string, error) {
	// Try with GITHUB_TOKEN if available
	token := os.Getenv("GITHUB_TOKEN")
	
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch GitHub file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	// Detect content type from URL
	contentType := detectContentType(rawURL, content)
	return content, contentType, nil
}

// extractReadableContent uses Mozilla Readability to extract main content from HTML
// For now, this is a simplified version that just returns the HTML
// In a full implementation, this would use a headless browser or Node.js with Readability
func extractReadableContent(htmlContent []byte) ([]byte, error) {
	// TODO: Implement proper Readability extraction
	// For now, just return the HTML as-is
	// A full implementation would:
	// 1. Use a headless browser (chromedp, rod) or
	// 2. Use Node.js with @mozilla/readability or
	// 3. Use a Go port of Readability
	return htmlContent, nil
}

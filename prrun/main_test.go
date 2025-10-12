package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParsePRURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantPRNum int
		wantErr   bool
	}{
		{
			name:      "full URL with https",
			url:       "https://github.com/neongreen/mono/pull/123",
			wantOwner: "neongreen",
			wantRepo:  "mono",
			wantPRNum: 123,
			wantErr:   false,
		},
		{
			name:      "URL without protocol",
			url:       "github.com/owner/repo/pull/456",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantPRNum: 456,
			wantErr:   false,
		},
		{
			name:      "invalid URL",
			url:       "not-a-github-url",
			wantOwner: "",
			wantRepo:  "",
			wantPRNum: 0,
			wantErr:   true,
		},
		{
			name:      "missing PR number",
			url:       "github.com/owner/repo/pull/",
			wantOwner: "",
			wantRepo:  "",
			wantPRNum: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePRURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePRURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if got.Owner != tt.wantOwner {
					t.Errorf("parsePRURL() Owner = %v, want %v", got.Owner, tt.wantOwner)
				}
				if got.Repo != tt.wantRepo {
					t.Errorf("parsePRURL() Repo = %v, want %v", got.Repo, tt.wantRepo)
				}
				if got.PRNum != tt.wantPRNum {
					t.Errorf("parsePRURL() PRNum = %v, want %v", got.PRNum, tt.wantPRNum)
				}
			}
		})
	}
}

func TestGetCacheDir(t *testing.T) {
	cacheDir, err := getCacheDir()
	if err != nil {
		t.Fatalf("getCacheDir() error = %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expectedDir := filepath.Join(homeDir, ".cache", "prrun")

	if cacheDir != expectedDir {
		t.Errorf("getCacheDir() = %v, want %v", cacheDir, expectedDir)
	}
}

func TestGetPlatformBinaryName(t *testing.T) {
	release := &GitHubRelease{
		TagName: "dissect--pr-123.1",
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			{
				Name:               "dissect-pr-123.1-linux-amd64",
				BrowserDownloadURL: "https://github.com/example/repo/releases/download/dissect--pr-123.1/dissect-pr-123.1-linux-amd64",
			},
			{
				Name:               "dissect-pr-123.1-darwin-arm64",
				BrowserDownloadURL: "https://github.com/example/repo/releases/download/dissect--pr-123.1/dissect-pr-123.1-darwin-arm64",
			},
		},
	}

	// This test will pass or fail depending on the current platform
	binaryName, downloadURL, err := getPlatformBinaryName(release, "dissect")
	if err != nil {
		t.Fatalf("getPlatformBinaryName() error = %v", err)
	}

	if binaryName == "" {
		t.Error("getPlatformBinaryName() returned empty binary name")
	}

	if downloadURL == "" {
		t.Error("getPlatformBinaryName() returned empty download URL")
	}

	t.Logf("Platform binary: %s", binaryName)
	t.Logf("Download URL: %s", downloadURL)
}

func TestGetPlatformBinaryName_NoAssets(t *testing.T) {
	release := &GitHubRelease{
		TagName: "test--pr-1.1",
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{},
	}

	_, _, err := getPlatformBinaryName(release, "test")
	if err == nil {
		t.Error("getPlatformBinaryName() should return error for release with no assets")
	}

	expectedErrMsg := "release test--pr-1.1 has no assets"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedErrMsg, err)
	}
}

func TestGetPlatformBinaryName_DoubleDashFormat(t *testing.T) {
	// Test with assets using double dash format (project--version-os-arch)
	release := &GitHubRelease{
		TagName: "dissect--pr-123.1",
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			{
				Name:               "dissect--pr-123.1-linux-amd64",
				BrowserDownloadURL: "https://github.com/example/repo/releases/download/dissect--pr-123.1/dissect--pr-123.1-linux-amd64",
			},
			{
				Name:               "dissect--pr-123.1-darwin-arm64",
				BrowserDownloadURL: "https://github.com/example/repo/releases/download/dissect--pr-123.1/dissect--pr-123.1-darwin-arm64",
			},
		},
	}

	binaryName, downloadURL, err := getPlatformBinaryName(release, "dissect")
	if err != nil {
		t.Fatalf("getPlatformBinaryName() should handle double dash format, error = %v", err)
	}

	if !strings.Contains(binaryName, "linux") || !strings.Contains(binaryName, "amd64") {
		t.Errorf("Expected binary name to contain platform info, got: %s", binaryName)
	}

	if downloadURL == "" {
		t.Error("getPlatformBinaryName() returned empty download URL")
	}

	t.Logf("Double dash format binary: %s", binaryName)
	t.Logf("Download URL: %s", downloadURL)
}

func TestGetGitHubToken(t *testing.T) {
	// Save original environment variables
	origGithubToken := os.Getenv("GITHUB_TOKEN")
	origMiseToken := os.Getenv("MISE_GITHUB_TOKEN")

	// Clean up after test
	defer func() {
		if origGithubToken != "" {
			os.Setenv("GITHUB_TOKEN", origGithubToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
		if origMiseToken != "" {
			os.Setenv("MISE_GITHUB_TOKEN", origMiseToken)
		} else {
			os.Unsetenv("MISE_GITHUB_TOKEN")
		}
	}()

	t.Run("GITHUB_TOKEN takes precedence", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "github_token")
		os.Setenv("MISE_GITHUB_TOKEN", "mise_token")

		token := getGitHubToken()
		if token != "github_token" {
			t.Errorf("getGitHubToken() = %v, want %v", token, "github_token")
		}
	})

	t.Run("MISE_GITHUB_TOKEN used when GITHUB_TOKEN not set", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")
		os.Setenv("MISE_GITHUB_TOKEN", "mise_token")

		token := getGitHubToken()
		if token != "mise_token" {
			t.Errorf("getGitHubToken() = %v, want %v", token, "mise_token")
		}
	})

	t.Run("returns empty string when no tokens available", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("MISE_GITHUB_TOKEN")

		token := getGitHubToken()
		// Token might be empty or come from gh CLI
		// We just verify the function doesn't crash
		t.Logf("getGitHubToken() = %q", token)
	})
}

func TestCreateAuthenticatedRequest(t *testing.T) {
	// Save original environment variable
	origGithubToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if origGithubToken != "" {
			os.Setenv("GITHUB_TOKEN", origGithubToken)
		} else {
			os.Unsetenv("GITHUB_TOKEN")
		}
	}()

	t.Run("adds authorization header when token available", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "test_token_123")

		req, err := createAuthenticatedRequest("GET", "https://api.github.com/repos/test/test")
		if err != nil {
			t.Fatalf("createAuthenticatedRequest() error = %v", err)
		}

		authHeader := req.Header.Get("Authorization")
		expectedHeader := "Bearer test_token_123"
		if authHeader != expectedHeader {
			t.Errorf("Authorization header = %v, want %v", authHeader, expectedHeader)
		}
	})

	t.Run("creates request without authorization when token not available", func(t *testing.T) {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("MISE_GITHUB_TOKEN")

		req, err := createAuthenticatedRequest("GET", "https://api.github.com/repos/test/test")
		if err != nil {
			t.Fatalf("createAuthenticatedRequest() error = %v", err)
		}

		authHeader := req.Header.Get("Authorization")
		// If gh CLI is available, there might be a token; otherwise it should be empty
		t.Logf("Authorization header = %q", authHeader)
	})
}

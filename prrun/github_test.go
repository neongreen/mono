package main

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/neongreen/mono/lib/ghrelease"
)

func TestGetPlatformBinaryName(t *testing.T) {
	release := &GitHubRelease{TagName: "dissect--pr-123.1", Assets: []struct {
		Name               string `json:"name"`
		URL                string `json:"url"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}{{Name: "dissect-pr-123.1-linux-amd64", URL: "https://api.github.com/repos/example/repo/releases/assets/123", BrowserDownloadURL: "https://github.com/example/repo/releases/download/dissect--pr-123.1/dissect-pr-123.1-linux-amd64"}, {Name: "dissect-pr-123.1-darwin-arm64", URL: "https://api.github.com/repos/example/repo/releases/assets/124", BrowserDownloadURL: "https://github.com/example/repo/releases/download/dissect--pr-123.1/dissect-pr-123.1-darwin-arm64"}}}
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
	release := &GitHubRelease{TagName: "test--pr-1.1",
		Assets: []struct {
			Name               string `json:"name"`
			URL                string `json:"url"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{}}
	_, _, err := getPlatformBinaryName(release, "test")
	if err == nil {
		t.Error("getPlatformBinaryName() should return error for release with no assets")
	}
	expectedErrMsg := "release test--pr-1.1 has no assets"
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("Expected error to contain '%s', got: %v",
			expectedErrMsg, err)
	}
}
func TestGetPlatformBinaryName_DoubleDashFormat(t *testing.T) {
	release := &GitHubRelease{TagName: "dissect--pr-123.1", Assets: []struct {
		Name               string `json:"name"`
		URL                string `json:"url"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}{{Name: "dissect--pr-123.1-linux-amd64", URL: "https://api.github.com/repos/example/repo/releases/assets/125", BrowserDownloadURL: "https://github.com/example/repo/releases/download/dissect--pr-123.1/dissect--pr-123.1-linux-amd64"}, {Name: "dissect--pr-123.1-darwin-arm64", URL: "https://api.github.com/repos/example/repo/releases/assets/126", BrowserDownloadURL: "https://github.com/example/repo/releases/download/dissect--pr-123.1/dissect--pr-123.1-darwin-arm64"}}}
	binaryName, downloadURL, err := getPlatformBinaryName(release, "dissect")
	if err != nil {
		t.Fatalf("getPlatformBinaryName() should handle double dash format, error = %v",

			err)
	}
	if !strings.Contains(binaryName, runtime.GOOS) ||
		!strings.Contains(binaryName, runtime.GOARCH) {
		t.Errorf("Expected binary name to contain platform info, got: %s",
			binaryName,
		)
	}
	if downloadURL == "" {
		t.Error("getPlatformBinaryName() returned empty download URL")
	}
	t.
		Logf("Double dash format binary: %s", binaryName)
	t.Logf("Download URL: %s", downloadURL)
}
func TestGetGitHubToken(t *testing.T) {
	origGithubToken := os.Getenv("GITHUB_TOKEN")
	origMiseToken := os.Getenv("MISE_GITHUB_TOKEN")
	defer func() {
		if origGithubToken !=
			"" {
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
		os.Setenv("MISE_GITHUB_TOKEN",
			"mise_token")
		token := ghrelease.GetGitHubToken()
		if token != "github_token" {
			t.Errorf("GetGitHubToken() = %v, want %v",
				token, "github_token")
		}
	})
	t.Run("MISE_GITHUB_TOKEN used when GITHUB_TOKEN not set",
		func(t *testing.T) {
			os.
				Unsetenv("GITHUB_TOKEN")
			os.Setenv("MISE_GITHUB_TOKEN", "mise_token")
			token := ghrelease.GetGitHubToken()
			if token != "mise_token" {
				t.
					Errorf("GetGitHubToken() = %v, want %v",
						token,
						"mise_token")
			}
		})
	t.Run("returns empty string when no tokens available",
		func(t *testing.T) {
			os.
				Unsetenv("GITHUB_TOKEN")
			os.Unsetenv("MISE_GITHUB_TOKEN")
			token := ghrelease.GetGitHubToken()

			t.Logf("GetGitHubToken() = %q", token)
		})
}
func TestCreateAuthenticatedRequest(t *testing.T) {
	origGithubToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		if origGithubToken != "" {
			os.
				Setenv("GITHUB_TOKEN", origGithubToken)
		} else {
			os.
				Unsetenv("GITHUB_TOKEN")
		}
	}()
	t.Run("adds authorization header when token available", func(t *testing.T) {
		os.Setenv("GITHUB_TOKEN", "test_token_123")
		req, err := ghrelease.CreateAuthenticatedRequest("GET", "https://api.github.com/repos/test/test")
		if err != nil {
			t.Fatalf("CreateAuthenticatedRequest() error = %v", err)
		}
		authHeader := req.Header.Get("Authorization")
		expectedHeader := "Bearer test_token_123"
		if authHeader !=
			expectedHeader {
			t.Errorf("Authorization header = %v, want %v", authHeader,

				expectedHeader)
		}
	})
	t.Run("creates request without authorization when token not available",

		func(t *testing.
			T) {
			os.Unsetenv("GITHUB_TOKEN")
			os.Unsetenv("MISE_GITHUB_TOKEN")
			req, err := ghrelease.CreateAuthenticatedRequest("GET", "https://api.github.com/repos/test/test")
			if err != nil {
				t.Fatalf("CreateAuthenticatedRequest() error = %v",
					err)
			}
			authHeader := req.Header.Get("Authorization")
			t.
				Logf("Authorization header = %q", authHeader)
		})
}

package main

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetPlatformBinaryName(t *testing.T) {
	release := &GitHubRelease{TagName: "dissect--pr-123.1", Assets: []struct {
		Name               string `json:"name"`
		URL                string `json:"url"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}{
		{Name: "dissect-pr-123.1-linux-amd64", URL: "https://api.github.com/repos/example/repo/releases/assets/123", BrowserDownloadURL: "https://github.com/example/repo/releases/download/dissect--pr-123.1/dissect-pr-123.1-linux-amd64"},
		{Name: "dissect-pr-123.1-linux-arm64", URL: "https://api.github.com/repos/example/repo/releases/assets/124", BrowserDownloadURL: "https://github.com/example/repo/releases/download/dissect--pr-123.1/dissect-pr-123.1-linux-arm64"},
		{Name: "dissect-pr-123.1-darwin-arm64", URL: "https://api.github.com/repos/example/repo/releases/assets/125", BrowserDownloadURL: "https://github.com/example/repo/releases/download/dissect--pr-123.1/dissect-pr-123.1-darwin-arm64"},
		{Name: "dissect-pr-123.1-darwin-amd64", URL: "https://api.github.com/repos/example/repo/releases/assets/126", BrowserDownloadURL: "https://github.com/example/repo/releases/download/dissect--pr-123.1/dissect-pr-123.1-darwin-amd64"},
	}}
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
			URL                string `json:"url"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{},
	}
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
	}{
		{Name: "dissect--pr-123.1-linux-amd64", URL: "https://api.github.com/repos/example/repo/releases/assets/127", BrowserDownloadURL: "https://github.com/example/repo/releases/download/dissect--pr-123.1/dissect--pr-123.1-linux-amd64"},
		{Name: "dissect--pr-123.1-linux-arm64", URL: "https://api.github.com/repos/example/repo/releases/assets/128", BrowserDownloadURL: "https://github.com/example/repo/releases/download/dissect--pr-123.1/dissect--pr-123.1-linux-arm64"},
		{Name: "dissect--pr-123.1-darwin-arm64", URL: "https://api.github.com/repos/example/repo/releases/assets/129", BrowserDownloadURL: "https://github.com/example/repo/releases/download/dissect--pr-123.1/dissect--pr-123.1-darwin-arm64"},
		{Name: "dissect--pr-123.1-darwin-amd64", URL: "https://api.github.com/repos/example/repo/releases/assets/130", BrowserDownloadURL: "https://github.com/example/repo/releases/download/dissect--pr-123.1/dissect--pr-123.1-darwin-amd64"},
	}}
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

// TestGetGitHubToken and TestCreateAuthenticatedRequest removed - these are library
// functions already tested in lib/ghrelease/ghrelease_test.go

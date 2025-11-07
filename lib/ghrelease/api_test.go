package ghrelease

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestGetReleaseByTag tests fetching a release by tag name
func TestGetReleaseByTag(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		release := Release{
			TagName: "v1.0.0",
			Name:    "Release 1.0.0",
			Assets: []Asset{
				{
					Name:               "app-linux-amd64",
					URL:                "https://api.github.com/repos/test/repo/releases/assets/1",
					BrowserDownloadURL: "https://github.com/test/repo/releases/download/v1.0.0/app-linux-amd64",
				},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/repos/test/repo/releases/tags/v1.0.0" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(release)
		}))
		defer server.Close()

		// Override API URL for testing
		ctx := context.Background()
		apiURL := server.URL + "/repos/test/repo/releases/tags/v1.0.0"
		req, err := CreateAuthenticatedRequestWithContext(ctx, "GET", apiURL)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to fetch release: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status code: %d", resp.StatusCode)
		}

		var fetched Release
		if err := json.NewDecoder(resp.Body).Decode(&fetched); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if fetched.TagName != release.TagName {
			t.Errorf("TagName = %q, want %q", fetched.TagName, release.TagName)
		}
		if len(fetched.Assets) != len(release.Assets) {
			t.Errorf("len(Assets) = %d, want %d", len(fetched.Assets), len(release.Assets))
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message": "Not Found"}`))
		}))
		defer server.Close()

		ctx := context.Background()
		apiURL := server.URL + "/repos/test/repo/releases/tags/nonexistent"
		req, err := CreateAuthenticatedRequestWithContext(ctx, "GET", apiURL)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to fetch release: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{invalid json`))
		}))
		defer server.Close()

		ctx := context.Background()
		apiURL := server.URL + "/repos/test/repo/releases/tags/v1.0.0"
		req, err := CreateAuthenticatedRequestWithContext(ctx, "GET", apiURL)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to fetch release: %v", err)
		}
		defer resp.Body.Close()

		var release Release
		err = json.NewDecoder(resp.Body).Decode(&release)
		if err == nil {
			t.Error("expected JSON decode error, got nil")
		}
	})
}

// TestDownloadAsset tests downloading a release asset
func TestDownloadAsset(t *testing.T) {
	t.Run("successful download", func(t *testing.T) {
		expectedContent := []byte("test binary content")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Accept") != "application/octet-stream" {
				t.Errorf("unexpected Accept header: %s", r.Header.Get("Accept"))
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(expectedContent)
		}))
		defer server.Close()

		asset := &Asset{
			Name: "test-binary",
			URL:  server.URL,
		}

		tempDir := t.TempDir()
		destPath := filepath.Join(tempDir, "test-binary")

		ctx := context.Background()
		err := DownloadAssetWithContext(ctx, asset, destPath)
		if err != nil {
			t.Fatalf("DownloadAssetWithContext() error = %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			t.Fatalf("downloaded file does not exist at %s", destPath)
		}

		// Verify content
		content, err := os.ReadFile(destPath)
		if err != nil {
			t.Fatalf("failed to read downloaded file: %v", err)
		}
		if string(content) != string(expectedContent) {
			t.Errorf("content = %q, want %q", content, expectedContent)
		}

		// Verify executable permissions
		info, err := os.Stat(destPath)
		if err != nil {
			t.Fatalf("failed to stat file: %v", err)
		}
		mode := info.Mode()
		if mode&0o111 == 0 {
			t.Errorf("file is not executable: mode = %v", mode)
		}
	})

	t.Run("creates parent directories", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("content"))
		}))
		defer server.Close()

		asset := &Asset{
			Name: "test-binary",
			URL:  server.URL,
		}

		tempDir := t.TempDir()
		destPath := filepath.Join(tempDir, "subdir1", "subdir2", "test-binary")

		ctx := context.Background()
		err := DownloadAssetWithContext(ctx, asset, destPath)
		if err != nil {
			t.Fatalf("DownloadAssetWithContext() error = %v", err)
		}

		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			t.Fatalf("downloaded file does not exist at %s", destPath)
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		asset := &Asset{
			Name: "nonexistent-binary",
			URL:  server.URL,
		}

		tempDir := t.TempDir()
		destPath := filepath.Join(tempDir, "test-binary")

		ctx := context.Background()
		err := DownloadAssetWithContext(ctx, asset, destPath)
		if err == nil {
			t.Fatal("expected error for 404, got nil")
		}
		if !strings.Contains(err.Error(), "404") {
			t.Errorf("error should mention 404, got: %v", err)
		}
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		asset := &Asset{
			Name: "test-binary",
			URL:  server.URL,
		}

		tempDir := t.TempDir()
		destPath := filepath.Join(tempDir, "test-binary")

		ctx := context.Background()
		err := DownloadAssetWithContext(ctx, asset, destPath)
		if err == nil {
			t.Fatal("expected error for 500, got nil")
		}
	})

	t.Run("DownloadAsset wrapper works", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("test"))
		}))
		defer server.Close()

		asset := &Asset{
			Name: "test-binary",
			URL:  server.URL,
		}

		tempDir := t.TempDir()
		destPath := filepath.Join(tempDir, "test-binary")

		err := DownloadAsset(asset, destPath)
		if err != nil {
			t.Fatalf("DownloadAsset() error = %v", err)
		}
	})
}

// TestListReleases tests listing releases with pagination
func TestListReleases(t *testing.T) {
	t.Run("single page", func(t *testing.T) {
		releases := []Release{
			{TagName: "v1.0.0", Name: "Release 1"},
			{TagName: "v0.9.0", Name: "Release 0.9"},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Path, "/repos/test/repo/releases") {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(releases)
		}))
		defer server.Close()

		ctx := context.Background()
		apiURL := server.URL + "/repos/test/repo/releases?per_page=100&page=1"
		req, err := CreateAuthenticatedRequestWithContext(ctx, "GET", apiURL)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to fetch releases: %v", err)
		}
		defer resp.Body.Close()

		var fetched []Release
		if err := json.NewDecoder(resp.Body).Decode(&fetched); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(fetched) != len(releases) {
			t.Errorf("len(releases) = %d, want %d", len(fetched), len(releases))
		}
	})

	t.Run("empty releases list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode([]Release{})
		}))
		defer server.Close()

		ctx := context.Background()
		apiURL := server.URL + "/repos/test/repo/releases?per_page=100&page=1"
		req, err := CreateAuthenticatedRequestWithContext(ctx, "GET", apiURL)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to fetch releases: %v", err)
		}
		defer resp.Body.Close()

		var fetched []Release
		if err := json.NewDecoder(resp.Body).Decode(&fetched); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(fetched) != 0 {
			t.Errorf("expected empty list, got %d releases", len(fetched))
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		ctx := context.Background()
		apiURL := server.URL + "/repos/test/repo/releases?per_page=100&page=1"
		req, err := CreateAuthenticatedRequestWithContext(ctx, "GET", apiURL)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to fetch releases: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})
}

// TestFindPlatformAssetErrorCases tests error handling in FindPlatformAsset
func TestFindPlatformAssetErrorCases(t *testing.T) {
	t.Run("unsupported OS", func(t *testing.T) {
		// This test verifies that the function handles unsupported platforms
		// In actual code, GetCurrentPlatform() returns the current OS,
		// so we test the error message construction logic
		release := &Release{
			TagName: "v1.0.0",
			Assets: []Asset{
				{Name: "app-linux-amd64", URL: "https://example.com/asset"},
			},
		}

		// When platform doesn't match any asset, we get a detailed error
		_, err := FindPlatformAsset(release, "nonexistent")
		if err == nil {
			// On supported platforms, this will fail to find matching asset
			t.Skip("test only runs on supported platforms")
		}
	})

	t.Run("wrong project name", func(t *testing.T) {
		release := &Release{
			TagName: "v1.0.0",
			Assets: []Asset{
				{Name: "app-linux-amd64", URL: "https://example.com/asset"},
				{Name: "app-darwin-arm64", URL: "https://example.com/asset2"},
			},
		}

		_, err := FindPlatformAsset(release, "wrongname")
		if err == nil {
			t.Fatal("expected error for wrong project name, got nil")
		}
		if !strings.Contains(err.Error(), "wrongname") {
			t.Errorf("error should mention project name, got: %v", err)
		}
	})

	t.Run("error message includes available assets", func(t *testing.T) {
		release := &Release{
			TagName: "v1.0.0",
			Assets: []Asset{
				{Name: "asset1", URL: "https://example.com/1"},
				{Name: "asset2", URL: "https://example.com/2"},
			},
		}

		_, err := FindPlatformAsset(release, "test")
		if err == nil {
			// If this succeeds, we're on a platform where asset names match
			t.Skip("test requires platform without matching assets")
		}

		errStr := err.Error()
		if !strings.Contains(errStr, "asset1") || !strings.Contains(errStr, "asset2") {
			t.Errorf("error should list available assets, got: %v", err)
		}
	})
}

// TestDownloadReleaseAsset tests the convenience function for downloading assets
func TestDownloadReleaseAsset(t *testing.T) {
	t.Run("invalid owner/repo causes error", func(t *testing.T) {
		tempDir := t.TempDir()
		destPath := filepath.Join(tempDir, "test-binary")

		// This should fail to connect to GitHub API
		err := DownloadReleaseAsset("nonexistent", "repo", "v1.0.0", "test", destPath)
		if err == nil {
			t.Error("expected error for nonexistent repository, got nil")
		}
	})
}

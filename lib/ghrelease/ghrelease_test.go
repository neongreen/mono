package ghrelease

import (
	"runtime"
	"strings"
	"testing"

	"github.com/neongreen/mono/lib/testhelpers"
)

// Test helper functions moved to lib/testhelpers

func TestGetCurrentPlatform(t *testing.T) {
	platform := GetCurrentPlatform()

	if platform.OS == "" {
		t.Error("GetCurrentPlatform() returned empty OS")
	}
	if platform.Arch == "" {
		t.Error("GetCurrentPlatform() returned empty Arch")
	}

	if platform.OS != runtime.GOOS {
		t.Errorf("Expected OS to be %s, got %s", runtime.GOOS, platform.OS)
	}
	if platform.Arch != runtime.GOARCH {
		t.Errorf("Expected Arch to be %s, got %s", runtime.GOARCH, platform.Arch)
	}

	t.Logf("Platform: %s-%s", platform.OS, platform.Arch)
}

func TestGetGitHubToken(t *testing.T) {
	t.Run("GITHUB_TOKEN takes precedence", func(t *testing.T) {
		buf := testhelpers.SetupTestLogger(t)
		t.Setenv("GITHUB_TOKEN", "github_token")
		t.Setenv("MISE_GITHUB_TOKEN", "mise_token")

		token := GetGitHubToken()
		if token != "github_token" {
			t.Fatalf("GetGitHubToken() = %q, want %q", token, "github_token")
		}

		output := buf.String()
		if !strings.Contains(output, "source=GITHUB_TOKEN") {
			t.Fatalf("expected log to mention GITHUB_TOKEN source, got %q", output)
		}
		if strings.Contains(output, "github_token") {
			t.Fatalf("logs must not contain actual token, got %q", output)
		}
	})

	t.Run("MISE_GITHUB_TOKEN used when GITHUB_TOKEN not set", func(t *testing.T) {
		buf := testhelpers.SetupTestLogger(t)
		t.Setenv("GITHUB_TOKEN", "")
		t.Setenv("MISE_GITHUB_TOKEN", "mise_token")

		token := GetGitHubToken()
		if token != "mise_token" {
			t.Fatalf("GetGitHubToken() = %q, want %q", token, "mise_token")
		}

		output := buf.String()
		if !strings.Contains(output, "source=MISE_GITHUB_TOKEN") {
			t.Fatalf("expected log to mention MISE_GITHUB_TOKEN source, got %q", output)
		}
		if strings.Contains(output, "mise_token") {
			t.Fatalf("logs must not contain actual token, got %q", output)
		}
	})

	t.Run("returns empty string when no tokens available", func(t *testing.T) {
		buf := testhelpers.SetupTestLogger(t)
		t.Setenv("GITHUB_TOKEN", "")
		t.Setenv("MISE_GITHUB_TOKEN", "")
		testhelpers.InstallGhStub(t, "", 1)

		token := GetGitHubToken()
		if token != "" {
			t.Fatalf("GetGitHubToken() = %q, want empty string", token)
		}

		output := buf.String()
		if !strings.Contains(output, "GitHub token unavailable after checking all sources") {
			t.Fatalf("expected log to note unavailable token, got %q", output)
		}
		if !strings.Contains(output, "source=gh_cli") {
			t.Fatalf("expected log to mention gh_cli source check, got %q", output)
		}
	})

	t.Run("gh CLI token used when available", func(t *testing.T) {
		buf := testhelpers.SetupTestLogger(t)
		t.Setenv("GITHUB_TOKEN", "")
		t.Setenv("MISE_GITHUB_TOKEN", "")
		testhelpers.InstallGhStub(t, "cli_token", 0)

		token := GetGitHubToken()
		if token != "cli_token" {
			t.Fatalf("GetGitHubToken() = %q, want %q", token, "cli_token")
		}

		output := buf.String()
		if !strings.Contains(output, "source=gh_cli") {
			t.Fatalf("expected log to mention gh_cli source, got %q", output)
		}
		if strings.Contains(output, "cli_token") {
			t.Fatalf("logs must not contain actual token, got %q", output)
		}
	})
}

func TestCreateAuthenticatedRequest(t *testing.T) {
	t.Run("adds authorization header when token available", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "test_token_123")

		req, err := CreateAuthenticatedRequest("GET", "https://api.github.com/repos/test/test")
		if err != nil {
			t.Fatalf("CreateAuthenticatedRequest() error = %v", err)
		}

		authHeader := req.Header.Get("Authorization")
		if authHeader != "Bearer test_token_123" {
			t.Fatalf("Authorization header = %q, want %q", authHeader, "Bearer test_token_123")
		}
	})

	t.Run("creates request without authorization when token not available", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "")
		t.Setenv("MISE_GITHUB_TOKEN", "")

		req, err := CreateAuthenticatedRequest("GET", "https://api.github.com/repos/test/test")
		if err != nil {
			t.Fatalf("CreateAuthenticatedRequest() error = %v", err)
		}

		authHeader := req.Header.Get("Authorization")
		t.Logf("Authorization header = %q", authHeader)
	})
}

func TestFindPlatformAsset(t *testing.T) {
	t.Run("finds asset with matching platform", func(t *testing.T) {
		release := &Release{
			TagName: "dissect--pr-123.1",
			Assets: []Asset{
				{Name: "dissect-pr-123.1-linux-amd64", URL: "https://api.github.com/repos/example/repo/releases/assets/123"},
				{Name: "dissect-pr-123.1-linux-arm64", URL: "https://api.github.com/repos/example/repo/releases/assets/124"},
				{Name: "dissect-pr-123.1-darwin-arm64", URL: "https://api.github.com/repos/example/repo/releases/assets/125"},
				{Name: "dissect-pr-123.1-darwin-amd64", URL: "https://api.github.com/repos/example/repo/releases/assets/126"},
			},
		}

		asset, err := FindPlatformAsset(release, "dissect")
		if err != nil {
			t.Fatalf("FindPlatformAsset() error = %v", err)
		}
		if asset == nil {
			t.Fatal("FindPlatformAsset() returned nil asset")
		}
		if asset.Name == "" {
			t.Error("FindPlatformAsset() returned empty asset name")
		}
		if asset.URL == "" {
			t.Error("FindPlatformAsset() returned empty asset URL")
		}

		t.Logf("Platform asset: %s", asset.Name)
	})

	t.Run("returns error for release with no assets", func(t *testing.T) {
		release := &Release{
			TagName: "test--pr-1.1",
			Assets:  []Asset{},
		}

		_, err := FindPlatformAsset(release, "test")
		if err == nil {
			t.Fatal("FindPlatformAsset() should return error for release with no assets")
		}
		if !strings.Contains(err.Error(), "release test--pr-1.1 has no assets") {
			t.Fatalf("expected error to mention missing assets, got %v", err)
		}
	})

	t.Run("handles double dash format", func(t *testing.T) {
		release := &Release{
			TagName: "dissect--pr-123.1",
			Assets: []Asset{
				{Name: "dissect--pr-123.1-linux-amd64", URL: "https://api.github.com/repos/example/repo/releases/assets/127"},
				{Name: "dissect--pr-123.1-linux-arm64", URL: "https://api.github.com/repos/example/repo/releases/assets/128"},
				{Name: "dissect--pr-123.1-darwin-arm64", URL: "https://api.github.com/repos/example/repo/releases/assets/129"},
				{Name: "dissect--pr-123.1-darwin-amd64", URL: "https://api.github.com/repos/example/repo/releases/assets/130"},
			},
		}

		asset, err := FindPlatformAsset(release, "dissect")
		if err != nil {
			t.Fatalf("FindPlatformAsset() should handle double dash format, error = %v", err)
		}
		if !strings.Contains(asset.Name, runtime.GOOS) || !strings.Contains(asset.Name, runtime.GOARCH) {
			t.Fatalf("Expected asset name to contain platform info, got %s", asset.Name)
		}

		t.Logf("Double dash format asset: %s", asset.Name)
	})

	t.Run("finds asset without project name", func(t *testing.T) {
		release := &Release{
			TagName: "want--main.3",
			Assets: []Asset{
				{Name: "want-main.3-linux-amd64", URL: "https://api.github.com/repos/example/repo/releases/assets/131"},
				{Name: "want-main.3-linux-arm64", URL: "https://api.github.com/repos/example/repo/releases/assets/132"},
				{Name: "want-main.3-darwin-arm64", URL: "https://api.github.com/repos/example/repo/releases/assets/133"},
				{Name: "want-main.3-darwin-amd64", URL: "https://api.github.com/repos/example/repo/releases/assets/134"},
			},
		}

		asset, err := FindPlatformAsset(release, "")
		if err != nil {
			t.Fatalf("FindPlatformAsset() error = %v", err)
		}
		if asset == nil {
			t.Fatal("FindPlatformAsset() returned nil asset")
		}

		t.Logf("Asset without project name: %s", asset.Name)
	})
}

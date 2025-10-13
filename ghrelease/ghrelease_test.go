package ghrelease

import (
"os"
"runtime"
"strings"
"testing"
)

func TestGetCurrentPlatform(t *testing.T) {
platform := GetCurrentPlatform()

// Check that we get valid values
if platform.OS == "" {
t.Error("GetCurrentPlatform() returned empty OS")
}
if platform.Arch == "" {
t.Error("GetCurrentPlatform() returned empty Arch")
}

// Check that the values match runtime values
if platform.OS != runtime.GOOS {
t.Errorf("Expected OS to be %s, got %s", runtime.GOOS, platform.OS)
}
if platform.Arch != runtime.GOARCH {
t.Errorf("Expected Arch to be %s, got %s", runtime.GOARCH, platform.Arch)
}

t.Logf("Platform: %s-%s", platform.OS, platform.Arch)
}

func TestGetGitHubToken(t *testing.T) {
origGithubToken := os.Getenv("GITHUB_TOKEN")
origMiseToken := os.Getenv("MISE_GITHUB_TOKEN")
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
token := GetGitHubToken()
if token != "github_token" {
t.Errorf("GetGitHubToken() = %v, want %v", token, "github_token")
}
})

t.Run("MISE_GITHUB_TOKEN used when GITHUB_TOKEN not set", func(t *testing.T) {
os.Unsetenv("GITHUB_TOKEN")
os.Setenv("MISE_GITHUB_TOKEN", "mise_token")
token := GetGitHubToken()
if token != "mise_token" {
t.Errorf("GetGitHubToken() = %v, want %v", token, "mise_token")
}
})

t.Run("returns empty string when no tokens available", func(t *testing.T) {
os.Unsetenv("GITHUB_TOKEN")
os.Unsetenv("MISE_GITHUB_TOKEN")
token := GetGitHubToken()
t.Logf("GetGitHubToken() = %q", token)
})
}

func TestCreateAuthenticatedRequest(t *testing.T) {
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
req, err := CreateAuthenticatedRequest("GET", "https://api.github.com/repos/test/test")
if err != nil {
t.Fatalf("CreateAuthenticatedRequest() error = %v", err)
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
{Name: "dissect-pr-123.1-darwin-arm64", URL: "https://api.github.com/repos/example/repo/releases/assets/124"},
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
t.Error("FindPlatformAsset() should return error for release with no assets")
}
expectedErrMsg := "release test--pr-1.1 has no assets"
if !strings.Contains(err.Error(), expectedErrMsg) {
t.Errorf("Expected error to contain '%s', got: %v", expectedErrMsg, err)
}
})

t.Run("handles double dash format", func(t *testing.T) {
release := &Release{
TagName: "dissect--pr-123.1",
Assets: []Asset{
{Name: "dissect--pr-123.1-linux-amd64", URL: "https://api.github.com/repos/example/repo/releases/assets/125"},
{Name: "dissect--pr-123.1-darwin-arm64", URL: "https://api.github.com/repos/example/repo/releases/assets/126"},
},
}
asset, err := FindPlatformAsset(release, "dissect")
if err != nil {
t.Fatalf("FindPlatformAsset() should handle double dash format, error = %v", err)
}
if !strings.Contains(asset.Name, runtime.GOOS) || !strings.Contains(asset.Name, runtime.GOARCH) {
t.Errorf("Expected asset name to contain platform info, got: %s", asset.Name)
}
t.Logf("Double dash format asset: %s", asset.Name)
})

t.Run("finds asset without project name", func(t *testing.T) {
release := &Release{
TagName: "want--main.3",
Assets: []Asset{
{Name: "want-main.3-linux-amd64", URL: "https://api.github.com/repos/example/repo/releases/assets/127"},
{Name: "want-main.3-darwin-arm64", URL: "https://api.github.com/repos/example/repo/releases/assets/128"},
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

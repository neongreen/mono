package ghrelease

import (
"encoding/json"
"fmt"
"io"
"net/http"
"os"
"os/exec"
"path/filepath"
"runtime"
"strings"
)

// Asset represents a GitHub release asset
type Asset struct {
Name               string `json:"name"`
URL                string `json:"url"`
BrowserDownloadURL string `json:"browser_download_url"`
}

// Release represents a GitHub release
type Release struct {
TagName    string  `json:"tag_name"`
Name       string  `json:"name"`
Prerelease bool    `json:"prerelease"`
Assets     []Asset `json:"assets"`
}

// Platform represents OS and architecture
type Platform struct {
OS   string
Arch string
}

// GetCurrentPlatform returns the current OS and architecture
func GetCurrentPlatform() Platform {
osName := runtime.GOOS
archName := runtime.GOARCH

// Normalize OS name
if osName == "darwin" {
osName = "darwin"
} else if osName == "linux" {
osName = "linux"
}

// Normalize architecture name
if archName == "amd64" {
archName = "amd64"
} else if archName == "arm64" {
archName = "arm64"
}

return Platform{OS: osName, Arch: archName}
}

// GetGitHubToken retrieves GitHub token from environment or gh CLI
func GetGitHubToken() string {
// Check GITHUB_TOKEN first
if token := os.Getenv("GITHUB_TOKEN"); token != "" {
return token
}
// Check MISE_GITHUB_TOKEN
if token := os.Getenv("MISE_GITHUB_TOKEN"); token != "" {
return token
}
// Try gh CLI
cmd := exec.Command("gh", "auth", "token")
output, err := cmd.Output()
if err == nil && len(output) > 0 {
return strings.TrimSpace(string(output))
}
return ""
}

// CreateAuthenticatedRequest creates an HTTP request with GitHub authentication
func CreateAuthenticatedRequest(method, url string) (*http.Request, error) {
req, err := http.NewRequest(method, url, nil)
if err != nil {
return nil, err
}

if token := GetGitHubToken(); token != "" {
req.Header.Set("Authorization", "Bearer "+token)
}
return req, nil
}

// GetReleaseByTag fetches a GitHub release by its tag name
func GetReleaseByTag(owner, repo, tag string) (*Release, error) {
apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", owner, repo, tag)
req, err := CreateAuthenticatedRequest("GET", apiURL)
if err != nil {
return nil, fmt.Errorf("failed to create request: %w", err)
}

client := &http.Client{}
resp, err := client.Do(req)
if err != nil {
return nil, fmt.Errorf("failed to fetch release: %w", err)
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
}

var release Release
if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
return nil, fmt.Errorf("failed to decode release: %w", err)
}

return &release, nil
}

// FindPlatformAsset finds the asset matching the current platform
// projectName can be empty to match any asset with the platform
func FindPlatformAsset(release *Release, projectName string) (*Asset, error) {
platform := GetCurrentPlatform()

if platform.OS != "darwin" && platform.OS != "linux" {
return nil, fmt.Errorf("unsupported OS: %s", platform.OS)
}
if platform.Arch != "amd64" && platform.Arch != "arm64" {
return nil, fmt.Errorf("unsupported architecture: %s", platform.Arch)
}

if len(release.Assets) == 0 {
return nil, fmt.Errorf("release %s has no assets (the build may have failed)", release.TagName)
}

for _, asset := range release.Assets {
// Check if asset contains platform identifiers
if !strings.Contains(asset.Name, platform.OS) || !strings.Contains(asset.Name, platform.Arch) {
continue
}

// If no project name specified, return first matching platform
if projectName == "" {
return &asset, nil
}

// If project name specified, check if asset starts with it
if strings.HasPrefix(asset.Name, projectName) {
return &asset, nil
}
}

var assetNames []string
for _, asset := range release.Assets {
assetNames = append(assetNames, asset.Name)
}

if projectName != "" {
return nil, fmt.Errorf(
"no binary found for %s/%s in release %s (expected name pattern: %s-*-%s-%s or %s--*-%s-%s). Available assets: %v",
platform.OS, platform.Arch,
release.TagName, projectName, platform.OS, platform.Arch, projectName, platform.OS, platform.Arch,
assetNames,
)
}

return nil, fmt.Errorf(
"no binary found for %s/%s in release %s. Available assets: %v",
platform.OS, platform.Arch, release.TagName, assetNames,
)
}

// DownloadAsset downloads a GitHub release asset to the specified path
func DownloadAsset(asset *Asset, destPath string) error {
// For private releases, we must use asset.URL (GitHub API URL) not asset.BrowserDownloadURL
// asset.URL works with authentication, asset.BrowserDownloadURL returns 404 for private releases
downloadURL := asset.URL

req, err := CreateAuthenticatedRequest("GET", downloadURL)
if err != nil {
return fmt.Errorf("failed to create request: %w", err)
}

req.Header.Set("Accept", "application/octet-stream")
client := &http.Client{}
resp, err := client.Do(req)
if err != nil {
return fmt.Errorf("failed to download binary: %w", err)
}
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
if resp.StatusCode == 404 {
return fmt.Errorf("download failed with status 404 (asset not found). This may mean:\n"+
"  1. The release exists but has no assets (build may have failed)\n"+
"  2. The asset name doesn't match what was expected\n"+
"  3. The release is private and requires authentication\n"+
"  Download URL: %s", downloadURL)
}
return fmt.Errorf("download failed with status %d", resp.StatusCode)
}

if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
return fmt.Errorf("failed to create cache directory: %w", err)
}

outFile, err := os.Create(destPath)
if err != nil {
return fmt.Errorf("failed to create file: %w", err)
}
defer outFile.Close()

if _, err := io.Copy(outFile, resp.Body); err != nil {
return fmt.Errorf("failed to write binary: %w", err)
}

if err := os.Chmod(destPath, 0755); err != nil {
return fmt.Errorf("failed to make binary executable: %w", err)
}

return nil
}

// DownloadReleaseAsset is a convenience function that downloads an asset from a GitHub release
// by tag name for the current platform
func DownloadReleaseAsset(owner, repo, tag, projectName, destPath string) error {
release, err := GetReleaseByTag(owner, repo, tag)
if err != nil {
return err
}

asset, err := FindPlatformAsset(release, projectName)
if err != nil {
return err
}

return DownloadAsset(asset, destPath)
}

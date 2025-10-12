package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		URL                string `json:"url"`
	} `json:"assets"`
}

// PRInfo contains parsed PR information
type PRInfo struct {
	Owner   string
	Repo    string
	PRNum   int
	Project string
}

// parsePRURL extracts owner, repo, and PR number from a GitHub PR URL
func parsePRURL(prURL string) (*PRInfo, error) {
	// Support various PR URL formats:
	// https://github.com/owner/repo/pull/123
	// github.com/owner/repo/pull/123
	re := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/pull/(\d+)`)
	matches := re.FindStringSubmatch(prURL)
	if matches == nil || len(matches) < 4 {
		return nil, fmt.Errorf("invalid GitHub PR URL: %s", prURL)
	}

	prNum, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid PR number: %s", matches[3])
	}

	return &PRInfo{
		Owner: matches[1],
		Repo:  matches[2],
		PRNum: prNum,
	}, nil
}

// getCacheDir returns the cache directory path
func getCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	cacheDir := filepath.Join(homeDir, ".cache", "prrun")
	return cacheDir, nil
}

// getGitHubToken attempts to get a GitHub token from multiple sources
// It tries in order: GITHUB_TOKEN, MISE_GITHUB_TOKEN, gh CLI tool
func getGitHubToken() string {
	// Try GITHUB_TOKEN environment variable first
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token
	}

	// Try MISE_GITHUB_TOKEN environment variable
	if token := os.Getenv("MISE_GITHUB_TOKEN"); token != "" {
		return token
	}

	// Try gh CLI tool
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output))
	}

	// No token found
	return ""
}

// createAuthenticatedRequest creates an HTTP request with authentication if available
func createAuthenticatedRequest(method, url string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication token if available
	if token := getGitHubToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}

// findPRRelease finds the latest release for a specific PR
func findPRRelease(owner, repo string, prNum int, project string) (*GitHubRelease, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)

	req, err := createAuthenticatedRequest("GET", apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode releases: %w", err)
	}

	// Find releases matching the PR pattern: project--pr-N.X or project/pr-N.X
	prPattern := fmt.Sprintf("pr-%d.", prNum)
	var matchingReleases []GitHubRelease

	for _, release := range releases {
		// Check if this release is for the specified project and PR
		// Format: project--pr-N.X or project/pr-N.X
		tagName := release.TagName
		if project != "" {
			// If project is specified, look for project--pr-N.X or project/pr-N.X
			if strings.Contains(tagName, project) && strings.Contains(tagName, prPattern) {
				matchingReleases = append(matchingReleases, release)
			}
		} else {
			// If no project specified, look for any pr-N.X release
			if strings.Contains(tagName, prPattern) {
				matchingReleases = append(matchingReleases, release)
			}
		}
	}

	if len(matchingReleases) == 0 {
		if project != "" {
			return nil, fmt.Errorf("no releases found for PR #%d and project %s", prNum, project)
		}
		return nil, fmt.Errorf("no releases found for PR #%d", prNum)
	}

	// Return the first (latest) matching release
	return &matchingReleases[0], nil
}

// getPlatformBinaryName returns the binary name for the current platform
func getPlatformBinaryName(release *GitHubRelease, projectName string) (string, string, error) {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Map Go arch names to the names used in releases
	if osName == "darwin" {
		osName = "darwin"
	} else if osName == "linux" {
		osName = "linux"
	} else {
		return "", "", fmt.Errorf("unsupported OS: %s", osName)
	}

	if archName == "amd64" {
		archName = "amd64"
	} else if archName == "arm64" {
		archName = "arm64"
	} else {
		return "", "", fmt.Errorf("unsupported architecture: %s", archName)
	}

	// Debug: Show available assets
	if len(release.Assets) == 0 {
		return "", "", fmt.Errorf("release %s has no assets (the build may have failed)", release.TagName)
	}

	fmt.Printf("Available assets (%d):\n", len(release.Assets))
	for _, asset := range release.Assets {
		fmt.Printf("  - %s\n", asset.Name)
	}
	fmt.Println()

	// Find the matching asset
	// Expected formats:
	// 1. project-version-os-arch (e.g., dissect-pr-123.1-linux-amd64)
	// 2. project--version-os-arch (e.g., dissect--pr-123.1-linux-amd64)
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, osName) && strings.Contains(asset.Name, archName) {
			if projectName == "" {
				// No project filter, return first matching asset
				return asset.Name, asset.URL, nil
			}
			// Check if asset starts with project name (handles both single and double dash)
			if strings.HasPrefix(asset.Name, projectName) {
				return asset.Name, asset.URL, nil
			}
		}
	}

	return "", "", fmt.Errorf("no binary found for %s/%s in release %s (expected name pattern: %s-*-%s-%s or %s--*-%s-%s)", osName, archName, release.TagName, projectName, osName, archName, projectName, osName, archName)
}

// downloadBinary downloads a binary from a URL to a local path
func downloadBinary(downloadURL, destPath string) error {
	fmt.Printf("Downloading binary from %s...\n", downloadURL)

	req, err := createAuthenticatedRequest("GET", downloadURL)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set Accept header to get binary data instead of JSON metadata
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

	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create the destination file
	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	// Copy the downloaded content
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write binary: %w", err)
	}

	// Make the binary executable
	if err := os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	fmt.Printf("✓ Binary cached at %s\n", destPath)
	return nil
}

// runBinary executes the binary with the given arguments
func runBinary(binaryPath string, args []string) error {
	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func printUsage() {
	fmt.Println("Usage: prrun <github-pr-url> [project-name] [-- args...]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  prrun https://github.com/neongreen/mono/pull/123 dissect")
	fmt.Println("  prrun https://github.com/neongreen/mono/pull/123 dissect -- --help")
	fmt.Println("  prrun github.com/neongreen/mono/pull/123 markdown-format -- file.md")
	fmt.Println()
	fmt.Println("The tool will:")
	fmt.Println("  1. Find the GitHub release for the PR")
	fmt.Println("  2. Download the binary to ~/.cache/prrun/")
	fmt.Println("  3. Run the binary with any arguments after --")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Parse arguments
	prURL := os.Args[1]
	var projectName string
	var binaryArgs []string

	// Find the separator "--" which separates our args from binary args
	separatorIdx := -1
	for i, arg := range os.Args {
		if arg == "--" {
			separatorIdx = i
			break
		}
	}

	if separatorIdx > 0 {
		// Arguments after "--" go to the binary
		binaryArgs = os.Args[separatorIdx+1:]
		// Check if there's a project name before "--"
		if separatorIdx > 2 {
			projectName = os.Args[2]
		}
	} else {
		// No separator, check if there's a project name
		if len(os.Args) > 2 {
			projectName = os.Args[2]
		}
	}

	// Parse the PR URL
	prInfo, err := parsePRURL(prURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// If project name is provided, use it; otherwise use prInfo.Project if available
	if projectName == "" && prInfo.Project != "" {
		projectName = prInfo.Project
	}

	fmt.Printf("Looking for PR #%d in %s/%s", prInfo.PRNum, prInfo.Owner, prInfo.Repo)
	if projectName != "" {
		fmt.Printf(" (project: %s)", projectName)
	}
	fmt.Println()

	// Find the PR release
	release, err := findPRRelease(prInfo.Owner, prInfo.Repo, prInfo.PRNum, projectName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found release: %s\n", release.TagName)

	// Get the binary name for the current platform
	binaryName, downloadURL, err := getPlatformBinaryName(release, projectName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Get cache directory
	cacheDir, err := getCacheDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to get cache directory: %v\n", err)
		os.Exit(1)
	}

	// Create a cache path based on the release tag and binary name
	cachePath := filepath.Join(cacheDir, release.TagName, binaryName)

	// Check if binary is already cached
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		// Download the binary
		if err := downloadBinary(downloadURL, cachePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("✓ Using cached binary at %s\n", cachePath)
	}

	// Run the binary
	fmt.Println()
	if len(binaryArgs) > 0 {
		fmt.Printf("Running: %s %s\n", binaryName, strings.Join(binaryArgs, " "))
	} else {
		fmt.Printf("Running: %s\n", binaryName)
	}
	fmt.Println(strings.Repeat("-", 50))

	if err := runBinary(cachePath, binaryArgs); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error running binary: %v\n", err)
		os.Exit(1)
	}
}

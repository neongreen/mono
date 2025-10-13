package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func parsePRURL(prURL string) (*PRInfo, error) {
	re := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/pull/(\d+)`)
	matches := re.FindStringSubmatch(prURL)
	if matches == nil || len(matches) < 4 {
		return nil, fmt.Errorf("invalid GitHub PR URL: %s", prURL)
	}
	prNum, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid PR number: %s", matches[3])
	}
	return &PRInfo{Owner: matches[1], Repo: matches[2], PRNum: prNum}, nil
}

func findPRRelease(owner, repo string,
	prNum int, project string) (*GitHubRelease, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases",
		owner, repo)
	req,
		err := createAuthenticatedRequest("GET", apiURL)
	if err !=
		nil {
		return nil, fmt.Errorf("failed to create request: %w",
			err,
		)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err !=

		nil { // parsePRURL extracts owner, repo, and PR number from a GitHub PR URL
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
	prPattern := fmt.Sprintf("pr-%d.", prNum)
	var matchingReleases []GitHubRelease
	for _, release := range releases {
		tagName := release.TagName
		if project != "" {
			if strings.Contains(tagName, project) && strings.Contains(tagName, prPattern) {
				matchingReleases = append(matchingReleases, release)
			}
		} else {
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
	return &matchingReleases[0], nil
}

func getPlatformBinaryName(release *GitHubRelease, projectName string) (string, string, error) {
	osName := runtime.GOOS
	archName := runtime.GOARCH
	if osName == "darwin" {
		osName = "darwin"
	} else if osName ==
		"linux" {
		osName = "linux"
	} else {
		return "", "", fmt.Errorf("unsupported OS: %s",

			osName)
	}
	if archName == "amd64" {
		archName = "amd64"
	} else if archName == "arm64" {
		archName = "arm64"
	} else {
		return "", "", fmt.Errorf("unsupported architecture: %s",
			archName)
	}
	if len(
		release.Assets) == 0 {
		return "", "", fmt.Errorf(
			"release %s has no assets (the build may have failed)",

			release.TagName)
	}

	for _, asset := range release.Assets {
		if strings.Contains(
			asset.Name, osName) && strings.Contains(asset.Name, archName) {
			if projectName == "" {
				// For private releases, we must use asset.URL (GitHub API URL) not asset.BrowserDownloadURL
				// asset.URL works with authentication, asset.BrowserDownloadURL returns 404 for private releases
				return asset.Name, asset.URL, nil
			}
			if strings.
				HasPrefix(asset.Name, projectName) {
				// For private releases, we must use asset.URL (GitHub API URL) not asset.BrowserDownloadURL
				// asset.URL works with authentication, asset.BrowserDownloadURL returns 404 for private releases
				return asset.Name, asset.URL, nil
			}
		}
	}
	return "", "", fmt.Errorf(
		"no binary found for %s/%s in release %s (expected name pattern: %s-*-%s-%s or %s--*-%s-%s). Available assets: %v",

		osName, archName,
		release.TagName, projectName, osName, archName, projectName,

		osName, archName,
		func() []string {
			var names []string
			for _, asset := range release.Assets {
				names = append(names,

					asset.
						Name)
			}
			return names
		}(),
	)
}

// findAllPRReleases finds all releases for a given PR number
func findAllPRReleases(owner, repo string, prNum int) ([]GitHubRelease, error) {
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
	
	// Find all releases for this PR
	prPattern := fmt.Sprintf("pr-%d.", prNum)
	var matchingReleases []GitHubRelease
	for _, release := range releases {
		if strings.Contains(release.TagName, prPattern) {
			matchingReleases = append(matchingReleases, release)
		}
	}
	
	return matchingReleases, nil
}

// extractProjectFromTag extracts the project name from a release tag
// e.g., "dissect--pr-123.1" -> "dissect"
func extractProjectFromTag(tag string) string {
	// Tag format: project--pr-N.X or project--main.X
	parts := strings.Split(tag, "--")
	if len(parts) >= 1 {
		return parts[0]
	}
	return ""
}

// extractUniqueProjects extracts unique project names from a list of releases
func extractUniqueProjects(releases []GitHubRelease) []string {
	projectSet := make(map[string]bool)
	var uniqueProjects []string
	for _, r := range releases {
		project := extractProjectFromTag(r.TagName)
		if !projectSet[project] {
			projectSet[project] = true
			uniqueProjects = append(uniqueProjects, project)
		}
	}
	return uniqueProjects
}

// checkWorkflowApproval checks if the workflow run for a PR is pending approval
func checkWorkflowApproval(owner, repo string, prNum int) {
	// Get the PR details to find associated workflow runs
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, prNum)
	req, err := createAuthenticatedRequest("GET", apiURL)
	if err != nil {
		// Silently fail - this is just a warning feature
		return
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return
	}
	
	var prDetails struct {
		Head struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&prDetails); err != nil {
		return
	}
	
	// Now get workflow runs for this commit
	apiURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s/check-runs", owner, repo, prDetails.Head.SHA)
	req, err = createAuthenticatedRequest("GET", apiURL)
	if err != nil {
		return
	}
	
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return
	}
	
	var checkRuns struct {
		CheckRuns []struct {
			Name       string `json:"name"`
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
		} `json:"check_runs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&checkRuns); err != nil {
		return
	}
	
	// Check if any workflow run is waiting for approval
	hasWaitingRelease := false
	for _, run := range checkRuns.CheckRuns {
		if strings.Contains(strings.ToLower(run.Name), "release") {
			if run.Status == "waiting" || run.Status == "action_required" || run.Status == "queued" {
				hasWaitingRelease = true
				break
			}
		}
	}
	
	if hasWaitingRelease {
		fmt.Fprintf(os.Stderr, "Warning: The release workflow for PR #%d may be pending approval.\n", prNum)
		fmt.Fprintf(os.Stderr, "         Check GitHub Actions at https://github.com/%s/%s/pull/%d/checks to approve it.\n\n", owner, repo, prNum)
	}
}

func downloadBinary(downloadURL, destPath string) error {
	req, err := createAuthenticatedRequest("GET", downloadURL)
	if err !=
		nil {
		return fmt.Errorf("failed to create request: %w",
			err)
	}

	req.Header.
		Set("Accept",
			"application/octet-stream")
	client := &http.Client{}
	resp,
		err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w",

			err)
	}
	defer resp.Body.
		Close()
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

	if err := os.
		MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w",

			err)
	}
	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w",

			err)
	}
	defer outFile.Close()
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.
			Errorf("failed to write binary: %w",
				err)
	}
	if err := os.Chmod(destPath,
		0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w",

			err)
	}
	return nil
}

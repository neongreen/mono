package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/neongreen/mono/ghrelease"
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
		err := ghrelease.CreateAuthenticatedRequest("GET", apiURL)
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
	// Convert GitHubRelease to ghrelease.Release
	ghrRelease := &ghrelease.Release{
		TagName:    release.TagName,
		Name:       release.Name,
		Prerelease: release.Prerelease,
		Assets:     make([]ghrelease.Asset, len(release.Assets)),
	}
	for i, asset := range release.Assets {
		ghrRelease.Assets[i] = ghrelease.Asset{
			Name:               asset.Name,
			URL:                asset.URL,
			BrowserDownloadURL: asset.BrowserDownloadURL,
		}
	}

	asset, err := ghrelease.FindPlatformAsset(ghrRelease, projectName)
	if err != nil {
		return "", "", err
	}

	return asset.Name, asset.URL, nil
}

// findAllPRReleases finds all releases for a given PR number
func findAllPRReleases(owner, repo string, prNum int) ([]GitHubRelease, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)
	req, err := ghrelease.CreateAuthenticatedRequest("GET", apiURL)
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
	req, err := ghrelease.CreateAuthenticatedRequest("GET", apiURL)
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
	req, err = ghrelease.CreateAuthenticatedRequest("GET", apiURL)
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
		fmt.Printf("Warning: The release workflow for PR #%d may be pending approval.\n", prNum)
		fmt.Printf("         Check GitHub Actions at https://github.com/%s/%s/pull/%d/checks to approve it.\n\n", owner, repo, prNum)
	}
}

func downloadBinary(downloadURL, destPath string) error {
	// Create a temporary asset object for the download
	asset := &ghrelease.Asset{
		URL: downloadURL,
	}
	return ghrelease.DownloadAsset(asset, destPath)
}

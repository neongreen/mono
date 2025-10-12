package main

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

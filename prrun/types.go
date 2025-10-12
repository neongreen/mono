package main

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name string `json:"name"`
		// URL is the GitHub API URL for the asset (e.g., https://api.github.com/repos/.../assets/123)
		// This URL works with authentication and is REQUIRED for private releases
		URL string `json:"url"`
		// BrowserDownloadURL is the direct download URL (e.g., https://github.com/.../releases/download/...)
		// This URL does NOT work for private releases - only use URL field
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// PRInfo contains parsed PR information
type PRInfo struct {
	Owner   string
	Repo    string
	PRNum   int
	Project string
}

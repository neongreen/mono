package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// These variables are set at build time using -ldflags
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// printVersion prints version information and checks for updates
func printVersion() {
	fmt.Printf("prrun version: %s\n", Version)
	fmt.Printf("Git commit: %s\n", GitCommit)
	fmt.Printf("Build time: %s\n", BuildTime)

	// Check for newer version
	checkForNewerVersion()
}

// checkForNewerVersion checks if there's a newer version available in main
func checkForNewerVersion() {
	// Only check if we have a valid commit hash
	if GitCommit == "unknown" || GitCommit == "" {
		return
	}

	// Get commits from main branch
	apiURL := "https://api.github.com/repos/neongreen/mono/commits?sha=main&per_page=1"
	req, err := createAuthenticatedRequest("GET", apiURL)
	if err != nil {
		// Silently fail - this is just informational
		return
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		// Silently fail
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var commits []struct {
		SHA    string `json:"sha"`
		Commit struct {
			Author struct {
				Date string `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return
	}

	if len(commits) == 0 {
		return
	}

	latestCommit := commits[0]
	latestSHA := latestCommit.SHA

	// Compare commit SHAs (case-insensitive, allowing for short hashes)
	currentCommit := strings.ToLower(GitCommit)
	latestCommitLower := strings.ToLower(latestSHA)

	// Check if current commit is a prefix of latest, or exact match
	if strings.HasPrefix(latestCommitLower, currentCommit) || currentCommit == latestCommitLower {
		fmt.Println("\nYou are using the latest version from main.")
	} else {
		fmt.Printf("\n⚠️  A newer version is available!\n")
		fmt.Printf("   Latest commit: %s\n", latestSHA[:12])
		fmt.Printf("   Your commit:   %s\n", GitCommit)
		fmt.Printf("\nTo update prrun, rebuild from the latest main branch:\n")
		fmt.Printf("  cd prrun && git pull && go build -ldflags=\"-X main.GitCommit=$(git rev-parse HEAD) -X main.BuildTime=$(date -u +%%Y-%%m-%%dT%%H:%%M:%%SZ)\" -o prrun .\n")
	}
}

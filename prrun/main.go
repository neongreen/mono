package main

import (
	"fmt"

	"os"
	"os/exec"
	"path/filepath"
)

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

	// Only show debug info on errors, not on success

	// Find the PR release
	release, err := findPRRelease(prInfo.Owner, prInfo.Repo, prInfo.PRNum, projectName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Looking for PR #%d in %s/%s", prInfo.PRNum, prInfo.Owner, prInfo.Repo)
		if projectName != "" {
			fmt.Fprintf(os.Stderr, " (project: %s)", projectName)
		}
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}

	// Get the binary name for the current platform
	binaryName, downloadURL, err := getPlatformBinaryName(release, projectName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Found release: %s\n", release.TagName)
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
	}

	// Run the binary (no debug output on success)

	if err := runBinary(cachePath, binaryArgs); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error running binary: %v\n", err)
		os.Exit(1)
	}
}

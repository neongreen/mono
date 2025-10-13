package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Check for --help or -h flag (only as first argument)
	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		printUsage()
		os.Exit(0)
	}
	if os.Args[1] == "--version" || os.Args[1] == "-v" {
		printVersion()
		os.Exit(0)
	}

	// Parse arguments
	prURL := os.Args[1]
	var projectName string
	var binaryArgs []string
	var explicitProject bool

	// Parse flags and arguments
	i := 2
	for i < len(os.Args) {
		arg := os.Args[i]
		if arg == "--project" || arg == "-p" {
			// Next argument should be project name
			if i+1 >= len(os.Args) {
				fmt.Fprintf(os.Stderr, "Error: %s flag requires a project name\n", arg)
				os.Exit(1)
			}
			projectName = os.Args[i+1]
			explicitProject = true
			i += 2
		} else if arg == "--" {
			// Everything after -- goes to binary
			binaryArgs = os.Args[i+1:]
			break
		} else {
			// Everything else goes to binary
			binaryArgs = os.Args[i:]
			break
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

	// Find the PR release (or all releases if no project specified)
	var release *GitHubRelease
	if projectName != "" {
		// Project explicitly specified
		release, err = findPRRelease(prInfo.Owner, prInfo.Repo, prInfo.PRNum, projectName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Looking for PR #%d in %s/%s (project: %s)", prInfo.PRNum, prInfo.Owner, prInfo.Repo, projectName)
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
			os.Exit(1)
		}
	} else {
		// No project specified - try to detect from releases
		releases, err := findAllPRReleases(prInfo.Owner, prInfo.Repo, prInfo.PRNum)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Looking for PR #%d in %s/%s", prInfo.PRNum, prInfo.Owner, prInfo.Repo)
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
			os.Exit(1)
		}

		if len(releases) == 0 {
			fmt.Fprintf(os.Stderr, "Error: no releases found for PR #%d\n", prInfo.PRNum)
			os.Exit(1)
		}

		if len(releases) > 1 && !explicitProject {
			// Multiple projects detected - extract unique project names
			uniqueProjects := extractUniqueProjects(releases)

			// Only error if there are truly multiple different projects
			if len(uniqueProjects) > 1 {
				fmt.Fprintf(os.Stderr, "Error: multiple projects found for PR #%d:\n", prInfo.PRNum)
				for _, project := range uniqueProjects {
					fmt.Fprintf(os.Stderr, "  - %s\n", project)
				}
				fmt.Fprintf(os.Stderr, "\nPlease specify a project with --project or -p flag:\n")
				fmt.Fprintf(os.Stderr, "  prrun %s --project <project-name> %s\n", prURL, strings.Join(binaryArgs, " "))
				os.Exit(1)
			}
		}

		release = &releases[0]
		projectName = extractProjectFromTag(release.TagName)
	}

	// Check if workflow is pending approval
	checkWorkflowApproval(prInfo.Owner, prInfo.Repo, prInfo.PRNum)

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

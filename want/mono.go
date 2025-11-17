package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/neongreen/mono/lib/cli"
	"github.com/neongreen/mono/lib/ghrelease"
	"github.com/neongreen/mono/want/cmd"
)

// Type aliases for cmd package types
type PRInfo = cmd.PRInfo

// installMonoRelease installs a specific version of a project from neongreen/mono
func installMonoRelease(project, version string, dryRun bool, planJson bool) {
	// Special handling for local builds
	if version == "local" {
		buildMonoFromLocal(project, dryRun, planJson)
		return
	}

	// Special handling for tk-vscode extension
	if project == "tk-vscode" {
		if after, ok := strings.CutPrefix(version, "pr-"); ok {
			prStr := after
			parts := strings.Split(prStr, ".")
			var prNumber int
			n, err := fmt.Sscanf(parts[0], "%d", &prNumber)
			if err != nil || n != 1 {
				fmt.Printf("Error: Invalid PR number in '%s'\n", version)
				os.Exit(1)
			}
			buildVSCodeExtensionFromPR(project, prNumber, dryRun, planJson)
			return
		}

		var refDescription string
		isCommitSHA := false
		if version == "main" || version == "master" {
			refDescription = fmt.Sprintf("latest commit on %s branch", version)
		} else if len(version) >= 7 && len(version) <= 40 && isHexString(version) {
			refDescription = fmt.Sprintf("commit %s", version)
			isCommitSHA = true
		} else {
			refDescription = fmt.Sprintf("branch %s", version)
		}

		buildVSCodeExtensionFromSource(project, version, refDescription, isCommitSHA, dryRun, planJson)
		return
	}

	if after, ok := strings.CutPrefix(version, "pr-"); ok {

		prStr := after

		parts := strings.Split(prStr, ".")
		var prNumber int
		n, err := fmt.Sscanf(parts[0], "%d", &prNumber)
		if err != nil || n != 1 {
			fmt.Printf("Error: Invalid PR number in '%s'\n", version)
			os.Exit(1)
		}

		tag := fmt.Sprintf("%s--%s", project, version)
		_, err = ghrelease.GetReleaseByTag("neongreen", "mono", tag)
		if err != nil {

			if !planJson && !dryRun {
				fmt.Printf("No release found for %s (would be tagged as %s)\n", version, tag)
				fmt.Printf("Building from PR #%d instead...\n", prNumber)
				fmt.Println()
			}
			buildMonoFromPR(project, prNumber, dryRun, planJson)
			return
		}

	}

	tag := fmt.Sprintf("%s--%s", project, version)
	_, err := ghrelease.GetReleaseByTag("neongreen", "mono", tag)
	if err != nil {

		var refDescription string
		isCommitSHA := false
		if version == "main" || version == "master" {
			refDescription = fmt.Sprintf("latest commit on %s branch", version)
		} else if len(version) >= 7 && len(version) <= 40 && isHexString(version) {
			refDescription = fmt.Sprintf("commit %s", version)
			isCommitSHA = true
		} else {
			refDescription = fmt.Sprintf("branch %s", version)
		}

		if !planJson && !dryRun {
			fmt.Printf("No release found for %s (would be tagged as %s)\n", version, tag)
			fmt.Printf("Building from %s instead...\n", refDescription)
			fmt.Println()
		}
		buildMonoFromSource(project, version, refDescription, isCommitSHA, dryRun, planJson)
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	destDir := filepath.Join(homeDir, ".local", "bin")
	destPath := filepath.Join(destDir, project)

	platform := ghrelease.GetCurrentPlatform()
	plan := FulfillmentPlan{
		Requirement: fmt.Sprintf("mono %s@%s", project, version),
		Steps: []PlanStep{
			{
				Type:        "download",
				Description: fmt.Sprintf("Fetch release information from GitHub (platform: %s-%s)", platform.OS, platform.Arch),
				Command:     fmt.Sprintf("GET https://api.github.com/repos/neongreen/mono/releases/tags/%s", tag),
				Automatic:   true,
			},
			{
				Type:        "download",
				Description: fmt.Sprintf("Download %s version %s to %s", project, version, destPath),
				Command:     fmt.Sprintf("download asset matching platform to %s", destPath),
				Automatic:   true,
			},
			{
				Type:        "configure",
				Description: "Make binary executable",
				Command:     fmt.Sprintf("chmod +x %s", destPath),
				Automatic:   true,
			},
		},
	}

	if planJson {
		jsonStr, err := plan.ToJSON()
		if err != nil {
			fmt.Printf("Error: Failed to generate JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(jsonStr)
		return
	}

	if dryRun {
		fmt.Printf("Installing %s version %s from neongreen/mono...\n", project, version)
		fmt.Println()
		plan.PrintPlan()
		return
	}

	fmt.Printf("Installing %s version %s from neongreen/mono...\n", project, version)
	fmt.Println()
	plan.PrintPlan()

	fmt.Println("Fetching release information...")
	err = ghrelease.DownloadReleaseAsset("neongreen", "mono", tag, project, destPath)
	if err != nil {
		fmt.Printf("\nError: Failed to download release: %v\n", err)
		fmt.Println("\nTroubleshooting:")
		fmt.Println("  • Check that the release exists")
		fmt.Println("  • Check that there's an asset for your platform")
		platform := ghrelease.GetCurrentPlatform()
		fmt.Printf("  • Your platform: %s-%s\n", platform.OS, platform.Arch)
		fmt.Println()
		fmt.Println("To see available releases:")
		fmt.Printf("  want mono %s --list\n", project)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("%s Installed %s version %s to: %s\n", cli.Success("✓"), cli.Key(project), cli.Key(version), cli.Path(destPath))
	fmt.Println()

	printPathInfo(project, destPath)
}

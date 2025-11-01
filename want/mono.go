package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/neongreen/mono/lib/ghclient"
	"github.com/neongreen/mono/lib/ghrelease"
)

// listOpenPRs fetches open PRs that modify the given project
func listOpenPRs(project string) ([]PRInfo, error) {
	ctx := context.Background()
	client := ghclient.NewClient(ctx)

	prs, _, err := client.PullRequests.List(ctx, "neongreen", "mono", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list PRs: %w", err)
	}

	var projectPRs []PRInfo
	for _, pr := range prs {
		if pr.Number == nil || pr.Title == nil || pr.Head == nil || pr.Head.Ref == nil {
			continue
		}

		files, _, err := client.PullRequests.ListFiles(ctx, "neongreen", "mono", *pr.Number, nil)
		if err != nil {
			continue
		}

		modifiesProject := false
		projectPrefix := project + "/"
		for _, file := range files {
			if file.Filename == nil {
				continue
			}
			if strings.HasPrefix(*file.Filename, projectPrefix) {
				modifiesProject = true
				break
			}
		}

		if modifiesProject {
			projectPRs = append(projectPRs, PRInfo{
				Number: *pr.Number,
				Title:  *pr.Title,
				Branch: *pr.Head.Ref,
			})
		}
	}

	return projectPRs, nil
}

// listMonoReleases lists all releases for a project from neongreen/mono
func listMonoReleases(project string) {
	fmt.Printf("Fetching releases for %s from neongreen/mono...\n", project)
	fmt.Println()

	releases, err := ghrelease.ListReleases("neongreen", "mono")
	if err != nil {
		fmt.Printf("Error: Failed to fetch releases: %v\n", err)
		os.Exit(1)
	}

	// Filter releases for the specified project
	var projectReleases []ghrelease.Release
	prefix := project + "--"
	for _, release := range releases {
		if strings.HasPrefix(release.TagName, prefix) {
			projectReleases = append(projectReleases, release)
		}
	}

	fmt.Printf("Fetching open PRs for %s...\n", project)
	openPRs, err := listOpenPRs(project)
	if err != nil {
		fmt.Printf("Warning: Failed to fetch open PRs: %v\n", err)
		fmt.Println()
	}

	if len(projectReleases) == 0 && len(openPRs) == 0 {
		fmt.Printf("No releases or open PRs found for %s\n", project)
		fmt.Println("\nAvailable projects in mono:")
		fmt.Println("  printpdf, dissect, want, prrun, markdown-format, ingest, conf, claude-trace, tk")
		os.Exit(1)
	}

	if len(projectReleases) > 0 {
		fmt.Printf("\nAvailable releases for %s:\n", project)
		fmt.Println()
		for _, release := range projectReleases {

			version := strings.TrimPrefix(release.TagName, prefix)

			status := ""
			if release.Prerelease {
				status = " (prerelease)"
			}

			fmt.Printf("  %s%s\n", version, status)
		}
	}

	if len(openPRs) > 0 {
		fmt.Printf("\nOpen PRs for %s:\n", project)
		fmt.Println()
		for _, pr := range openPRs {
			fmt.Printf("  pr-%d: %s\n", pr.Number, pr.Title)
		}
	}

	fmt.Println()
	fmt.Println("To install a specific version:")
	fmt.Printf("  want mono %s@<version>\n", project)
	fmt.Println()
	fmt.Println("Examples:")
	if len(projectReleases) > 0 {
		version := strings.TrimPrefix(projectReleases[0].TagName, prefix)
		fmt.Printf("  want mono %s@%s\n", project, version)
	}
	if len(openPRs) > 0 {
		fmt.Printf("  want mono %s@pr-%d\n", project, openPRs[0].Number)
	}
}

// createGoBuildCommand creates a command to run 'go build' with the given arguments.
// If 'go' is not in PATH, it uses 'mise exec go@<version> -- go build' instead.
func createGoBuildCommand(args ...string) *exec.Cmd {
	if isToolAvailable("go") {
		return exec.Command("go", args...)
	}

	if !isMiseAvailable() {

		return exec.Command("go", args...)
	}

	miseArgs := []string{"exec", fmt.Sprintf("go@%s", goVersion), "--", "go"}
	miseArgs = append(miseArgs, args...)
	return exec.Command("mise", miseArgs...)
}

// getBuildPath determines the correct build path for a Go project.
// Some projects have their main.go in a cmd subdirectory, while others have it in the root.
// Returns the relative path to use with 'go build' (either "." or "./cmd").
func getBuildPath(projectDir string) string {

	cmdDir := filepath.Join(projectDir, "cmd")
	if _, err := os.Stat(cmdDir); err == nil {

		entries, err := os.ReadDir(cmdDir)
		if err != nil {

			return "."
		}
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".go" {

				return "./cmd"
			}
		}
	}

	return "."
}

// buildMonoFromSource builds a project from a branch or commit in the mono repository
func buildMonoFromSource(project, refSpec, refDescription string, isCommitSHA bool, dryRun bool, planJson bool) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	destDir := filepath.Join(homeDir, ".local", "bin")
	destPath := filepath.Join(destDir, project)

	// Build the plan - different commands for commits vs branches
	var cloneCmd string
	if isCommitSHA {
		cloneCmd = fmt.Sprintf("git clone https://github.com/neongreen/mono.git <tmpdir> && cd <tmpdir> && git checkout %s", refSpec)
	} else {
		cloneCmd = fmt.Sprintf("git clone --depth=1 --branch %s https://github.com/neongreen/mono.git <tmpdir>", refSpec)
	}

	plan := FulfillmentPlan{
		Requirement: fmt.Sprintf("mono %s@%s", project, refSpec),
		Steps: []PlanStep{
			{
				Type:        "download",
				Description: fmt.Sprintf("Clone neongreen/mono repository (%s)", refDescription),
				Command:     cloneCmd,
				Automatic:   true,
			},
			{
				Type:        "install",
				Description: fmt.Sprintf("Build %s from source", project),
				Command:     fmt.Sprintf("go build -o %s .", destPath),
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
		fmt.Printf("Building %s from %s...\n", project, refDescription)
		fmt.Println()
		plan.PrintPlan()
		return
	}

	fmt.Printf("Building %s from %s...\n", project, refDescription)
	fmt.Println()
	plan.PrintPlan()
	if !plan.ConfirmPlan() {
		fmt.Println("Cancelled.")
		os.Exit(0)
	}
	fmt.Println()

	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("want-mono-%s-*", project))
	if err != nil {
		fmt.Printf("Error: Failed to create temporary directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Cloning neongreen/mono (%s)...\n", refDescription)
	var cmd *exec.Cmd
	if isCommitSHA {

		cmd = exec.Command("git", "clone", "https://github.com/neongreen/mono.git", tmpDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("\nError: Failed to clone repository: %v\n", err)
			os.Exit(1)
		}

		cmd = exec.Command("git", "checkout", refSpec)
		cmd.Dir = tmpDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("\nError: Failed to checkout commit %s: %v\n", refSpec, err)
			fmt.Printf("Note: Make sure the commit '%s' exists in neongreen/mono\n", refSpec)
			os.Exit(1)
		}
	} else {

		cmd = exec.Command("git", "clone", "--depth=1", "--branch", refSpec,
			"https://github.com/neongreen/mono.git", tmpDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("\nError: Failed to clone repository: %v\n", err)
			fmt.Printf("Note: Make sure the branch or tag '%s' exists in neongreen/mono\n", refSpec)
			os.Exit(1)
		}
	}

	projectDir := filepath.Join(tmpDir, project)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		fmt.Printf("\nError: Project '%s' not found in repository\n", project)
		os.Exit(1)
	}

	fmt.Printf("\nBuilding %s...\n", project)
	buildPath := getBuildPath(projectDir)
	cmd = createGoBuildCommand("build", "-o", destPath, buildPath)
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: Failed to build %s: %v\n", project, err)
		os.Exit(1)
	}

	if err := os.Chmod(destPath, 0755); err != nil {
		fmt.Printf("Warning: Failed to make binary executable: %v\n", err)
	}

	fmt.Println()
	fmt.Printf("✓ Built and installed %s from %s to: %s\n", project, refDescription, destPath)
	fmt.Println()

	pathEnv := os.Getenv("PATH")
	if !strings.Contains(pathEnv, destDir) {
		fmt.Printf("Note: %s is not in your PATH\n", destDir)
		fmt.Println()
		fmt.Println("To use the binary, either:")
		fmt.Println("  1. Run it with the full path:")
		fmt.Printf("     %s\n", destPath)
		fmt.Println()
		fmt.Println("  2. Add the directory to your PATH:")
		configFile := getShellConfigFile()
		fmt.Printf("     echo 'export PATH=\"$PATH:%s\"' >> %s\n", destDir, configFile)
		fmt.Printf("     source %s\n", configFile)
	} else {
		fmt.Printf("✓ Binary is available in your PATH as: %s\n", project)
	}
}

// buildMonoFromPR builds a project from a PR branch
func buildMonoFromPR(project string, prNumber int, dryRun bool, planJson bool) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	destDir := filepath.Join(homeDir, ".local", "bin")
	destPath := filepath.Join(destDir, project)

	plan := FulfillmentPlan{
		Requirement: fmt.Sprintf("mono %s@pr-%d", project, prNumber),
		Steps: []PlanStep{
			{
				Type:        "download",
				Description: fmt.Sprintf("Fetch PR #%d information from GitHub", prNumber),
				Command:     fmt.Sprintf("GET https://api.github.com/repos/neongreen/mono/pulls/%d", prNumber),
				Automatic:   true,
			},
			{
				Type:        "download",
				Description: "Clone neongreen/mono repository (PR branch)",
				Command:     "git clone --depth=1 --branch <pr-branch> https://github.com/neongreen/mono.git <tmpdir>",
				Automatic:   true,
			},
			{
				Type:        "install",
				Description: fmt.Sprintf("Build %s from source", project),
				Command:     fmt.Sprintf("go build -o %s .", destPath),
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
		fmt.Printf("Building %s from PR #%d...\n", project, prNumber)
		fmt.Println()
		plan.PrintPlan()
		return
	}

	fmt.Printf("Building %s from PR #%d...\n", project, prNumber)
	fmt.Println()
	plan.PrintPlan()
	if !plan.ConfirmPlan() {
		fmt.Println("Cancelled.")
		os.Exit(0)
	}
	fmt.Println()

	ctx := context.Background()
	client := ghclient.NewClient(ctx)
	pr, _, err := client.PullRequests.Get(ctx, "neongreen", "mono", prNumber)
	if err != nil {
		fmt.Printf("Error: Failed to fetch PR #%d: %v\n", prNumber, err)
		os.Exit(1)
	}

	if pr.Head == nil || pr.Head.Ref == nil {
		fmt.Printf("Error: PR #%d has no head branch\n", prNumber)
		os.Exit(1)
	}

	branch := *pr.Head.Ref
	fmt.Printf("PR #%d: %s\n", prNumber, *pr.Title)
	fmt.Printf("Branch: %s\n", branch)
	fmt.Println()

	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("want-mono-%s-pr-%d-*", project, prNumber))
	if err != nil {
		fmt.Printf("Error: Failed to create temporary directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Cloning neongreen/mono (branch: %s)...\n", branch)
	cmd := exec.Command("git", "clone", "--depth=1", "--branch", branch,
		"https://github.com/neongreen/mono.git", tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: Failed to clone repository: %v\n", err)
		os.Exit(1)
	}

	projectDir := filepath.Join(tmpDir, project)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		fmt.Printf("\nError: Project '%s' not found in repository\n", project)
		os.Exit(1)
	}

	fmt.Printf("\nBuilding %s...\n", project)
	buildPath := getBuildPath(projectDir)
	cmd = createGoBuildCommand("build", "-o", destPath, buildPath)
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: Failed to build %s: %v\n", project, err)
		os.Exit(1)
	}

	if err := os.Chmod(destPath, 0755); err != nil {
		fmt.Printf("Warning: Failed to make binary executable: %v\n", err)
	}

	fmt.Println()
	fmt.Printf("✓ Built and installed %s from PR #%d to: %s\n", project, prNumber, destPath)
	fmt.Println()

	pathEnv := os.Getenv("PATH")
	if !strings.Contains(pathEnv, destDir) {
		fmt.Printf("Note: %s is not in your PATH\n", destDir)
		fmt.Println()
		fmt.Println("To use the binary, either:")
		fmt.Println("  1. Run it with the full path:")
		fmt.Printf("     %s\n", destPath)
		fmt.Println()
		fmt.Println("  2. Add the directory to your PATH:")
		configFile := getShellConfigFile()
		fmt.Printf("     echo 'export PATH=\"$PATH:%s\"' >> %s\n", destDir, configFile)
		fmt.Printf("     source %s\n", configFile)
	} else {
		fmt.Printf("✓ Binary is available in your PATH as: %s\n", project)
	}
}

// installMonoRelease installs a specific version of a project from neongreen/mono
func installMonoRelease(project, version string, dryRun bool, planJson bool) {
	// Special handling for tk-vscode extension
	if project == "tk-vscode" {
		if strings.HasPrefix(version, "pr-") {
			prStr := strings.TrimPrefix(version, "pr-")
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

		refDescription := version
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

	if strings.HasPrefix(version, "pr-") {

		prStr := strings.TrimPrefix(version, "pr-")

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

		refDescription := version
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

	alreadyInstalled := false
	if _, err := os.Stat(destPath); err == nil {
		alreadyInstalled = true
	}

	if alreadyInstalled {
		fmt.Println()
		fmt.Printf("⚠ %s is already installed at %s\n", project, destPath)
		fmt.Println()
		fmt.Print("Overwrite? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading response")
			os.Exit(1)
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			os.Exit(0)
		}
		fmt.Println()
	} else if !plan.ConfirmPlan() {
		fmt.Println("Cancelled.")
		os.Exit(0)
	}
	fmt.Println()

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
	fmt.Printf("✓ Installed %s version %s to: %s\n", project, version, destPath)
	fmt.Println()

	pathEnv := os.Getenv("PATH")
	if !strings.Contains(pathEnv, destDir) {
		fmt.Printf("Note: %s is not in your PATH\n", destDir)
		fmt.Println()
		fmt.Println("To use the binary, either:")
		fmt.Println("  1. Run it with the full path:")
		fmt.Printf("     %s\n", destPath)
		fmt.Println()
		fmt.Println("  2. Add the directory to your PATH:")
		configFile := getShellConfigFile()
		fmt.Printf("     echo 'export PATH=\"$PATH:%s\"' >> %s\n", destDir, configFile)
		fmt.Printf("     source %s\n", configFile)
	} else {
		fmt.Printf("✓ Binary is available in your PATH as: %s\n", project)
	}
}

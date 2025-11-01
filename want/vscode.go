package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/neongreen/mono/lib/ghclient"
)

const pnpmVersion = "9.15.4"

// createPnpmCommand creates a command to run 'pnpm' with the given arguments.
// If 'pnpm' is not in PATH, it uses 'mise exec pnpm@<version> -- pnpm' instead.
func createPnpmCommand(args ...string) *exec.Cmd {
	if isToolAvailable("pnpm") {
		return exec.Command("pnpm", args...)
	}

	if !isMiseAvailable() {
		// If mise is not available, fall back to pnpm anyway (will fail with a helpful error)
		return exec.Command("pnpm", args...)
	}

	miseArgs := []string{"exec", fmt.Sprintf("pnpm@%s", pnpmVersion), "--", "pnpm"}
	miseArgs = append(miseArgs, args...)
	return createMiseCommand(miseArgs...)
}

// buildVSCodeExtensionFromSource builds a VS Code extension from a branch or commit
func buildVSCodeExtensionFromSource(project, refSpec, refDescription string, isCommitSHA bool, dryRun bool, planJson bool) {
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
				Description: "Install dependencies",
				Command:     "pnpm install",
				Automatic:   true,
			},
			{
				Type:        "install",
				Description: "Compile TypeScript",
				Command:     "pnpm run compile",
				Automatic:   true,
			},
			{
				Type:        "install",
				Description: "Package extension",
				Command:     "pnpm run package",
				Automatic:   true,
			},
			{
				Type:        "install",
				Description: "Install VS Code extension",
				Command:     "code --install-extension tk-*.vsix --force",
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

	fmt.Println("\nInstalling dependencies...")
	cmd = createPnpmCommand("install")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: Failed to install dependencies: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nCompiling TypeScript...")
	cmd = createPnpmCommand("run", "compile")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: Failed to compile: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nPackaging extension...")
	cmd = createPnpmCommand("run", "package")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: Failed to package extension: %v\n", err)
		os.Exit(1)
	}

	// Find the .vsix file
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		fmt.Printf("\nError: Failed to read project directory: %v\n", err)
		os.Exit(1)
	}

	var vsixFile string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".vsix" {
			vsixFile = filepath.Join(projectDir, entry.Name())
			break
		}
	}

	if vsixFile == "" {
		fmt.Printf("\nError: No .vsix file found in %s\n", projectDir)
		os.Exit(1)
	}

	fmt.Printf("\nInstalling extension from %s...\n", filepath.Base(vsixFile))
	cmd = exec.Command("code", "--install-extension", vsixFile, "--force")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: Failed to install extension: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("✓ Built and installed %s from %s\n", project, refDescription)
}

// buildVSCodeExtensionFromPR builds a VS Code extension from a PR branch
func buildVSCodeExtensionFromPR(project string, prNumber int, dryRun bool, planJson bool) {
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
				Description: "Install dependencies",
				Command:     "pnpm install",
				Automatic:   true,
			},
			{
				Type:        "install",
				Description: "Compile TypeScript",
				Command:     "pnpm run compile",
				Automatic:   true,
			},
			{
				Type:        "install",
				Description: "Package extension",
				Command:     "pnpm run package",
				Automatic:   true,
			},
			{
				Type:        "install",
				Description: "Install VS Code extension",
				Command:     "code --install-extension tk-*.vsix --force",
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

	fmt.Println("\nInstalling dependencies...")
	cmd = createPnpmCommand("install")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: Failed to install dependencies: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nCompiling TypeScript...")
	cmd = createPnpmCommand("run", "compile")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: Failed to compile: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nPackaging extension...")
	cmd = createPnpmCommand("run", "package")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: Failed to package extension: %v\n", err)
		os.Exit(1)
	}

	// Find the .vsix file
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		fmt.Printf("\nError: Failed to read project directory: %v\n", err)
		os.Exit(1)
	}

	var vsixFile string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".vsix" {
			vsixFile = filepath.Join(projectDir, entry.Name())
			break
		}
	}

	if vsixFile == "" {
		fmt.Printf("\nError: No .vsix file found in %s\n", projectDir)
		os.Exit(1)
	}

	fmt.Printf("\nInstalling extension from %s...\n", filepath.Base(vsixFile))
	cmd = exec.Command("code", "--install-extension", vsixFile, "--force")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: Failed to install extension: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("✓ Built and installed %s from PR #%d\n", project, prNumber)
}

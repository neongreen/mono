package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/neongreen/mono/lib/cli"
	"github.com/neongreen/mono/lib/ghclient"
)

// NOTE: Build tasks for tk-vscode are defined in the top-level mise.toml
// If you update the build process here, also update /mise.toml to keep them in sync.
// The mise tasks define the canonical build process.
// Task names used: "tk-vscode:install-deps", "tk-vscode:build", "tk-vscode:install"
// See: /mise.toml (search for "tk-vscode")

// createMiseRunCommand creates a command to run a mise task.
// Example: createMiseRunCommand("tk-vscode:install-deps") runs the tk-vscode:install-deps task.
func createMiseRunCommand(taskPath string) *exec.Cmd {
	if !isMiseAvailable() {
		// If mise is not available, fail with a helpful error
		fmt.Println("Error: mise is required but not available")
		fmt.Println("Please install mise first with:")
		fmt.Println("  want mise")
		os.Exit(1)
	}
	return createMiseCommand("run", taskPath)
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
				Command:     "cd tk-vscode && mise run tk-vscode:install-deps",
				Automatic:   true,
			},
			{
				Type:        "install",
				Description: "Build extension (compile & package)",
				Command:     "cd tk-vscode && mise run tk-vscode:build",
				Automatic:   true,
			},
			{
				Type:        "install",
				Description: "Install VS Code extension",
				Command:     "cd tk-vscode && mise run tk-vscode:install",
				Automatic:   true,
			},
		},
	}

	if planJson {
		jsonStr, err := plan.ToJSON()
		if err != nil {
			cli.PrintError(fmt.Sprintf("Error: Failed to generate JSON: %v", err))
			os.Exit(1)
		}
		fmt.Println(jsonStr)
		return
	}

	if dryRun {
		fmt.Printf("Building %s from %s...\n", cli.Key(project), refDescription)
		fmt.Println()
		plan.PrintPlan()
		return
	}

	fmt.Printf("Building %s from %s...\n", cli.Key(project), refDescription)
	fmt.Println()
	plan.PrintPlan()
	fmt.Println()

	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("want-mono-%s-*", project))
	if err != nil {
		cli.PrintError(fmt.Sprintf("Error: Failed to create temporary directory: %v", err))
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
			cli.PrintError(fmt.Sprintf("\nError: Failed to clone repository: %v", err))
			os.Exit(1)
		}

		cmd = exec.Command("git", "checkout", refSpec)
		cmd.Dir = tmpDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			cli.PrintError(fmt.Sprintf("\nError: Failed to checkout commit %s: %v", cli.Key(refSpec), err))
			fmt.Printf("%s Make sure the commit '%s' exists in neongreen/mono\n", cli.Warning("Note:"), refSpec)
			os.Exit(1)
		}
	} else {
		cmd = exec.Command("git", "clone", "--depth=1", "--branch", refSpec,
			"https://github.com/neongreen/mono.git", tmpDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			cli.PrintError(fmt.Sprintf("\nError: Failed to clone repository: %v", err))
			fmt.Printf("%s Make sure the branch or tag '%s' exists in neongreen/mono\n", cli.Warning("Note:"), refSpec)
			os.Exit(1)
		}
	}

	projectDir := filepath.Join(tmpDir, project)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		cli.PrintError(fmt.Sprintf("\nError: Project '%s' not found in repository", cli.Key(project)))
		os.Exit(1)
	}

	fmt.Println("\nInstalling dependencies...")
	cmd = createMiseRunCommand("tk-vscode:install-deps")
	cmd.Dir = projectDir
	setMiseTrustedPath(cmd, tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		cli.PrintError(fmt.Sprintf("\nError: Failed to install dependencies: %v", err))
		os.Exit(1)
	}

	fmt.Println("\nBuilding extension (compile & package)...")
	cmd = createMiseRunCommand("tk-vscode:build")
	cmd.Dir = projectDir
	setMiseTrustedPath(cmd, tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		cli.PrintError(fmt.Sprintf("\nError: Failed to build extension: %v", err))
		os.Exit(1)
	}

	// Find the .vsix file
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		cli.PrintError(fmt.Sprintf("\nError: Failed to read project directory: %v", err))
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
		cli.PrintError(fmt.Sprintf("\nError: No .vsix file found in %s", cli.Path(projectDir)))
		os.Exit(1)
	}

	fmt.Printf("\nInstalling extension from %s...\n", filepath.Base(vsixFile))
	cmd = createMiseRunCommand("tk-vscode:install")
	cmd.Dir = projectDir
	setMiseTrustedPath(cmd, tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		cli.PrintError(fmt.Sprintf("\nError: Failed to install extension: %v", err))
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("%s Built and installed %s from %s\n", cli.Success("✓"), cli.Key(project), refDescription)
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
				Command:     "cd tk-vscode && mise run tk-vscode:install-deps",
				Automatic:   true,
			},
			{
				Type:        "install",
				Description: "Build extension (compile & package)",
				Command:     "cd tk-vscode && mise run tk-vscode:build",
				Automatic:   true,
			},
			{
				Type:        "install",
				Description: "Install VS Code extension",
				Command:     "cd tk-vscode && mise run tk-vscode:install",
				Automatic:   true,
			},
		},
	}

	if planJson {
		jsonStr, err := plan.ToJSON()
		if err != nil {
			cli.PrintError(fmt.Sprintf("Error: Failed to generate JSON: %v", err))
			os.Exit(1)
		}
		fmt.Println(jsonStr)
		return
	}

	if dryRun {
		fmt.Printf("Building %s from PR #%d...\n", cli.Key(project), prNumber)
		fmt.Println()
		plan.PrintPlan()
		return
	}

	fmt.Printf("Building %s from PR #%d...\n", cli.Key(project), prNumber)
	fmt.Println()
	plan.PrintPlan()
	fmt.Println()

	ctx := context.Background()
	client := ghclient.NewClient(ctx)
	pr, _, err := client.PullRequests.Get(ctx, "neongreen", "mono", prNumber)
	if err != nil {
		cli.PrintError(fmt.Sprintf("Error: Failed to fetch PR #%d: %v", prNumber, err))
		os.Exit(1)
	}

	if pr.Head == nil || pr.Head.Ref == nil {
		cli.PrintError(fmt.Sprintf("Error: PR #%d has no head branch", prNumber))
		os.Exit(1)
	}

	branch := *pr.Head.Ref
	fmt.Printf("PR #%d: %s\n", prNumber, *pr.Title)
	fmt.Printf("Branch: %s\n", cli.Key(branch))
	fmt.Println()

	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("want-mono-%s-pr-%d-*", project, prNumber))
	if err != nil {
		cli.PrintError(fmt.Sprintf("Error: Failed to create temporary directory: %v", err))
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Cloning neongreen/mono (branch: %s)...\n", cli.Key(branch))
	cmd := exec.Command("git", "clone", "--depth=1", "--branch", branch,
		"https://github.com/neongreen/mono.git", tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		cli.PrintError(fmt.Sprintf("\nError: Failed to clone repository: %v", err))
		os.Exit(1)
	}

	projectDir := filepath.Join(tmpDir, project)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		cli.PrintError(fmt.Sprintf("\nError: Project '%s' not found in repository", cli.Key(project)))
		os.Exit(1)
	}

	fmt.Println("\nInstalling dependencies...")
	cmd = createMiseRunCommand("tk-vscode:install-deps")
	cmd.Dir = projectDir
	setMiseTrustedPath(cmd, tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		cli.PrintError(fmt.Sprintf("\nError: Failed to install dependencies: %v", err))
		os.Exit(1)
	}

	fmt.Println("\nBuilding extension (compile & package)...")
	cmd = createMiseRunCommand("tk-vscode:build")
	cmd.Dir = projectDir
	setMiseTrustedPath(cmd, tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		cli.PrintError(fmt.Sprintf("\nError: Failed to build extension: %v", err))
		os.Exit(1)
	}

	// Find the .vsix file
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		cli.PrintError(fmt.Sprintf("\nError: Failed to read project directory: %v", err))
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
		cli.PrintError(fmt.Sprintf("\nError: No .vsix file found in %s", cli.Path(projectDir)))
		os.Exit(1)
	}

	fmt.Printf("\nInstalling extension from %s...\n", filepath.Base(vsixFile))
	cmd = createMiseRunCommand("tk-vscode:install")
	cmd.Dir = projectDir
	setMiseTrustedPath(cmd, tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		cli.PrintError(fmt.Sprintf("\nError: Failed to install extension: %v", err))
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("%s Built and installed %s from PR #%d\n", cli.Success("✓"), cli.Key(project), prNumber)
}

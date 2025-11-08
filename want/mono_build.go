package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/neongreen/mono/lib/cli"
)

// buildWithMiseAndCopy builds a project using mise and copies the binary to destination
func buildWithMiseAndCopy(project, workDir, destPath string) error {
	projectDir := filepath.Join(workDir, project)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return fmt.Errorf("project '%s' not found in repository", project)
	}

	fmt.Printf("\nBuilding %s using mise...\n", project)

	cmd := createMiseCommand("run", fmt.Sprintf("%s:build", project))
	cmd.Dir = workDir
	setMiseTrustedPath(cmd, workDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build %s: %w", project, err)
	}

	// Copy the binary from _build to destination
	buildSource := filepath.Join(workDir, "_build", project)
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	sourceData, err := os.ReadFile(buildSource)
	if err != nil {
		return fmt.Errorf("failed to read built binary from %s: %w", buildSource, err)
	}

	if err := os.WriteFile(destPath, sourceData, 0o755); err != nil {
		return fmt.Errorf("failed to write binary to %s: %w", destPath, err)
	}

	return nil
}

// printPathInfo prints information about whether the binary is in PATH
func printPathInfo(project, destPath string) {
	destDir := filepath.Dir(destPath)
	pathEnv := os.Getenv("PATH")

	if !strings.Contains(pathEnv, destDir) {
		fmt.Printf("%s %s is not in your PATH\n", cli.Warning("Note:"), cli.Path(destDir))
		fmt.Println()
		fmt.Println("To use the binary, either:")
		fmt.Println("  1. Run it with the full path:")
		fmt.Printf("     %s\n", cli.Path(destPath))
		fmt.Println()
		fmt.Println("  2. Add the directory to your PATH:")
		configFile := getShellConfigFile()
		fmt.Printf("     echo 'export PATH=\"$PATH:%s\"' >> %s\n", destDir, configFile)
		fmt.Printf("     source %s\n", configFile)
	} else {
		fmt.Printf("%s Binary is available in your PATH as: %s\n", cli.Success("✓"), cli.Key(project))
	}
}

// buildMonoFromLocal builds a project from the local mono repository checkout
func buildMonoFromLocal(project string, dryRun bool, planJson bool) {
	// Special handling for tk-vscode extension (local build)
	if project == "tk-vscode" {
		buildVSCodeExtensionFromSource(project, "local", "local checkout", false, dryRun, planJson)
		return
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error: Failed to get current directory: %v\n", err)
		os.Exit(1)
	}

	// Check if we're in a mono repository
	// Look for common project directories to confirm this is mono
	commonProjects := []string{"tk", "want", "dissect", "printpdf", "conf"}
	monoRepoDetected := false
	for _, proj := range commonProjects {
		projPath := filepath.Join(cwd, proj)
		if stat, err := os.Stat(projPath); err == nil && stat.IsDir() {
			monoRepoDetected = true
			break
		}
	}

	if !monoRepoDetected {
		fmt.Printf("Error: Current directory does not appear to be the mono repository\n")
		fmt.Printf("  Current directory: %s\n", cwd)
		fmt.Printf("  Expected to find project directories like: %v\n", commonProjects)
		os.Exit(1)
	}

	// Check if the project directory exists
	projectDir := filepath.Join(cwd, project)
	if stat, err := os.Stat(projectDir); os.IsNotExist(err) || !stat.IsDir() {
		fmt.Printf("Error: Project '%s' not found in current directory\n", project)
		fmt.Printf("  Looking for: %s\n", projectDir)
		os.Exit(1)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	destDir := filepath.Join(homeDir, ".local", "bin")
	destPath := filepath.Join(destDir, project)
	buildSource := filepath.Join(cwd, "_build", project)

	plan := FulfillmentPlan{
		Requirement: fmt.Sprintf("mono %s@local", project),
		Steps: []PlanStep{
			{
				Type:        "install",
				Description: fmt.Sprintf("Build %s from local checkout", project),
				Command:     fmt.Sprintf("mise run %s:build", project),
				Automatic:   true,
			},
			{
				Type:        "install",
				Description: fmt.Sprintf("Copy binary to %s", destPath),
				Command:     fmt.Sprintf("cp %s %s", buildSource, destPath),
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
		fmt.Printf("Building %s from local checkout...\n", project)
		fmt.Println()
		plan.PrintPlan()
		return
	}

	fmt.Printf("Building %s from local checkout...\n", project)
	fmt.Println()
	plan.PrintPlan()
	fmt.Println()

	if err := buildWithMiseAndCopy(project, cwd, destPath); err != nil {
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("%s Built and installed %s from local checkout to: %s\n", cli.Success("✓"), cli.Key(project), cli.Path(destPath))
	fmt.Println()

	printPathInfo(project, destPath)
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
				Command:     fmt.Sprintf("mise run %s:build", project),
				Automatic:   true,
			},
			{
				Type:        "install",
				Description: fmt.Sprintf("Copy binary to %s", destPath),
				Command:     fmt.Sprintf("cp <tmpdir>/_build/%s %s", project, destPath),
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

	if err := buildWithMiseAndCopy(project, tmpDir, destPath); err != nil {
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("%s Built and installed %s from %s to: %s\n", cli.Success("✓"), cli.Key(project), refDescription, cli.Path(destPath))
	fmt.Println()

	printPathInfo(project, destPath)
}

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/neongreen/mono/lib/ghrelease"
)

func handleWant(args []string) {

	fs := flag.NewFlagSet("want", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be done without doing it")
	planJson := fs.Bool("plan-json", false, "Output the fulfillment plan as JSON")
	fs.Parse(args)

	if fs.NArg() == 0 {
		fmt.Println("Error: no requirement specified")
		fmt.Println("Usage: want [--dry-run] [--plan-json] <requirement>")
		os.Exit(1)
	}

	requirement := fs.Arg(0)
	remainingArgs := fs.Args()[1:]

	if handler, ok := getCompoundHandler(requirement); ok {
		handler(remainingArgs, *dryRun, *planJson)
		return
	}

	if strings.Contains(requirement, "github.com") && (strings.Contains(requirement, "/releases/download/") || strings.Contains(requirement, "/releases/tag/")) {
		handleGitHubAsset(requirement, *dryRun, *planJson)
		return
	}

	if strings.Contains(requirement, "/") && (strings.Contains(requirement, "github.com") || strings.Contains(requirement, ".git")) {
		fmt.Printf("Error: Git repository cloning is not yet implemented\n")
		fmt.Printf("Requirement: %s\n", requirement)
		os.Exit(1)
	}

	installToolViaMise(requirement, *dryRun, *planJson)
}

func handleList() {
	fmt.Println("MVP: No requirements tracked yet")
	fmt.Println("\nThis command will show:")
	fmt.Println("  • Tools installed via want")
	fmt.Println("  • Repositories cloned via want")
	fmt.Println("  • Their current status")
}

func handleCheck() {
	fmt.Println("MVP: No requirements to check")
	fmt.Println("\nThis command will verify:")
	fmt.Println("  • Whether tracked requirements are still available")
	fmt.Println("  • Whether repositories are still cloned")
	fmt.Println("  • Status of each requirement")
}

func handleForget() {
	if len(os.Args) < 3 {
		fmt.Println("Error: no requirement specified")
		fmt.Println("Usage: want forget <requirement>")
		os.Exit(1)
	}

	requirement := os.Args[2]
	fmt.Printf("MVP: Would forget: %s\n", requirement)
	fmt.Println("This command will remove the requirement from tracking without uninstalling.")
}

// handleJsonCommand handles "want json <command>" - converts command output to JSON
func handleJsonCommand(args []string, dryRun bool, planJson bool) {
	if len(args) == 0 {
		fmt.Println("Error: no command specified")
		fmt.Println("Usage: want json <command>")
		fmt.Println("\nExample:")
		fmt.Println("  want json ps    # Get running processes as JSON")
		os.Exit(1)
	}

	commandStr := strings.Join(args, " ")

	plan := FulfillmentPlan{
		Requirement: fmt.Sprintf("json %s", commandStr),
		Steps:       buildToolInstallationPlan("jc"),
	}

	plan.Steps = append(plan.Steps, PlanStep{
		Type:        "execute",
		Description: "Execute command with jc",
		Command:     fmt.Sprintf("jc %s", commandStr),
		Automatic:   true,
		Safe:        isCommandSafe(commandStr),
	})

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
		plan.PrintPlan()
		return
	}

	plan.PrintPlan()

	if !plan.HasOnlySafeSteps() {
		if !plan.ConfirmPlan() {
			fmt.Println("Cancelled.")
			os.Exit(0)
		}
	}
	fmt.Println()

	if !isToolAvailable("jc") {
		if !isMiseAvailable() {
			if err := performMiseInstallation(); err != nil {
				fmt.Printf("\nError: %v\n", err)
				os.Exit(1)
			}

			if !isMiseAvailable() {
				fmt.Println("Error: mise installed but not yet available in this shell.")
				fmt.Println("Please ensure mise is on your PATH or run:")
				shellName := getShellName()
				fmt.Printf("  eval \"$(mise activate %s)\"\n", shellName)
				fmt.Println("Then rerun:")
				fmt.Printf("  want json %s\n", commandStr)
				os.Exit(1)
			}
		}

		fmt.Println("Installing jc via mise...")
		cmd := createMiseCommand("use", "-g", "jc")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println()
			fmt.Println("Error: Failed to install jc via mise")
			fmt.Println("Command failed: mise use -g jc")
			os.Exit(1)
		}
		fmt.Println()

		if !isToolAvailable("jc") {
			fmt.Println("Error: jc installation succeeded but jc is still not in PATH")
			fmt.Println("You may need to restart your shell or run:")
			shellName := getShellName()
			fmt.Printf("  eval \"$(mise activate %s)\"\n", shellName)
			os.Exit(1)
		}
	}

	cmd := exec.Command("jc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("\nError: Command failed: jc %s\n", commandStr)
		os.Exit(1)
	}
}

// handleMarkdownCommand handles "want md <url>" - converts URL to markdown
func handleMarkdownCommand(args []string, dryRun bool, planJson bool) {
	if len(args) == 0 {
		fmt.Println("Error: no URL specified")
		fmt.Println("Usage: want md <url>")
		fmt.Println("\nExample:")
		fmt.Println("  want md https://example.com    # Convert webpage to markdown")
		os.Exit(1)
	}

	url := args[0]

	plan := FulfillmentPlan{
		Requirement: fmt.Sprintf("md %s", url),
		Steps:       []PlanStep{},
	}

	hasMarkitdown := isToolAvailable("markitdown")
	hasUvx := isToolAvailable("uvx")
	hasMise := isMiseAvailable()
	hasUv := isToolAvailable("uv")

	if hasMarkitdown {

		plan.Steps = append(plan.Steps, PlanStep{
			Type:        "execute",
			Description: "Convert URL to markdown",
			Command:     fmt.Sprintf("markitdown %s", url),
			Automatic:   true,
		})
	} else if hasUvx {

		plan.Steps = append(plan.Steps, PlanStep{
			Type:        "execute",
			Description: "Run markitdown via uvx",
			Command:     fmt.Sprintf("uvx markitdown %s", url),
			Automatic:   true,
		})
	} else if hasMise {

		plan.Steps = append(plan.Steps, buildToolInstallationPlan("markitdown")...)
		plan.Steps = append(plan.Steps, PlanStep{
			Type:        "execute",
			Description: "Convert URL to markdown",
			Command:     fmt.Sprintf("markitdown %s", url),
			Automatic:   true,
		})
	} else if hasUv {

		plan.Steps = append(plan.Steps, PlanStep{
			Type:        "execute",
			Description: "Run markitdown via uv",
			Command:     fmt.Sprintf("uv tool run markitdown %s", url),
			Automatic:   true,
		})
	} else {

		plan.Steps = append(plan.Steps, buildToolInstallationPlan("uv")...)
		plan.Steps = append(plan.Steps, PlanStep{
			Type:        "execute",
			Description: "Run markitdown via uvx",
			Command:     fmt.Sprintf("uvx markitdown %s", url),
			Automatic:   true,
		})
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
		plan.PrintPlan()
		return
	}

	plan.PrintPlan()
	if !plan.ConfirmPlan() {
		fmt.Println("Cancelled.")
		os.Exit(0)
	}
	fmt.Println()

	if !hasMarkitdown {
		if hasUvx {
			cmd := exec.Command("uvx", "markitdown", url)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Println("\nWarning: uvx markitdown failed")
				fmt.Println("Falling back to pure.md service...")
				usePureMd(url)
			}
			return
		} else if hasMise {
			fmt.Println("Installing markitdown via mise...")
			installToolViaMise("python:markitdown", false, false)
			fmt.Println()
			hasMarkitdown = isToolAvailable("markitdown")
		} else if hasUv {
			cmd := exec.Command("uv", "tool", "run", "markitdown", url)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Println("\nWarning: uv tool run markitdown failed")
				fmt.Println("Falling back to pure.md service...")
				usePureMd(url)
			}
			return
		} else {
			fmt.Println("Error: Neither uvx nor mise is available for installing markitdown")
			fmt.Println()
			fmt.Println("Please install uv (recommended):")
			fmt.Println("  curl -LsSf https://astral.sh/uv/install.sh | sh")
			fmt.Println()
			fmt.Println("Falling back to pure.md service...")
			usePureMd(url)
			return
		}

		if !hasMarkitdown {
			fmt.Println("markitdown not available, using pure.md service...")
			usePureMd(url)
			return
		}
	}

	cmd := exec.Command("markitdown", url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("\nError: Command failed: markitdown %s\n", url)
		fmt.Println("Trying pure.md service as fallback...")
		usePureMd(url)
	}
}

// handleExcalifontCommand handles "want excalifont" - downloads and installs Excalifont
func handleExcalifontCommand(args []string, dryRun bool, planJson bool) {
	if len(args) > 0 {
		fmt.Println("Error: excalifont command does not take arguments")
		fmt.Println("Usage: want excalifont")
		os.Exit(1)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	tmpDir := filepath.Join(homeDir, ".cache", "want")
	woff2Path := filepath.Join(tmpDir, "Excalifont Regular.woff2")
	ttfPath := filepath.Join(tmpDir, "Excalifont Regular.ttf")

	pythonScript := `from fontTools.ttLib import TTFont
font = TTFont('%s')
font.flavor = None
font.save('%s')`
	formattedPythonScript := fmt.Sprintf(pythonScript, woff2Path, ttfPath)

	plan := FulfillmentPlan{
		Requirement: "excalifont",
		Steps:       buildToolInstallationPlan("uv"),
	}

	plan.Steps = append(plan.Steps, PlanStep{
		Type:        "download",
		Description: "Download Excalifont Regular (woff2) from excalidraw.com",
		Command:     fmt.Sprintf("curl -L -o '%s' https://excalidraw.com/Excalifont-Regular.woff2", woff2Path),
		Automatic:   true,
	})

	plan.Steps = append(plan.Steps, PlanStep{
		Type:        "execute",
		Description: "Convert woff2 to ttf using fontTools",
		Command:     fmt.Sprintf("uv run --with fonttools --with brotli python3 -c \"%s\"", formattedPythonScript),
		Automatic:   true,
	})

	if runtime.GOOS == "darwin" {
		plan.Steps = append(plan.Steps, PlanStep{
			Type:        "execute",
			Description: "Open the font file (macOS Font Book will handle installation)",
			Command:     fmt.Sprintf("open '%s'", ttfPath),
			Automatic:   true,
		})
	} else {
		plan.Steps = append(plan.Steps, PlanStep{
			Type:        "configure",
			Description: "Font saved and ready to install",
			Command:     fmt.Sprintf("Font saved to: %s", ttfPath),
			Automatic:   false,
		})
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
		plan.PrintPlan()
		return
	}

	plan.PrintPlan()
	if !plan.ConfirmPlan() {
		fmt.Println("Cancelled.")
		os.Exit(0)
	}
	fmt.Println()

	if !isToolAvailable("uv") {
		fmt.Println("Error: uv is not installed")
		fmt.Println()
		fmt.Println("Please install uv first:")
		fmt.Println("  curl -LsSf https://astral.sh/uv/install.sh | sh")
		fmt.Println()
		fmt.Println("Then rerun:")
		fmt.Println("  want excalifont")
		os.Exit(1)
	}

	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		fmt.Printf("Error: Failed to create cache directory %s: %v\n", tmpDir, err)
		os.Exit(1)
	}

	fmt.Println("Downloading Excalifont Regular...")
	cmd := exec.Command("curl", "-L", "-o", woff2Path, "https://excalidraw.com/Excalifont-Regular.woff2")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: Failed to download Excalifont: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	fmt.Println("Converting woff2 to ttf...")
	cmd = exec.Command("uv", "run", "--with", "fonttools", "--with", "brotli", "python3", "-c", formattedPythonScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: Failed to convert font: %v\n", err)
		fmt.Println("\nTroubleshooting:")
		fmt.Println("  • Make sure uv is installed and in PATH")
		fmt.Println("  • The conversion uses fonttools and brotli Python packages")
		os.Exit(1)
	}
	fmt.Println()

	if runtime.GOOS == "darwin" {
		fmt.Println("Opening font in Font Book...")
		cmd = exec.Command("open", ttfPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("\nWarning: Failed to open font: %v\n", err)
			fmt.Printf("\nThe font has been converted and saved to: %s\n", ttfPath)
			fmt.Println("You can manually double-click it to install.")
			return
		}

		fmt.Println()
		fmt.Printf("✓ Excalifont Regular has been opened in Font Book\n")
		fmt.Println()
		fmt.Println("Summary:")
		fmt.Printf("  ✓ Downloaded from excalidraw.com\n")
		fmt.Printf("  ✓ Converted woff2 to ttf\n")
		fmt.Printf("  ✓ Font file saved to: %s\n", ttfPath)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  • Font Book should now be open")
		fmt.Println("  • Click 'Install Font' to add it to your system")
	} else {
		fmt.Println()
		fmt.Printf("✓ Excalifont Regular has been converted successfully\n")
		fmt.Println()
		fmt.Println("Summary:")
		fmt.Printf("  ✓ Downloaded from excalidraw.com\n")
		fmt.Printf("  ✓ Converted woff2 to ttf\n")
		fmt.Printf("  ✓ Font file saved to: %s\n", ttfPath)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  • Double-click the font file to install it")
		fmt.Printf("  • Or copy it to your system fonts directory\n")
	}
}

// handleGitHubAsset downloads a binary from a GitHub release
func handleGitHubAsset(url string, dryRun bool, planJson bool) {
	// Parse the URL to extract owner, repo, and tag
	var owner, repo, tag, projectName string

	if strings.Contains(url, "/releases/download/") {

		re := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/releases/download/([^/]+)/([^/]+)`)
		matches := re.FindStringSubmatch(url)
		if matches == nil || len(matches) < 5 {
			fmt.Printf("Error: Invalid GitHub release download URL: %s\n", url)
			fmt.Println("\nExpected format:")
			fmt.Println("  https://github.com/owner/repo/releases/download/tag/asset-name")
			os.Exit(1)
		}
		owner = matches[1]
		repo = matches[2]
		tag = matches[3]
		assetName := matches[4]

		parts := strings.Split(assetName, "-")
		if len(parts) > 0 {
			projectName = parts[0]
		}
	} else if strings.Contains(url, "/releases/tag/") {

		re := regexp.MustCompile(`github\.com/([^/]+)/([^/]+)/releases/tag/([^/]+)`)
		matches := re.FindStringSubmatch(url)
		if matches == nil || len(matches) < 4 {
			fmt.Printf("Error: Invalid GitHub release tag URL: %s\n", url)
			fmt.Println("\nExpected format:")
			fmt.Println("  https://github.com/owner/repo/releases/tag/tag-name")
			os.Exit(1)
		}
		owner = matches[1]
		repo = matches[2]
		tag = matches[3]

		parts := strings.Split(tag, "--")
		if len(parts) > 0 {
			projectName = parts[0]
		}
	} else {
		fmt.Printf("Error: URL doesn't contain /releases/download/ or /releases/tag/\n")
		fmt.Println("\nSupported formats:")
		fmt.Println("  https://github.com/owner/repo/releases/download/tag/asset-name")
		fmt.Println("  https://github.com/owner/repo/releases/tag/tag-name")
		os.Exit(1)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	destDir := filepath.Join(homeDir, ".local", "bin")
	destFile := projectName
	if destFile == "" {
		destFile = tag
	}
	destPath := filepath.Join(destDir, destFile)

	platform := ghrelease.GetCurrentPlatform()
	plan := FulfillmentPlan{
		Requirement: url,
		Steps:       []PlanStep{},
	}

	plan.Steps = append(plan.Steps, PlanStep{
		Type:        "download",
		Description: fmt.Sprintf("Fetch release information from GitHub (platform: %s-%s)", platform.OS, platform.Arch),
		Command:     fmt.Sprintf("GET https://api.github.com/repos/%s/%s/releases/tags/%s", owner, repo, tag),
		Automatic:   true,
	})

	plan.Steps = append(plan.Steps, PlanStep{
		Type:        "download",
		Description: fmt.Sprintf("Download asset to %s", destPath),
		Command:     fmt.Sprintf("download asset to %s", destPath),
		Automatic:   true,
	})

	plan.Steps = append(plan.Steps, PlanStep{
		Type:        "configure",
		Description: "Make executable",
		Command:     fmt.Sprintf("chmod +x %s", destPath),
		Automatic:   true,
	})

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
		fmt.Printf("Downloading GitHub release asset from: %s\n", url)
		fmt.Println()
		fmt.Printf("Owner: %s\n", owner)
		fmt.Printf("Repo: %s\n", repo)
		fmt.Printf("Tag: %s\n", tag)
		if projectName != "" {
			fmt.Printf("Project: %s\n", projectName)
		}
		fmt.Println()
		plan.PrintPlan()
		return
	}

	fmt.Printf("Downloading GitHub release asset from: %s\n", url)
	fmt.Println()
	fmt.Printf("Owner: %s\n", owner)
	fmt.Printf("Repo: %s\n", repo)
	fmt.Printf("Tag: %s\n", tag)
	if projectName != "" {
		fmt.Printf("Project: %s\n", projectName)
	}
	fmt.Println()

	plan.PrintPlan()
	if !plan.ConfirmPlan() {
		fmt.Println("Cancelled.")
		os.Exit(0)
	}
	fmt.Println()

	fmt.Println("Fetching release information...")
	err = ghrelease.DownloadReleaseAsset(owner, repo, tag, projectName, destPath)
	if err != nil {
		fmt.Printf("\nError: Failed to download release asset: %v\n", err)
		fmt.Println("\nTroubleshooting:")
		fmt.Println("  • Check that the release exists")
		fmt.Println("  • Check that there's an asset for your platform")
		fmt.Println("  • For private repositories, set GITHUB_TOKEN environment variable")
		fmt.Printf("  • Your platform: %s-%s\n", platform.OS, platform.Arch)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("✓ Downloaded to: %s\n", destPath)
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
		fmt.Printf("✓ Binary is available in your PATH as: %s\n", destFile)
	}
}

// handleMono handles installation of tools from the neongreen/mono repository
func handleMono() {

	fs := flag.NewFlagSet("mono", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be done without doing it")
	planJson := fs.Bool("plan-json", false, "Output the fulfillment plan as JSON")
	listFlag := fs.Bool("list", false, "List all releases and open PRs")

	fs.Parse(os.Args[2:])

	if fs.NArg() == 0 {
		fmt.Println("Error: no project specified")
		fmt.Println("Usage: want mono [--dry-run] [--plan-json] <project> [--list]")
		fmt.Println("       want mono [--dry-run] [--plan-json] <project@version>")
		fmt.Println("\nExamples:")
		fmt.Println("  want mono printpdf --list              # List all releases and open PRs for printpdf")
		fmt.Println("  want mono printpdf@main.1              # Install printpdf version main.1")
		fmt.Println("  want mono printpdf@main                # Build printpdf from latest commit on main")
		fmt.Println("  want mono dissect@feature-branch       # Build dissect from a specific branch")
		fmt.Println("  want mono want@abc1234                 # Build want from a specific commit")
		fmt.Println("  want mono --dry-run dissect@pr-42      # Preview building from PR #42")
		fmt.Println("  want mono --plan-json printpdf@main.1  # Show installation plan as JSON")
		os.Exit(1)
	}

	arg := fs.Arg(0)

	if *listFlag {
		listMonoReleases(arg)
		return
	}

	parts := strings.Split(arg, "@")
	if len(parts) == 1 {

		listMonoReleases(parts[0])
		return
	}

	if len(parts) != 2 {
		fmt.Printf("Error: Invalid format '%s'\n", arg)
		fmt.Println("Expected: <project>@<version> or <project> --list")
		fmt.Println("\nExamples:")
		fmt.Println("  want mono printpdf@main.1")
		fmt.Println("  want mono dissect@pr-42")
		fmt.Println("  want mono printpdf --list")
		os.Exit(1)
	}

	project := parts[0]
	version := parts[1]

	installMonoRelease(project, version, *dryRun, *planJson)
}

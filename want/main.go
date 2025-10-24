package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/neongreen/mono/lib/ghrelease"
)

const version = "0.1.0-mvp"

// PlanStep represents a single step in a fulfillment plan
type PlanStep struct {
	Type        string `json:"type"`        // "install", "download", "execute", "configure"
	Description string `json:"description"` // Human-readable description
	Command     string `json:"command"`     // Command to execute
	Automatic   bool   `json:"automatic"`   // Whether this step is automatic
	Safe        bool   `json:"safe"`        // Whether this step is read-only/safe
}

// FulfillmentPlan represents a complete plan to fulfill a requirement
type FulfillmentPlan struct {
	Requirement string     `json:"requirement"` // What the user requested
	Steps       []PlanStep `json:"steps"`       // Steps to fulfill the requirement
}

// buildMiseInstallationSteps returns the steps needed to install mise
func buildMiseInstallationSteps() []PlanStep {
	steps := []PlanStep{
		{
			Type:        "install",
			Description: "Download and run mise installation script",
			Command:     "curl https://mise.run | sh",
			Automatic:   true,
		},
	}

	// Check if mise activation is needed
	if !isMiseActivated() {
		configFile := getShellConfigFile()
		shellName := getShellName()
		steps = append(steps, PlanStep{
			Type:        "configure",
			Description: fmt.Sprintf("Add mise activation to %s", configFile),
			Command:     fmt.Sprintf("echo 'eval \"$(mise activate %s)\"' >> %s", shellName, configFile),
			Automatic:   true,
		})
		steps = append(steps, PlanStep{
			Type:        "configure",
			Description: "Activate mise in current shell (or restart shell)",
			Command:     fmt.Sprintf("eval \"$(mise activate %s)\"", shellName),
			Automatic:   false,
		})
	}

	return steps
}

// PrintPlan displays the plan in human-readable format
func (p *FulfillmentPlan) PrintPlan() {
	fmt.Println("Fulfillment plan:")
	fmt.Println()
	for i, step := range p.Steps {
		// Only show label for manual steps
		if !step.Automatic {
			fmt.Printf("Step %d (MANUAL): %s\n", i+1, step.Description)
		} else {
			fmt.Printf("Step %d: %s\n", i+1, step.Description)
		}
		fmt.Printf("  $ %s\n", step.Command)
		fmt.Println()
	}
}

// ToJSON returns the plan as a JSON string
func (p *FulfillmentPlan) ToJSON() (string, error) {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// HasOnlySafeSteps returns true if all steps in the plan are marked as safe
func (p *FulfillmentPlan) HasOnlySafeSteps() bool {
	for _, step := range p.Steps {
		if !step.Safe {
			return false
		}
	}
	return len(p.Steps) > 0
}

// ConfirmPlan asks the user to confirm the plan
func (p *FulfillmentPlan) ConfirmPlan() bool {
	fmt.Print("Proceed with this plan? [Y/n]: ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "" || response == "y" || response == "yes"
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	if err := ensureConfigDirectory(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "mono":
		handleMono()
	case "list":
		handleList()
	case "check":
		handleCheck()
	case "forget":
		handleForget()
	case "version", "--version", "-v":
		fmt.Printf("want version %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		handleWant(os.Args[1:])
	}
}

func printUsage() {
	fmt.Println(`want - Interactive task fulfillment tool for macOS

Usage:
  want [--dry-run] [--plan-json] <requirement>  Get a tool, repository, or resource
  want mono <project> [--list]                  Install tools from neongreen/mono repo
  want json <command>                           Convert command output to JSON
  want md <url>                                 Convert URL to markdown
  want list                                     Show what you have
  want check                                    Check status of requirements
  want forget <name>                            Remove from tracking (doesn't uninstall)
  want version                                  Show version
  want help                                     Show this help

Flags:
  --dry-run                       Show what would be done without actually doing it
  --plan-json                     Output the fulfillment plan as JSON

Examples:
  want jujutsu                    # Install a tool (asks for confirmation)
  want mise                       # Install mise itself
  want --dry-run jujutsu          # Preview installation without confirmation
  want --plan-json jujutsu        # Show installation plan as JSON
  want json ps                    # Get running processes as JSON (uses jc)
  want md https://example.com     # Convert webpage to markdown
  want mono printpdf --list       # List all releases of printpdf from mono
  want mono printpdf@main.1       # Install printpdf version main.1 from mono
  want https://github.com/org/repo/releases/tag/v1.0.0  # Download GitHub release
  want github.com/user/repo       # Clone a repository (not yet implemented)

Configuration is stored at ~/.config/want/

For more information, see README.md`)
}

func ensureConfigDirectory() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "want")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create want config directory %s: %w", configDir, err)
	}

	return nil
}

func handleWant(args []string) {
	// Parse flags
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

	// Check for compound commands (e.g., "want json ps" or "want md url")
	if handler, ok := getCompoundHandler(requirement); ok {
		handler(remainingArgs, *dryRun, *planJson)
		return
	}

	// Check if requirement looks like a GitHub release download URL
	if strings.Contains(requirement, "github.com") && (strings.Contains(requirement, "/releases/download/") || strings.Contains(requirement, "/releases/tag/")) {
		handleGitHubAsset(requirement, *dryRun, *planJson)
		return
	}

	// Check if requirement looks like a git repository
	if strings.Contains(requirement, "/") && (strings.Contains(requirement, "github.com") || strings.Contains(requirement, ".git")) {
		fmt.Printf("Error: Git repository cloning is not yet implemented\n")
		fmt.Printf("Requirement: %s\n", requirement)
		os.Exit(1)
	}

	// Try to install as a tool via mise
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

// isToolAvailable checks if a tool is already available in PATH
func isToolAvailable(tool string) bool {
	_, err := exec.LookPath(tool)
	return err == nil
}

// isMiseAvailable checks if mise is installed
func isMiseAvailable() bool {
	_, err := exec.LookPath("mise")
	return err == nil
}

// getShellConfigFile returns the shell configuration file path based on the current shell
func getShellConfigFile() string {
	shell := os.Getenv("SHELL")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Determine config file based on shell
	if strings.Contains(shell, "zsh") {
		zshrc := filepath.Join(homeDir, ".zshrc")
		if _, err := os.Stat(zshrc); err == nil {
			return zshrc
		}
		zprofile := filepath.Join(homeDir, ".zprofile")
		return zprofile
	} else if strings.Contains(shell, "bash") {
		bashrc := filepath.Join(homeDir, ".bashrc")
		if _, err := os.Stat(bashrc); err == nil {
			return bashrc
		}
		bashProfile := filepath.Join(homeDir, ".bash_profile")
		if _, err := os.Stat(bashProfile); err == nil {
			return bashProfile
		}
		profile := filepath.Join(homeDir, ".profile")
		return profile
	}

	// Default to .profile for other shells
	return filepath.Join(homeDir, ".profile")
}

// isMiseActivated checks if mise activation is present in the shell config
func isMiseActivated() bool {
	configFile := getShellConfigFile()
	if configFile == "" {
		return false
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return false
	}

	content := string(data)
	// Check for mise activation command
	return strings.Contains(content, "mise activate") || strings.Contains(content, "mise.jdx.dev")
}

// getShellName returns a human-readable name for the current shell
func getShellName() string {
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		return "zsh"
	} else if strings.Contains(shell, "bash") {
		return "bash"
	}
	return filepath.Base(shell)
}

// installMise installs mise using the official installation script
func installMise(dryRun bool, planJson bool) error {
	plan := FulfillmentPlan{
		Requirement: "mise",
		Steps:       buildMiseInstallationSteps(),
	}

	if planJson {
		jsonStr, err := plan.ToJSON()
		if err != nil {
			return fmt.Errorf("failed to generate JSON: %w", err)
		}
		fmt.Println(jsonStr)
		return nil
	}

	if dryRun {
		plan.PrintPlan()
		return nil
	}

	plan.PrintPlan()
	if !plan.ConfirmPlan() {
		fmt.Println("Cancelled.")
		os.Exit(0)
	}
	fmt.Println()

	return performMiseInstallation()
}

func performMiseInstallation() error {
	fmt.Println("Installing mise...")
	fmt.Println()

	cmd := exec.Command("sh", "-c", "curl https://mise.run | sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println()
		fmt.Println("Error: Failed to download and install mise")
		fmt.Println()
		fmt.Println("You can try installing manually:")
		fmt.Println("  curl https://mise.run | sh")
		fmt.Println()
		fmt.Println("Or visit: https://mise.jdx.dev/getting-started.html")
		return fmt.Errorf("failed to install mise: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ mise installed successfully")
	fmt.Println()

	if !isMiseActivated() {
		configFile := getShellConfigFile()
		shellName := getShellName()

		fmt.Printf("Adding mise activation to %s...\n", configFile)

		activationLine := fmt.Sprintf("\n# Added by want - enables mise\neval \"$(mise activate %s)\"\n", shellName)

		file, err := os.OpenFile(configFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			fmt.Printf("⚠ Could not automatically add mise activation to %s: %v\n", configFile, err)
			fmt.Println()
			fmt.Println("Please add this line manually:")
			fmt.Printf("  eval \"$(mise activate %s)\"\n", shellName)
			fmt.Println()
			fmt.Printf("You can add it with:\n")
			fmt.Printf("  echo 'eval \"$(mise activate %s)\"' >> %s\n", shellName, configFile)
			return nil
		}
		defer file.Close()

		if _, err = file.WriteString(activationLine); err != nil {
			fmt.Printf("⚠ Could not write to %s: %v\n", configFile, err)
			fmt.Println()
			fmt.Println("Please add this line manually:")
			fmt.Printf("  eval \"$(mise activate %s)\"\n", shellName)
			return nil
		}

		fmt.Println("✓ mise activation added to your shell configuration")
		fmt.Println()
		fmt.Println("Summary:")
		fmt.Println("  ✓ mise installed")
		fmt.Println("  ✓ Shell configuration updated")
		fmt.Println()
		fmt.Println()
		fmt.Println("Manual step required:")
		fmt.Println("  To activate mise in your current shell, run:")
		fmt.Printf("    eval \"$(mise activate %s)\"\n", shellName)
		fmt.Println()
		fmt.Println("  Or restart your shell to automatically load mise.")
	} else {
		fmt.Println("✓ mise activation already configured in your shell")
		fmt.Println()
		fmt.Println("Summary:")
		fmt.Println("  ✓ mise installed")
		fmt.Println("  ✓ Shell configuration already present")
		fmt.Println()
		fmt.Println("No manual steps required.")
	}

	return nil
}

// CompoundHandler is a function that handles compound commands
type CompoundHandler func(args []string, dryRun bool, planJson bool)

// getCompoundHandler returns a handler for compound commands if one exists
func getCompoundHandler(command string) (CompoundHandler, bool) {
	handlers := map[string]CompoundHandler{
		"json": handleJsonCommand,
		"md":   handleMarkdownCommand,
	}
	handler, ok := handlers[command]
	return handler, ok
}

// isCommandSafe checks if a command is read-only/safe (doesn't modify system state)
func isCommandSafe(command string) bool {
	// List of safe/read-only commands that can be run without confirmation
	// NOTE: This is a conservative list. Commands that CAN modify state with certain flags
	// (like 'find -delete', 'date -s') are excluded for safety.
	// Future improvement: Implement argument inspection to allow safe uses of these commands.
	safeCommands := []string{
		"ps", "top", "uptime", "whoami", "id", "hostname", "uname",
		"df", "du", "free", "vmstat", "iostat", "netstat", "ss",
		"ls", "pwd", "cat", "head", "tail", "wc", "grep",
		"cal", "env", "printenv", "which", "whereis",
		"git status", "git log", "git diff", "git show",
	}

	// Check if the command starts with any safe command
	cmdLower := strings.ToLower(strings.TrimSpace(command))
	for _, safe := range safeCommands {
		if strings.HasPrefix(cmdLower, safe) {
			return true
		}
	}

	return false
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

	// Check if we need to install jc
	needsJcInstall := !isToolAvailable("jc")

	// Build the plan
	plan := FulfillmentPlan{
		Requirement: fmt.Sprintf("json %s", commandStr),
		Steps:       []PlanStep{},
	}

	// Check if jc is available
	if needsJcInstall {
		// Check if mise is available - if not, add mise installation steps first
		if !isMiseAvailable() {
			plan.Steps = append(plan.Steps, buildMiseInstallationSteps()...)
		}

		plan.Steps = append(plan.Steps, PlanStep{
			Type:        "install",
			Description: "Install jc via mise",
			Command:     "mise use -g jc",
			Automatic:   true,
		})
	}

	plan.Steps = append(plan.Steps, PlanStep{
		Type:        "execute",
		Description: "Execute command with jc",
		Command:     fmt.Sprintf("jc %s", commandStr),
		Automatic:   true,
		Safe:        isCommandSafe(commandStr),
	})

	// Handle plan output modes
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

	// Execute the plan
	plan.PrintPlan()

	// Skip confirmation if all steps are safe
	if !plan.HasOnlySafeSteps() {
		if !plan.ConfirmPlan() {
			fmt.Println("Cancelled.")
			os.Exit(0)
		}
	}
	fmt.Println()

	// Install jc if needed
	if needsJcInstall {
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
		cmd := exec.Command("mise", "use", "-g", "jc")
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

	// Execute jc with the provided command
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

	// Build the plan based on available tools
	plan := FulfillmentPlan{
		Requirement: fmt.Sprintf("md %s", url),
		Steps:       []PlanStep{},
	}

	hasMarkitdown := isToolAvailable("markitdown")
	hasUvx := isToolAvailable("uvx")
	hasMise := isMiseAvailable()
	hasUv := isToolAvailable("uv")

	// Determine the best method and build the plan
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
		// mise is available, use it to install markitdown
		plan.Steps = append(plan.Steps, PlanStep{
			Type:        "install",
			Description: "Install markitdown via mise",
			Command:     "mise use -g python:markitdown",
			Automatic:   true,
		})
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
		plan.Steps = append(plan.Steps, PlanStep{
			Type:        "install",
			Description: "Install uv (recommended)",
			Command:     "curl -LsSf https://astral.sh/uv/install.sh | sh",
			Automatic:   false,
		})
		plan.Steps = append(plan.Steps, PlanStep{
			Type:        "execute",
			Description: "Run markitdown via uvx",
			Command:     fmt.Sprintf("uvx markitdown %s", url),
			Automatic:   true,
		})
	}

	// Handle plan output modes
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

	// Execute the plan
	plan.PrintPlan()
	if !plan.ConfirmPlan() {
		fmt.Println("Cancelled.")
		os.Exit(0)
	}
	fmt.Println()

	// Execute based on available tools
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

	// Execute markitdown
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

// usePureMd uses the pure.md service to convert URL to markdown
func usePureMd(url string) {
	if !isToolAvailable("curl") {
		fmt.Println("Error: curl is required but not available")
		os.Exit(1)
	}

	pureMdURL := "https://pure.md/" + url
	fmt.Printf("Fetching from pure.md: %s\n\n", pureMdURL)

	cmd := exec.Command("curl", "-s", pureMdURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("\nError: Failed to fetch from pure.md\n")
		os.Exit(1)
	}
}

// installToolViaMise installs a tool using mise
func installToolViaMise(tool string, dryRun bool, planJson bool) {
	// Special case: handle mise installation request
	if tool == "mise" {
		if isMiseAvailable() {
			fmt.Printf("✓ mise is already available\n")
			fmt.Println()

			// Show where it's from
			cmd := exec.Command("which", "mise")
			output, err := cmd.Output()
			if err == nil {
				fmt.Printf("  Location: %s", string(output))
			}

			// Check activation status
			if !isMiseActivated() {
				configFile := getShellConfigFile()
				shellName := getShellName()
				fmt.Println()
				fmt.Printf("⚠ mise activation not found in %s\n", configFile)
				fmt.Println()
				fmt.Println("To make mise-installed tools available in your PATH, add this line:")
				fmt.Printf("  eval \"$(mise activate %s)\"\n", shellName)
				fmt.Println()
				fmt.Printf("You can add it automatically with:\n")
				fmt.Printf("  echo 'eval \"$(mise activate %s)\"' >> %s\n", shellName, configFile)
			} else {
				fmt.Println()
				fmt.Println("✓ mise activation is configured in your shell")
			}
			return
		}

		// Install mise using the official installation script
		err := installMise(dryRun, planJson)
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check if tool is already available
	if isToolAvailable(tool) {
		fmt.Printf("✓ %s is already available\n", tool)
		fmt.Println()

		// Show where it's from
		cmd := exec.Command("which", tool)
		output, err := cmd.Output()
		if err == nil {
			fmt.Printf("  Location: %s", string(output))
		}

		fmt.Println("\nNote: The tool is already available, so no installation is needed.")
		fmt.Println("If you want to install it via mise specifically, you can run:")
		fmt.Printf("  mise use -g %s\n", tool)
		return
	}

	// Build the plan
	plan := FulfillmentPlan{
		Requirement: tool,
		Steps:       []PlanStep{},
	}

	// Check if mise is available and add to plan if needed
	if !isMiseAvailable() {
		plan.Steps = append(plan.Steps, buildMiseInstallationSteps()...)
	}

	plan.Steps = append(plan.Steps, PlanStep{
		Type:        "install",
		Description: fmt.Sprintf("Install %s via mise", tool),
		Command:     fmt.Sprintf("mise use -g %s", tool),
		Automatic:   true,
	})

	// Handle plan output modes
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

	// Execute the plan
	plan.PrintPlan()
	if !plan.ConfirmPlan() {
		fmt.Println("Cancelled.")
		os.Exit(0)
	}
	fmt.Println()

	// Check if mise is available before executing
	if !isMiseAvailable() {
		fmt.Println()
		fmt.Println("Error: mise is not installed")
		fmt.Println("Please install mise first with:")
		fmt.Println("  want mise")
		os.Exit(1)
	}

	// Execute installation
	fmt.Printf("Installing %s via mise...\n", tool)
	cmd := exec.Command("mise", "use", "-g", tool)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("\nError: Failed to install %s via mise\n", tool)
		fmt.Printf("Command failed: mise use -g %s\n", tool)
		fmt.Println("\nPossible reasons:")
		fmt.Printf("  • The tool '%s' might not be available in mise's registry\n", tool)
		fmt.Println("  • There might be a network issue")
		fmt.Println("  • mise might need to be updated")
		fmt.Println("\nYou can try manually:")
		fmt.Printf("  mise use -g %s\n", tool)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("✓ %s installed successfully\n", tool)

	// Verify installation
	if isToolAvailable(tool) {
		fmt.Printf("✓ %s is now available in your PATH\n", tool)
	} else {
		fmt.Printf("⚠ %s was installed but may not be in your PATH yet\n", tool)
		shellName := getShellName()
		fmt.Println("You may need to restart your shell or run:")
		fmt.Printf("  eval \"$(mise activate %s)\"\n", shellName)
	}
}

// handleGitHubAsset downloads a binary from a GitHub release
func handleGitHubAsset(url string, dryRun bool, planJson bool) {
	// Parse the URL to extract owner, repo, and tag
	var owner, repo, tag, projectName string

	// Handle two URL formats:
	// 1. https://github.com/owner/repo/releases/download/tag/asset-name
	// 2. https://github.com/owner/repo/releases/tag/tag-name

	if strings.Contains(url, "/releases/download/") {
		// Direct download URL
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

		// Try to extract project name from asset name (e.g., "want-main.3-darwin-arm64" -> "want")
		parts := strings.Split(assetName, "-")
		if len(parts) > 0 {
			projectName = parts[0]
		}
	} else if strings.Contains(url, "/releases/tag/") {
		// Release tag URL - will detect platform automatically
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

		// Try to extract project name from tag (e.g., "want--main.3" -> "want")
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

	// Determine destination path
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

	// Build the plan
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

	// Handle plan output modes
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

	// Execute the plan
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

	// Check if destDir is in PATH
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
	if len(os.Args) < 3 {
		fmt.Println("Error: no project specified")
		fmt.Println("Usage: want mono <project> [--list]")
		fmt.Println("       want mono <project@version>")
		fmt.Println("\nExamples:")
		fmt.Println("  want mono printpdf --list     # List all releases for printpdf")
		fmt.Println("  want mono printpdf@main.1     # Install printpdf version main.1")
		os.Exit(1)
	}

	arg := os.Args[2]

	// Check for --list flag
	if len(os.Args) == 4 && os.Args[3] == "--list" {
		listMonoReleases(arg)
		return
	}

	// Parse project@version syntax
	parts := strings.Split(arg, "@")
	if len(parts) == 1 {
		// No version specified - list releases
		listMonoReleases(parts[0])
		return
	}

	if len(parts) != 2 {
		fmt.Printf("Error: Invalid format '%s'\n", arg)
		fmt.Println("Expected: <project>@<version> or <project> --list")
		fmt.Println("\nExamples:")
		fmt.Println("  want mono printpdf@main.1")
		fmt.Println("  want mono printpdf --list")
		os.Exit(1)
	}

	project := parts[0]
	version := parts[1]

	installMonoRelease(project, version)
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

	if len(projectReleases) == 0 {
		fmt.Printf("No releases found for %s\n", project)
		fmt.Println("\nAvailable projects in mono:")
		fmt.Println("  printpdf, dissect, want, prrun, markdown-format, ingest, conf, claude-trace")
		os.Exit(1)
	}

	fmt.Printf("Available releases for %s:\n", project)
	fmt.Println()
	for _, release := range projectReleases {
		// Extract version from tag (e.g., "printpdf--main.1" -> "main.1")
		version := strings.TrimPrefix(release.TagName, prefix)

		status := ""
		if release.Prerelease {
			status = " (prerelease)"
		}

		fmt.Printf("  %s%s\n", version, status)
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
}

// installMonoRelease installs a specific version of a project from neongreen/mono
func installMonoRelease(project, version string) {
	tag := fmt.Sprintf("%s--%s", project, version)

	fmt.Printf("Installing %s version %s from neongreen/mono...\n", project, version)
	fmt.Println()

	// Determine destination path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	destDir := filepath.Join(homeDir, ".local", "bin")
	destPath := filepath.Join(destDir, project)

	// Check if already installed
	if _, err := os.Stat(destPath); err == nil {
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
	}

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

	// Check if destDir is in PATH
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

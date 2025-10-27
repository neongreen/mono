package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/neongreen/mono/lib/ghclient"
	"github.com/neongreen/mono/lib/ghrelease"
)

const version = "0.1.0-mvp"

// goVersion is the Go version to use when building projects via mise
const goVersion = "1.24.7"

// Tool represents an installable tool or dependency
type Tool struct {
	Name        string   // Name of the tool
	CheckCmd    string   // Command to check if tool is available (usually just the tool name)
	InstallStep PlanStep // Step to install the tool
	// For tools installed via mise, this is the mise package name
	MisePackage string
}

// ToolRegistry is a catalogue of installable tools and their installation methods
var ToolRegistry = map[string]Tool{
	"uv": {
		Name:     "uv",
		CheckCmd: "uv",
		InstallStep: PlanStep{
			Type:        "install",
			Description: "Install uv (Python package manager)",
			Command:     "curl -LsSf https://astral.sh/uv/install.sh | sh",
			Automatic:   false,
		},
	},
	"uvx": {
		Name:     "uvx",
		CheckCmd: "uvx",
		InstallStep: PlanStep{
			Type:        "install",
			Description: "Install uv (includes uvx)",
			Command:     "curl -LsSf https://astral.sh/uv/install.sh | sh",
			Automatic:   false,
		},
	},
	"mise": {
		Name:     "mise",
		CheckCmd: "mise",
		InstallStep: PlanStep{
			Type:        "install",
			Description: "Install mise",
			Command:     "curl https://mise.run | sh",
			Automatic:   true,
		},
	},
	"jc": {
		Name:        "jc",
		CheckCmd:    "jc",
		MisePackage: "jc",
		InstallStep: PlanStep{
			Type:        "install",
			Description: "Install jc via mise",
			Command:     "mise use -g jc",
			Automatic:   true,
		},
	},
	"markitdown": {
		Name:        "markitdown",
		CheckCmd:    "markitdown",
		MisePackage: "python:markitdown",
		InstallStep: PlanStep{
			Type:        "install",
			Description: "Install markitdown via mise",
			Command:     "mise use -g python:markitdown",
			Automatic:   true,
		},
	},
}

// ensureToolAvailable checks if a tool is available and returns installation steps if not
func ensureToolAvailable(toolName string) (available bool, installSteps []PlanStep) {
	if isToolAvailable(toolName) {
		return true, nil
	}

	tool, exists := ToolRegistry[toolName]
	if !exists {
		// Tool not in registry, return empty steps
		return false, nil
	}

	return false, []PlanStep{tool.InstallStep}
}

// buildToolInstallationPlan creates a complete installation plan for a tool
// It handles dependencies (like mise) automatically based on the registry
func buildToolInstallationPlan(toolName string) []PlanStep {
	var steps []PlanStep

	// Check if tool is already available
	if isToolAvailable(toolName) {
		return steps
	}

	tool, exists := ToolRegistry[toolName]
	if !exists {
		// Tool not in registry, can't generate a plan
		return steps
	}

	// If this tool is installed via mise, ensure mise is available first
	if tool.MisePackage != "" && !isMiseAvailable() {
		// Add mise installation steps
		steps = append(steps, buildMiseInstallationSteps()...)
	}

	// Add the tool's installation step
	steps = append(steps, tool.InstallStep)

	return steps
}

// buildShellConfigStep creates a step to configure shell for mise activation
func buildShellConfigStep() PlanStep {
	configFile := getShellConfigFile()
	shellName := getShellName()
	return PlanStep{
		Type:        "configure",
		Description: fmt.Sprintf("Add mise activation to %s", configFile),
		Command:     fmt.Sprintf("echo 'eval \"$(mise activate %s)\"' >> %s", shellName, configFile),
		Automatic:   true,
	}
}

// buildManualActivationStep creates a manual step for activating mise in current shell
func buildManualActivationStep() PlanStep {
	shellName := getShellName()
	return PlanStep{
		Type:        "configure",
		Description: "Activate mise in current shell (or restart shell)",
		Command:     fmt.Sprintf("eval \"$(mise activate %s)\"", shellName),
		Automatic:   false,
	}
}

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
	var steps []PlanStep

	// Get mise tool from registry
	miseTool, exists := ToolRegistry["mise"]
	if !exists {
		// This should never happen - mise must be in registry
		panic("BUG: 'mise' tool not found in ToolRegistry. Please ensure ToolRegistry includes a 'mise' entry with proper installation steps.")
	}

	steps = append(steps, miseTool.InstallStep)

	// Check if mise activation is needed
	if !isMiseActivated() {
		steps = append(steps, buildShellConfigStep())
		steps = append(steps, buildManualActivationStep())
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
  want [--dry-run] [--plan-json] <requirement>          Get a tool, repository, or resource
  want mono [--dry-run] [--plan-json] <project> [--list]  Install tools from neongreen/mono repo
  want mono [--dry-run] [--plan-json] <project@version>   Install specific version or PR
  want json <command>                                   Convert command output to JSON
  want md <url>                                         Convert URL to markdown
  want excalifont                                       Download and install Excalifont
  want list                                             Show what you have
  want check                                            Check status of requirements
  want forget <name>                                    Remove from tracking (doesn't uninstall)
  want version                                          Show version
  want help                                             Show this help

Flags:
  --dry-run                       Show what would be done without actually doing it
  --plan-json                     Output the fulfillment plan as JSON

Examples:
  want jujutsu                         # Install a tool (asks for confirmation)
  want mise                            # Install mise itself
  want --dry-run jujutsu               # Preview installation without confirmation
  want --plan-json jujutsu             # Show installation plan as JSON
  want json ps                         # Get running processes as JSON (uses jc)
  want md https://example.com          # Convert webpage to markdown
  want excalifont                      # Download and install Excalifont font
  want mono printpdf --list            # List all releases and open PRs of printpdf
  want mono printpdf@main.1            # Install printpdf version main.1 from mono
  want mono printpdf@main              # Build printpdf from latest commit on main branch
  want mono dissect@feature-branch     # Build dissect from a specific branch
  want mono want@abc1234               # Build want from a specific commit
  want mono --dry-run dissect@pr-42    # Preview building from PR #42
  want mono --plan-json dissect@pr-42  # Show build plan as JSON for PR #42
  want https://github.com/org/repo/releases/tag/v1.0.0  # Download GitHub release
  want github.com/user/repo            # Clone a repository (not yet implemented)

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
		"json":       handleJsonCommand,
		"md":         handleMarkdownCommand,
		"excalifont": handleExcalifontCommand,
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

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
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

	// Build the plan using the tool registry
	plan := FulfillmentPlan{
		Requirement: fmt.Sprintf("json %s", commandStr),
		Steps:       buildToolInstallationPlan("jc"),
	}

	// Add execution step
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

	// Determine the best method and build the plan using algorithm
	if hasMarkitdown {
		// Already installed, just execute
		plan.Steps = append(plan.Steps, PlanStep{
			Type:        "execute",
			Description: "Convert URL to markdown",
			Command:     fmt.Sprintf("markitdown %s", url),
			Automatic:   true,
		})
	} else if hasUvx {
		// Use uvx to run without installation
		plan.Steps = append(plan.Steps, PlanStep{
			Type:        "execute",
			Description: "Run markitdown via uvx",
			Command:     fmt.Sprintf("uvx markitdown %s", url),
			Automatic:   true,
		})
	} else if hasMise {
		// mise is available, install markitdown via registry
		plan.Steps = append(plan.Steps, buildToolInstallationPlan("markitdown")...)
		plan.Steps = append(plan.Steps, PlanStep{
			Type:        "execute",
			Description: "Convert URL to markdown",
			Command:     fmt.Sprintf("markitdown %s", url),
			Automatic:   true,
		})
	} else if hasUv {
		// Use uv tool run without installation
		plan.Steps = append(plan.Steps, PlanStep{
			Type:        "execute",
			Description: "Run markitdown via uv",
			Command:     fmt.Sprintf("uv tool run markitdown %s", url),
			Automatic:   true,
		})
	} else {
		// Neither uvx nor mise available - need to install uv via registry
		plan.Steps = append(plan.Steps, buildToolInstallationPlan("uv")...)
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

// handleExcalifontCommand handles "want excalifont" - downloads and installs Excalifont
func handleExcalifontCommand(args []string, dryRun bool, planJson bool) {
	if len(args) > 0 {
		fmt.Println("Error: excalifont command does not take arguments")
		fmt.Println("Usage: want excalifont")
		os.Exit(1)
	}

	// Determine destination directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	// Use a temporary directory for downloading
	tmpDir := filepath.Join(homeDir, ".cache", "want")
	woff2Path := filepath.Join(tmpDir, "Excalifont Regular.woff2")
	ttfPath := filepath.Join(tmpDir, "Excalifont Regular.ttf")

	// Python script for converting woff2 to ttf
	pythonScript := `from fontTools.ttLib import TTFont
font = TTFont('%s')
font.flavor = None
font.save('%s')`
	formattedPythonScript := fmt.Sprintf(pythonScript, woff2Path, ttfPath)

	// Build the plan using registry
	plan := FulfillmentPlan{
		Requirement: "excalifont",
		Steps:       buildToolInstallationPlan("uv"),
	}

	// Download the font
	plan.Steps = append(plan.Steps, PlanStep{
		Type:        "download",
		Description: "Download Excalifont Regular (woff2) from excalidraw.com",
		Command:     fmt.Sprintf("curl -L -o '%s' https://excalidraw.com/Excalifont-Regular.woff2", woff2Path),
		Automatic:   true,
	})

	// Convert woff2 to ttf using Python script with uv
	plan.Steps = append(plan.Steps, PlanStep{
		Type:        "execute",
		Description: "Convert woff2 to ttf using fontTools",
		Command:     fmt.Sprintf("uv run --with fonttools --with brotli python3 -c \"%s\"", formattedPythonScript),
		Automatic:   true,
	})

	// On macOS, open the font to install it
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

	// Check if uv is available before execution
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

	// Create cache directory
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		fmt.Printf("Error: Failed to create cache directory %s: %v\n", tmpDir, err)
		os.Exit(1)
	}

	// Download the font
	fmt.Println("Downloading Excalifont Regular...")
	cmd := exec.Command("curl", "-L", "-o", woff2Path, "https://excalidraw.com/Excalifont-Regular.woff2")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: Failed to download Excalifont: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	// Convert woff2 to ttf
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

	// Open the font on macOS or show instructions for other platforms
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

	// Build the plan using registry if tool exists there, otherwise manual
	plan := FulfillmentPlan{
		Requirement: tool,
		Steps:       []PlanStep{},
	}

	// Check if tool is in registry and uses mise
	if registryTool, inRegistry := ToolRegistry[tool]; inRegistry && registryTool.MisePackage != "" {
		// Use registry-based plan generation for mise-installed tools
		plan.Steps = buildToolInstallationPlan(tool)
	} else {
		// Tool not in registry or doesn't use mise, build plan manually
		if !isMiseAvailable() {
			plan.Steps = append(plan.Steps, buildMiseInstallationSteps()...)
		}

		plan.Steps = append(plan.Steps, PlanStep{
			Type:        "install",
			Description: fmt.Sprintf("Install %s via mise", tool),
			Command:     fmt.Sprintf("mise use -g %s", tool),
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
	// Parse flags
	fs := flag.NewFlagSet("mono", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be done without doing it")
	planJson := fs.Bool("plan-json", false, "Output the fulfillment plan as JSON")
	listFlag := fs.Bool("list", false, "List all releases and open PRs")

	// Parse remaining args (skip "mono" command)
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

	// Check for --list flag
	if *listFlag {
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
		fmt.Println("  want mono dissect@pr-42")
		fmt.Println("  want mono printpdf --list")
		os.Exit(1)
	}

	project := parts[0]
	version := parts[1]

	installMonoRelease(project, version, *dryRun, *planJson)
}

// PRInfo holds information about a pull request
type PRInfo struct {
	Number int
	Title  string
	Branch string
}

// listOpenPRs fetches open PRs that modify the given project
func listOpenPRs(project string) ([]PRInfo, error) {
	ctx := context.Background()
	client := ghclient.NewClient(ctx)

	// List open PRs
	prs, _, err := client.PullRequests.List(ctx, "neongreen", "mono", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list PRs: %w", err)
	}

	var projectPRs []PRInfo
	for _, pr := range prs {
		if pr.Number == nil || pr.Title == nil || pr.Head == nil || pr.Head.Ref == nil {
			continue
		}

		// Check if PR modifies the project
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

	// Also fetch open PRs
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
			// Extract version from tag (e.g., "printpdf--main.1" -> "main.1")
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

	// go is not available, check if mise is available
	if !isMiseAvailable() {
		// Neither go nor mise is available - return go command anyway
		// so the error message will be clear to the user
		return exec.Command("go", args...)
	}

	// Use mise exec to run go
	miseArgs := []string{"exec", fmt.Sprintf("go@%s", goVersion), "--", "go"}
	miseArgs = append(miseArgs, args...)
	return exec.Command("mise", miseArgs...)
}

// getBuildPath determines the correct build path for a Go project.
// Some projects have their main.go in a cmd subdirectory, while others have it in the root.
// Returns the relative path to use with 'go build' (either "." or "./cmd").
func getBuildPath(projectDir string) string {
	// Check if there's a cmd subdirectory with Go files
	cmdDir := filepath.Join(projectDir, "cmd")
	if _, err := os.Stat(cmdDir); err == nil {
		// cmd directory exists, check if it has Go files
		entries, err := os.ReadDir(cmdDir)
		if err != nil {
			// If we can't read the directory (e.g., permissions issue),
			// fall back to default. The actual error will surface when go build runs.
			return "."
		}
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".go" {
				// Found Go files in cmd directory
				return "./cmd"
			}
		}
	}
	// Default to current directory (either cmd exists but has no .go files, or cmd doesn't exist)
	return "."
}

// buildMonoFromSource builds a project from a branch or commit in the mono repository
func buildMonoFromSource(project, refSpec, refDescription string, isCommitSHA bool, dryRun bool, planJson bool) {
	// Determine destination path
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
		fmt.Printf("Building %s from %s...\n", project, refDescription)
		fmt.Println()
		plan.PrintPlan()
		return
	}

	// Execute the plan
	fmt.Printf("Building %s from %s...\n", project, refDescription)
	fmt.Println()
	plan.PrintPlan()
	if !plan.ConfirmPlan() {
		fmt.Println("Cancelled.")
		os.Exit(0)
	}
	fmt.Println()

	// Create a temporary directory for cloning
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("want-mono-%s-*", project))
	if err != nil {
		fmt.Printf("Error: Failed to create temporary directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	// Clone the repository - different approach for commits vs branches
	fmt.Printf("Cloning neongreen/mono (%s)...\n", refDescription)
	var cmd *exec.Cmd
	if isCommitSHA {
		// For commit SHAs, we need a full clone (not shallow) to ensure the commit is available
		cmd = exec.Command("git", "clone", "https://github.com/neongreen/mono.git", tmpDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("\nError: Failed to clone repository: %v\n", err)
			os.Exit(1)
		}

		// Now checkout the specific commit
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
		// For branches and tags, use --branch flag
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

	// Check if project directory exists
	projectDir := filepath.Join(tmpDir, project)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		fmt.Printf("\nError: Project '%s' not found in repository\n", project)
		os.Exit(1)
	}

	// Build the project
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

	// Make executable
	if err := os.Chmod(destPath, 0755); err != nil {
		fmt.Printf("Warning: Failed to make binary executable: %v\n", err)
	}

	fmt.Println()
	fmt.Printf("✓ Built and installed %s from %s to: %s\n", project, refDescription, destPath)
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

// buildMonoFromPR builds a project from a PR branch
func buildMonoFromPR(project string, prNumber int, dryRun bool, planJson bool) {
	// Determine destination path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	destDir := filepath.Join(homeDir, ".local", "bin")
	destPath := filepath.Join(destDir, project)

	// Build the plan first
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
		fmt.Printf("Building %s from PR #%d...\n", project, prNumber)
		fmt.Println()
		plan.PrintPlan()
		return
	}

	// Execute the plan
	fmt.Printf("Building %s from PR #%d...\n", project, prNumber)
	fmt.Println()
	plan.PrintPlan()
	if !plan.ConfirmPlan() {
		fmt.Println("Cancelled.")
		os.Exit(0)
	}
	fmt.Println()

	// Get PR info to find the branch
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

	// Create a temporary directory for cloning
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("want-mono-%s-pr-%d-*", project, prNumber))
	if err != nil {
		fmt.Printf("Error: Failed to create temporary directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	// Clone the repository
	fmt.Printf("Cloning neongreen/mono (branch: %s)...\n", branch)
	cmd := exec.Command("git", "clone", "--depth=1", "--branch", branch,
		"https://github.com/neongreen/mono.git", tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nError: Failed to clone repository: %v\n", err)
		os.Exit(1)
	}

	// Check if project directory exists
	projectDir := filepath.Join(tmpDir, project)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		fmt.Printf("\nError: Project '%s' not found in repository\n", project)
		os.Exit(1)
	}

	// Build the project
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

	// Make executable
	if err := os.Chmod(destPath, 0755); err != nil {
		fmt.Printf("Warning: Failed to make binary executable: %v\n", err)
	}

	fmt.Println()
	fmt.Printf("✓ Built and installed %s from PR #%d to: %s\n", project, prNumber, destPath)
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

// installMonoRelease installs a specific version of a project from neongreen/mono
func installMonoRelease(project, version string, dryRun bool, planJson bool) {
	// Check if this is a PR reference (e.g., "pr-42" or "pr-42.1")
	if strings.HasPrefix(version, "pr-") {
		// Extract PR number
		prStr := strings.TrimPrefix(version, "pr-")
		// Remove any version suffix like ".1"
		parts := strings.Split(prStr, ".")
		var prNumber int
		n, err := fmt.Sscanf(parts[0], "%d", &prNumber)
		if err != nil || n != 1 {
			fmt.Printf("Error: Invalid PR number in '%s'\n", version)
			os.Exit(1)
		}

		// Check if there's a release for this PR
		tag := fmt.Sprintf("%s--%s", project, version)
		_, err = ghrelease.GetReleaseByTag("neongreen", "mono", tag)
		if err != nil {
			// No release found, build from PR
			if !planJson && !dryRun {
				fmt.Printf("No release found for %s (would be tagged as %s)\n", version, tag)
				fmt.Printf("Building from PR #%d instead...\n", prNumber)
				fmt.Println()
			}
			buildMonoFromPR(project, prNumber, dryRun, planJson)
			return
		}
		// Release exists, install it normally
	}

	// Try to fetch as a release tag first
	tag := fmt.Sprintf("%s--%s", project, version)
	_, err := ghrelease.GetReleaseByTag("neongreen", "mono", tag)
	if err != nil {
		// Not a release tag - treat as branch name or commit SHA
		// Common branch names: main, develop, feature-xyz, etc.
		// Commit SHAs are typically 40 hex chars, but can be abbreviated (7+ chars)
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

	// Determine destination path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: Failed to get home directory: %v\n", err)
		os.Exit(1)
	}

	destDir := filepath.Join(homeDir, ".local", "bin")
	destPath := filepath.Join(destDir, project)

	// Build the plan
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
		fmt.Printf("Installing %s version %s from neongreen/mono...\n", project, version)
		fmt.Println()
		plan.PrintPlan()
		return
	}

	// Execute the plan
	fmt.Printf("Installing %s version %s from neongreen/mono...\n", project, version)
	fmt.Println()
	plan.PrintPlan()

	// Check if already installed and ask for confirmation
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

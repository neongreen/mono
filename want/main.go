package main

import (
	"bufio"

	"encoding/json"

	"fmt"
	"os"

	"path/filepath"

	"strings"
)

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
		// Check for --json flag
		jsonOutput := false
		if len(os.Args) > 2 {
			for _, arg := range os.Args[2:] {
				if arg == "--json" {
					jsonOutput = true
					break
				}
			}
		}
		printVersionWithJSON(jsonOutput)
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
  want mono tk@local                   # Build tk from current directory (must be in mono repo)
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

// CompoundHandler is a function that handles compound commands
type CompoundHandler func(args []string, dryRun bool, planJson bool)

// PRInfo holds information about a pull request
type PRInfo struct {
	Number int
	Title  string
	Branch string
}

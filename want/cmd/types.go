package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// GoVersion is the Go version to use when building projects via mise
const GoVersion = "1.24.7"

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

// PRInfo holds information about a pull request
type PRInfo struct {
	Number int
	Title  string
	Branch string
}

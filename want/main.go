package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const version = "0.1.0-mvp"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
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
  want [--dry-run] <requirement>  Get a tool, repository, or resource
  want list                       Show what you have
  want check                      Check status of requirements
  want forget <name>              Remove from tracking (doesn't uninstall)
  want version                    Show version
  want help                       Show this help

Flags:
  --dry-run                       Show what would be done without actually doing it

Examples:
  want jujutsu                    # Get a tool
  want --dry-run jujutsu          # See what would be done
  want github.com/user/repo       # Clone a repository (not yet implemented)

Configuration is stored at ~/.config/want/

For more information, see README.md`)
}

func handleWant(args []string) {
	// Parse flags
	fs := flag.NewFlagSet("want", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be done without doing it")
	fs.Parse(args)

	if fs.NArg() == 0 {
		fmt.Println("Error: no requirement specified")
		fmt.Println("Usage: want [--dry-run] <requirement>")
		os.Exit(1)
	}

	requirement := fs.Arg(0)

	// Check if requirement looks like a git repository
	if strings.Contains(requirement, "/") && (strings.Contains(requirement, "github.com") || strings.Contains(requirement, ".git")) {
		fmt.Printf("Error: Git repository cloning is not yet implemented\n")
		fmt.Printf("Requirement: %s\n", requirement)
		os.Exit(1)
	}

	// Try to install as a tool via mise
	installToolViaMise(requirement, *dryRun)
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

// installToolViaMise installs a tool using mise
func installToolViaMise(tool string, dryRun bool) {
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

	// Try to install via mise
	fmt.Printf("Installing %s via mise...\n", tool)
	
	if dryRun {
		fmt.Println()
		fmt.Println("DRY RUN: Would execute:")
		fmt.Printf("  mise use -g %s\n", tool)
		fmt.Println()
		fmt.Println("This would:")
		fmt.Printf("  • Install %s globally via mise\n", tool)
		fmt.Printf("  • Add %s to your mise global config\n", tool)
		fmt.Printf("  • Make %s available in your PATH\n", tool)
		return
	}

	// Check if mise is available (only for actual installation)
	if !isMiseAvailable() {
		fmt.Println()
		fmt.Println("Error: mise is not installed")
		fmt.Println("Please install mise first: https://mise.jdx.dev/")
		os.Exit(1)
	}

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
		fmt.Println("You may need to restart your shell or run:")
		fmt.Println("  eval \"$(mise activate bash)\"  # for bash")
		fmt.Println("  eval \"$(mise activate zsh)\"   # for zsh")
	}
}

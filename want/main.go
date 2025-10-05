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
  want json <command>             Convert command output to JSON
  want md <url>                   Convert URL to markdown
  want list                       Show what you have
  want check                      Check status of requirements
  want forget <name>              Remove from tracking (doesn't uninstall)
  want version                    Show version
  want help                       Show this help

Flags:
  --dry-run                       Show what would be done without actually doing it

Examples:
  want jujutsu                    # Install a tool
  want --dry-run jujutsu          # Preview installation
  want json ps                    # Get running processes as JSON (uses jc)
  want md https://example.com     # Convert webpage to markdown
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
	remainingArgs := fs.Args()[1:]

	// Check for compound commands (e.g., "want json ps" or "want md url")
	if handler, ok := getCompoundHandler(requirement); ok {
		handler(remainingArgs, *dryRun)
		return
	}

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

// CompoundHandler is a function that handles compound commands
type CompoundHandler func(args []string, dryRun bool)

// getCompoundHandler returns a handler for compound commands if one exists
func getCompoundHandler(command string) (CompoundHandler, bool) {
	handlers := map[string]CompoundHandler{
		"json": handleJsonCommand,
		"md":   handleMarkdownCommand,
	}
	handler, ok := handlers[command]
	return handler, ok
}

// handleJsonCommand handles "want json <command>" - converts command output to JSON
func handleJsonCommand(args []string, dryRun bool) {
	if len(args) == 0 {
		fmt.Println("Error: no command specified")
		fmt.Println("Usage: want json <command>")
		fmt.Println("\nExample:")
		fmt.Println("  want json ps    # Get running processes as JSON")
		os.Exit(1)
	}

	// Check if jc is available
	if !isToolAvailable("jc") {
		fmt.Println("The 'json' command requires 'jc' (JSON CLI output formatter)")
		fmt.Println()
		
		if dryRun {
			fmt.Println("DRY RUN - Execution plan:")
			fmt.Println()
			fmt.Println("Step 1: Install jc")
			fmt.Println("  $ mise use -g jc")
			fmt.Println()
			fmt.Println("Step 2: Execute command")
			fmt.Printf("  $ jc %s\n", strings.Join(args, " "))
			return
		}

		fmt.Println("Installing jc...")
		installToolViaMise("jc", false)
		fmt.Println()
		
		// Check if installation succeeded
		if !isToolAvailable("jc") {
			fmt.Println("Error: jc installation failed")
			os.Exit(1)
		}
	}

	// Build and execute the command
	commandStr := strings.Join(args, " ")
	
	if dryRun {
		fmt.Println("DRY RUN - Execution plan:")
		fmt.Println()
		fmt.Println("Step 1: Execute command")
		fmt.Printf("  $ jc %s\n", commandStr)
		return
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
func handleMarkdownCommand(args []string, dryRun bool) {
	if len(args) == 0 {
		fmt.Println("Error: no URL specified")
		fmt.Println("Usage: want md <url>")
		fmt.Println("\nExample:")
		fmt.Println("  want md https://example.com    # Convert webpage to markdown")
		os.Exit(1)
	}

	url := args[0]

	// Check which tool is available (prefer markitdown, fallback to pure.md via curl)
	hasMarkitdown := isToolAvailable("markitdown")
	
	if !hasMarkitdown {
		fmt.Println("The 'md' command requires 'markitdown' (Microsoft's HTML to Markdown converter)")
		fmt.Println("Repository: https://github.com/microsoft/markitdown")
		fmt.Println()
		
		if dryRun {
			fmt.Println("DRY RUN - Execution plan:")
			fmt.Println()
			if isToolAvailable("pip") || isToolAvailable("pip3") {
				pipCmd := "pip3"
				if !isToolAvailable("pip3") {
					pipCmd = "pip"
				}
				fmt.Println("Step 1: Install markitdown")
				fmt.Printf("  $ %s install markitdown\n", pipCmd)
				fmt.Println()
				fmt.Println("Step 2: Convert URL to markdown")
				fmt.Printf("  $ markitdown %s\n", url)
				fmt.Println()
				fmt.Println("  (If step 1 fails, fallback to pure.md)")
			} else {
				fmt.Println("Step 1: Convert URL to markdown using pure.md")
				fmt.Printf("  $ curl -s https://pure.md/%s\n", url)
			}
			return
		}

		// Try to install markitdown via pip if available
		if isToolAvailable("pip") || isToolAvailable("pip3") {
			fmt.Println("Installing markitdown via pip...")
			
			pipCmd := "pip3"
			if !isToolAvailable("pip3") {
				pipCmd = "pip"
			}
			
			cmd := exec.Command(pipCmd, "install", "markitdown")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			
			if err := cmd.Run(); err != nil {
				fmt.Println("\nWarning: markitdown installation failed")
				fmt.Println("Falling back to pure.md service...")
				usePureMd(url)
				return
			}
			
			fmt.Println()
			hasMarkitdown = isToolAvailable("markitdown")
		}
		
		// If still not available, use pure.md
		if !hasMarkitdown {
			fmt.Println("markitdown not available, using pure.md service...")
			usePureMd(url)
			return
		}
	}

	if dryRun {
		fmt.Println("DRY RUN - Execution plan:")
		fmt.Println()
		fmt.Println("Step 1: Convert URL to markdown")
		fmt.Printf("  $ markitdown %s\n", url)
		return
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
		fmt.Println("DRY RUN - Execution plan:")
		fmt.Println()
		fmt.Println("Step 1: Install tool via mise")
		fmt.Printf("  $ mise use -g %s\n", tool)
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

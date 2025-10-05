package main

import (
	"fmt"
	"os"
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
  want <requirement>     Get a tool, repository, or resource
  want list              Show what you have
  want check             Check status of requirements
  want forget <name>     Remove from tracking (doesn't uninstall)
  want version           Show version
  want help              Show this help

Examples:
  want jujutsu                    # Get a tool
  want github.com/user/repo       # Clone a repository

Configuration is stored at ~/.config/want/

For more information, see README.md`)
}

func handleWant(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: no requirement specified")
		fmt.Println("Usage: want <requirement>")
		os.Exit(1)
	}

	requirement := args[0]
	fmt.Printf("MVP: Would install/get: %s\n", requirement)
	fmt.Println("This is a minimal viable product. Full implementation coming soon.")
	fmt.Println("\nPlanned features:")
	fmt.Println("  • Install tools via mise, Homebrew, or other providers")
	fmt.Println("  • Clone git repositories")
	fmt.Println("  • Interactive option selection")
	fmt.Println("  • Preference learning")
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

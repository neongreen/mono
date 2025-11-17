package cmd

import (
	"fmt"
	"os"
	"strings"
)

// handleRequirement processes a general requirement (tool, URL, etc.)
func handleRequirement(args []string, dryRun, planJSON bool) error {
	if len(args) == 0 {
		return fmt.Errorf("no requirement specified")
	}

	requirement := args[0]
	remainingArgs := args[1:]

	// Check for compound handlers (json, md, excalifont, etc.)
	if handler, ok := getCompoundHandler(requirement); ok {
		handler(remainingArgs, dryRun, planJSON)
		return nil
	}

	// Check for GitHub release URLs
	if strings.Contains(requirement, "github.com") && (strings.Contains(requirement, "/releases/download/") || strings.Contains(requirement, "/releases/tag/")) {
		handleGitHubAsset(requirement, dryRun, planJSON)
		return nil
	}

	// Check for git repositories
	if strings.Contains(requirement, "/") && (strings.Contains(requirement, "github.com") || strings.Contains(requirement, ".git")) {
		fmt.Printf("Error: Git repository cloning is not yet implemented\n")
		fmt.Printf("Requirement: %s\n", requirement)
		os.Exit(1)
	}

	// Default: try to install as a tool via mise
	installToolViaMise(requirement, dryRun, planJSON)
	return nil
}

// Function references that will be set by main package
var (
	GetCompoundHandlerFunc func(string) (CompoundHandlerFunc, bool)
	HandleGitHubAssetFunc  func(string, bool, bool)
	InstallToolViaMiseFunc func(string, bool, bool)
)

func getCompoundHandler(requirement string) (CompoundHandlerFunc, bool) {
	if GetCompoundHandlerFunc != nil {
		return GetCompoundHandlerFunc(requirement)
	}
	return nil, false
}

func handleGitHubAsset(requirement string, dryRun, planJSON bool) {
	if HandleGitHubAssetFunc != nil {
		HandleGitHubAssetFunc(requirement, dryRun, planJSON)
		return
	}
	fmt.Printf("Would handle GitHub asset: %s\n", requirement)
}

func installToolViaMise(requirement string, dryRun, planJSON bool) {
	if InstallToolViaMiseFunc != nil {
		InstallToolViaMiseFunc(requirement, dryRun, planJSON)
		return
	}
	fmt.Printf("Would install tool via mise: %s\n", requirement)
}

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ensureToolAvailable checks if a tool is available and returns installation steps if not
func ensureToolAvailable(toolName string) (available bool, installSteps []PlanStep) {
	if isToolAvailable(toolName) {
		return true, nil
	}

	tool, exists := ToolRegistry[toolName]
	if !exists {

		return false, nil
	}

	return false, []PlanStep{tool.InstallStep}
}

// buildToolInstallationPlan creates a complete installation plan for a tool
// It handles dependencies (like mise) automatically based on the registry
func buildToolInstallationPlan(toolName string) []PlanStep {
	var steps []PlanStep

	if isToolAvailable(toolName) {
		return steps
	}

	tool, exists := ToolRegistry[toolName]
	if !exists {

		return steps
	}

	if tool.MisePackage != "" && !isMiseAvailable() {

		steps = append(steps, buildMiseInstallationSteps()...)
	}

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

// buildMiseInstallationSteps returns the steps needed to install mise
func buildMiseInstallationSteps() []PlanStep {
	var steps []PlanStep

	miseTool, exists := ToolRegistry["mise"]
	if !exists {

		panic("BUG: 'mise' tool not found in ToolRegistry. Please ensure ToolRegistry includes a 'mise' entry with proper installation steps.")
	}

	steps = append(steps, miseTool.InstallStep)

	if !isMiseActivated() {
		steps = append(steps, buildShellConfigStep())
		steps = append(steps, buildManualActivationStep())
	}

	return steps
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

// installToolViaMise installs a tool using mise
func installToolViaMise(tool string, dryRun bool, planJson bool) {

	if tool == "mise" {
		if isMiseAvailable() {
			fmt.Printf("✓ mise is already available\n")
			fmt.Println()

			cmd := exec.Command("which", "mise")
			output, err := cmd.Output()
			if err == nil {
				fmt.Printf("  Location: %s", string(output))
			}

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

		err := installMise(dryRun, planJson)
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if isToolAvailable(tool) {
		fmt.Printf("✓ %s is already available\n", tool)
		fmt.Println()

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

	plan := FulfillmentPlan{
		Requirement: tool,
		Steps:       []PlanStep{},
	}

	if registryTool, inRegistry := ToolRegistry[tool]; inRegistry && registryTool.MisePackage != "" {

		plan.Steps = buildToolInstallationPlan(tool)
	} else {

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

	if !isMiseAvailable() {
		fmt.Println()
		fmt.Println("Error: mise is not installed")
		fmt.Println("Please install mise first with:")
		fmt.Println("  want mise")
		os.Exit(1)
	}

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

	if isToolAvailable(tool) {
		fmt.Printf("✓ %s is now available in your PATH\n", tool)
	} else {
		fmt.Printf("⚠ %s was installed but may not be in your PATH yet\n", tool)
		shellName := getShellName()
		fmt.Println("You may need to restart your shell or run:")
		fmt.Printf("  eval \"$(mise activate %s)\"\n", shellName)
	}
}

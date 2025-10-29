package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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

	safeCommands := []string{
		"ps", "top", "uptime", "whoami", "id", "hostname", "uname",
		"df", "du", "free", "vmstat", "iostat", "netstat", "ss",
		"ls", "pwd", "cat", "head", "tail", "wc", "grep",
		"cal", "env", "printenv", "which", "whereis",
		"git status", "git log", "git diff", "git show",
	}

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

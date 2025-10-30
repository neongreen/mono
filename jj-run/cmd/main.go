package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// Change represents a jj change with its metadata
type Change struct {
	CommitID    string   `json:"commit_id"`
	ChangeID    string   `json:"change_id"`
	Description string   `json:"description"`
	Parents     []string `json:"parents"`
}

// ErrorStrategy defines how to handle command failures
type ErrorStrategy string

const (
	ErrorContinue ErrorStrategy = "continue"
	ErrorStop     ErrorStrategy = "stop"
	ErrorFatal    ErrorStrategy = "fatal"
)

var (
	revset      string
	errStrategy string
)

var rootCmd = &cobra.Command{
	Use:   "jj-run [flags] <command>",
	Short: "Execute shell commands across multiple repository changes in isolated workspaces using jj",
	Long: `A script to execute shell commands across multiple repository changes in isolated workspaces using jj.

Runs arbitrary shell commands for each change in a revset, in isolation.
Uses a temporary workspace for each run, so your main repo doesn't change while the script is running.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCommand,
}

func init() {
	rootCmd.Flags().StringVarP(&revset, "revset", "r", "reachable(@, mutable())", "Revset to process")
	rootCmd.Flags().StringVarP(&errStrategy, "err-strategy", "e", "continue", "Error handling strategy (continue|stop|fatal)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runCommand(cmd *cobra.Command, args []string) error {
	command := strings.Join(args, " ")

	// Validate error strategy
	strategy := ErrorStrategy(errStrategy)
	if strategy != ErrorContinue && strategy != ErrorStop && strategy != ErrorFatal {
		return fmt.Errorf("invalid error strategy: %s (must be continue, stop, or fatal)", errStrategy)
	}

	// Get current operation ID
	beforeOp, err := getCurrentOpID()
	if err != nil {
		return fmt.Errorf("failed to get current operation ID: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Current operation: %s. To revert, run:\n", beforeOp[:12])
	fmt.Fprintf(os.Stderr, "  jj op restore %s\n\n", beforeOp[:12])

	// Create and manage workspace
	workspacePath, workspaceName, err := createWorkspace()
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}
	defer forgetWorkspace(workspaceName)

	// Get workspace change
	workspaceChanges, err := getChangeList(fmt.Sprintf("%s@", workspaceName), workspacePath)
	if err != nil {
		return fmt.Errorf("failed to get workspace change: %w", err)
	}
	if len(workspaceChanges) != 1 {
		return fmt.Errorf("expected exactly 1 workspace change, got %d", len(workspaceChanges))
	}
	workspaceChange := workspaceChanges[0]

	// Get changes to process
	changes, err := getChangeList(fmt.Sprintf("(%s) ~ %s ~ root()", revset, workspaceChange.ChangeID), workspacePath)
	if err != nil {
		return fmt.Errorf("failed to get change list: %w", err)
	}

	totalChanges := len(changes)
	if totalChanges == 0 {
		fmt.Fprintf(os.Stderr, "No changes found to process.\n")
		abandonChanges([]*Change{workspaceChange})
		return nil
	}

	// Process changes
	newChanges, allSuccessful := processChanges(workspacePath, changes, command, strategy)

	// Rewrite parents
	modifiedCount := rewriteParents(workspacePath, newChanges)

	// Update stale workspaces
	runJJ([]string{"workspace", "update-stale"}, ".")
	runJJ([]string{"workspace", "update-stale"}, workspacePath)

	// Abandon all created changes
	allChanges := append([]*Change{}, newChanges...)
	allChanges = append(allChanges, workspaceChange)
	abandonChanges(allChanges)

	fmt.Fprintf(os.Stderr, "Rewrote %d/%d commits.\n", modifiedCount, totalChanges)
	if !allSuccessful {
		fmt.Fprintf(os.Stderr, "Not all changes were processed successfully.\n")
	}

	// Get after operation ID
	afterOp, err := getCurrentOpID()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't get current operation ID. Likely a bug in jj-run.\n")
		return fmt.Errorf("failed to get after operation ID: %w", err)
	}

	if modifiedCount > 0 {
		fmt.Fprintf(os.Stderr, "To compare the changes between the 'before' and 'after' repo states, run:\n")
		fmt.Fprintf(os.Stderr, "  jj operation diff --from %s --to %s -p\n\n", beforeOp[:12], afterOp[:12])
	}

	return nil
}

func getCurrentOpID() (string, error) {
	out, err := runJJOutput([]string{"op", "log", "-n1", "-Tid", "--no-graph", "--no-pager"}, ".")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func createWorkspace() (string, string, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "jj-run-")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate workspace name from directory basename
	workspaceName := filepath.Base(tempDir)
	workspacePath := filepath.Join(tempDir, workspaceName)

	// Add the workspace
	if _, err := runJJOutput([]string{"workspace", "add", workspacePath}, "."); err != nil {
		return "", "", fmt.Errorf("failed to add workspace: %w", err)
	}

	return workspacePath, workspaceName, nil
}

func forgetWorkspace(workspaceName string) {
	runJJ([]string{"workspace", "forget", workspaceName}, ".")
}

func getChangeList(revset string, cwd string) ([]*Change, error) {
	args := []string{
		"log",
		"-r", revset,
		"-T", "json(self)",
		"--no-graph",
	}

	out, err := runJJOutput(args, cwd)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(out) == "" {
		return []*Change{}, nil
	}

	// Parse concatenated JSON objects
	var changes []*Change
	decoder := json.NewDecoder(strings.NewReader(out))
	for decoder.More() {
		var change Change
		if err := decoder.Decode(&change); err != nil {
			return nil, fmt.Errorf("failed to decode change: %w", err)
		}
		changes = append(changes, &change)
	}

	return changes, nil
}

func processChanges(workspacePath string, changes []*Change, command string, strategy ErrorStrategy) ([]*Change, bool) {
	var newChanges []*Change
	allSuccessful := true
	exitEarly := false

	totalChanges := len(changes)
	for idx, change := range changes {
		message := strings.TrimSpace(change.Description)
		if message == "" {
			message = "(no description set)"
		}

		changeID := change.ChangeID
		fmt.Fprintf(os.Stderr, "Processing change %d/%d %s: %s\n", idx+1, totalChanges, changeID[:12], message)

		// Create new change based on this one
		if _, err := runJJOutput([]string{"new", changeID}, workspacePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating new change: %v\n", err)
			allSuccessful = false
			if exitEarly = handleError(strategy, changeID[:12], err); exitEarly {
				break
			}
			continue
		}

		// Run the command
		result, err := runShellCommand(command, workspacePath)
		printCommandResult(result, err)

		if err != nil {
			allSuccessful = false
			if exitEarly = handleError(strategy, changeID[:12], err); exitEarly {
				break
			}
		}

		// Get the newly created change
		newChangeList, err := getChangeList("@", workspacePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to get new change: %v\n", err)
		} else {
			newChanges = append(newChanges, newChangeList...)
		}

		if exitEarly {
			break
		}
	}

	return newChanges, allSuccessful
}

func getChangeMetadata(changeID string, cwd string) (*Change, error) {
	changes, err := getChangeList(changeID, cwd)
	if err != nil {
		return nil, err
	}
	if len(changes) == 0 {
		return nil, fmt.Errorf("no change found for ID: %s", changeID)
	}
	return changes[0], nil
}

func isChangeEmpty(workspacePath string, changeID string) (bool, error) {
	out, err := runJJOutput(
		[]string{"log", "-T", "json(empty)", "-r", fmt.Sprintf("present(%s)", changeID), "--no-graph"},
		workspacePath,
	)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "false", nil
}

func rewriteParents(workspacePath string, changes []*Change) int {
	modifiedCount := 0
	for _, change := range changes {
		if len(change.Parents) == 0 {
			continue
		}

		// Check if change is empty
		isEmpty, err := isChangeEmpty(workspacePath, change.ChangeID)
		if err != nil || isEmpty {
			continue
		}

		// Edit the parent
		if _, err := runJJOutput([]string{"edit", change.Parents[0]}, workspacePath); err != nil {
			continue
		}

		// Restore from the change
		if _, err := runJJOutput([]string{"restore", "--from", change.ChangeID, "--restore-descendants"}, workspacePath); err != nil {
			continue
		}

		modifiedCount++
	}

	return modifiedCount
}

func abandonChanges(changes []*Change) {
	for _, change := range changes {
		// Only print first 12 chars when abandoning
		shortID := change.ChangeID
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}
		runJJ([]string{"abandon", fmt.Sprintf("present(%s)", shortID), "--ignore-working-copy"}, ".")
	}
}

func handleError(strategy ErrorStrategy, changeID string, err error) bool {
	errorMsg := formatError(changeID, err)

	switch strategy {
	case ErrorContinue:
		fmt.Fprintf(os.Stderr, "%s\n", errorMsg)
		return false
	case ErrorStop:
		fmt.Fprintf(os.Stderr, "Stopped on change [with fail] %s:\n%s\n", changeID, errorMsg)
		// Exit with error status but return true to stop processing
		os.Exit(1)
		return true
	case ErrorFatal:
		fmt.Fprintf(os.Stderr, "Fatal error at change [%s]:\n%s\n", changeID, errorMsg)
		os.Exit(1)
		return true
	}
	return false
}

func formatError(changeID string, err error) string {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return fmt.Sprintf("Error while processing change [%s]:\nReturn code: %d\nSTDERR:\n%s",
			changeID,
			exitErr.ExitCode(),
			string(exitErr.Stderr))
	}
	return fmt.Sprintf("Error while processing change [%s]:\n%v", changeID, err)
}

func printCommandResult(result *exec.Cmd, err error) {
	// For successful commands, the output is already printed by runShellCommand
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(os.Stderr, "Command failed with return code %d\n", exitErr.ExitCode())
		} else {
			fmt.Fprintf(os.Stderr, "Command failed: %v\n", err)
		}
	}
	fmt.Println()
}

func runShellCommand(command string, cwd string) (*exec.Cmd, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	return cmd, err
}

func runJJ(args []string, cwd string) error {
	_, err := runJJOutput(args, cwd)
	return err
}

func runJJOutput(args []string, cwd string) (string, error) {
	cmd := exec.Command("jj", args...)
	cmd.Dir = cwd

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitErr.Stderr = stderr.Bytes()
		}
		return "", err
	}

	return stdout.String(), nil
}

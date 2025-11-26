package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/neongreen/mono/lib/cli"
	"github.com/neongreen/mono/lib/version"
	"github.com/spf13/cobra"
)

// Change represents a jj change with its metadata
type Change struct {
	CommitID    string   `json:"commit_id"`
	ChangeID    string   `json:"change_id"`
	Description string   `json:"description"`
	Parents     []string `json:"parents"`
}

// Job represents a change processing job
type Job struct {
	Index  int
	Change *Change
}

// Result represents the result of processing a job
type Result struct {
	Index      int
	NewChanges []*Change
	Success    bool
	ShouldStop bool
	Error      error
}

// ErrorStrategy defines how to handle command failures
type ErrorStrategy string

const (
	ErrorContinue ErrorStrategy = "continue"
	ErrorStop     ErrorStrategy = "stop"
	ErrorFatal    ErrorStrategy = "fatal"
)

var (
	revset          string
	errStrategy     string
	directMode      bool
	showVersion     bool
	ignoreImmutable bool
	jobs            int
	repoURL         string
)

var rootCmd = &cobra.Command{
	Use:   "jj-run [flags] <command>",
	Short: "Execute shell commands across multiple repository changes in isolated workspaces using jj",
	Long: `A script to execute shell commands across multiple repository changes in isolated workspaces using jj.

Runs arbitrary shell commands for each change in a revset, in isolation.
Uses a temporary workspace for each run, so your main repo doesn't change while the script is running.

Direct mode (--direct): Instead of using temporary workspaces, directly edits each revision in place.
Useful for metadata changes (e.g., changing commit descriptions, authors) that don't require file isolation.

Parallel processing (-j/--jobs): Process changes concurrently using multiple workers (only available in workspace mode).
Set to 1 (default) for sequential processing, or higher numbers for parallel execution.`,
	Args: func(cmd *cobra.Command, args []string) error {
		// Allow 0 args if --version is specified
		if showVersion {
			return nil
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	RunE: runCommand,
}

func init() {
	rootCmd.Flags().StringVarP(&revset, "revset", "r", "reachable(@, mutable())", "Revset to process")
	rootCmd.Flags().StringVarP(&errStrategy, "err-strategy", "e", "continue", "Error handling strategy (continue|stop|fatal)")
	rootCmd.Flags().BoolVarP(&directMode, "direct", "d", false, "Direct mode: edit each revision in place without worktrees")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Show version information")
	rootCmd.Flags().BoolVar(&ignoreImmutable, "ignore-immutable", false, "Allow rewriting immutable commits")
	rootCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "Number of parallel workers (only for workspace mode, not direct mode)")
	rootCmd.Flags().StringVar(&repoURL, "repo-untested", "", "Clone and work on a remote repository (e.g., https://github.com/user/repo) [AI-written, untested]")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runCommand(cmd *cobra.Command, args []string) error {
	// Handle --version flag
	if showVersion {
		version.PrintVersion("jj-run")
		return nil
	}

	// Check if we have at least one argument
	if len(args) == 0 {
		return fmt.Errorf("command is required")
	}

	command := strings.Join(args, " ")

	// Validate error strategy
	strategy := ErrorStrategy(errStrategy)
	if strategy != ErrorContinue && strategy != ErrorStop && strategy != ErrorFatal {
		return fmt.Errorf("invalid error strategy: %s (must be continue, stop, or fatal)", errStrategy)
	}

	// Validate jobs
	if jobs < 1 {
		return fmt.Errorf("jobs must be at least 1, got %d", jobs)
	}
	if directMode && jobs > 1 {
		return fmt.Errorf("parallel processing is not supported in direct mode")
	}

	// Handle --repo flag: clone repository to temp directory
	var repoDir string
	var cleanupRepo func()
	if repoURL != "" {
		var err error
		repoDir, cleanupRepo, err = cloneRepository(repoURL)
		if err != nil {
			return fmt.Errorf("failed to clone repository: %w", err)
		}
		defer cleanupRepo()
		fmt.Fprintf(os.Stderr, "Cloned repository to %s\n\n", cli.Key(repoDir))
	} else {
		repoDir = "."
	}

	// Get current operation ID
	beforeOp, err := getCurrentOpID(repoDir)
	if err != nil {
		return fmt.Errorf("failed to get current operation ID: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Current operation: %s. To revert, run:\n", cli.Key(beforeOp[:12]))
	fmt.Fprintf(os.Stderr, "  jj op restore %s\n\n", cli.Key(beforeOp[:12]))

	// Use direct mode if requested
	if directMode {
		err := runDirectMode(command, strategy, beforeOp, repoDir)
		if err != nil {
			return err
		}
		if repoURL != "" {
			return promptPush(repoDir)
		}
		return nil
	}

	// Create and manage workspace
	workspacePath, workspaceName, err := createWorkspace(repoDir)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}
	defer forgetWorkspace(workspaceName, repoDir)

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
		abandonChanges([]*Change{workspaceChange}, repoDir)
		return nil
	}

	// Process changes (either sequentially or in parallel)
	var newChanges []*Change
	var allSuccessful bool
	var processErr error

	if jobs > 1 {
		newChanges, allSuccessful, processErr = processChangesParallel(changes, command, strategy, jobs, repoDir)

		// In parallel mode, workers create changes in their own workspaces, so the base workspace
		// is now stale. Update it before attempting to rewrite parents, otherwise jj edit/restore
		// commands in rewriteParents will fail with "workspace is outdated" errors.
		runJJ([]string{"workspace", "update-stale"}, workspacePath)
	} else {
		newChanges, allSuccessful, processErr = processChanges(workspacePath, changes, command, strategy)
	}

	// Rewrite parents
	modifiedCount := rewriteParents(workspacePath, newChanges)

	// Update stale workspaces
	runJJ([]string{"workspace", "update-stale"}, repoDir)
	runJJ([]string{"workspace", "update-stale"}, workspacePath)

	// Abandon all created changes
	allChanges := append([]*Change{}, newChanges...)
	allChanges = append(allChanges, workspaceChange)
	abandonChanges(allChanges, repoDir)

	fmt.Fprintf(os.Stderr, "%s %d/%d commits.\n", cli.Success("Rewrote"), modifiedCount, totalChanges)
	if !allSuccessful {
		fmt.Fprintf(os.Stderr, "%s\n", cli.Warning("Not all changes were processed successfully."))
	}

	// Get after operation ID
	afterOp, err := getCurrentOpID(repoDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't get current operation ID. Likely a bug in jj-run.\n")
		return fmt.Errorf("failed to get after operation ID: %w", err)
	}

	if modifiedCount > 0 {
		fmt.Fprintf(os.Stderr, "To compare the changes between the 'before' and 'after' repo states, run:\n")
		fmt.Fprintf(os.Stderr, "  jj operation diff --from %s --to %s -p\n\n", beforeOp[:12], afterOp[:12])
	}

	// Return error from processChanges if one occurred (stop/fatal)
	if processErr != nil {
		return processErr
	}

	// Prompt to push if working on a remote repo
	if repoURL != "" {
		return promptPush(repoDir)
	}

	return nil
}

func runDirectMode(command string, strategy ErrorStrategy, beforeOp string, repoDir string) error {
	// Get changes to process (exclude root)
	changes, err := getChangeList(fmt.Sprintf("(%s) ~ root()", revset), repoDir)
	if err != nil {
		return fmt.Errorf("failed to get change list: %w", err)
	}

	totalChanges := len(changes)
	if totalChanges == 0 {
		fmt.Fprintf(os.Stderr, "No changes found to process.\n")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Processing %s in direct mode...\n", cli.Header(fmt.Sprintf("%d changes", totalChanges)))

	allSuccessful := true
	processedCount := 0

	for idx, change := range changes {
		message := strings.TrimSpace(change.Description)
		if message == "" {
			message = "(no description set)"
		}

		changeID := change.ChangeID
		fmt.Fprintf(os.Stderr, "Processing change %d/%d %s: %s\n", idx+1, totalChanges, changeID[:12], message)

		// Edit this change (make it the working copy)
		if _, err := runJJOutput([]string{"edit", changeID}, repoDir); err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", cli.Error("Error editing change:"), err)
			allSuccessful = false
			exitEarly, handlerErr := handleError(strategy, changeID[:12], err)
			if exitEarly {
				return handlerErr
			}
			continue
		}

		// Run the command in the repository
		err := runShellCommand(command, repoDir)
		printCommandResult(err)

		if err != nil {
			allSuccessful = false
			exitEarly, handlerErr := handleError(strategy, changeID[:12], err)
			if exitEarly {
				return handlerErr
			}
		} else {
			processedCount++
		}
	}

	fmt.Fprintf(os.Stderr, "%s %d/%d changes.\n", cli.Success("Processed"), processedCount, totalChanges)
	if !allSuccessful {
		fmt.Fprintf(os.Stderr, "%s\n", cli.Warning("Not all changes were processed successfully."))
	}

	// Get after operation ID
	afterOp, err := getCurrentOpID(repoDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't get current operation ID. Likely a bug in jj-run.\n")
		return fmt.Errorf("failed to get after operation ID: %w", err)
	}

	if processedCount > 0 {
		fmt.Fprintf(os.Stderr, "To compare the changes between the 'before' and 'after' repo states, run:\n")
		fmt.Fprintf(os.Stderr, "  jj operation diff --from %s --to %s -p\n\n", beforeOp[:12], afterOp[:12])
	}

	if !allSuccessful {
		return fmt.Errorf("some changes failed to process")
	}

	return nil
}

func cloneRepository(url string) (string, func(), error) {
	// Create a temporary directory for the cloned repo
	tempDir, err := os.MkdirTemp("", "jj-run-repo-")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	// Clone using jj git clone
	fmt.Fprintf(os.Stderr, "Cloning repository from %s...\n", cli.Key(url))
	cmd := exec.Command("jj", "git", "clone", url, tempDir)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	return tempDir, cleanup, nil
}

func promptPush(repoDir string) error {
	fmt.Fprintf(os.Stderr, "\n%s\n", cli.Header("Ready to push changes"))
	fmt.Fprintf(os.Stderr, "Do you want to push the changes? [y/N]: ")

	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	if response == "y" || response == "yes" {
		fmt.Fprintf(os.Stderr, "Pushing changes...\n")
		cmd := exec.Command("jj", "git", "push")
		cmd.Dir = repoDir
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to push: %w", err)
		}
		fmt.Fprintf(os.Stderr, "%s\n", cli.Success("Changes pushed successfully"))
	} else {
		fmt.Fprintf(os.Stderr, "Push skipped. You can manually push from: %s\n", cli.Key(repoDir))
	}

	return nil
}

func getCurrentOpID(cwd string) (string, error) {
	out, err := runJJOutput([]string{"op", "log", "-n1", "-Tid", "--no-graph", "--no-pager"}, cwd)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func createWorkspace(repoDir string) (string, string, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "jj-run-")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate workspace name from directory basename
	workspaceName := filepath.Base(tempDir)
	workspacePath := filepath.Join(tempDir, workspaceName)

	// Add the workspace
	if _, err := runJJOutput([]string{"workspace", "add", workspacePath}, repoDir); err != nil {
		return "", "", fmt.Errorf("failed to add workspace: %w", err)
	}

	return workspacePath, workspaceName, nil
}

func forgetWorkspace(workspaceName string, repoDir string) {
	runJJ([]string{"workspace", "forget", workspaceName}, repoDir)
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

func processChanges(workspacePath string, changes []*Change, command string, strategy ErrorStrategy) ([]*Change, bool, error) {
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
		fmt.Fprintf(os.Stderr, "Processing change %d/%d %s: %s\n", idx+1, totalChanges, cli.Key(changeID[:12]), message)

		// Create new change based on this one
		if _, err := runJJOutput([]string{"new", changeID}, workspacePath); err != nil {
			fmt.Fprintf(os.Stderr, cli.Error("Error creating new change:")+" %v\n", err)
			allSuccessful = false
			var handlerErr error
			if exitEarly, handlerErr = handleError(strategy, changeID[:12], err); exitEarly {
				return newChanges, allSuccessful, handlerErr
			}
			continue
		}

		// Run the command
		err := runShellCommand(command, workspacePath)
		printCommandResult(err)

		if err != nil {
			allSuccessful = false
			var handlerErr error
			if exitEarly, handlerErr = handleError(strategy, changeID[:12], err); exitEarly {
				return newChanges, allSuccessful, handlerErr
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

	return newChanges, allSuccessful, nil
}

func processChangesParallel(changes []*Change, command string, strategy ErrorStrategy, numWorkers int, repoDir string) ([]*Change, bool, error) {
	totalChanges := len(changes)

	// Create channels for job distribution and result collection
	jobs := make(chan Job, totalChanges)
	results := make(chan Result, totalChanges)
	stopSignal := make(chan struct{})

	var wg sync.WaitGroup

	// Start workers
	for w := range numWorkers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Create workspace for this worker
			workspacePath, workspaceName, err := createWorkspace(repoDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, cli.Error("Worker %d: failed to create workspace:")+" %v\n", workerID, err)
				return
			}
			defer forgetWorkspace(workspaceName, repoDir)

			// Process jobs
			for job := range jobs {
				// Check if we should stop
				select {
				case <-stopSignal:
					return
				default:
				}

				change := job.Change
				message := strings.TrimSpace(change.Description)
				if message == "" {
					message = "(no description set)"
				}

				changeID := change.ChangeID
				fmt.Fprintf(os.Stderr, "Processing change %d/%d %s: %s\n", job.Index+1, totalChanges, cli.Key(changeID[:12]), message)

				result := Result{Index: job.Index, Success: true}

				// Create new change based on this one
				if _, err := runJJOutput([]string{"new", changeID}, workspacePath); err != nil {
					fmt.Fprintf(os.Stderr, cli.Error("Error creating new change:")+" %v\n", err)
					result.Success = false
					result.Error = err
					exitEarly, handlerErr := handleError(strategy, changeID[:12], err)
					if exitEarly {
						result.ShouldStop = true
						result.Error = handlerErr
					}
					results <- result
					continue
				}

				// Run the command
				err := runShellCommand(command, workspacePath)
				printCommandResult(err)

				if err != nil {
					result.Success = false
					result.Error = err
					exitEarly, handlerErr := handleError(strategy, changeID[:12], err)
					if exitEarly {
						result.ShouldStop = true
						result.Error = handlerErr
					}
				}

				// Get the newly created change
				newChangeList, err := getChangeList("@", workspacePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to get new change: %v\n", err)
				} else {
					result.NewChanges = newChangeList
				}

				results <- result
			}
		}(w)
	}

	// Send jobs to workers
	go func() {
		for idx, change := range changes {
			jobs <- Job{Index: idx, Change: change}
		}
		close(jobs)
	}()

	// Wait for workers to finish in a separate goroutine
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	resultSlice := make([]Result, totalChanges)
	shouldStop := false
	var stopErr error

	for result := range results {
		resultSlice[result.Index] = result

		// If any result signals we should stop, broadcast stop to all workers
		if result.ShouldStop && !shouldStop {
			shouldStop = true
			stopErr = result.Error
			close(stopSignal)
		}
	}

	// Collect all new changes in order
	var newChanges []*Change
	allSuccessful := true
	for _, result := range resultSlice {
		if !result.Success {
			allSuccessful = false
		}
		newChanges = append(newChanges, result.NewChanges...)
	}

	return newChanges, allSuccessful, stopErr
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

func abandonChanges(changes []*Change, repoDir string) {
	for _, change := range changes {
		// Only print first 12 chars when abandoning
		shortID := change.ChangeID
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}
		runJJ([]string{"abandon", fmt.Sprintf("present(%s)", shortID), "--ignore-working-copy"}, repoDir)
	}
}

func handleError(strategy ErrorStrategy, changeID string, err error) (bool, error) {
	errorMsg := formatError(changeID, err)

	switch strategy {
	case ErrorContinue:
		fmt.Fprintf(os.Stderr, "%s\n", errorMsg)
		return false, nil
	case ErrorStop:
		fmt.Fprintf(os.Stderr, "Stopped on change [with fail] %s:\n%s\n", changeID, errorMsg)
		// Return error to stop processing and allow cleanup
		return true, fmt.Errorf("stopped on error at change %s", changeID)
	case ErrorFatal:
		fmt.Fprintf(os.Stderr, "Fatal error at change [%s]:\n%s\n", changeID, errorMsg)
		return true, fmt.Errorf("fatal error at change %s", changeID)
	}
	return false, nil
}

func formatError(changeID string, err error) string {
	exitErr := &exec.ExitError{}
	if errors.As(err, &exitErr) {
		return fmt.Sprintf("Error while processing change [%s]:\nReturn code: %d\nSTDERR:\n%s",
			changeID,
			exitErr.ExitCode(),
			string(exitErr.Stderr))
	}
	return fmt.Sprintf("Error while processing change [%s]:\n%v", changeID, err)
}

func printCommandResult(err error) {
	// For successful commands, the output is already printed by runShellCommand
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			fmt.Fprintf(os.Stderr, "Command failed with return code %d\n", exitErr.ExitCode())
		} else {
			fmt.Fprintf(os.Stderr, "Command failed: %v\n", err)
		}
	}
	fmt.Println()
}

func runShellCommand(command string, cwd string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func runJJ(args []string, cwd string) error {
	_, err := runJJOutput(args, cwd)
	return err
}

func runJJOutput(args []string, cwd string) (string, error) {
	// Add --ignore-immutable flag if requested
	if ignoreImmutable {
		args = append([]string{"--ignore-immutable"}, args...)
	}

	cmd := exec.Command("jj", args...)
	cmd.Dir = cwd

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			exitErr.Stderr = stderr.Bytes()
		}
		return "", err
	}

	return stdout.String(), nil
}

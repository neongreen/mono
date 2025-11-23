package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/neongreen/mono/aihook/pkg/validator"
	"github.com/neongreen/mono/lib/version"
)

var rootCmd = &cobra.Command{
	Use:   "aihook",
	Short: "Claude Code hook validator",
	Long:  `aihook validates shell commands and code patterns for Claude Code hooks.`,
}

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Shell hook that validates shell commands",
	Long:  `Shell hook that parses shell syntax and forbids 'cd' invocations outside subshells.`,
	RunE:  runShell,
}

var claudeFlag bool
var blockOnCdFlag bool

func init() {
	rootCmd.AddCommand(shellCmd)
	rootCmd.AddCommand(version.NewVersionCommand("aihook"))

	shellCmd.Flags().BoolVar(&claudeFlag, "claude", false, "Output in Claude Code hook format")
	shellCmd.Flags().BoolVar(&blockOnCdFlag, "block-on-cd", false, "Block when cd commands are found outside subshells")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runShell(cmd *cobra.Command, args []string) error {
	// Read and parse JSON input from stdin
	var input struct {
		Command     string `json:"command"`
		Timeout     int    `json:"timeout,omitempty"`
		Description string `json:"description,omitempty"`
	}

	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		return formatOutput(fmt.Sprintf("failed to parse JSON input: %v", err), 1)
	}

	// Validate the command string
	v := validator.New()
	commandReader := strings.NewReader(input.Command)
	violations, err := v.ValidateScript(commandReader)
	if err != nil {
		return formatOutput(err.Error(), 1)
	}

	if len(violations) > 0 {
		msg := validator.FormatViolations(violations)
		if blockOnCdFlag {
			return formatOutput(msg, 2)
		}
		return formatOutput(msg, 0)
	}

	return formatOutput("No violations found", 0)
}

// formatOutput formats the output according to the --claude flag
func formatOutput(message string, exitCode int) error {
	if claudeFlag {
		// Claude Code hook format
		output := map[string]interface{}{
			"message":   message,
			"exit_code": exitCode,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			return fmt.Errorf("failed to encode JSON output: %w", err)
		}
		if exitCode != 0 {
			os.Exit(exitCode)
		}
		return nil
	}

	// Regular output
	if exitCode == 0 {
		fmt.Println(message)
		return nil
	}

	fmt.Fprintln(os.Stderr, message)
	os.Exit(exitCode)
	return nil
}

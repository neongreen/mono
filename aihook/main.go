package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/neongreen/mono/aihook/pkg/validator"
	"github.com/neongreen/mono/lib/version"
)

var rootCmd = &cobra.Command{
	Use:   "aihook",
	Short: "Claude Code hook validator",
	Long:  `aihook validates shell commands and code patterns for Claude Code hooks.`,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop hook that validates shell commands",
	Long:  `Stop hook that parses shell syntax and forbids 'cd' invocations outside subshells.`,
	RunE:  runStop,
}

var claudeFlag bool

func init() {
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(version.NewVersionCommand("aihook"))

	stopCmd.Flags().BoolVar(&claudeFlag, "claude", false, "Output in Claude Code hook format")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runStop(cmd *cobra.Command, args []string) error {
	v := validator.New()
	violations, err := v.ValidateScript(os.Stdin)
	if err != nil {
		return formatOutput(err.Error(), 1)
	}

	if len(violations) > 0 {
		msg := validator.FormatViolations(violations)
		return formatOutput(msg, 2)
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

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"mvdan.cc/sh/v3/syntax"

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
	// Read shell script from stdin
	parser := syntax.NewParser()
	file, err := parser.Parse(os.Stdin, "")
	if err != nil {
		return formatOutput(fmt.Sprintf("Failed to parse shell script: %v", err), 1)
	}

	// Check for cd commands outside subshells
	violations := checkCdCommands(file)

	if len(violations) > 0 {
		msg := formatViolations(violations)
		return formatOutput(msg, 2)
	}

	return formatOutput("No violations found", 0)
}

// checkCdCommands traverses the AST and finds cd commands outside subshells
func checkCdCommands(file *syntax.File) []string {
	var violations []string
	inSubshell := false

	var walker func(syntax.Node) bool
	walker = func(node syntax.Node) bool {
		// Handle subshells and command substitution the same way
		if isSubshellNode(node) {
			oldInSubshell := inSubshell
			inSubshell = true
			syntax.Walk(node, func(innerNode syntax.Node) bool {
				return checkNodeForCd(innerNode, &violations, inSubshell)
			})
			inSubshell = oldInSubshell
			return false
		}
		return checkNodeForCd(node, &violations, inSubshell)
	}

	syntax.Walk(file, walker)
	return violations
}

// isSubshellNode checks if a node represents a subshell context
func isSubshellNode(node syntax.Node) bool {
	switch node.(type) {
	case *syntax.Subshell, *syntax.CmdSubst:
		return true
	}
	return false
}

// checkNodeForCd checks if a node is a cd command and records violations
func checkNodeForCd(node syntax.Node, violations *[]string, inSubshell bool) bool {
	if callExpr, ok := node.(*syntax.CallExpr); ok {
		// Check if this is a 'cd' command
		if len(callExpr.Args) > 0 {
			// Get the command name from the first word
			if word := callExpr.Args[0]; word != nil {
				cmdName := getWordText(word)
				if cmdName == "cd" && !inSubshell {
					// Found a cd command outside a subshell
					pos := callExpr.Pos()
					*violations = append(*violations, fmt.Sprintf("Line %d: 'cd' command found outside subshell", pos.Line()))
				}
			}
		}
	}
	return true
}

// getWordText extracts the text from a Word node
func getWordText(word *syntax.Word) string {
	if word == nil || len(word.Parts) == 0 {
		return ""
	}

	// For simple literals, extract the text
	if lit, ok := word.Parts[0].(*syntax.Lit); ok {
		return lit.Value
	}

	return ""
}

// formatViolations formats the list of violations into a message
func formatViolations(violations []string) string {
	msg := "Found cd commands outside subshells:\n"
	for _, v := range violations {
		msg += "  " + v + "\n"
	}
	msg += "\nAll 'cd' commands must be in a subshell. Example:\n"
	msg += "  # Bad:  cd /tmp && ls\n"
	msg += "  # Good: (cd /tmp && ls)\n"
	return msg
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

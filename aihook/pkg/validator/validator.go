package validator

import (
	"fmt"
	"io"

	"mvdan.cc/sh/v3/syntax"
)

// Validator performs shell script validation
type Validator struct{}

// New creates a new Validator instance
func New() *Validator {
	return &Validator{}
}

// ValidateScript parses and validates a shell script from the given reader
func (v *Validator) ValidateScript(r io.Reader) ([]string, error) {
	parser := syntax.NewParser()
	file, err := parser.Parse(r, "")
	if err != nil {
		return nil, fmt.Errorf("failed to parse shell script: %w", err)
	}

	violations := v.checkCdCommands(file)
	return violations, nil
}

// checkCdCommands traverses the AST and finds cd commands outside subshells
func (v *Validator) checkCdCommands(file *syntax.File) []string {
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

// FormatViolations formats the list of violations into a user-friendly message
func FormatViolations(violations []string) string {
	msg := "Found cd commands outside subshells:\n"
	for _, v := range violations {
		msg += "  " + v + "\n"
	}
	msg += "\nAll 'cd' commands must be in a subshell. Example:\n"
	msg += "  # Bad:  cd /tmp && ls\n"
	msg += "  # Good: (cd /tmp && ls)\n"
	return msg
}

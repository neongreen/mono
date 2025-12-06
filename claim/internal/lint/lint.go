// Package lint provides linting rules for claims.
package lint

import (
	"regexp"

	"github.com/neongreen/mono/claim/internal/parse"
)

// Issue represents a lint warning or error
type Issue struct {
	ClaimID  string
	Rule     string
	Message  string
	Severity string // "warning" or "error"
}

// Options configures which lint rules to run
type Options struct {
	// WarnLineNumbers warns when proofs reference line numbers (e.g., "line 42")
	// Line numbers are fragile and break when code changes.
	WarnLineNumbers bool
}

// DefaultOptions returns the default lint options (all warnings enabled)
func DefaultOptions() Options {
	return Options{
		WarnLineNumbers: true,
	}
}

var lineNumberRegex = regexp.MustCompile(`\bline\s+\d+\b`)

// LintClaim checks a claim for lint issues
func LintClaim(claim parse.Claim, opts Options) []Issue {
	var issues []Issue

	if opts.WarnLineNumbers {
		issues = append(issues, lintLineNumbers(claim)...)
	}

	return issues
}

// lintLineNumbers checks for references to line numbers in proof text
func lintLineNumbers(claim parse.Claim) []Issue {
	var issues []Issue

	if lineNumberRegex.MatchString(claim.Proof) {
		issues = append(issues, Issue{
			ClaimID:  claim.ID,
			Rule:     "no-line-numbers",
			Message:  "Avoid referencing line numbers in proofs - they break when code changes. Use function names, code snippets, or structural references instead.",
			Severity: "warning",
		})
	}

	return issues
}

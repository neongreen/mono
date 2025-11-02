package diff

import (
	"fmt"
	"strings"

	"github.com/neongreen/mono/lib/cli"
	"github.com/pmezard/go-difflib/difflib"
)

// DisplayUnifiedDiff displays a unified diff with color highlighting
// Returns true if there were changes, false otherwise
func DisplayUnifiedDiff(before, after, filename string) bool {
	// Use go-difflib to generate unified diff
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(before),
		B:        difflib.SplitLines(after),
		FromFile: filename + " (before)",
		ToFile:   filename + " (after)",
		Context:  3,
	}

	diffText, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		// If there's an error generating the diff, fall back to simple comparison
		if before == after {
			return false
		}
		fmt.Printf("Error generating diff: %v\n", err)
		return true
	}

	// If no diff output, there were no changes
	if diffText == "" {
		return false
	}

	// Display the diff with colors
	displayColoredDiff(diffText)

	return true
}

// displayColoredDiff displays a unified diff string with color highlighting
func displayColoredDiff(diffText string) {

	lines := strings.Split(diffText, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++"):
			// File headers
			fmt.Println()
			fmt.Println(line)
		case strings.HasPrefix(line, "@@"):
			// Hunk headers (skip these for cleaner output)
			continue
		case strings.HasPrefix(line, "-"):
			// Deletions
			fmt.Println(cli.Error(line))
		case strings.HasPrefix(line, "+"):
			// Additions
			fmt.Println(cli.Success(line))
		case strings.HasPrefix(line, " "):
			// Context lines
			fmt.Println(line)
		default:
			// Anything else (shouldn't happen in unified diff)
			fmt.Println(line)
		}
	}
	fmt.Println()
}

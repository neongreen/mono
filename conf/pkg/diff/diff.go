package diff

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// DisplayDiff computes and displays a colored diff between before and after content
// Returns true if there were changes, false otherwise
func DisplayDiff(before, after string) bool {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(before, after, false)

	// Check if there are any actual changes
	hasChanges := false
	for _, diff := range diffs {
		if diff.Type != diffmatchpatch.DiffEqual {
			hasChanges = true
			break
		}
	}

	if !hasChanges {
		return false
	}

	// Display the diff with colors
	displayColoredDiff(diffs)

	return true
}

// displayColoredDiff displays diffs with color highlighting
func displayColoredDiff(diffs []diffmatchpatch.Diff) {
	red := color.New(color.FgRed)
	green := color.New(color.FgGreen)

	for _, diff := range diffs {
		text := diff.Text

		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			// Green for additions
			for _, line := range strings.Split(text, "\n") {
				if line != "" {
					green.Printf("+ %s\n", line)
				}
			}
		case diffmatchpatch.DiffDelete:
			// Red for deletions
			for _, line := range strings.Split(text, "\n") {
				if line != "" {
					red.Printf("- %s\n", line)
				}
			}
		case diffmatchpatch.DiffEqual:
			// No color for unchanged lines
			// Only show a few context lines around changes
			lines := strings.Split(text, "\n")
			for _, line := range lines {
				if line != "" {
					fmt.Printf("  %s\n", line)
				}
			}
		}
	}
}

// DisplayUnifiedDiff displays a more compact unified diff format
func DisplayUnifiedDiff(before, after, filename string) bool {
	dmp := diffmatchpatch.New()

	// Split into lines for unified diff
	beforeLines := strings.Split(before, "\n")
	afterLines := strings.Split(after, "\n")

	// Create diffs at line level
	diffs := dmp.DiffMain(before, after, false)

	// Check if there are any actual changes
	hasChanges := false
	for _, diff := range diffs {
		if diff.Type != diffmatchpatch.DiffEqual {
			hasChanges = true
			break
		}
	}

	if !hasChanges {
		return false
	}

	// Display unified diff header
	fmt.Printf("\n--- %s (before)\n", filename)
	fmt.Printf("+++ %s (after)\n", filename)

	// Display the colored diff
	displayColoredUnifiedDiff(beforeLines, afterLines)
	fmt.Println()

	return true
}

// displayColoredUnifiedDiff displays a unified diff with line-by-line coloring
func displayColoredUnifiedDiff(beforeLines, afterLines []string) {
	red := color.New(color.FgRed)
	green := color.New(color.FgGreen)
	cyan := color.New(color.FgCyan)

	// Simple line-by-line diff for clarity
	maxLen := len(beforeLines)
	if len(afterLines) > maxLen {
		maxLen = len(afterLines)
	}

	i, j := 0, 0
	contextLines := 3
	lastDiffLine := -10

	for i < len(beforeLines) || j < len(afterLines) {
		// Determine if we're in a diff region
		inDiffRegion := false

		if i < len(beforeLines) && j < len(afterLines) {
			if beforeLines[i] == afterLines[j] {
				// Lines match
				// Show context around diffs
				if i-lastDiffLine <= contextLines {
					// Only show non-empty context lines
					if beforeLines[i] != "" {
						fmt.Printf("  %s\n", beforeLines[i])
					}
				} else if i-lastDiffLine == contextLines+1 {
					cyan.Println("  ...")
				}
				i++
				j++
				continue
			} else {
				inDiffRegion = true
				lastDiffLine = i
			}
		} else {
			inDiffRegion = true
			lastDiffLine = i
		}

		if inDiffRegion {
			// Lines differ or one side is shorter
			if i < len(beforeLines) && (j >= len(afterLines) || beforeLines[i] != afterLines[j]) {
				red.Printf("- %s\n", beforeLines[i])
				i++
			}
			if j < len(afterLines) && (i >= len(beforeLines) || (i > 0 && beforeLines[i-1] != afterLines[j])) {
				green.Printf("+ %s\n", afterLines[j])
				j++
			}
		}
	}
}

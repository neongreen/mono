package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/neongreen/mono/claim/internal/index"
	"github.com/neongreen/mono/claim/internal/parse"
	"github.com/neongreen/mono/claim/internal/scan"
)

var seeRefPattern = regexp.MustCompile(`@see\[([^\]]+)\]`)

var (
	showRoot string
)

var showCmd = &cobra.Command{
	Use:   "show <claim-id>",
	Short: "Display a claim and all its dependencies",
	Long: `Shows a claim's statement, context, proof, and recursively displays
all claims it depends on via @see references.

This is a read-only operation that doesn't verify anything.`,
	Args: cobra.ExactArgs(1),
	RunE: runShow,
}

func init() {
	showCmd.Flags().StringVar(&showRoot, "root", ".", "Root directory to scan")
	RootCmd.AddCommand(showCmd)
}

func runShow(cmd *cobra.Command, args []string) error {
	claimID := args[0]

	// Scan files
	files, err := scan.ScanFiles(showRoot)
	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	// Build index
	idx, err := index.Build(files)
	if err != nil {
		return fmt.Errorf("failed to build index: %w", err)
	}

	// Track visited claims to avoid cycles
	visited := make(map[string]bool)

	// Show the claim and its dependencies
	return showClaimRecursive(idx, claimID, visited, 0)
}

func showClaimRecursive(idx *index.Index, claimID string, visited map[string]bool, depth int) error {
	if visited[claimID] {
		return nil
	}
	visited[claimID] = true

	claim, ok := idx.GetClaim(claimID)
	if !ok {
		fmt.Fprintf(os.Stderr, "Warning: claim %q not found\n", claimID)
		return nil
	}

	// Print separator for non-root claims
	if depth > 0 {
		fmt.Println()
		fmt.Println(strings.Repeat("─", 70))
	}

	// Header
	header := color.New(color.Bold, color.FgBlue)
	muted := color.New(color.FgHiBlack)
	label := color.New(color.FgYellow)

	header.Printf("@claim[%s]", claimID)
	muted.Printf(" (%s:%d)\n", claim.File, claim.Line)
	fmt.Println(claim.Statement)
	fmt.Println()

	// Context (if present)
	if claim.Context != "" {
		label.Println("@context:")
		printIndented(claim.Context)
		fmt.Println()
	}

	// Proof (if present)
	if claim.Proof != "" {
		label.Println("@proof:")
		printIndented(claim.Proof)
		fmt.Println()
	}

	// Show dependencies
	if len(claim.SeeRefs) > 0 {
		label.Printf("Dependencies: ")
		fmt.Println(strings.Join(claim.SeeRefs, ", "))
	}

	// Recursively show dependencies
	for _, ref := range claim.SeeRefs {
		if err := showClaimRecursive(idx, ref, visited, depth+1); err != nil {
			return err
		}
	}

	return nil
}

func printIndented(text string) {
	dedented := dedentText(text)
	lines := strings.Split(dedented, "\n")
	for _, line := range lines {
		fmt.Printf("  %s\n", highlightSeeRefs(line))
	}
}

// dedentText removes common leading whitespace from all lines
func dedentText(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return text
	}

	// Find minimum indentation (ignoring empty lines)
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return text
	}

	// Remove common indentation
	var result []string
	for _, line := range lines {
		if len(line) >= minIndent {
			result = append(result, line[minIndent:])
		} else {
			result = append(result, strings.TrimLeft(line, " \t"))
		}
	}

	return strings.Join(result, "\n")
}

func highlightSeeRefs(text string) string {
	seeStyle := color.New(color.FgBlue).SprintFunc()
	return seeRefPattern.ReplaceAllStringFunc(text, func(match string) string {
		return seeStyle(match)
	})
}

// Ensure parse is used (for SeeRefs)
var _ = parse.Claim{}

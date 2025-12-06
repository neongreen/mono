package check

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	claimcontext "github.com/neongreen/mono/claim/internal/context"
	"github.com/neongreen/mono/claim/internal/index"
	"github.com/neongreen/mono/claim/internal/logger"
	"github.com/neongreen/mono/claim/internal/prompt"
	"github.com/neongreen/mono/claim/internal/runner"
)

// Helper functions for file I/O
func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// renderMarkdown renders markdown text for terminal output
// Returns content with leading/trailing blank lines removed - caller controls spacing
func renderMarkdown(text string) (string, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return "", err
	}
	rendered, err := r.Render(text)
	if err != nil {
		return "", err
	}
	return trimBlankLines(rendered), nil
}

// trimBlankLines removes leading and trailing blank lines (whitespace-only lines)
// Also handles lines that contain only ANSI escape codes
func trimBlankLines(text string) string {
	lines := strings.Split(text, "\n")

	// Find first non-blank line
	start := 0
	for start < len(lines) && isBlankLine(lines[start]) {
		start++
	}

	// Find last non-blank line
	end := len(lines) - 1
	for end >= start && isBlankLine(lines[end]) {
		end--
	}

	if start > end {
		return ""
	}

	return strings.Join(lines[start:end+1], "\n")
}

// isBlankLine returns true if line is empty after removing ANSI codes and whitespace
func isBlankLine(line string) bool {
	// Strip ANSI escape codes: ESC [ ... m
	stripped := stripANSI(line)
	return strings.TrimSpace(stripped) == ""
}

// stripANSI removes ANSI escape sequences from text
func stripANSI(text string) string {
	var result strings.Builder
	i := 0
	for i < len(text) {
		if text[i] == 0x1b && i+1 < len(text) && text[i+1] == '[' {
			// Skip until 'm' (end of color code)
			j := i + 2
			for j < len(text) && text[j] != 'm' {
				j++
			}
			i = j + 1
		} else {
			result.WriteByte(text[i])
			i++
		}
	}
	return result.String()
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

// ProofCheckResult contains the result of checking a proof using the new syntax
type ProofCheckResult struct {
	ClaimID                string
	Result                 string         // "proven", "unproven"
	Issues                 []runner.Issue // Problems found
	InterpretationNarrowed bool           // True if context changed claim meaning
	Stats                  *runner.Stats  // Execution stats (timing, cost, tokens)
}

// CheckProofOptions contains options for CheckProof
type CheckProofOptions struct {
	DebugPrompt     bool
	ProgressWriter  io.Writer
	ContextResolver *claimcontext.Resolver // Optional resolver for @context blocks
	ArtifactsDir    string                 // Directory to save prompts, outputs, etc.
}

// CheckProof verifies a claim using the new @proof syntax
// It evaluates the whole proof at once instead of incrementally
func CheckProof(
	ctx context.Context,
	idx *index.Index,
	lenses map[string]string,
	claimID string,
	r runner.Runner,
	opts CheckProofOptions,
) (*ProofCheckResult, error) {
	logger.Debug("checking proof", "id", claimID)

	// Get the claim
	claim, ok := idx.GetClaim(claimID)
	if !ok {
		return nil, fmt.Errorf("claim %q not found", claimID)
	}

	// Check that the claim has a proof
	if claim.Proof == "" {
		return nil, fmt.Errorf("claim %q has no @proof block (use old format or add @proof[%s]:)", claimID, claimID)
	}

	// Collect axioms from @see references
	axioms := make(map[string]string)
	for _, ref := range claim.SeeRefs {
		refClaim, ok := idx.GetClaim(ref)
		if !ok {
			return nil, fmt.Errorf("@see[%s] references unknown claim", ref)
		}
		axioms[ref] = refClaim.Statement
	}

	logger.Debug("collected axioms", "claim", claimID, "axioms", len(axioms))

	// Resolve context if @context block is present
	var resolvedContext string
	if claim.Context != "" && opts.ContextResolver != nil {
		if opts.ProgressWriter != nil {
			fmt.Fprintf(opts.ProgressWriter, "Resolving @context[%s]...\n", claimID)
		}
		resolved, err := opts.ContextResolver.Resolve(ctx, claim.Context)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve @context: %w", err)
		}
		if resolved != nil {
			resolvedContext = resolved.Content
		}
	}

	// Define styles for progress output
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	claimStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14"))

	// Print what we're checking
	if opts.ProgressWriter != nil {
		axiomIDStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
		axiomStmtStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

		fmt.Fprintf(opts.ProgressWriter, "\n%s\n", strings.Repeat("─", 70))
		fmt.Fprintf(opts.ProgressWriter, "%s:\n", claimStyle.Render("@claim["+claimID+"]"))
		stmtText := dedentText(claim.Statement)
		if rendered, err := renderMarkdown(stmtText); err == nil {
			fmt.Fprintf(opts.ProgressWriter, "%s\n\n", rendered)
		} else {
			fmt.Fprintf(opts.ProgressWriter, "%s\n\n", stmtText)
		}

		// Show context specification if present
		if claim.Context != "" {
			fmt.Fprintf(opts.ProgressWriter, "%s\n", labelStyle.Render("Context requested:"))
			contextText := dedentText(claim.Context)
			if rendered, err := renderMarkdown(contextText); err == nil {
				fmt.Fprintf(opts.ProgressWriter, "%s\n\n", rendered)
			} else {
				fmt.Fprintf(opts.ProgressWriter, "%s\n\n", contextText)
			}
		}

		// Show the proof body
		if claim.Proof != "" {
			fmt.Fprintf(opts.ProgressWriter, "%s\n", labelStyle.Render("Proof:"))
			proofText := dedentText(claim.Proof)
			logger.Debug("proof text after dedent", "text", proofText)
			if rendered, err := renderMarkdown(proofText); err == nil {
				fmt.Fprintf(opts.ProgressWriter, "%s\n\n", rendered)
			} else {
				fmt.Fprintf(opts.ProgressWriter, "%s\n\n", proofText)
			}
		}

		if len(axioms) > 0 {
			fmt.Fprintf(opts.ProgressWriter, "%s\n", labelStyle.Render("Axioms (assumed true):"))
			for id, stmt := range axioms {
				fmt.Fprintf(opts.ProgressWriter, "  %s: %s\n",
					axiomIDStyle.Render("@claim["+id+"]"),
					axiomStmtStyle.Render(stmt))
			}
			fmt.Fprintf(opts.ProgressWriter, "\n")
		}
	}

	// Build the prompt
	promptText := prompt.BuildProofPrompt(claim, axioms, lenses, resolvedContext)

	// Save prompt to artifacts dir
	if opts.ArtifactsDir != "" {
		// Sanitize claimID for use as filename (replace / with _)
		safeClaimID := strings.ReplaceAll(claimID, "/", "_")
		promptPath := fmt.Sprintf("%s/%s-prompt.md", opts.ArtifactsDir, safeClaimID)
		if err := writeFile(promptPath, []byte(promptText)); err != nil {
			logger.Debug("failed to save prompt", "error", err)
		} else if opts.ProgressWriter != nil {
			fmt.Fprintf(opts.ProgressWriter, "%s %s\n",
				labelStyle.Render("Prompt saved:"),
				promptPath)
		}
	}

	if opts.DebugPrompt {
		fmt.Println("=== PROOF PROMPT ===")
		fmt.Println(promptText)
		fmt.Println("=== END PROOF PROMPT ===")
		fmt.Println()
	}

	// Run Claude
	proofResult, err := r.RunProof(ctx, promptText)
	if err != nil {
		return nil, fmt.Errorf("failed to check proof: %w", err)
	}

	// Save Claude output to artifacts dir
	if opts.ArtifactsDir != "" {
		// Copy the temp output file to artifacts
		safeClaimID := strings.ReplaceAll(claimID, "/", "_")
		outputPath := fmt.Sprintf("%s/%s-claude-output.jsonl", opts.ArtifactsDir, safeClaimID)
		if data, readErr := readFile("/tmp/claim-claude-output.jsonl"); readErr == nil {
			if writeErr := writeFile(outputPath, data); writeErr != nil {
				logger.Debug("failed to save claude output", "error", writeErr)
			} else if opts.ProgressWriter != nil {
				fmt.Fprintf(opts.ProgressWriter, "%s %s\n",
					labelStyle.Render("Claude output:"),
					outputPath)
			}
		}
	}

	// Build result
	result := &ProofCheckResult{
		ClaimID:                claimID,
		Result:                 proofResult.Result,
		Issues:                 proofResult.Issues,
		InterpretationNarrowed: proofResult.InterpretationNarrowed,
		Stats:                  proofResult.Stats,
	}

	// If interpretation was narrowed, the claim is ambiguous - mark as unproven
	if result.InterpretationNarrowed && result.Result == "proven" {
		result.Result = "unproven"
		result.Issues = append(result.Issues, runner.Issue{
			Title:       "Ambiguous claim",
			Description: "The claim's meaning depends on context; it should be unambiguous on its own.",
		})
	}

	// Print result
	if opts.ProgressWriter != nil {
		PrintProofResult(opts.ProgressWriter, result)
	}

	return result, nil
}

// PrintProofResult prints the result of a proof check
func PrintProofResult(w io.Writer, result *ProofCheckResult) {
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	statusOK := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	statusBad := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	fmt.Fprintf(w, "\n%s\n", labelStyle.Render("═══════════════════════════════════════════════════════════"))

	if result.InterpretationNarrowed {
		fmt.Fprintf(w, "%s %s\n", labelStyle.Render("Warning:"), warnStyle.Render("Context narrows claim interpretation"))
	}

	if result.Result == "proven" {
		fmt.Fprintf(w, "%s %s\n", labelStyle.Render("Result:"), statusOK.Render("PROVEN"))
	} else {
		fmt.Fprintf(w, "%s %s\n\n", labelStyle.Render("Result:"), statusBad.Render("UNPROVEN"))
		if len(result.Issues) > 0 {
			fmt.Fprintf(w, "%s\n", labelStyle.Render("Issues:"))
			renderIssues(w, result.Issues)
		}
	}

	// Print stats if available
	if result.Stats != nil {
		fmt.Fprintf(w, "\n%s\n", labelStyle.Render("───────────────────────────────────────────────────────────"))
		s := result.Stats
		durationSec := float64(s.DurationMs) / 1000.0
		fmt.Fprintf(w, "%s %s\n", labelStyle.Render("Stats:"),
			statsStyle.Render(fmt.Sprintf("%.1fs, $%.4f, %d turns", durationSec, s.TotalCostUSD, s.NumTurns)))
		fmt.Fprintf(w, "%s %s\n", labelStyle.Render("Tokens:"),
			statsStyle.Render(fmt.Sprintf("in=%d out=%d cache_read=%d cache_create=%d",
				s.InputTokens, s.OutputTokens, s.CacheRead, s.CacheCreation)))
	}
}

// renderIssues prints a list of issues with formatting
func renderIssues(w io.Writer, issues []runner.Issue) {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	for _, issue := range issues {
		fmt.Fprintf(w, "  • %s\n", titleStyle.Render(issue.Title))
		fmt.Fprintf(w, "    %s\n\n", descStyle.Render(issue.Description))
	}
}

// HasProof returns true if the claim uses the new @proof syntax
func HasProof(idx *index.Index, claimID string) bool {
	claim, ok := idx.GetClaim(claimID)
	if !ok {
		return false
	}
	return claim.Proof != ""
}

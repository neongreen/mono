package check

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/neongreen/mono/claim/internal/index"
	"github.com/neongreen/mono/claim/internal/logger"
	"github.com/neongreen/mono/claim/internal/parse"
	"github.com/neongreen/mono/claim/internal/prompt"
	"github.com/neongreen/mono/claim/internal/runner"
)

// Check runs the claim checker for a specific claim ID
func Check(
	ctx context.Context,
	idx *index.Index,
	lenses map[string]string,
	claimID string,
	maxRefDepth int,
	debugPrompt bool,
	anonymizeFiles bool,
	r runner.Runner,
) (*runner.ClaimResult, error) {
	logger.Debug("checking claim", "id", claimID, "max_ref_depth", maxRefDepth)

	// Get target claim
	claim, ok := idx.GetClaim(claimID)
	if !ok {
		return nil, fmt.Errorf("claim %q not found", claimID)
	}

	// Expand referenced claims
	referencedClaims, err := idx.ExpandReferences(claimID, maxRefDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to expand references: %w", err)
	}

	logger.Debug("expanded references", "claim", claimID, "refs", len(referencedClaims))

	// Get all claim IDs for the prompt
	var allClaimIDs []string
	for id := range idx.Claims {
		allClaimIDs = append(allClaimIDs, id)
	}

	// Build prompt
	promptText := prompt.Build(claim, referencedClaims, lenses, allClaimIDs, anonymizeFiles)

	logger.Debug("built prompt", "length", len(promptText), "lenses", len(lenses))

	if debugPrompt {
		fmt.Println("=== PROMPT ===")
		fmt.Println(promptText)
		fmt.Println("=== END PROMPT ===")
		fmt.Println()
	}

	// Run Claude
	result, err := r.Run(ctx, promptText)
	if err != nil {
		return nil, fmt.Errorf("failed to run Claude: %w", err)
	}

	logger.Debug("enforcing post-rules", "claim", claimID)

	// Enforce post-rules
	if err := enforcePostRules(claim, result); err != nil {
		return nil, fmt.Errorf("post-rule violation: %w", err)
	}

	logger.Debug("check complete", "claim", claimID, "result", result.Result)

	return result, nil
}

// enforcePostRules validates that Claude's response follows all requirements
func enforcePostRules(claim parse.Claim, result *runner.ClaimResult) error {
	// Collect all bullet paths
	bulletPaths := make(map[string]parse.Bullet)
	var collectPaths func([]parse.Bullet)
	collectPaths = func(bullets []parse.Bullet) {
		for _, b := range bullets {
			bulletPaths[b.Path] = b
			collectPaths(b.Children)
		}
	}
	collectPaths(claim.Bullets)

	// Check that every bullet has a verdict
	verdictPaths := make(map[string]bool)
	for i := range result.Bullets {
		v := &result.Bullets[i]
		// Normalize path - strip common LLM hallucination patterns
		// Examples: "file.go:0" -> "0", "[0]" -> "0", "path 0" -> "0"
		path := v.Path

		// Strip file prefix (e.g., "file.go:0" -> "0")
		if idx := strings.LastIndex(path, ":"); idx >= 0 {
			path = path[idx+1:]
		}

		// Strip square brackets (e.g., "[0]" -> "0")
		path = strings.Trim(path, "[]")

		// Strip "path" prefix (e.g., "path 0" -> "0")
		path = strings.TrimPrefix(path, "path ")
		path = strings.TrimSpace(path)

		// Update the verdict path to normalized form
		v.Path = path
		verdictPaths[path] = true
	}

	for path := range bulletPaths {
		if !verdictPaths[path] {
			return fmt.Errorf("missing verdict for bullet path %q", path)
		}
	}

	// Check that bullets with @claim[ref] have those refs in required_claims
	for _, verdict := range result.Bullets {
		bullet, ok := bulletPaths[verdict.Path]
		if !ok {
			// Verdict for non-existent bullet - warn but don't fail
			continue
		}

		// If bullet has references, they must be in required_claims
		for _, ref := range bullet.References {
			hasRef := false
			for _, req := range verdict.RequiredClaims {
				if req == ref {
					hasRef = true
					break
				}
			}
			if !hasRef {
				// Force unproven if reference not required
				result.Result = "unproven"
				if result.Counterexample == "" {
					result.Counterexample = fmt.Sprintf("Bullet %s references @claim[%s] but doesn't require it", verdict.Path, ref)
				}
			}
		}
	}

	// Check @sorry bullets
	hasSorry := false
	for path, bullet := range bulletPaths {
		if bullet.IsSorry {
			hasSorry = true
			// Verify verdict status is "sorry"
			for i, v := range result.Bullets {
				if v.Path == path && v.Status != "sorry" {
					result.Bullets[i].Status = "sorry"
				}
			}
		}
	}

	// If any bullet is @sorry, result must not be proven
	if hasSorry && result.Result == "proven" {
		result.Result = "sorry"
	}

	return nil
}

// PrintReport prints a human-readable report of the check result
func PrintReport(w io.Writer, result *runner.ClaimResult) {
	// Define styles
	claimIDStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")) // Cyan
	resultStyle := getResultStyle(result.Result)
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
	statusStyle := func(status string) lipgloss.Style { return getStatusStyle(status) }
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))                  // Muted
	suggestionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Italic(true) // Yellow

	// Print claim and result
	fmt.Fprintf(w, "%s %s\n", labelStyle.Render("Claim:"), claimIDStyle.Render(result.ClaimID))
	fmt.Fprintf(w, "%s %s\n\n", labelStyle.Render("Result:"), resultStyle.Render(result.Result))

	if len(result.Bullets) > 0 {
		fmt.Fprintf(w, "%s\n", labelStyle.Render("Bullet Verdicts:"))
		for _, v := range result.Bullets {
			fmt.Fprintf(w, "  %s %s", pathStyle.Render("["+v.Path+"]"), statusStyle(v.Status).Render(v.Status))
			if len(v.RequiredClaims) > 0 {
				fmt.Fprintf(w, " %s", labelStyle.Render("requires="+strings.Join(v.RequiredClaims, ",")))
			}
			fmt.Fprintf(w, "\n")
			if len(v.SuggestedRewrite) > 0 {
				for _, suggestion := range v.SuggestedRewrite {
					fmt.Fprintf(w, "    %s %s\n", suggestionStyle.Render("→"), suggestion)
				}
			}
		}
		fmt.Fprintf(w, "\n")
	}

	if result.Counterexample != "" {
		counterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Red
		fmt.Fprintf(w, "%s\n%s\n", labelStyle.Render("Counterexample:"), counterStyle.Render(result.Counterexample))
	}
}

// getResultStyle returns the appropriate style for a result
func getResultStyle(result string) lipgloss.Style {
	switch result {
	case "proven":
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")) // Green
	case "unproven":
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")) // Yellow
	case "sorry":
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9")) // Red
	default:
		return lipgloss.NewStyle().Bold(true)
	}
}

// getStatusStyle returns the appropriate style for a bullet status
func getStatusStyle(status string) lipgloss.Style {
	switch status {
	case "ok":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green
	case "trivial":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // Blue
	case "needs_split", "needs_claim":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
	case "contradicts":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Red
	case "sorry":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("13")) // Magenta
	default:
		return lipgloss.NewStyle()
	}
}

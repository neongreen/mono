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
//
// @claim[every-bullet-gets-verdict]: Every bullet in the claim receives exactly one verdict
// - All bullet paths are collected including nested bullets (lines 82-91)
//   - collectPaths recursively walks the bullet tree
//   - Each bullet is stored in bulletPaths map by its path
// - LLM verdict paths are normalized to handle hallucinations (lines 94-116)
//   - File prefixes are stripped: "file.go:0" becomes "0"
//   - Square brackets are stripped: "[0]" becomes "0"
//   - "path" prefix is stripped: "path 0" becomes "0"
//   - The normalized path is stored back in v.Path
// - Every bullet path is checked for a corresponding verdict (lines 118-121)
//   - If any bullet path is missing from verdictPaths, an error is returned
//   - This ensures the LLM provided verdicts for all bullets
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
	//
	// @claim[references-must-be-required]: Bullets with @claim[ref] must list that ref in required_claims
	// - Each verdict is checked against its corresponding bullet (lines 138-162)
	// - For each claim reference in bullet.References (line 146)
	//   - We check if that reference appears in verdict.RequiredClaims (lines 147-153)
	//   - If the reference is missing, the result is forced to "unproven" (line 156)
	//   - A counterexample is added explaining which reference was missing (lines 157-159)
	// - This prevents the LLM from marking a bullet as proven without verifying its dependencies
	//   - If a bullet says "see @claim[foo]", the verdict must require claim foo
	//   - Otherwise we can't trust that the bullet's assertion is actually supported
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
	//
	// @claim[sorry-prevents-proven]: A claim with @sorry bullets cannot have result "proven"
	// - All bullets are checked for the IsSorry flag (lines 166-176)
	// - If any bullet has IsSorry=true, hasSorry is set to true
	// - The verdict status is corrected to "sorry" if the LLM got it wrong (lines 170-174)
	//   - This handles cases where LLM marks @sorry bullets with other statuses
	// - If hasSorry is true and result is "proven", result is changed to "sorry" (lines 179-181)
	//   - This enforces the rule that @sorry means "accepted without proof"
	//   - A claim cannot be fully proven if it contains unproven bullets
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
//
// @claim[output-shows-original-bullets]: The output shows the user's original bullet text for each verdict
// - PrintReport receives the original claim with bullets (parameter: claim parse.Claim)
// - A map is built from bullet path to bullet text (lines 200-209)
//   - collectBullets recursively walks all bullets including nested ones
//   - Each bullet's path and text are stored: bulletTexts[b.Path] = b.Text
// - For each verdict, the original bullet text is printed (lines 217-218)
//   - bulletText is retrieved from bulletTexts[v.Path]
//   - It's displayed with label "Your text:" followed by the bullet content
// - The bullet text appears before any issue explanations or suggestions
//   - Users see what they wrote before seeing what's wrong with it
func PrintReport(w io.Writer, claim parse.Claim, result *runner.ClaimResult) {
	// Define styles
	claimIDStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")) // Cyan
	resultStyle := getResultStyle(result.Result)
	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))                     // Gray
	statusStyle := func(status string) lipgloss.Style { return getStatusStyle(status) }
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))                  // Muted
	suggestionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Italic(true) // Yellow
	bulletTextStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))              // Bright white

	// Print claim header with statement
	fmt.Fprintf(w, "%s %s\n", labelStyle.Render("Claim:"), claimIDStyle.Render(result.ClaimID))
	fmt.Fprintf(w, "%s %s\n", labelStyle.Render("Statement:"), claim.Statement)
	fmt.Fprintf(w, "%s %s:%d\n", labelStyle.Render("Location:"), claim.File, claim.Line)
	
	// Print result with explanation
	resultExplanation := ""
	switch result.Result {
	case "proven":
		resultExplanation = " (all bullets are acceptable proof)"
	case "unproven":
		resultExplanation = " (expected - this tool catches false positives)"
	case "sorry":
		resultExplanation = " (contains @sorry bullets)"
	}
	fmt.Fprintf(w, "%s %s%s\n\n", labelStyle.Render("Result:"), resultStyle.Render(result.Result), labelStyle.Render(resultExplanation))

	// Build a map of bullet path -> bullet text
	bulletTexts := make(map[string]string)
	var collectBullets func(bullets []parse.Bullet)
	collectBullets = func(bullets []parse.Bullet) {
		for _, b := range bullets {
			bulletTexts[b.Path] = b.Text
			collectBullets(b.Children)
		}
	}
	collectBullets(claim.Bullets)

	if len(result.Bullets) > 0 {
		fmt.Fprintf(w, "%s\n", labelStyle.Render("Bullet Analysis:"))
		fmt.Fprintf(w, "%s\n\n", labelStyle.Render("Each bullet from your claim was checked. Fix the issues below:"))
		
		for _, v := range result.Bullets {
			// Get the original bullet text
			bulletText := bulletTexts[v.Path]
			if bulletText == "" {
				bulletText = "(bullet text not found)"
			}
			
			// Status badge with original bullet text
			statusBadge := statusStyle(v.Status).Render(strings.ToUpper(v.Status))
			fmt.Fprintf(w, "  %s %s\n", pathStyle.Render("Bullet ["+v.Path+"]:"), statusBadge)
			fmt.Fprintf(w, "    %s %s\n", labelStyle.Render("Your text:"), bulletTextStyle.Render(bulletText))
			
			// Explanation with clear formatting
			explanation := getStatusExplanation(v.Status)
			if explanation != "" {
				fmt.Fprintf(w, "    %s %s\n", labelStyle.Render("Issue:"), explanation)
			}
			
			// Required claims with clear context
			if len(v.RequiredClaims) > 0 {
				fmt.Fprintf(w, "    %s This bullet depends on these other claims being proven:\n", 
					labelStyle.Render("Dependencies:"))
				for _, claim := range v.RequiredClaims {
					fmt.Fprintf(w, "                 @claim[%s]\n", claim)
				}
			}
			
			// Suggested rewrites with ultra-clear labeling
			//
			// @claim[multiple-suggestions-are-alternatives]: When multiple suggestions are shown, users are told to choose one
			// - The number of suggestions is checked (line 287)
			// - If there's exactly one suggestion, label is "How to fix:" (line 288)
			// - If there are multiple suggestions, label is "How to fix (choose ONE approach):" (line 290)
			//   - The "(choose ONE approach)" text makes it explicit that suggestions are alternatives
			//   - Users don't have to guess whether they should do all suggestions or pick one
			// - Each suggestion is numbered (1, 2, 3...) to reinforce they are separate options (lines 293-315)
			if len(v.SuggestedRewrite) > 0 {
				if len(v.SuggestedRewrite) == 1 {
					fmt.Fprintf(w, "    %s\n", labelStyle.Render("How to fix:"))
				} else {
					fmt.Fprintf(w, "    %s\n", labelStyle.Render("How to fix (choose ONE approach):"))
				}
				
				for i, suggestion := range v.SuggestedRewrite {
					// Check if suggestion is asking for a new claim
					if strings.Contains(suggestion, "@claim[") {
						// Extract the claim ID and statement
						parts := strings.SplitN(suggestion, "@claim[", 2)
						if len(parts) == 2 {
							claimPart := strings.SplitN(parts[1], "]", 2)
							if len(claimPart) == 2 {
								claimID := claimPart[0]
								statement := strings.TrimSpace(strings.TrimPrefix(claimPart[1], ":"))
								fmt.Fprintf(w, "      %d. Extract this to a new claim and prove it:\n", i+1)
								fmt.Fprintf(w, "         @claim[%s]: %s\n", suggestionStyle.Render(claimID), statement)
								continue
							}
						}
						// Fallback if parsing fails
						fmt.Fprintf(w, "      %d. Extract to a new claim:\n", i+1)
						fmt.Fprintf(w, "         %s\n", suggestionStyle.Render(suggestion))
					} else {
						fmt.Fprintf(w, "      %d. Replace your bullet with:\n", i+1)
						fmt.Fprintf(w, "         - %s\n", suggestionStyle.Render(suggestion))
					}
				}
			}
			
			fmt.Fprintf(w, "\n")
		}
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

// getStatusExplanation returns a helpful explanation for bullet statuses
func getStatusExplanation(status string) string {
	switch status {
	case "ok":
		return ""
	case "trivial":
		return "this is obvious and doesn't need to be stated"
	case "needs_split":
		return "this bullet is too vague or hand-wavy"
	case "needs_claim":
		return "this makes an unsupported assertion - extract it to a separate @claim and prove it"
	case "contradicts":
		return "this contradicts the claim statement or other bullets"
	case "sorry":
		return "explicitly marked as accepted without proof (@sorry)"
	default:
		return ""
	}
}

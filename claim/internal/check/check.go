package check

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/neongreen/mono/claim/internal/index"
	"github.com/neongreen/mono/claim/internal/logger"
	"github.com/neongreen/mono/claim/internal/parse"
	"github.com/neongreen/mono/claim/internal/prompt"
	"github.com/neongreen/mono/claim/internal/runner"
)

// TODO: Add lipgloss for colorful output

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
	for _, v := range result.Bullets {
		verdictPaths[v.Path] = true
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
	fmt.Fprintf(w, "Claim: %s\n", result.ClaimID)
	fmt.Fprintf(w, "Result: %s\n\n", result.Result)

	if len(result.Bullets) > 0 {
		fmt.Fprintf(w, "Bullet Verdicts:\n")
		for _, v := range result.Bullets {
			fmt.Fprintf(w, "  [%s] status=%s", v.Path, v.Status)
			if len(v.RequiredClaims) > 0 {
				fmt.Fprintf(w, " requires=%s", strings.Join(v.RequiredClaims, ","))
			}
			fmt.Fprintf(w, "\n")
			if len(v.SuggestedRewrite) > 0 {
				for _, suggestion := range v.SuggestedRewrite {
					fmt.Fprintf(w, "    suggestion: %s\n", suggestion)
				}
			}
		}
		fmt.Fprintf(w, "\n")
	}

	if result.Counterexample != "" {
		fmt.Fprintf(w, "Counterexample:\n%s\n", result.Counterexample)
	}
}

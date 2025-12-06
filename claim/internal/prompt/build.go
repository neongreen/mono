package prompt

import (
	"fmt"
	"strings"

	"github.com/neongreen/mono/claim/internal/parse"
)

// Build constructs the prompt for Claude to check a claim
func Build(
	claim parse.Claim,
	referencedClaims map[string]parse.Claim,
	lenses map[string]string,
	allClaimIDs []string,
	anonymizeFiles bool,
) string {
	var b strings.Builder

	// List of available claim IDs (so Claude doesn't invent IDs)
	b.WriteString("# Available Claim IDs\n\n")
	b.WriteString("These are all the claim IDs in the codebase. Do not reference any IDs not in this list:\n\n")
	for _, id := range allClaimIDs {
		fmt.Fprintf(&b, "- %s\n", id)
	}
	b.WriteString("\n")

	// Include selected lenses
	b.WriteString("# Lens Instructions\n\n")
	b.WriteString("Apply these lenses when checking the claim:\n\n")

	// Always include default lens if it exists
	if defaultLens, ok := lenses["default"]; ok {
		b.WriteString("## Default Lens\n\n")
		b.WriteString(defaultLens)
		b.WriteString("\n\n")
	}

	// Include lenses matching claim tags
	for _, tag := range claim.Tags {
		if tagLens, ok := lenses[tag]; ok {
			fmt.Fprintf(&b, "## %s Lens\n\n", tag)
			b.WriteString(tagLens)
			b.WriteString("\n\n")
		}
	}

	// Output contract
	b.WriteString("# Output Contract\n\n")
	b.WriteString("You must respond with a JSON object matching this exact structure:\n\n")
	b.WriteString("```json\n")
	b.WriteString(`{
  "claim_id": "string",
  "result": "proven|unproven|sorry",
  "bullets": [
    {
      "path": "string",
      "status": "ok|trivial|needs_split|needs_claim|contradicts|sorry",
      "required_claims": ["string"],
      "suggested_rewrite": ["string"]
    }
  ],
  "counterexample": "string"
}
`)
	b.WriteString("```\n\n")

	// Rules
	b.WriteString("# Rules\n\n")
	b.WriteString("1. Do NOT say \"proven\" unless every bullet is acceptable\n")
	b.WriteString("2. A bullet is acceptable if it is trivial, supported by nested bullets, depends on other claims, or is @sorry\n")
	b.WriteString("3. If a bullet contains @claim[id], you MUST list that ID in required_claims for that bullet\n")
	b.WriteString("4. If a bullet contains @claim[id], do NOT mark it as trivial - it depends on another claim\n")
	b.WriteString("5. If ANY bullet is exactly @sorry, the overall result MUST NOT be \"proven\"\n")
	b.WriteString("6. If unsure, return \"unproven\" with either a counterexample or list of missing cases\n")
	b.WriteString("7. Every bullet path must have exactly one verdict entry in the bullets array\n")
	b.WriteString("\n")

	// Target claim
	b.WriteString("# Target Claim to Check\n\n")
	b.WriteString(formatClaim(claim, anonymizeFiles))
	b.WriteString("\n")

	// Referenced claims
	if len(referencedClaims) > 0 {
		b.WriteString("# Referenced Claims\n\n")
		b.WriteString("These claims are referenced by the target claim:\n\n")
		for _, refClaim := range referencedClaims {
			b.WriteString(formatClaim(refClaim, anonymizeFiles))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// formatClaim renders a claim with its bullets and paths
func formatClaim(claim parse.Claim, anonymizeFiles bool) string {
	var b strings.Builder

	fmt.Fprintf(&b, "@claim[%s]", claim.ID)
	if len(claim.Tags) > 0 {
		for _, tag := range claim.Tags {
			fmt.Fprintf(&b, " @%s", tag)
		}
	}
	fmt.Fprintf(&b, ": %s\n", claim.Statement)

	// Optionally anonymize filename to prevent Claude from cheating
	if anonymizeFiles {
		fmt.Fprintf(&b, "Location: <source>:%d\n\n", claim.Line)
	} else {
		fmt.Fprintf(&b, "Location: %s:%d\n\n", claim.File, claim.Line)
	}

	// Include source code context
	if claim.SourceBefore != "" || claim.SourceAfter != "" {
		b.WriteString("Source context:\n```\n")
		if claim.SourceBefore != "" {
			b.WriteString(claim.SourceBefore)
			b.WriteString("\n")
		}
		b.WriteString("// ... @claim header and bullets here ...\n")
		if claim.SourceAfter != "" {
			b.WriteString(claim.SourceAfter)
			b.WriteString("\n")
		}
		b.WriteString("```\n\n")
	}

	formatBullets(&b, claim.Bullets, 0)

	return b.String()
}

// formatBullets recursively formats bullets with their paths
func formatBullets(b *strings.Builder, bullets []parse.Bullet, indent int) {
	for _, bullet := range bullets {
		// Indentation
		for i := 0; i < indent; i++ {
			b.WriteString("  ")
		}

		// Bullet with path (use format "path: text" not "[path] text" to avoid confusion)
		fmt.Fprintf(b, "- %s: %s\n", bullet.Path, bullet.Text)

		// Recursively format children
		if len(bullet.Children) > 0 {
			formatBullets(b, bullet.Children, indent+1)
		}
	}
}

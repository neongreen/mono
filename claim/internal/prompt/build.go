package prompt

import (
	"fmt"
	"strings"

	"github.com/neongreen/mono/claim/internal/parse"
)

// BuildProofPrompt constructs a prompt for the @proof syntax
// This evaluates a whole proof at once
// resolvedContext is the content fetched for the @context block (empty if no context)
func BuildProofPrompt(claim parse.Claim, axioms map[string]string, lenses map[string]string, resolvedContext string) string {
	var b strings.Builder

	// Role and verification approach
	b.WriteString("# Incremental Proof Verification\n\n")
	b.WriteString("We are doing **bottom-up proof verification**. Complex proofs are broken into lemmas, and each lemma is verified separately before being used.\n\n")
	b.WriteString("You are checking ONE STEP in this process. The lemmas listed below have ALREADY BEEN VERIFIED in previous passes. Your job is to check whether the current proof correctly uses these verified lemmas to establish its claim.\n\n")
	b.WriteString("This is like mathematical proof: once a theorem is proven, you cite it without re-proving it. Re-checking already-verified lemmas would be redundant and wasteful.\n\n")

	b.WriteString("# Your Task\n\n")
	b.WriteString("Check if the proof's conclusion follows from:\n")
	b.WriteString("1. The verified lemmas (listed below as \"Proven Lemmas\")\n")
	b.WriteString("2. The verified context (grep results, code snippets)\n")
	b.WriteString("3. The reasoning in the proof\n\n")
	b.WriteString("If the logic is valid and the proof correctly applies the lemmas, return 'proven'.\n")
	b.WriteString("If there are gaps in reasoning or incorrect use of lemmas, return 'unproven' with specific issues.\n\n")

	b.WriteString("# Common Mistakes\n\n")
	b.WriteString("- **Don't re-verify lemmas**: They're already proven. Just check if they're used correctly.\n")
	b.WriteString("- **Don't count grep matches in comments**: Focus on code, not claim/proof text.\n")
	b.WriteString("- **Don't invent hypothetical gaps**: Only flag specific issues you can identify.\n")
	b.WriteString("- **Grep is exhaustive**: When context provides grep results, those ARE all the matches. Grep is a complete text search - if a pattern isn't in the results, it doesn't exist in the searched scope.\n\n")

	b.WriteString("# Undefined Terms\n\n")
	b.WriteString("**CRITICAL**: If the CLAIM or PROOF references entities that are NOT defined in the verified context or proven lemmas, you MUST reject as UNPROVEN.\n\n")
	b.WriteString("This applies to ALL undefined terms, including:\n")
	b.WriteString("- Domain concepts in the claim itself (e.g., \"viewers\", \"workspace\", \"widgets\")\n")
	b.WriteString("- Code entities (functions, variables, files, permissions)\n")
	b.WriteString("- Any noun that refers to something in a codebase\n\n")
	b.WriteString("Example: A claim \"Viewers cannot invite anyone to the workspace\" is UNVERIFIABLE if:\n")
	b.WriteString("- No context defines what a \"viewer\" is\n")
	b.WriteString("- No context defines what a \"workspace\" is\n")
	b.WriteString("- No context defines what \"invite\" means in this system\n\n")
	b.WriteString("You cannot verify claims about concepts you have no information about. A proof cannot define terms that the claim uses - the definitions must come from verified context.\n\n")
	b.WriteString("Do NOT accept a proof's claims about code structure, permission mappings, function behavior, etc. unless that information is provided in the verified context. A proof cannot be its own evidence.\n\n")

	// Include additional lenses matching claim tags (pedantic, local, etc.)
	for _, tag := range claim.Tags {
		if tagLens, ok := lenses[tag]; ok {
			fmt.Fprintf(&b, "## %s Mode\n\n", tag)
			b.WriteString(tagLens)
			b.WriteString("\n\n")
		}
	}

	// Output format
	b.WriteString("# Output Format\n\n")
	b.WriteString("Respond with a JSON object:\n\n")
	b.WriteString("```json\n")
	b.WriteString(`{
  "result": "proven|unproven",
  "issues": [
    {
      "title": "Short issue title",
      "description": "Explanation of the problem"
    }
  ],
  "interpretation_narrowed": true|false
}
`)
	b.WriteString("```\n\n")

	b.WriteString("## Result Values\n\n")
	b.WriteString("- **proven**: The proof justifies the claim. The reasoning is sound.\n")
	b.WriteString("- **unproven**: The proof has gaps, errors, or doesn't establish the claim.\n\n")

	b.WriteString("## Issues Format\n\n")
	b.WriteString("When result is 'unproven', provide a list of issues:\n")
	b.WriteString("- **title**: Short phrase (3-6 words) identifying the problem\n")
	b.WriteString("- **description**: One sentence explaining the issue\n\n")
	b.WriteString("When result is 'proven', provide an empty issues array.\n\n")

	b.WriteString("## Interpretation Narrowed\n\n")
	b.WriteString("Set `interpretation_narrowed` to **true** if the provided context changes or narrows\n")
	b.WriteString("how the claim should be interpreted (e.g., an ambiguous 'the function' becomes specific).\n")
	b.WriteString("Set to **false** if the claim is unambiguous on its own and context only provides evidence.\n\n")

	// Verification guidance
	b.WriteString("# Verification\n\n")
	b.WriteString("When a proof claims something was verified by a search (grep, find, etc.):\n")
	b.WriteString("1. If the Verified Context section contains the search results, **use those results directly**\n")
	b.WriteString("2. If no context is provided for that search, run it yourself to verify\n")
	b.WriteString("3. If you find different results than claimed, mark as unproven and explain\n")
	b.WriteString("4. If the search methodology is unclear and no context provided, mark as unproven\n\n")

	// The claim
	b.WriteString("# Claim to Verify\n\n")
	fmt.Fprintf(&b, "@claim[%s]: %s\n\n", claim.ID, claim.Statement)

	// Context (if provided via @context block)
	if resolvedContext != "" {
		b.WriteString("# Verified Context (treat as ground truth)\n\n")
		b.WriteString("The following was gathered by a trusted verification process immediately before this check.\n")
		b.WriteString("**Treat these results as factual evidence. Do NOT re-run these searches or verifications.**\n")
		b.WriteString("Just as you assume axioms are true, assume this context is accurate and fresh.\n\n")
		b.WriteString("Use this context ONLY for evidence verification, NOT to narrow the claim's interpretation.\n")
		b.WriteString("If the claim is ambiguous without this context, mark as unproven with an 'Ambiguous claim' issue.\n\n")
		b.WriteString(resolvedContext)
		b.WriteString("\n\n")
	}

	// Proven lemmas (referenced claims)
	if len(axioms) > 0 {
		b.WriteString("# Proven Lemmas (verified in previous passes)\n\n")
		b.WriteString("These lemmas have already been verified. Use them as established facts:\n\n")
		for id, statement := range axioms {
			fmt.Fprintf(&b, "- **%s**: %s\n", id, statement)
		}
		b.WriteString("\n")
	}

	// The proof
	b.WriteString("# Proof\n\n")
	b.WriteString(claim.Proof)
	b.WriteString("\n\n")

	// Task
	b.WriteString("# Task\n\n")
	b.WriteString("Does this proof justify the claim?\n\n")
	b.WriteString("Consider:\n")
	b.WriteString("1. Does the proof use the axioms correctly?\n")
	b.WriteString("2. Are all logical steps valid?\n")
	b.WriteString("3. Are there gaps in the reasoning?\n")
	b.WriteString("4. Does the conclusion follow from the premises?\n")

	return b.String()
}

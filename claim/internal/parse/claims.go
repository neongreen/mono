package parse

import (
	"regexp"
	"strings"

	"github.com/neongreen/mono/claim/internal/logger"
)

// Claim represents a parsed claim with its context and proof
type Claim struct {
	ID          string
	Statement   string
	Tags        []string
	Context     string   // @context[id] block - specifies what context to fetch
	Proof       string   // @proof[id] block - the proof body
	SeeRefs     []string // @see[id] references in proof
	File        string
	Line        int
	ContextLine int // Line number where @context[id] starts (0 if no context)
	ProofLine   int // Line number where @proof[id] starts (0 if no proof)
}

var seeRefRegex = regexp.MustCompile(`@see\[([^\]]+)\]`)

// ParseClaims extracts all claims from file content
func ParseClaims(content, filename string) ([]Claim, error) {
	logger.Debug("parsing claims", "file", filename)

	// Parse all blocks using syntax module
	specs := []BlockSpec{ClaimSpec, ContextSpec, ProofSpec}
	blocks := ParseBlocks(content, filename, specs)

	// Group blocks by ID
	claimBlocks := make(map[string]Block)
	contextBlocks := make(map[string]Block)
	proofBlocks := make(map[string]Block)

	for _, block := range blocks {
		switch block.Type {
		case "claim":
			claimBlocks[block.ID] = block
		case "context":
			contextBlocks[block.ID] = block
		case "proof":
			proofBlocks[block.ID] = block
		}
	}

	// Build claims by associating context and proof blocks
	var claims []Claim
	for id, claimBlock := range claimBlocks {
		claim := Claim{
			ID:        id,
			Statement: claimBlock.Body, // For claims, Body holds the statement
			Tags:      parseTags(claimBlock.Args),
			File:      filename,
			Line:      claimBlock.Line,
		}

		if ctx, ok := contextBlocks[id]; ok {
			claim.Context = ctx.Body
			claim.ContextLine = ctx.Line
			logger.Debug("associated context", "claim", id)
		}

		if proof, ok := proofBlocks[id]; ok {
			claim.Proof = proof.Body
			claim.ProofLine = proof.Line
			claim.SeeRefs = extractSeeRefs(proof.Body)
			logger.Debug("associated proof", "claim", id, "see_refs", len(claim.SeeRefs))
		}

		claims = append(claims, claim)
	}

	logger.Debug("parsed claims", "file", filename, "count", len(claims))
	return claims, nil
}

// parseTags extracts tags from the args portion of a claim header
func parseTags(args string) []string {
	var tags []string
	if args == "" {
		return tags
	}
	for _, tag := range strings.Fields(args) {
		if strings.HasPrefix(tag, "@") {
			tags = append(tags, strings.TrimPrefix(tag, "@"))
		}
	}
	return tags
}

// extractSeeRefs finds all @see[id] or @see[id1, id2, ...] references in text
func extractSeeRefs(text string) []string {
	matches := seeRefRegex.FindAllStringSubmatch(text, -1)
	seen := make(map[string]bool)
	var refs []string
	for _, match := range matches {
		// Split by comma to support @see[foo, bar, baz]
		ids := strings.Split(match[1], ",")
		for _, id := range ids {
			id = strings.TrimSpace(id)
			if id != "" && !seen[id] {
				seen[id] = true
				refs = append(refs, id)
			}
		}
	}
	return refs
}

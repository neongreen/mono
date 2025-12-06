package parse

import (
	"regexp"
	"strings"
)

// Block represents a parsed block with a header like @type[id]: and optional body
type Block struct {
	Type string // "claim", "proof", "context", "lens"
	ID   string
	Args string // Additional args after ID (e.g., tags for claims, statement for claims)
	Body string // Body text for multi-line blocks
	Line int    // 1-indexed line number
}

// BlockSpec defines how to parse a block type
type BlockSpec struct {
	Name      string
	Pattern   *regexp.Regexp
	HasBody   bool // Whether this block has a multi-line body
	ArgsGroup int  // Regex capture group for args (0 = no args)
}

var (
	// @claim[id] tags...: statement (statement can be on same line or following lines)
	// Must start at beginning of line (after comment leader stripped) to avoid matching references
	ClaimSpec = BlockSpec{
		Name:      "claim",
		Pattern:   regexp.MustCompile(`^@claim\[([^\]]+)\]([^:]*):(.*)$`),
		HasBody:   true, // Now supports multi-line body for statement
		ArgsGroup: 2,    // tags are in group 2, statement (if inline) in group 3
	}

	// @context[id]: - must start at beginning
	ContextSpec = BlockSpec{
		Name:    "context",
		Pattern: regexp.MustCompile(`^@context\[([^\]]+)\]:`),
		HasBody: true,
	}

	// @proof[id]: - must start at beginning
	ProofSpec = BlockSpec{
		Name:    "proof",
		Pattern: regexp.MustCompile(`^@proof\[([^\]]+)\]:`),
		HasBody: true,
	}

	// @lens[id] - must start at beginning
	LensSpec = BlockSpec{
		Name:    "lens",
		Pattern: regexp.MustCompile(`^@lens\[([^\]]+)\]`),
		HasBody: true,
	}

	// All block specs for termination detection
	AllSpecs = []BlockSpec{ClaimSpec, ContextSpec, ProofSpec, LensSpec}
)

// ParseBlocks extracts all blocks of specified types from content
func ParseBlocks(content, filename string, specs []BlockSpec) []Block {
	lines := strings.Split(content, "\n")
	var blocks []Block

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		stripped := strings.TrimLeft(StripCommentLeader(line), " \t")

		for _, spec := range specs {
			matches := spec.Pattern.FindStringSubmatch(stripped)
			if matches == nil {
				continue
			}

			block := Block{
				Type: spec.Name,
				ID:   strings.TrimSpace(matches[1]),
				Line: i + 1,
			}

			// Handle special case for claims which have inline args and optional multi-line statement
			if spec.Name == "claim" && len(matches) >= 4 {
				block.Args = strings.TrimSpace(matches[2]) // tags
				inlineStatement := strings.TrimSpace(matches[3])
				if inlineStatement != "" {
					// Statement on same line
					block.Body = inlineStatement
				} else {
					// Statement on following lines
					body, nextLine := parseBlockBody(lines, i+1)
					block.Body = strings.TrimSpace(body)
					i = nextLine - 1
				}
			} else if spec.HasBody {
				body, nextLine := parseBlockBody(lines, i+1)
				block.Body = body
				i = nextLine - 1
			}

			blocks = append(blocks, block)
			break
		}
	}

	return blocks
}

// parseBlockBody extracts block text starting from startLine
// Returns the block body and the line number where it ends
func parseBlockBody(lines []string, startLine int) (string, int) {
	var bodyLines []string

	for i := startLine; i < len(lines); i++ {
		line := lines[i]

		// Stop at completely blank lines (not even a comment marker)
		if strings.TrimSpace(line) == "" {
			return strings.Join(bodyLines, "\n"), i
		}

		// Stop at non-comment lines (code)
		if !IsCommentLine(line) {
			return strings.Join(bodyLines, "\n"), i
		}

		stripped := StripCommentLeader(line)

		// End at any block header (trim for header detection, but preserve original for body)
		trimmedForHeader := strings.TrimLeft(stripped, " \t")
		for _, spec := range AllSpecs {
			if spec.Pattern.MatchString(trimmedForHeader) {
				return strings.Join(bodyLines, "\n"), i
			}
		}

		bodyLines = append(bodyLines, stripped)
	}

	return strings.Join(bodyLines, "\n"), len(lines)
}

// IsCommentLine checks if a line is a comment (not code)
func IsCommentLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return true
	}
	prefixes := []string{"//", "#", "/*", "*", "--", ";"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	return false
}

// StripCommentLeader removes comment characters from the start of a line
// but preserves indentation after the comment leader.
func StripCommentLeader(line string) string {
	prefixes := []string{"//", "#", "/*", "*", "--", ";"}

	for _, prefix := range prefixes {
		if idx := strings.Index(line, prefix); idx >= 0 {
			before := line[:idx]
			if strings.TrimSpace(before) == "" {
				return line[idx+len(prefix):]
			}
		}
	}

	return line
}

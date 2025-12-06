package parse

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/neongreen/mono/claim/internal/logger"
)

// Claim represents a parsed claim with its bullets
type Claim struct {
	ID           string
	Statement    string
	Tags         []string
	Bullets      []Bullet
	File         string
	Line         int
	SourceBefore string // Lines before the claim (for context)
	SourceAfter  string // Lines after the claim (for context)
}

// Bullet represents a single bullet point with possible children
type Bullet struct {
	Path         string   // e.g. "0", "1.2", "3.0.1"
	Text         string   // Bullet text content
	IsSorry      bool     // True if bullet is exactly "@sorry"
	References   []string // Referenced claim IDs (@claim[id])
	Children     []Bullet
	Indent       int // Indentation level
	OriginalLine int // Line number in file
}

var (
	// @claim[id] tags...: statement
	claimHeaderRegex = regexp.MustCompile(`@claim\[([^\]]+)\]([^:]*):(.+)`)
	// @claim[id] in bullet text
	claimRefRegex = regexp.MustCompile(`@claim\[([^\]]+)\]`)
)

// ParseClaims extracts all claims from file content
func ParseClaims(content, filename string) ([]Claim, error) {
	return ParseClaimsWithContext(content, filename, 10)
}

// ParseClaimsWithContext extracts claims with surrounding source code context
func ParseClaimsWithContext(content, filename string, contextLines int) ([]Claim, error) {
	logger.Debug("parsing claims", "file", filename)
	lines := strings.Split(content, "\n")
	var claims []Claim

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Try to find @claim header
		stripped := stripCommentLeader(line)
		if matches := claimHeaderRegex.FindStringSubmatch(stripped); matches != nil {
			id := strings.TrimSpace(matches[1])
			tagsPart := strings.TrimSpace(matches[2])
			statement := strings.TrimSpace(matches[3])

			logger.Debug("found claim", "id", id, "file", filename, "line", i+1)

			// Parse tags
			var tags []string
			if tagsPart != "" {
				for _, tag := range strings.Fields(tagsPart) {
					if strings.HasPrefix(tag, "@") {
						tags = append(tags, strings.TrimPrefix(tag, "@"))
					}
				}
			}

			// Parse bullets starting from next line
			bullets, nextLine := parseBullets(lines, i+1, filename)
			logger.Debug("parsed bullets", "claim", id, "count", len(bullets))

			// Capture source context
			beforeStart := max(0, i-contextLines)
			afterEnd := min(len(lines), nextLine+contextLines)

			sourceBefore := strings.Join(lines[beforeStart:i], "\n")
			sourceAfter := ""
			if nextLine < len(lines) {
				sourceAfter = strings.Join(lines[nextLine:afterEnd], "\n")
			}

			claims = append(claims, Claim{
				ID:           id,
				Statement:    statement,
				Tags:         tags,
				Bullets:      bullets,
				File:         filename,
				Line:         i + 1, // 1-indexed
				SourceBefore: sourceBefore,
				SourceAfter:  sourceAfter,
			})

			// Continue from after the bullets
			i = nextLine - 1 // -1 because loop will increment
		}
	}

	logger.Debug("parsed claims", "file", filename, "count", len(claims))
	return claims, nil
}

// parseBullets parses bullet points starting from startLine
// Returns bullets and the line number where bullets end
func parseBullets(lines []string, startLine int, filename string) ([]Bullet, int) {
	var rawBullets []struct {
		text   string
		indent int
		line   int
	}

	for i := startLine; i < len(lines); i++ {
		line := lines[i]
		stripped := stripCommentLeader(line)

		// Check if this is a new claim header - if so, end bullet parsing
		if claimHeaderRegex.MatchString(stripped) {
			break
		}

		// Check if this is a bullet
		trimmed := strings.TrimLeft(stripped, " \t")
		if strings.HasPrefix(trimmed, "- ") {
			// Measure indentation
			indent := len(stripped) - len(trimmed)
			text := strings.TrimPrefix(trimmed, "- ")

			rawBullets = append(rawBullets, struct {
				text   string
				indent int
				line   int
			}{text, indent, i + 1})
		} else if strings.TrimSpace(stripped) == "" {
			// Empty line - continue
			continue
		} else if strings.TrimSpace(stripped) != "" && len(rawBullets) > 0 {
			// Non-bullet, non-empty line after bullets started - end parsing
			break
		}
	}

	// Build nested structure
	bullets := buildBulletTree(rawBullets)
	assignPaths(bullets, "")

	return bullets, startLine + len(rawBullets)
}

// buildBulletTree converts flat list with indentation into nested structure
func buildBulletTree(raw []struct {
	text   string
	indent int
	line   int
}) []Bullet {
	if len(raw) == 0 {
		return nil
	}

	var bullets []Bullet
	i := 0

	for i < len(raw) {
		current := raw[i]
		bullet := Bullet{
			Text:         current.text,
			IsSorry:      strings.TrimSpace(current.text) == "@sorry",
			References:   extractReferences(current.text),
			Indent:       current.indent,
			OriginalLine: current.line,
		}

		// Find children (items with greater indentation)
		childStart := i + 1
		childEnd := childStart
		for childEnd < len(raw) && raw[childEnd].indent > current.indent {
			childEnd++
		}

		if childEnd > childStart {
			// Recursively parse children
			bullet.Children = buildBulletTree(raw[childStart:childEnd])
		}

		bullets = append(bullets, bullet)
		i = childEnd
	}

	return bullets
}

// assignPaths assigns stable paths like "0", "1.2", "3.0.1" to bullets
func assignPaths(bullets []Bullet, parentPath string) {
	for i := range bullets {
		var path string
		if parentPath == "" {
			path = fmt.Sprintf("%d", i)
		} else {
			path = fmt.Sprintf("%s.%d", parentPath, i)
		}
		bullets[i].Path = path
		assignPaths(bullets[i].Children, path)
	}
}

// extractReferences finds all @claim[id] references in text
func extractReferences(text string) []string {
	matches := claimRefRegex.FindAllStringSubmatch(text, -1)
	var refs []string
	for _, match := range matches {
		refs = append(refs, match[1])
	}
	return refs
}

// stripCommentLeader removes comment characters from the start of a line
// but preserves indentation after the comment leader.
// Accepts lines like "// @claim[id]", "#   - bullet", "/*  @lens[x]", etc.
func stripCommentLeader(line string) string {
	// Strategy: Find common comment prefixes and remove them while preserving
	// the indentation that comes after them.

	// For bullets specifically, we want to preserve space before the "-"
	// Example: "//   - child" should become "  - child" (preserving 2 spaces)

	// Common comment prefixes
	prefixes := []string{"//", "#", "/*", "*", "--", ";"}

	for _, prefix := range prefixes {
		if idx := strings.Index(line, prefix); idx >= 0 {
			// Check if everything before prefix is whitespace
			before := line[:idx]
			if strings.TrimSpace(before) == "" {
				// Strip the prefix and everything before it, keep everything after
				return line[idx+len(prefix):]
			}
		}
	}

	// No comment prefix found, return as-is
	return line
}

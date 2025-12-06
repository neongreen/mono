package parse

import (
	"regexp"
	"strings"
)

var (
	// @lens[name]
	lensHeaderRegex = regexp.MustCompile(`@lens\[([^\]]+)\]`)
)

// ParseLenses extracts all lenses from file content
func ParseLenses(content, filename string) (map[string]string, error) {
	lines := strings.Split(content, "\n")
	lenses := make(map[string]string)

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		stripped := stripCommentLeader(line)

		// Try to find @lens header
		if matches := lensHeaderRegex.FindStringSubmatch(stripped); matches != nil {
			name := strings.TrimSpace(matches[1])

			// Collect all following lines until next @lens or @claim header
			var body []string
			for j := i + 1; j < len(lines); j++ {
				nextLine := lines[j]
				nextStripped := stripCommentLeader(nextLine)

				// Stop at next lens or claim header
				if lensHeaderRegex.MatchString(nextStripped) || claimHeaderRegex.MatchString(nextStripped) {
					break
				}

				// Add line to body (strip comment leaders but keep content)
				if trimmed := strings.TrimSpace(nextStripped); trimmed != "" {
					body = append(body, trimmed)
				}
			}

			lenses[name] = strings.Join(body, "\n")
		}
	}

	return lenses, nil
}


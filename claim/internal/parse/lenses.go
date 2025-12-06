package parse

import (
	"strings"
)

// ParseLenses extracts all lenses from file content
func ParseLenses(content, filename string) (map[string]string, error) {
	blocks := ParseBlocks(content, filename, []BlockSpec{LensSpec})
	lenses := make(map[string]string)

	for _, block := range blocks {
		// Clean up body - remove empty lines and trim
		var lines []string
		for _, line := range strings.Split(block.Body, "\n") {
			if trimmed := strings.TrimSpace(line); trimmed != "" {
				lines = append(lines, trimmed)
			}
		}
		lenses[block.ID] = strings.Join(lines, "\n")
	}

	return lenses, nil
}

package parser

import "strings"

// isJSONL detects if content is in JSONL format by checking if the first few lines
// contain valid JSON objects
func isJSONL(content string) bool {
	lines := strings.Split(content, "\n")
	jsonLineCount := 0
	for i, line := range lines {
		if i > 5 {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
			jsonLineCount++
		}
	}
	return jsonLineCount >= 2
}

// parseSimpleTrace creates a simple representation for non-JSONL traces
func parseSimpleTrace(content string) (*ParsedTrace, error) {
	return &ParsedTrace{
		Summary: "Legacy trace format",
		Items:   []ConversationItem{},
	}, nil
}

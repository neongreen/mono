package beads

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ReadBeadsFile reads and parses a beads JSONL file
func ReadBeadsFile(path string) ([]BeadsIssue, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var issues []BeadsIssue
	scanner := bufio.NewScanner(file)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		var issue BeadsIssue
		if err := json.Unmarshal([]byte(line), &issue); err != nil {
			return nil, fmt.Errorf("failed to parse line %d: %w", lineNum, err)
		}

		issues = append(issues, issue)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return issues, nil
}

// ExtractPrefixesFromBeads groups beads issues by their prefix
func ExtractPrefixesFromBeads(issues []BeadsIssue) map[string][]BeadsIssue {
	grouped := make(map[string][]BeadsIssue)

	for _, issue := range issues {
		// Extract prefix from ID (e.g., "mono-123" → "mono")
		parts := strings.Split(issue.ID, "-")
		if len(parts) < 2 {
			// Skip malformed IDs
			continue
		}
		prefix := parts[0]
		grouped[prefix] = append(grouped[prefix], issue)
	}

	return grouped
}

// ParseBeadsNumber extracts the number from a beads ID
func ParseBeadsNumber(id string) (int64, error) {
	// mono-123 → 123
	parts := strings.Split(id, "-")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid beads ID format: %s", id)
	}
	return strconv.ParseInt(parts[1], 10, 64)
}

// MapBeadsStatus maps beads status to tk status
func MapBeadsStatus(beadsStatus string) string {
	switch beadsStatus {
	case "in_progress":
		return "in-progress"
	case "closed":
		return "done"
	case "open":
		return "todo"
	default:
		return beadsStatus
	}
}

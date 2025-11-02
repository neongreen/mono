package types

import "strings"

// ExtractPrefix extracts the prefix from a TaskID (format: prefix-number-node).
// Returns the prefix part, or empty string if the TaskID is invalid.
func ExtractPrefix(taskID string) string {
	parts := strings.Split(taskID, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

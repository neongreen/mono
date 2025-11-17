package status

import (
	"fmt"
	"slices"
	"strings"
)

// PredefinedStatuses is the list of standard statuses that tk encourages users to use
// Empty string represents "no status" and is the default state
var PredefinedStatuses = []string{
	"next",
	"wip",
	"done",
	"closed",
}

// IsValidPredefinedStatus checks if a status is in the predefined list
func IsValidPredefinedStatus(status string) bool {
	normalized := strings.ToLower(strings.TrimSpace(status))
	return slices.Contains(PredefinedStatuses, normalized)
}

// ValidateStatus validates a status value. If allowCustom is false, only predefined
// statuses are allowed. Returns an error with helpful message if validation fails.
// existingCustomStatuses should contain any custom statuses already in use in the project.
func ValidateStatus(status string, allowCustom bool, existingCustomStatuses []string) error {
	if status == "" {
		return nil // Empty status is allowed (for unsetting)
	}

	if allowCustom {
		return nil // Any status is allowed with custom flag
	}

	normalized := NormalizeStatus(status)

	// Check if it's a predefined status
	if IsValidPredefinedStatus(normalized) {
		return nil
	}

	// Check if it's an existing custom status in the project
	for _, existing := range existingCustomStatuses {
		if strings.EqualFold(existing, normalized) {
			return nil
		}
	}

	// Build the error message with both predefined and existing custom statuses
	allStatuses := make([]string, 0, len(PredefinedStatuses)+len(existingCustomStatuses))
	allStatuses = append(allStatuses, PredefinedStatuses...)
	allStatuses = append(allStatuses, existingCustomStatuses...)

	return fmt.Errorf(
		"This project uses: %s\nIf you want to use '%s' here, add: --custom-status",
		strings.Join(allStatuses, ", "),
		status,
	)
}

// NormalizeStatus converts a status to lowercase for consistent storage
func NormalizeStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

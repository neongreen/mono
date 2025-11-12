package status

import (
	"fmt"
	"strings"
)

// PredefinedStatuses is the list of standard statuses that tk encourages users to use
var PredefinedStatuses = []string{
	"todo",
	"wip",
	"next",
	"done",
	"blocked",
	"cancelled",
	"abandoned",
}

// IsValidPredefinedStatus checks if a status is in the predefined list
func IsValidPredefinedStatus(status string) bool {
	normalized := strings.ToLower(strings.TrimSpace(status))
	for _, s := range PredefinedStatuses {
		if s == normalized {
			return true
		}
	}
	return false
}

// ValidateStatus validates a status value. If allowCustom is false, only predefined
// statuses are allowed. Returns an error with helpful message if validation fails.
func ValidateStatus(status string, allowCustom bool) error {
	if status == "" {
		return nil // Empty status is allowed (for unsetting)
	}

	if allowCustom {
		return nil // Any status is allowed with custom flag
	}

	if !IsValidPredefinedStatus(status) {
		return fmt.Errorf(
			"invalid status: %q\nValid statuses are: %s\nUse --custom-status to set a non-standard status",
			status,
			strings.Join(PredefinedStatuses, ", "),
		)
	}

	return nil
}

// NormalizeStatus converts a status to lowercase for consistent storage
func NormalizeStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

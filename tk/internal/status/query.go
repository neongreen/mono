package status

import "github.com/neongreen/mono/tk/internal/types"

// GetExistingCustomStatusesFromTasks returns all custom statuses (non-predefined) currently in use
// in the given tasks list.
func GetExistingCustomStatusesFromTasks(tasks []*types.Task) []string {
	seenStatuses := make(map[string]bool)
	var customStatuses []string

	for _, task := range tasks {
		// Check the generic axis for status
		if axis, ok := task.Axes["generic"]; ok {
			status := axis.Effective
			normalized := NormalizeStatus(status)

			// Skip predefined statuses, empty values, and already seen
			if normalized == "" || IsValidPredefinedStatus(normalized) || seenStatuses[normalized] {
				continue
			}

			seenStatuses[normalized] = true
			customStatuses = append(customStatuses, normalized)
		}
	}

	return customStatuses
}

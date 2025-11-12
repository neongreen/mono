package types

import (
	"sort"
	"strings"
)

// SortTasks sorts tasks based on the specified sort order.
// Supported sort orders: "created", "created-desc", "id", "title", "status", or empty (defaults to "created").
// Add "-desc" suffix for descending order (e.g., "created-desc" for newest first).
func SortTasks(tasks []*Task, sortBy string) {
	// Check for descending order suffix
	descending := strings.HasSuffix(sortBy, "-desc")
	if descending {
		sortBy = strings.TrimSuffix(sortBy, "-desc")
	}

	switch sortBy {
	case "created":
		sort.Slice(tasks, func(i, j int) bool {
			if descending {
				return tasks[j].CreatedAt.Before(tasks[i].CreatedAt)
			}
			return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
		})
	case "id":
		sort.Slice(tasks, func(i, j int) bool {
			if descending {
				return tasks[j].TaskDisplayID < tasks[i].TaskDisplayID
			}
			return tasks[i].TaskDisplayID < tasks[j].TaskDisplayID
		})
	case "title":
		sort.Slice(tasks, func(i, j int) bool {
			if descending {
				return tasks[j].Title < tasks[i].Title
			}
			return tasks[i].Title < tasks[j].Title
		})
	case "status":
		sort.Slice(tasks, func(i, j int) bool {
			statusI := getEffectiveStatus(tasks[i])
			statusJ := getEffectiveStatus(tasks[j])
			if descending {
				return statusJ < statusI
			}
			return statusI < statusJ
		})
	default:
		// Default to created ascending
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
		})
	}
}

// getEffectiveStatus returns the effective status from the generic axis, or empty string
func getEffectiveStatus(task *Task) string {
	if axis, ok := task.Axes["generic"]; ok {
		return axis.Effective
	}
	return ""
}

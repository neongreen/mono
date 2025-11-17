package query

import (
	"regexp"
	"slices"
	"strings"

	"github.com/neongreen/mono/tk/internal/types"
)

// FilterOptions contains options for filtering tasks
type FilterOptions struct {
	Projects         []string // Project aliases to filter by
	AxisFilter       string   // Format: "axis:state" (deprecated, use AxisFilters)
	AxisFilters      []string // Multiple axis filters with OR logic. Format: "axis:state"
	NegatedAxisState string   // Negated axis:state to exclude (e.g., "generic:done")
	BlockedOnly      bool
	UnblockedOnly    bool
	GrepPattern      string   // Regex pattern to match against title and notes
	ItemKinds        []string // Item kinds to filter by (OR logic)
}

// FilterTasks filters a list of tasks based on the given options
func FilterTasks(tasks []*types.Task, taskUIDSet map[string]bool, opts FilterOptions) []*types.Task {
	var filtered []*types.Task

	// Compile grep pattern once if provided (case insensitive by default)
	var grepRe *regexp.Regexp
	if opts.GrepPattern != "" {
		var err error
		// Prepend (?i) for case-insensitive matching by default
		grepRe, err = regexp.Compile("(?i)" + opts.GrepPattern)
		if err != nil {
			// Return empty list if regex is invalid
			return filtered
		}
	}

	for _, task := range tasks {
		// Filter by project (if task UID set is provided)
		if taskUIDSet != nil && !taskUIDSet[task.TaskUUID] {
			continue
		}

		// Filter by item kind (if specified)
		if len(opts.ItemKinds) > 0 {
			matched := slices.Contains(opts.ItemKinds, task.ItemKind)
			if !matched {
				continue
			}
		}

		// Handle negated axis filter (exclude instead of include)
		if opts.NegatedAxisState != "" {
			parts := strings.Split(opts.NegatedAxisState, ":")
			if len(parts) == 2 {
				axisName := parts[0]
				stateName := parts[1]

				axis, ok := task.Axes[axisName]
				// Exclude tasks that match the negated state
				if ok && axis.Effective == stateName {
					continue
				}
			}
		}

		// Filter by axis (support both single and multiple filters)
		filters := opts.AxisFilters
		if len(filters) == 0 && opts.AxisFilter != "" {
			// Backward compatibility: use single AxisFilter if AxisFilters is empty
			filters = []string{opts.AxisFilter}
		}

		if len(filters) > 0 {
			// OR logic: task passes if it matches ANY of the filters
			matched := false
			for _, filter := range filters {
				parts := strings.Split(filter, ":")
				if len(parts) == 2 {
					axisName := parts[0]
					stateName := parts[1]

					axis, ok := task.Axes[axisName]
					if ok && axis.Effective == stateName {
						matched = true
						break
					}
				}
			}
			if !matched {
				continue
			}
		}

		// Filter by blocked status
		if opts.BlockedOnly && !task.Blocked {
			continue
		}
		if opts.UnblockedOnly && task.Blocked {
			continue
		}

		// Filter by grep pattern (regex)
		if grepRe != nil {
			matched := false

			// Check title
			if grepRe.MatchString(task.Title) {
				matched = true
			}

			// Check notes
			if !matched {
				for _, note := range task.Notes {
					if grepRe.MatchString(note.Markdown) {
						matched = true
						break
					}
				}
			}

			if !matched {
				continue
			}
		}

		filtered = append(filtered, task)
	}

	return filtered
}

// GroupTasks groups tasks by a specified key
func GroupTasks(tasks []*types.Task, groupBy string, getGroupKey func(*types.Task) string) (map[string][]*types.Task, []string) {
	grouped := make(map[string][]*types.Task)
	var groupOrder []string

	for _, task := range tasks {
		groupKey := getGroupKey(task)

		if _, exists := grouped[groupKey]; !exists {
			groupOrder = append(groupOrder, groupKey)
		}
		grouped[groupKey] = append(grouped[groupKey], task)
	}

	return grouped, groupOrder
}

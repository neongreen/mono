package query

import (
	"strings"

	"github.com/neongreen/mono/tk/internal/types"
)

// FilterOptions contains options for filtering tasks
type FilterOptions struct {
	Projects      []string // Project aliases to filter by
	AxisFilter    string   // Format: "axis:state"
	BlockedOnly   bool
	UnblockedOnly bool
}

// FilterTasks filters a list of tasks based on the given options
func FilterTasks(tasks []*types.Task, taskUIDSet map[string]bool, opts FilterOptions) []*types.Task {
	var filtered []*types.Task

	for _, task := range tasks {
		// Filter by project (if task UID set is provided)
		if taskUIDSet != nil && !taskUIDSet[task.TaskUUID] {
			continue
		}

		// Filter by axis
		if opts.AxisFilter != "" {
			parts := strings.Split(opts.AxisFilter, ":")
			if len(parts) == 2 {
				axisName := parts[0]
				stateName := parts[1]

				axis, ok := task.Axes[axisName]
				if !ok || axis.Effective != stateName {
					continue
				}
			}
		}

		// Filter by blocked status
		if opts.BlockedOnly && !task.Blocked {
			continue
		}
		if opts.UnblockedOnly && task.Blocked {
			continue
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

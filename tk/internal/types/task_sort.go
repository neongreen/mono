package types

import "sort"

// SortTasks sorts tasks based on the specified sort order.
// Supported sort orders: "created", "id", "title", or empty (defaults to "created").
func SortTasks(tasks []*Task, sortBy string) {
	switch sortBy {
	case "created":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
		})
	case "id":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].TaskID < tasks[j].TaskID
		})
	case "title":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Title < tasks[j].Title
		})
	default:
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
		})
	}
}

package database

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/reducer"
)

// GetProjectAliasForTask returns the project name for a task.
// Now uses reducer instead of direct SQL queries.
func GetProjectAliasForTask(db *DB, reducer *reducer.Reducer, taskUID string) (string, error) {
	// Get the task from reducer
	task, exists := reducer.GetTask(taskUID)
	if !exists {
		return "", fmt.Errorf("task %s not found", taskUID)
	}

	// Get project from reducer
	project, exists := reducer.GetProject(task.ProjectUUID)
	if !exists {
		// Fall back to project UID if project not found
		return task.ProjectUUID, nil
	}

	return project.Name, nil
}

// GetAllProjectDisplayNames returns a map of project UIDs to their display names (name).
// Now uses reducer instead of direct SQL queries.
func GetAllProjectDisplayNames(db *DB, reducer *reducer.Reducer) (map[string]string, error) {
	result := make(map[string]string)
	for _, project := range reducer.GetAllProjects() {
		result[project.ProjectUID] = project.Name
	}
	return result, nil
}

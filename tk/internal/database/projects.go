package database

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/types"
)

// GetProjectAliasForTask returns the preferred project alias for a task.
// If no alias exists, falls back to project name, then UID.
func GetProjectAliasForTask(db *DB, taskUID string) (string, error) {
	// Get project UID for this task
	var projectUID string
	err := db.Db.QueryRow(`
		SELECT project_uid FROM tasks WHERE task_uid = ?
	`, taskUID).Scan(&projectUID)
	if err != nil {
		return "", fmt.Errorf("failed to get project for task %s: %w", taskUID, err)
	}

	alias, err := PreferredAliasForProject(db, types.ProjectUID(projectUID))
	if err != nil {
		return "", err
	}

	if alias == "" {
		// If no alias exists, fall back to project name
		var projectName string
		err := db.Db.QueryRow(`
			SELECT name FROM projects WHERE project_uid = ?
		`, projectUID).Scan(&projectName)
		if err != nil {
			// If we can't get the name, fall back to UID
			return projectUID, nil
		}
		return projectName, nil
	}

	return alias, nil
}

// GetAllProjectDisplayNames returns a map of project UIDs to their display names (alias or name).
func GetAllProjectDisplayNames(db *DB) (map[string]string, error) {
	// Query all projects
	rows, err := db.Db.Query(`
		SELECT project_uid, name FROM projects ORDER BY created_at
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var projectUID, name string
		if err := rows.Scan(&projectUID, &name); err != nil {
			return nil, err
		}

		// Try to get preferred alias
		alias, err := PreferredAliasForProject(db, types.ProjectUID(projectUID))
		if err != nil {
			// On error, fall back to name
			result[projectUID] = name
			continue
		}

		if alias != "" {
			result[projectUID] = alias
		} else {
			result[projectUID] = name
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

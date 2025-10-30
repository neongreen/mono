package main

import (
	"fmt"
	"os/user"
	"strings"
)

// extractPrefix extracts the prefix from a TaskID (format: prefix-number-node)
func extractPrefix(taskID string) string {
	parts := strings.Split(taskID, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// getProjectAliasForTask returns the preferred project alias for a task
func getProjectAliasForTask(db *DB, taskUID string) (string, error) {
	// Get project UID for this task
	var projectUID string
	err := db.db.QueryRow(`
		SELECT project_uid FROM tasks WHERE task_uid = ?
	`, taskUID).Scan(&projectUID)
	if err != nil {
		return "", fmt.Errorf("failed to get project for task %s: %w", taskUID, err)
	}

	alias, err := preferredAliasForProject(db, projectUID)
	if err != nil {
		return "", err
	}

	if alias == "" {

		return projectUID, nil
	}

	return alias, nil
}

func openExistingDB() (*DB, error) {
	path, err := GetDBPath()
	if err != nil {
		return nil, err
	}

	db, err := OpenDB(path)
	if err != nil {
		return nil, err
	}

	if err := db.InitDB(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database schema: %w", err)
	}

	return db, nil
}

func getCurrentUser() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	return currentUser.Username, nil
}

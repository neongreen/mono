package tasks

import (
	"database/sql"
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
)

// GetProjectForTask returns the project UID for a given task
func GetProjectForTask(db *database.DB, taskUID string) (string, error) {
	var projectUID string
	err := db.Db.QueryRow(`SELECT project_uid FROM tasks WHERE task_uid = ?`, taskUID).Scan(&projectUID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("task %s not found in tasks table", taskUID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to load task %s: %w", taskUID, err)
	}
	return projectUID, nil
}

// GetProjectAndNumberForTask returns both the project UID and number for a given task
func GetProjectAndNumberForTask(db *database.DB, taskUID string) (string, int64, error) {
	var projectUID string
	if err := db.Db.QueryRow(`SELECT project_uid FROM tasks WHERE task_uid = ?`, taskUID).Scan(&projectUID); err != nil {
		if err == sql.ErrNoRows {
			return "", 0, fmt.Errorf("task %s not found", taskUID)
		}
		return "", 0, fmt.Errorf("failed to lookup task %s: %w", taskUID, err)
	}

	var number int64
	if err := db.Db.QueryRow(`SELECT number FROM task_numbers WHERE task_uid = ?`, taskUID).Scan(&number); err != nil {
		if err == sql.ErrNoRows {
			number = 0
		} else {
			return "", 0, fmt.Errorf("failed to lookup task number: %w", err)
		}
	}

	return projectUID, number, nil
}

// CheckNumberCollision checks if a number already exists in a project (excluding a specific task)
func CheckNumberCollision(db *database.DB, projectUID string, number int64, excludeTaskUID string) (bool, error) {
	var count int
	if err := db.Db.QueryRow(`
		SELECT COUNT(*) FROM task_numbers
		WHERE project_uid = ? AND number = ? AND task_uid != ?
	`, projectUID, number, excludeTaskUID).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check number collision: %w", err)
	}
	return count > 0, nil
}

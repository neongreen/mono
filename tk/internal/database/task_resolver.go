package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/neongreen/mono/tk/internal/types"
)

// ResolveTaskReference resolves a user-supplied task reference into a task_uid.
// Supports direct task_uids and display IDs in the form <alias>-<number>[-<node_hint>].
func ResolveTaskReference(db *DB, ref types.TaskRef) (string, error) {
	refStr := strings.TrimSpace(string(ref))
	if refStr == "" {
		return "", fmt.Errorf("task reference cannot be empty")
	}

	// Direct task UID
	if ref.IsTaskUID() {
		var count int
		if err := db.Db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE task_uid = ?`, refStr).Scan(&count); err != nil {
			return "", fmt.Errorf("failed to lookup task %s: %w", refStr, err)
		}
		if count == 0 {
			return "", fmt.Errorf("task %s not found", refStr)
		}
		return refStr, nil
	}

	// Pure numeric references are ambiguous by design.
	if _, err := strconv.ParseInt(refStr, 10, 64); err == nil {
		return "", fmt.Errorf("ambiguous task reference %s: numeric references are not supported", refStr)
	}

	displayID := types.DisplayID(refStr)
	alias, number, nodeHint, err := displayID.Parse()
	if err != nil {
		return "", fmt.Errorf("invalid task reference %s: %w", refStr, err)
	}

	projectUID, err := ResolveProjectByAlias(db, alias)
	if err != nil {
		return "", err
	}

	rows, err := db.Db.Query(`
		SELECT task_uid FROM task_numbers
		WHERE project_uid = ? AND number = ?
	`, projectUID, number)
	if err != nil {
		return "", fmt.Errorf("failed to query tasks for %s: %w", refStr, err)
	}
	defer rows.Close()

	var candidateUIDs []string
	for rows.Next() {
		var taskUID string
		if err := rows.Scan(&taskUID); err != nil {
			return "", fmt.Errorf("failed to scan task UID: %w", err)
		}
		candidateUIDs = append(candidateUIDs, taskUID)
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("failed to iterate tasks for %s: %w", refStr, err)
	}

	if len(candidateUIDs) == 0 {
		return "", fmt.Errorf("task %s not found", refStr)
	}

	if len(candidateUIDs) == 1 {
		return candidateUIDs[0], nil
	}

	// Multiple matches – use node hint if present.
	if nodeHint != "" {
		for _, uid := range candidateUIDs {
			short, err := taskNodeHint(db, uid)
			if err != nil {
				return "", err
			}
			if short == nodeHint {
				return uid, nil
			}
		}
		return "", fmt.Errorf("node hint %s for %s does not match any task", nodeHint, refStr)
	}

	// Ambiguous without hint – return list of choices.
	displayIDs := make([]string, 0, len(candidateUIDs))
	for _, uid := range candidateUIDs {
		display, err := RenderTaskDisplayID(db, uid)
		if err != nil {
			return "", err
		}
		displayIDs = append(displayIDs, display)
	}

	return "", fmt.Errorf("ambiguous task reference %s; candidates: %s", refStr, strings.Join(displayIDs, ", "))
}

// RenderTaskDisplayID renders the preferred display string for a task.
func RenderTaskDisplayID(db *DB, taskUID string) (string, error) {
	var projectUID string
	var number int64
	err := db.Db.QueryRow(`
		SELECT project_uid, number FROM task_numbers WHERE task_uid = ?
	`, taskUID).Scan(&projectUID, &number)
	if errors.Is(err, sql.ErrNoRows) {
		return taskUID, nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to load task number for %s: %w", taskUID, err)
	}

	// Use project name for display (aliases removed)
	var projectName string
	err = db.Db.QueryRow(`
		SELECT name FROM projects WHERE project_uid = ?
	`, projectUID).Scan(&projectName)
	if err != nil {
		// If we can't get the name, fall back to UID
		return fmt.Sprintf("%s-%d", projectUID, number), nil
	}

	collision, err := HasNumberCollision(db, projectUID, number)
	if err != nil {
		return "", err
	}
	if !collision {
		return fmt.Sprintf("%s-%d", projectName, number), nil
	}

	hint, err := taskNodeHint(db, taskUID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%d-%s", projectName, number, hint), nil
}

// ResolveProjectRef resolves an unresolved project reference (UID or name) to a ProjectUID
func ResolveProjectRef(db *DB, ref types.ProjectRef) (types.ProjectUID, error) {
	// If it looks like a ProjectUID, validate and verify it exists
	if ref.IsProjectUID() {
		uid := types.ProjectUID(ref)
		if err := uid.Validate(); err != nil {
			return "", fmt.Errorf("invalid project UID: %w", err)
		}

		var count int
		if err := db.Db.QueryRow(`SELECT COUNT(*) FROM projects WHERE project_uid = ?`, uid).Scan(&count); err != nil {
			return "", fmt.Errorf("failed to lookup project %s: %w", uid, err)
		}
		if count == 0 {
			return "", fmt.Errorf("project %s not found", uid)
		}
		return uid, nil
	}

	// Find matches by project name only (aliases removed)
	matchedProjects := make(map[string]string) // uid -> name

	rows, err := db.Db.Query(`
		SELECT project_uid, name FROM projects
		WHERE name = ?
		ORDER BY created_at
	`, ref.String())
	if err != nil {
		return "", fmt.Errorf("failed to resolve project name %s: %w", ref, err)
	}
	defer rows.Close()

	for rows.Next() {
		var uid, name string
		if err := rows.Scan(&uid, &name); err != nil {
			return "", fmt.Errorf("failed to scan name match: %w", err)
		}
		matchedProjects[uid] = name
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("failed to iterate projects: %w", err)
	}

	// No matches - list available projects
	if len(matchedProjects) == 0 {
		availableProjects, err := listAvailableProjects(db)
		if err != nil {
			return "", fmt.Errorf("project %q not found", ref)
		}
		return "", fmt.Errorf("project %q not found. Available projects: %s", ref, strings.Join(availableProjects, ", "))
	}

	// Single unique project UID - success!
	if len(matchedProjects) == 1 {
		for uid := range matchedProjects {
			return types.ProjectUID(uid), nil
		}
	}

	// Multiple projects with same name - ambiguous
	var matchDescriptions []string
	for uid, name := range matchedProjects {
		matchDescriptions = append(matchDescriptions, fmt.Sprintf("%s (%s)", uid, name))
	}
	return "", fmt.Errorf("ambiguous project reference %q; matches: %s. Use full project UID to disambiguate", ref, strings.Join(matchDescriptions, ", "))
}

// listAvailableProjects returns a list of available project names
func listAvailableProjects(db *DB) ([]string, error) {
	var projects []string

	// Get project names only (aliases removed)
	rows, err := db.Db.Query(`SELECT name FROM projects ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		projects = append(projects, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

// ResolveProjectByAlias is deprecated - use ResolveProjectRef instead
// This is kept temporarily for backward compatibility during refactoring
// deprecated:v5 remove-after:v5-migration
func ResolveProjectByAlias(db *DB, alias string) (string, error) {
	uid, err := ResolveProjectRef(db, types.NewProjectRef(alias))
	return string(uid), err
}

// PreferredAliasForProject is deprecated - aliases have been removed.
// Kept for backward compatibility but always returns empty string.
// deprecated:v5 remove-after:v5-migration
func PreferredAliasForProject(db *DB, projectUID types.ProjectUID) (string, error) {
	return "", nil
}

func HasNumberCollision(db *DB, projectUID string, number int64) (bool, error) {
	var count int
	if err := db.Db.QueryRow(`
		SELECT COUNT(*) FROM task_numbers
		WHERE project_uid = ? AND number = ?
	`, projectUID, number).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check collisions: %w", err)
	}
	return count > 1, nil
}

func taskNodeHint(db *DB, taskUID string) (string, error) {
	var createdNode string
	if err := db.Db.QueryRow(`SELECT created_node FROM tasks WHERE task_uid = ?`, taskUID).Scan(&createdNode); err != nil {
		return "", fmt.Errorf("failed to lookup node for task %s: %w", taskUID, err)
	}
	return types.NodeID(createdNode).Short(), nil
}

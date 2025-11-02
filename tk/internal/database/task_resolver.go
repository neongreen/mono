package database

import (
	"database/sql"
	"fmt"
	"github.com/neongreen/mono/tk/internal/types"
	"strconv"
	"strings"
)

// ResolveTaskReference resolves a user-supplied task reference into a task_uid.
// Supports direct task_uids and display IDs in the form <alias>-<number>[-<node_hint>].
func ResolveTaskReference(db *DB, ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("task reference cannot be empty")
	}

	// Direct task UID
	if strings.HasPrefix(ref, "tsk_") {
		var count int
		if err := db.Db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE task_uid = ?`, ref).Scan(&count); err != nil {
			return "", fmt.Errorf("failed to lookup task %s: %w", ref, err)
		}
		if count == 0 {
			return "", fmt.Errorf("task %s not found", ref)
		}
		return ref, nil
	}

	// Pure numeric references are ambiguous by design.
	if _, err := strconv.ParseInt(ref, 10, 64); err == nil {
		return "", fmt.Errorf("ambiguous task reference %s: numeric references are not supported", ref)
	}

	displayID := types.DisplayID(ref)
	alias, number, nodeHint, err := displayID.Parse()
	if err != nil {
		return "", fmt.Errorf("invalid task reference %s: %w", ref, err)
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
		return "", fmt.Errorf("failed to query tasks for %s: %w", ref, err)
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
		return "", fmt.Errorf("failed to iterate tasks for %s: %w", ref, err)
	}

	if len(candidateUIDs) == 0 {
		return "", fmt.Errorf("task %s not found", ref)
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
		return "", fmt.Errorf("node hint %s for %s does not match any task", nodeHint, ref)
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

	return "", fmt.Errorf("ambiguous task reference %s; candidates: %s", ref, strings.Join(displayIDs, ", "))
}

// RenderTaskDisplayID renders the preferred display string for a task.
func RenderTaskDisplayID(db *DB, taskUID string) (string, error) {
	var projectUID string
	var number int64
	err := db.Db.QueryRow(`
		SELECT project_uid, number FROM task_numbers WHERE task_uid = ?
	`, taskUID).Scan(&projectUID, &number)
	if err == sql.ErrNoRows {
		return taskUID, nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to load task number for %s: %w", taskUID, err)
	}

	alias, err := PreferredAliasForProject(db, types.ProjectUID(projectUID))
	if err != nil {
		return "", err
	}

	// If no alias exists, fall back to project name
	prefix := alias
	if prefix == "" {
		var projectName string
		err := db.Db.QueryRow(`
			SELECT name FROM projects WHERE project_uid = ?
		`, projectUID).Scan(&projectName)
		if err != nil {
			// If we can't get the name, fall back to UID
			prefix = projectUID
		} else {
			prefix = projectName
		}
	}

	collision, err := HasNumberCollision(db, projectUID, number)
	if err != nil {
		return "", err
	}
	if !collision {
		return fmt.Sprintf("%s-%d", prefix, number), nil
	}

	hint, err := taskNodeHint(db, taskUID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%d-%s", prefix, number, hint), nil
}

// ResolveProjectRef resolves an unresolved project reference (UID, alias, or name) to a ProjectUID
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

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return "", err
	}

	// First try to resolve by alias
	var projectUID string
	err = db.Db.QueryRow(`
		SELECT project_uid FROM project_aliases
		WHERE alias = ?
		ORDER BY CASE WHEN node = ? THEN 0 ELSE 1 END
		LIMIT 1
	`, ref.String(), nodeID).Scan(&projectUID)
	if err == nil {
		return types.ProjectUID(projectUID), nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to resolve project alias %s: %w", ref, err)
	}

	// If no alias found, try to resolve by project name
	err = db.Db.QueryRow(`
		SELECT project_uid FROM projects
		WHERE name = ?
		ORDER BY created_at
		LIMIT 1
	`, ref.String()).Scan(&projectUID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("project/alias %s not found (checked aliases and project names)", ref)
	}
	if err != nil {
		return "", fmt.Errorf("failed to resolve project name %s: %w", ref, err)
	}
	return types.ProjectUID(projectUID), nil
}

// ResolveProjectByAlias is deprecated - use ResolveProjectRef instead
// This is kept temporarily for backward compatibility during refactoring
func ResolveProjectByAlias(db *DB, alias string) (string, error) {
	uid, err := ResolveProjectRef(db, types.NewProjectRef(alias))
	return string(uid), err
}

func PreferredAliasForProject(db *DB, projectUID types.ProjectUID) (string, error) {
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return "", err
	}

	var alias sql.NullString
	err = db.Db.QueryRow(`
		SELECT alias FROM project_aliases
		WHERE project_uid = ?
		ORDER BY CASE WHEN node = ? THEN 0 ELSE 1 END
		LIMIT 1
	`, projectUID.String(), nodeID).Scan(&alias)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to resolve alias for project %s: %w", projectUID, err)
	}
	if alias.Valid {
		return alias.String, nil
	}
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

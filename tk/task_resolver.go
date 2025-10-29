package main

import (
	"database/sql"
	"fmt"
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
		if err := db.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE task_uid = ?`, ref).Scan(&count); err != nil {
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

	displayID := DisplayID(ref)
	alias, number, nodeHint, err := displayID.Parse()
	if err != nil {
		return "", fmt.Errorf("invalid task reference %s: %w", ref, err)
	}

	projectUID, err := resolveProjectByAlias(db, alias)
	if err != nil {
		return "", err
	}

	rows, err := db.db.Query(`
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
	err := db.db.QueryRow(`
		SELECT project_uid, number FROM task_numbers WHERE task_uid = ?
	`, taskUID).Scan(&projectUID, &number)
	if err == sql.ErrNoRows {
		return taskUID, nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to load task number for %s: %w", taskUID, err)
	}

	alias, err := preferredAliasForProject(db, projectUID)
	if err != nil {
		return "", err
	}
	if alias == "" {
		return fmt.Sprintf("%s-%d", projectUID, number), nil
	}

	collision, err := hasNumberCollision(db, projectUID, number)
	if err != nil {
		return "", err
	}
	if !collision {
		return fmt.Sprintf("%s-%d", alias, number), nil
	}

	hint, err := taskNodeHint(db, taskUID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%d-%s", alias, number, hint), nil
}

func resolveProjectByAlias(db *DB, alias string) (string, error) {
	if strings.HasPrefix(alias, "prj_") {
		var count int
		if err := db.db.QueryRow(`SELECT COUNT(*) FROM projects WHERE project_uid = ?`, alias).Scan(&count); err != nil {
			return "", fmt.Errorf("failed to lookup project %s: %w", alias, err)
		}
		if count == 0 {
			return "", fmt.Errorf("project %s not found", alias)
		}
		return alias, nil
	}

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return "", err
	}

	// First try to resolve by alias
	var projectUID string
	err = db.db.QueryRow(`
		SELECT project_uid FROM project_aliases
		WHERE alias = ?
		ORDER BY CASE WHEN node = ? THEN 0 ELSE 1 END
		LIMIT 1
	`, alias, nodeID).Scan(&projectUID)
	if err == nil {
		return projectUID, nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to resolve project alias %s: %w", alias, err)
	}

	// If no alias found, try to resolve by project name
	err = db.db.QueryRow(`
		SELECT project_uid FROM projects
		WHERE name = ?
		ORDER BY created_at
		LIMIT 1
	`, alias).Scan(&projectUID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("project/alias %s not found (checked aliases and project names)", alias)
	}
	if err != nil {
		return "", fmt.Errorf("failed to resolve project name %s: %w", alias, err)
	}
	return projectUID, nil
}

func preferredAliasForProject(db *DB, projectUID string) (string, error) {
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return "", err
	}

	var alias sql.NullString
	err = db.db.QueryRow(`
		SELECT alias FROM project_aliases
		WHERE project_uid = ?
		ORDER BY CASE WHEN node = ? THEN 0 ELSE 1 END
		LIMIT 1
	`, projectUID, nodeID).Scan(&alias)
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

func hasNumberCollision(db *DB, projectUID string, number int64) (bool, error) {
	var count int
	if err := db.db.QueryRow(`
		SELECT COUNT(*) FROM task_numbers
		WHERE project_uid = ? AND number = ?
	`, projectUID, number).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check collisions: %w", err)
	}
	return count > 1, nil
}

func taskNodeHint(db *DB, taskUID string) (string, error) {
	var createdNode string
	if err := db.db.QueryRow(`SELECT created_node FROM tasks WHERE task_uid = ?`, taskUID).Scan(&createdNode); err != nil {
		return "", fmt.Errorf("failed to lookup node for task %s: %w", taskUID, err)
	}
	return NodeID(createdNode).Short(), nil
}

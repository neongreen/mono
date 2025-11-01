package main

import (
	"fmt"
	"strconv"
	"strings"
)

// ResolveTaskIDToUUID resolves a task reference to its UUID (legacy helper).
func (d *DB) ResolveTaskIDToUUID(taskID string) (string, error) {
	return ResolveTaskReference(d, taskID)
}

// ResolveTaskID resolves a short task ID to a full task ID
// Accepts formats: "1", "tk-1", "foo-2", "tk-1-abc123"
// Returns an error if the ID is ambiguous or doesn't exist
func (d *DB) ResolveTaskID(shortID string) (string, error) {

	hyphenCount := strings.Count(shortID, "-")
	if hyphenCount >= 2 {
		// Verify it exists
		var count int
		err := d.db.QueryRow(`
			SELECT COUNT(*)
			FROM events
			WHERE kind = 'task.created' AND json_extract(payload, '$.task_id') = ?
		`, shortID).Scan(&count)
		if err != nil {
			return "", fmt.Errorf("failed to query task ID: %w", err)
		}
		if count > 0 {
			return shortID, nil
		}
		return "", fmt.Errorf("task not found: %s", shortID)
	}

	if _, err := strconv.Atoi(shortID); err == nil {

		query := `
			SELECT DISTINCT json_extract(payload, '$.task_id') as task_id
			FROM events
			WHERE kind = 'task.created'
			  AND json_extract(payload, '$.task_id') LIKE '%-' || ? || '-%'
		`

		rows, err := d.db.Query(query, shortID)
		if err != nil {
			return "", fmt.Errorf("failed to query task IDs: %w", err)
		}
		defer rows.Close()

		var matches []string
		for rows.Next() {
			var taskID string
			if err := rows.Scan(&taskID); err != nil {
				return "", fmt.Errorf("failed to scan task ID: %w", err)
			}
			matches = append(matches, taskID)
		}

		if len(matches) == 0 {
			return "", fmt.Errorf("task not found: %s", shortID)
		}

		prefixNumberMap := make(map[string][]string)
		for _, taskID := range matches {
			parts := strings.Split(taskID, "-")
			if len(parts) >= 2 {
				prefixNumber := strings.Join(parts[:2], "-")
				prefixNumberMap[prefixNumber] = append(prefixNumberMap[prefixNumber], taskID)
			}
		}

		if len(prefixNumberMap) > 1 {
			prefixes := make([]string, 0, len(prefixNumberMap))
			for pn := range prefixNumberMap {
				prefixes = append(prefixes, pn)
			}
			return "", fmt.Errorf("ambiguous task ID %s (matches %v) — use <prefix>-%s instead", shortID, prefixes, shortID)
		}

		if len(matches) == 1 {
			return matches[0], nil
		}

		// Multiple matches with same prefix-number but different nodes - this is ambiguous
		// Extract the prefix-number for error message
		var prefixNumber string
		for pn := range prefixNumberMap {
			prefixNumber = pn
			break
		}
		return "", fmt.Errorf("ambiguous task ID %s (multiple nodes created %s) — use full ID like %s", shortID, prefixNumber, matches[0])
	}

	query := `
		SELECT DISTINCT json_extract(payload, '$.task_id') as task_id
		FROM events
		WHERE kind = 'task.created'
		  AND json_extract(payload, '$.task_id') LIKE ? || '-%'
	`

	rows, err := d.db.Query(query, shortID)
	if err != nil {
		return "", fmt.Errorf("failed to query task IDs: %w", err)
	}
	defer rows.Close()

	var matches []string
	for rows.Next() {
		var taskID string
		if err := rows.Scan(&taskID); err != nil {
			return "", fmt.Errorf("failed to scan task ID: %w", err)
		}

		parts := strings.Split(taskID, "-")
		if len(parts) >= 2 {
			shortForm := strings.Join(parts[:2], "-")
			if shortForm == shortID {
				matches = append(matches, taskID)
			}
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("task not found: %s", shortID)
	}

	if len(matches) > 1 {
		return "", fmt.Errorf("ambiguous task ID %s (multiple nodes created %s) — use full ID like %s", shortID, shortID, matches[0])
	}

	return matches[0], nil
}

// GetAllTaskIDs returns all task IDs in the database
func (d *DB) GetAllTaskIDs() ([]string, error) {

	query := `SELECT DISTINCT task_uid FROM tasks ORDER BY created_at`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query task UIDs: %w", err)
	}
	defer rows.Close()

	var taskUIDs []string
	for rows.Next() {
		var taskUID string
		if err := rows.Scan(&taskUID); err != nil {
			return nil, fmt.Errorf("failed to scan task UID: %w", err)
		}
		taskUIDs = append(taskUIDs, taskUID)
	}
	return taskUIDs, rows.Err()
}

// GetTaskIDsByProjects returns task IDs filtered by project identifiers (alias, name, or UID).
func (d *DB) GetTaskIDsByProjects(projects []string) ([]string, error) {
	if len(projects) == 0 {
		return d.GetAllTaskIDs()
	}

	seen := make(map[string]struct{})
	projectUIDs := make([]string, 0, len(projects))
	for _, spec := range projects {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			return nil, fmt.Errorf("project filter cannot be empty")
		}

		projectUID, err := resolveProjectByAlias(d, spec)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve project %s: %w", spec, err)
		}
		if _, exists := seen[projectUID]; exists {
			continue
		}
		seen[projectUID] = struct{}{}
		projectUIDs = append(projectUIDs, projectUID)
	}

	if len(projectUIDs) == 0 {
		return []string{}, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(projectUIDs)), ",")
	query := fmt.Sprintf(`
		SELECT DISTINCT task_uid
		FROM tasks
		WHERE project_uid IN (%s)
		ORDER BY created_at
	`, placeholders)

	args := make([]interface{}, len(projectUIDs))
	for i, projectUID := range projectUIDs {
		args[i] = projectUID
	}

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query task UIDs by project: %w", err)
	}
	defer rows.Close()

	var taskUIDs []string
	for rows.Next() {
		var taskUID string
		if err := rows.Scan(&taskUID); err != nil {
			return nil, fmt.Errorf("failed to scan task UID: %w", err)
		}
		taskUIDs = append(taskUIDs, taskUID)
	}
	return taskUIDs, rows.Err()
}

// FormatTaskID formats a task ID for display, hiding the suffix unless needed for disambiguation.
//
// The function returns the full ID unchanged in the following error cases:
//   - Malformed ID that cannot be parsed
//   - Alias that cannot be resolved to a project UID
//   - Database error when checking for collisions
//
// This ensures that task IDs are always displayable even when errors occur,
// at the cost of potentially showing longer IDs than necessary.
func FormatTaskID(db *DB, fullID string) string {
	// Parse the display ID to extract alias and number
	displayID := DisplayID(fullID)
	alias, number, _, err := displayID.Parse()
	if err != nil {
		// Malformed ID, return as-is
		return fullID
	}

	// Resolve the alias to get project_uid
	projectUID, err := resolveProjectByAlias(db, alias)
	if err != nil {
		// Cannot resolve alias, return as-is
		return fullID
	}

	// Check if there's a collision for this project/number combination
	collision, err := hasNumberCollision(db, projectUID, number)
	if err != nil {
		// Error checking collision, return as-is
		return fullID
	}

	// If collision exists, return full ID with suffix
	if collision {
		return fullID
	}

	// No collision, return short form (alias-number)
	return fmt.Sprintf("%s-%d", alias, number)
}

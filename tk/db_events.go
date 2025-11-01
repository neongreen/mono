package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

// InsertEvent adds an event to the database
func (d *DB) InsertEvent(e types.Event) error {
	query := `
		INSERT INTO events (id, ts, created_at, actor, role, kind, payload, ctx, repo_uuid, branch, commit_sha, jj_op_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := d.db.Exec(query, e.ID, e.TS, e.CreatedAt.UnixNano(), e.Actor, e.Role, e.Kind, e.Payload, e.Ctx, e.RepoUUID, e.Branch, e.Commit, e.JJOpID)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	d.reducerCache = nil
	d.reducerConfig = nil

	return nil
}

// GetEvents retrieves all events in chronological order
func (d *DB) GetEvents() ([]types.Event, error) {
	query := `SELECT id, ts, created_at, actor, role, kind, payload, ctx, repo_uuid, branch, commit_sha, jj_op_id
	          FROM events ORDER BY created_at, id`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []types.Event
	for rows.Next() {
		var e types.Event
		var ctx, repoUUID, branch, commit, jjOpID sql.NullString
		var createdAtNano int64

		err := rows.Scan(&e.ID, &e.TS, &createdAtNano, &e.Actor, &e.Role, &e.Kind, &e.Payload, &ctx, &repoUUID, &branch, &commit, &jjOpID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		e.CreatedAt = time.Unix(0, createdAtNano)
		if ctx.Valid {
			e.Ctx = json.RawMessage(ctx.String)
		}
		if repoUUID.Valid {
			e.RepoUUID = repoUUID.String
		}
		if branch.Valid {
			e.Branch = branch.String
		}
		if commit.Valid {
			e.Commit = commit.String
		}
		if jjOpID.Valid {
			e.JJOpID = jjOpID.String
		}

		events = append(events, e)
	}

	return events, rows.Err()
}

// GetEventsByTaskID retrieves events for a specific task
func (d *DB) GetEventsByTaskID(taskID string) ([]types.Event, error) {
	query := `
		SELECT id, ts, created_at, actor, role, kind, payload, ctx, repo_uuid, branch, commit_sha, jj_op_id
		FROM events
		WHERE json_extract(payload, '$.task_id') = ?
		ORDER BY created_at, id
	`

	rows, err := d.db.Query(query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events for task %s: %w", taskID, err)
	}
	defer rows.Close()

	var events []types.Event
	for rows.Next() {
		var e types.Event
		var ctx, repoUUID, branch, commit, jjOpID sql.NullString
		var createdAtNano int64

		err := rows.Scan(&e.ID, &e.TS, &createdAtNano, &e.Actor, &e.Role, &e.Kind, &e.Payload, &ctx, &repoUUID, &branch, &commit, &jjOpID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		e.CreatedAt = time.Unix(0, createdAtNano)
		if ctx.Valid {
			e.Ctx = json.RawMessage(ctx.String)
		}
		if repoUUID.Valid {
			e.RepoUUID = repoUUID.String
		}
		if branch.Valid {
			e.Branch = branch.String
		}
		if commit.Valid {
			e.Commit = commit.String
		}
		if jjOpID.Valid {
			e.JJOpID = jjOpID.String
		}

		events = append(events, e)
	}

	return events, rows.Err()
}

// GetEventsByTaskUUID retrieves events for a specific task UUID
func (d *DB) GetEventsByTaskUUID(taskUUID string) ([]types.Event, error) {
	query := `
		SELECT id, ts, created_at, actor, role, kind, payload, ctx, repo_uuid, branch, commit_sha, jj_op_id
		FROM events
		WHERE json_extract(payload, '$.task_uuid') = ?
		   OR json_extract(payload, '$.task_id') = ?
		ORDER BY created_at, id
	`

	rows, err := d.db.Query(query, taskUUID, taskUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events for task UUID %s: %w", taskUUID, err)
	}
	defer rows.Close()

	var events []types.Event
	for rows.Next() {
		var e types.Event
		var ctx, repoUUID, branch, commit, jjOpID sql.NullString
		var createdAtNano int64

		err := rows.Scan(&e.ID, &e.TS, &createdAtNano, &e.Actor, &e.Role, &e.Kind, &e.Payload, &ctx, &repoUUID, &branch, &commit, &jjOpID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		e.CreatedAt = time.Unix(0, createdAtNano)
		if ctx.Valid {
			e.Ctx = json.RawMessage(ctx.String)
		}
		if repoUUID.Valid {
			e.RepoUUID = repoUUID.String
		}
		if branch.Valid {
			e.Branch = branch.String
		}
		if commit.Valid {
			e.Commit = commit.String
		}
		if jjOpID.Valid {
			e.JJOpID = jjOpID.String
		}

		events = append(events, e)
	}

	return events, rows.Err()
}

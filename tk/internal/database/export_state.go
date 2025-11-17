package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ExportState tracks export state per remote/space
type ExportState struct {
	RemoteName          string    `json:"remote_name"`
	Space               string    `json:"space"`
	LastExportedEventID string    `json:"last_exported_event_id"`
	SegmentSeq          int64     `json:"segment_seq"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// GetExportState retrieves the export state for a remote/space combination
func (d *DB) GetExportState(remoteName, space string) (*ExportState, error) {
	var state ExportState
	var updatedAtUnix int64

	err := d.Db.QueryRow(`
		SELECT remote_name, space, last_exported_event_id, segment_seq, updated_at
		FROM export_state
		WHERE remote_name = ? AND space = ?
	`, remoteName, space).Scan(
		&state.RemoteName,
		&state.Space,
		&state.LastExportedEventID,
		&state.SegmentSeq,
		&updatedAtUnix,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // No state found, return nil without error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get export state: %w", err)
	}

	state.UpdatedAt = time.Unix(updatedAtUnix, 0)
	return &state, nil
}

// SaveExportState saves or updates the export state for a remote/space combination
func (d *DB) SaveExportState(state *ExportState) error {
	updatedAtUnix := state.UpdatedAt.Unix()

	_, err := d.Db.Exec(`
		INSERT OR REPLACE INTO export_state (remote_name, space, last_exported_event_id, segment_seq, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, state.RemoteName, state.Space, state.LastExportedEventID, state.SegmentSeq, updatedAtUnix)

	if err != nil {
		return fmt.Errorf("failed to save export state: %w", err)
	}

	return nil
}

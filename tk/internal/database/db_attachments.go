package database

import (
	"database/sql"
	"fmt"
	"time"
)

// InsertAttachment inserts an attachment blob into the database
func (d *DB) InsertAttachment(hash string, content []byte, mimeType string, createdBy string) error {
	_, err := d.Db.Exec(`
		INSERT OR IGNORE INTO attachments (hash, content, mime_type, size, created_at, created_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`, hash, content, mimeType, int64(len(content)), time.Now().Format(time.RFC3339), createdBy)
	if err != nil {
		return fmt.Errorf("failed to insert attachment: %w", err)
	}
	return nil
}

// GetAttachment retrieves an attachment by hash
func (d *DB) GetAttachment(hash string) (content []byte, mimeType string, err error) {
	err = d.Db.QueryRow(`
		SELECT content, mime_type FROM attachments WHERE hash = ?
	`, hash).Scan(&content, &mimeType)
	if err == sql.ErrNoRows {
		return nil, "", fmt.Errorf("attachment not found: %s", hash)
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to get attachment: %w", err)
	}
	return content, mimeType, nil
}

// AttachmentExists checks if an attachment with the given hash exists
func (d *DB) AttachmentExists(hash string) (bool, error) {
	var exists bool
	err := d.Db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM attachments WHERE hash = ?)
	`, hash).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check attachment existence: %w", err)
	}
	return exists, nil
}

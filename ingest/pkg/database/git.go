package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// CreateCommit creates a new commit record
func (d *Database) CreateCommit(runID int64, hash, author, authorEmail, committer, committerEmail string, date time.Time, message string, parentHashes []string) (int64, error) {
	result, err := d.db.Exec(
		"INSERT INTO commits (run_id, hash, author, author_email, committer, committer_email, date, message) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		runID,
		hash,
		author,
		authorEmail,
		committer,
		committerEmail,
		date,
		message,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create commit: %w", err)
	}

	commitID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get commit ID: %w", err)
	}

	for _, parentHash := range parentHashes {
		_, err := d.db.Exec(
			"INSERT INTO commit_parents (commit_id, parent_hash) VALUES (?, ?)",
			commitID,
			parentHash,
		)
		if err != nil {
			return 0, fmt.Errorf("failed to create commit parent: %w", err)
		}
	}

	return commitID, nil
}

// GetOrCreateBlob stores a blob if it doesn't exist and returns its ID
func (d *Database) GetOrCreateBlob(content []byte, sha256Hash string) (int64, error) {
	// Check if blob already exists
	var blobID int64
	err := d.db.QueryRow("SELECT id FROM blobs WHERE sha256 = ?", sha256Hash).Scan(&blobID)
	if err == nil {
		return blobID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("failed to check for existing blob: %w", err)
	}

	result, err := d.db.Exec(
		"INSERT INTO blobs (sha256, content, size) VALUES (?, ?, ?)",
		sha256Hash,
		content,
		len(content),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create blob: %w", err)
	}

	return result.LastInsertId()
}

// CreateFile creates a new file record
func (d *Database) CreateFile(commitID int64, path string, size int64, mode string, blobID *int64) error {
	_, err := d.db.Exec(
		"INSERT INTO files (commit_id, path, size, mode, blob_id) VALUES (?, ?, ?, ?, ?)",
		commitID,
		path,
		size,
		mode,
		blobID,
	)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

// CreateFSEntry creates a new filesystem entry record
func (d *Database) CreateFSEntry(runID int64, path string, isDir bool, size int64, mode string, modTime time.Time, blobID *int64) error {
	_, err := d.db.Exec(
		"INSERT INTO fs_entries (run_id, path, is_dir, size, mode, mod_time, blob_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		runID,
		path,
		isDir,
		size,
		mode,
		modTime,
		blobID,
	)
	if err != nil {
		return fmt.Errorf("failed to create fs entry: %w", err)
	}

	return nil
}

// CreateGitRef creates a new git reference record (branch or tag)
func (d *Database) CreateGitRef(runID int64, refType, name, targetHash string) error {
	_, err := d.db.Exec(
		"INSERT INTO git_refs (run_id, ref_type, name, target_hash) VALUES (?, ?, ?, ?)",
		runID,
		refType,
		name,
		targetHash,
	)
	if err != nil {
		return fmt.Errorf("failed to create git ref: %w", err)
	}

	return nil
}

// CreateGitRemote creates a new git remote record
func (d *Database) CreateGitRemote(runID int64, name, url string) error {
	_, err := d.db.Exec(
		"INSERT INTO git_remotes (run_id, name, url) VALUES (?, ?, ?)",
		runID,
		name,
		url,
	)
	if err != nil {
		return fmt.Errorf("failed to create git remote: %w", err)
	}

	return nil
}

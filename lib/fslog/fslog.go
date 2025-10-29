package fslog

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// FileSystem provides ACID-ish semantics over filesystem operations.
// It maintains an immutable operation log and supports rollback to any past state.
type FileSystem struct {
	mu      sync.RWMutex
	logDir  string
	log     *OperationLog
	basedir string
}

// New creates a new FileSystem that operates on basedir and stores its log in logDir.
// If logDir is empty, it defaults to basedir/.fslog
func New(basedir string, logDir string) (*FileSystem, error) {
	if basedir == "" {
		return nil, fmt.Errorf("basedir cannot be empty")
	}

	absBasedir, err := filepath.Abs(basedir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for basedir: %w", err)
	}

	if logDir == "" {
		logDir = filepath.Join(absBasedir, ".fslog")
	}

	absLogDir, err := filepath.Abs(logDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for logDir: %w", err)
	}

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(absLogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	log, err := OpenOperationLog(absLogDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open operation log: %w", err)
	}

	return &FileSystem{
		logDir:  absLogDir,
		log:     log,
		basedir: absBasedir,
	}, nil
}

// Close closes the filesystem and flushes any pending operations.
func (fs *FileSystem) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.log.Close()
}

// Begin starts a new transaction.
func (fs *FileSystem) Begin(ctx context.Context) *Transaction {
	return &Transaction{
		fs:         fs,
		ctx:        ctx,
		operations: make([]*Operation, 0),
		startTime:  time.Now(),
	}
}

// History returns all operations in chronological order.
func (fs *FileSystem) History() ([]*Operation, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.log.All()
}

// RollbackTo rolls back the filesystem to the state after the given operation ID.
// If opID is 0, rolls back to the initial state (before any operations).
func (fs *FileSystem) RollbackTo(opID int64) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	ops, err := fs.log.All()
	if err != nil {
		return fmt.Errorf("failed to get operation history: %w", err)
	}

	// Find operations to reverse
	var toReverse []*Operation
	for i := len(ops) - 1; i >= 0; i-- {
		if ops[i].ID > opID {
			toReverse = append(toReverse, ops[i])
		} else {
			break
		}
	}

	// Apply reversals
	for _, op := range toReverse {
		if err := fs.reverseOperation(op); err != nil {
			return fmt.Errorf("failed to reverse operation %d: %w", op.ID, err)
		}
	}

	return nil
}

// reverseOperation reverses a single operation by restoring the before state.
func (fs *FileSystem) reverseOperation(op *Operation) error {
	targetPath := filepath.Join(fs.basedir, op.Path)

	switch op.Type {
	case OpWrite, OpCreate:
		if op.BeforeExists {
			// Restore previous content
			if err := os.WriteFile(targetPath, op.BeforeContent, op.BeforeMode); err != nil {
				return fmt.Errorf("failed to restore file content: %w", err)
			}
		} else {
			// Delete the file
			if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove file: %w", err)
			}
		}

	case OpDelete:
		// Restore deleted file
		if err := os.WriteFile(targetPath, op.BeforeContent, op.BeforeMode); err != nil {
			return fmt.Errorf("failed to restore deleted file: %w", err)
		}

	case OpMkdir:
		// Remove created directory
		if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove directory: %w", err)
		}

	case OpRename:
		// Reverse the rename
		oldPath := filepath.Join(fs.basedir, op.Metadata["old_path"])
		if err := os.Rename(targetPath, oldPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to reverse rename: %w", err)
		}
	}

	return nil
}

// Transaction represents a group of filesystem operations that are applied atomically.
type Transaction struct {
	fs         *FileSystem
	ctx        context.Context
	operations []*Operation
	startTime  time.Time
	committed  bool
}

// OpType represents the type of filesystem operation.
type OpType string

const (
	OpWrite  OpType = "write"  // Write to existing file
	OpCreate OpType = "create" // Create new file
	OpDelete OpType = "delete" // Delete file
	OpMkdir  OpType = "mkdir"  // Create directory
	OpRename OpType = "rename" // Rename/move file
)

// WriteFile writes content to a file within the transaction.
func (tx *Transaction) WriteFile(path string, content []byte, mode os.FileMode) error {
	if tx.committed {
		return fmt.Errorf("transaction already committed")
	}

	targetPath := filepath.Join(tx.fs.basedir, path)

	// Record before state
	var beforeContent []byte
	var beforeMode os.FileMode
	var beforeExists bool

	info, err := os.Stat(targetPath)
	if err == nil {
		beforeExists = true
		beforeMode = info.Mode()
		beforeContent, err = os.ReadFile(targetPath)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	opType := OpWrite
	if !beforeExists {
		opType = OpCreate
	}

	op := &Operation{
		Type:          opType,
		Path:          path,
		BeforeExists:  beforeExists,
		BeforeContent: beforeContent,
		BeforeMode:    beforeMode,
		AfterContent:  content,
		AfterMode:     mode,
		Timestamp:     time.Now(),
		Metadata:      make(map[string]string),
	}

	tx.operations = append(tx.operations, op)
	return nil
}

// DeleteFile deletes a file within the transaction.
func (tx *Transaction) DeleteFile(path string) error {
	if tx.committed {
		return fmt.Errorf("transaction already committed")
	}

	targetPath := filepath.Join(tx.fs.basedir, path)

	// Record before state
	var beforeContent []byte
	var beforeMode os.FileMode

	info, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Already doesn't exist
		}
		return fmt.Errorf("failed to stat file: %w", err)
	}

	beforeMode = info.Mode()
	beforeContent, err = os.ReadFile(targetPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	op := &Operation{
		Type:          OpDelete,
		Path:          path,
		BeforeExists:  true,
		BeforeContent: beforeContent,
		BeforeMode:    beforeMode,
		Timestamp:     time.Now(),
		Metadata:      make(map[string]string),
	}

	tx.operations = append(tx.operations, op)
	return nil
}

// Mkdir creates a directory within the transaction.
func (tx *Transaction) Mkdir(path string, mode os.FileMode) error {
	if tx.committed {
		return fmt.Errorf("transaction already committed")
	}

	op := &Operation{
		Type:      OpMkdir,
		Path:      path,
		AfterMode: mode,
		Timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}

	tx.operations = append(tx.operations, op)
	return nil
}

// Rename renames/moves a file within the transaction.
func (tx *Transaction) Rename(oldPath, newPath string) error {
	if tx.committed {
		return fmt.Errorf("transaction already committed")
	}

	targetPath := filepath.Join(tx.fs.basedir, oldPath)

	// Record before state
	var beforeContent []byte
	var beforeMode os.FileMode

	info, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	beforeMode = info.Mode()
	if !info.IsDir() {
		beforeContent, err = os.ReadFile(targetPath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
	}

	op := &Operation{
		Type:          OpRename,
		Path:          newPath,
		BeforeExists:  true,
		BeforeContent: beforeContent,
		BeforeMode:    beforeMode,
		AfterContent:  beforeContent,
		AfterMode:     beforeMode,
		Timestamp:     time.Now(),
		Metadata: map[string]string{
			"old_path": oldPath,
		},
	}

	tx.operations = append(tx.operations, op)
	return nil
}

// Commit applies all operations in the transaction atomically.
func (tx *Transaction) Commit() error {
	if tx.committed {
		return fmt.Errorf("transaction already committed")
	}

	tx.fs.mu.Lock()
	defer tx.fs.mu.Unlock()

	// Apply all operations
	for _, op := range tx.operations {
		targetPath := filepath.Join(tx.fs.basedir, op.Path)

		switch op.Type {
		case OpWrite, OpCreate:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}
			if err := os.WriteFile(targetPath, op.AfterContent, op.AfterMode); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

		case OpDelete:
			if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to delete file: %w", err)
			}

		case OpMkdir:
			if err := os.MkdirAll(targetPath, op.AfterMode); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

		case OpRename:
			oldPath := filepath.Join(tx.fs.basedir, op.Metadata["old_path"])
			if err := os.Rename(oldPath, targetPath); err != nil {
				return fmt.Errorf("failed to rename: %w", err)
			}
		}

		// Record operation in log
		if err := tx.fs.log.Append(op); err != nil {
			return fmt.Errorf("failed to log operation: %w", err)
		}
	}

	tx.committed = true
	return nil
}

// Rollback discards all operations in the transaction without applying them.
func (tx *Transaction) Rollback() {
	tx.committed = true
	tx.operations = nil
}

// Operation represents a single filesystem operation with before/after state.
type Operation struct {
	ID            int64             `json:"id"`
	Type          OpType            `json:"type"`
	Path          string            `json:"path"`
	BeforeExists  bool              `json:"before_exists"`
	BeforeContent []byte            `json:"before_content,omitempty"`
	BeforeMode    os.FileMode       `json:"before_mode,omitempty"`
	AfterContent  []byte            `json:"after_content,omitempty"`
	AfterMode     os.FileMode       `json:"after_mode,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// Diff returns a human-readable diff of the operation.
func (op *Operation) Diff() string {
	switch op.Type {
	case OpCreate:
		return fmt.Sprintf("Created %s (%d bytes)", op.Path, len(op.AfterContent))
	case OpWrite:
		return fmt.Sprintf("Modified %s (%d -> %d bytes)", op.Path, len(op.BeforeContent), len(op.AfterContent))
	case OpDelete:
		return fmt.Sprintf("Deleted %s (%d bytes)", op.Path, len(op.BeforeContent))
	case OpMkdir:
		return fmt.Sprintf("Created directory %s", op.Path)
	case OpRename:
		return fmt.Sprintf("Renamed %s -> %s", op.Metadata["old_path"], op.Path)
	default:
		return fmt.Sprintf("Unknown operation on %s", op.Path)
	}
}

// OperationLog is an append-only, immutable log of filesystem operations stored in SQLite.
type OperationLog struct {
	mu     sync.RWMutex
	logDir string
	db     *sql.DB
}

// OpenOperationLog opens or creates an operation log in the given directory.
func OpenOperationLog(logDir string) (*OperationLog, error) {
	dbPath := filepath.Join(logDir, "operations.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create table if it doesn't exist
	schema := `
	CREATE TABLE IF NOT EXISTS operations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		path TEXT NOT NULL,
		before_exists INTEGER NOT NULL,
		before_content BLOB,
		before_mode INTEGER,
		after_content BLOB,
		after_mode INTEGER,
		timestamp TEXT NOT NULL,
		metadata TEXT
	);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON operations(timestamp);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &OperationLog{
		logDir: logDir,
		db:     db,
	}, nil
}

// Append adds an operation to the log.
func (ol *OperationLog) Append(op *Operation) error {
	ol.mu.Lock()
	defer ol.mu.Unlock()

	// Marshal metadata to JSON
	var metadataJSON []byte
	var err error
	if len(op.Metadata) > 0 {
		metadataJSON, err = json.Marshal(op.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
	INSERT INTO operations (type, path, before_exists, before_content, before_mode, 
	                        after_content, after_mode, timestamp, metadata)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := ol.db.Exec(query,
		string(op.Type),
		op.Path,
		op.BeforeExists,
		op.BeforeContent,
		int64(op.BeforeMode),
		op.AfterContent,
		int64(op.AfterMode),
		op.Timestamp.Format(time.RFC3339Nano),
		metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to insert operation: %w", err)
	}

	// Get the auto-incremented ID
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get operation ID: %w", err)
	}
	op.ID = id

	return nil
}

// All returns all operations in chronological order.
func (ol *OperationLog) All() ([]*Operation, error) {
	ol.mu.RLock()
	defer ol.mu.RUnlock()

	query := `
	SELECT id, type, path, before_exists, before_content, before_mode,
	       after_content, after_mode, timestamp, metadata
	FROM operations
	ORDER BY id ASC
	`

	rows, err := ol.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query operations: %w", err)
	}
	defer rows.Close()

	var ops []*Operation
	for rows.Next() {
		op := &Operation{
			Metadata: make(map[string]string),
		}

		var opType string
		var beforeMode, afterMode int64
		var timestampStr string
		var metadataJSON []byte

		err := rows.Scan(
			&op.ID,
			&opType,
			&op.Path,
			&op.BeforeExists,
			&op.BeforeContent,
			&beforeMode,
			&op.AfterContent,
			&afterMode,
			&timestampStr,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan operation: %w", err)
		}

		op.Type = OpType(opType)
		op.BeforeMode = os.FileMode(beforeMode)
		op.AfterMode = os.FileMode(afterMode)

		// Parse timestamp
		op.Timestamp, err = time.Parse(time.RFC3339Nano, timestampStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}

		// Unmarshal metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &op.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		ops = append(ops, op)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating operations: %w", err)
	}

	return ops, nil
}

// Close closes the operation log.
func (ol *OperationLog) Close() error {
	ol.mu.Lock()
	defer ol.mu.Unlock()

	if ol.db != nil {
		return ol.db.Close()
	}
	return nil
}

package fslog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
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

// OperationLog is an append-only, immutable log of filesystem operations.
type OperationLog struct {
	mu      sync.RWMutex
	logDir  string
	logFile *os.File
	nextID  int64
}

// OpenOperationLog opens or creates an operation log in the given directory.
func OpenOperationLog(logDir string) (*OperationLog, error) {
	logPath := filepath.Join(logDir, "operations.jsonl")

	// Try to open existing log
	var nextID int64 = 1
	if _, err := os.Stat(logPath); err == nil {
		// Read existing log to find next ID
		ops, err := readOperationLog(logPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read existing log: %w", err)
		}
		for _, op := range ops {
			if op.ID >= nextID {
				nextID = op.ID + 1
			}
		}
	}

	// Open log file for appending
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &OperationLog{
		logDir:  logDir,
		logFile: logFile,
		nextID:  nextID,
	}, nil
}

// readOperationLog reads all operations from a log file.
func readOperationLog(path string) ([]*Operation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ops []*Operation
	decoder := json.NewDecoder(bytes.NewReader(data))
	for decoder.More() {
		var op Operation
		if err := decoder.Decode(&op); err != nil {
			return nil, fmt.Errorf("failed to decode operation: %w", err)
		}
		ops = append(ops, &op)
	}

	return ops, nil
}

// Append adds an operation to the log.
func (ol *OperationLog) Append(op *Operation) error {
	ol.mu.Lock()
	defer ol.mu.Unlock()

	op.ID = ol.nextID
	ol.nextID++

	data, err := json.Marshal(op)
	if err != nil {
		return fmt.Errorf("failed to marshal operation: %w", err)
	}

	if _, err := ol.logFile.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to log: %w", err)
	}

	if err := ol.logFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync log: %w", err)
	}

	return nil
}

// All returns all operations in chronological order.
func (ol *OperationLog) All() ([]*Operation, error) {
	ol.mu.RLock()
	defer ol.mu.RUnlock()

	logPath := filepath.Join(ol.logDir, "operations.jsonl")
	return readOperationLog(logPath)
}

// Close closes the operation log.
func (ol *OperationLog) Close() error {
	ol.mu.Lock()
	defer ol.mu.Unlock()

	if ol.logFile != nil {
		return ol.logFile.Close()
	}
	return nil
}

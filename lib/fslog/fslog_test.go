package fslog

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()

	fs, err := New(tmpDir, "")
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer fs.Close()

	// Check that log directory was created
	logDir := filepath.Join(tmpDir, ".fslog")
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Errorf("Log directory was not created")
	}
}

func TestTransaction_WriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := New(tmpDir, "")
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer fs.Close()

	// Create and commit a transaction
	tx := fs.Begin(context.Background())
	content := []byte("hello world")
	if err := tx.WriteFile("test.txt", content, 0644); err != nil {
		t.Fatalf("Failed to write file in transaction: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}

	// Verify file was created
	filePath := filepath.Join(tmpDir, "test.txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(data) != "hello world" {
		t.Errorf("File content mismatch: got %q, want %q", string(data), "hello world")
	}
}

func TestTransaction_ModifyFile(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := New(tmpDir, "")
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer fs.Close()

	// Create initial file
	tx1 := fs.Begin(context.Background())
	if err := tx1.WriteFile("test.txt", []byte("version 1"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if err := tx1.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Modify file
	tx2 := fs.Begin(context.Background())
	if err := tx2.WriteFile("test.txt", []byte("version 2"), 0644); err != nil {
		t.Fatalf("Failed to modify file: %v", err)
	}
	if err := tx2.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Verify modification
	filePath := filepath.Join(tmpDir, "test.txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(data) != "version 2" {
		t.Errorf("File content mismatch: got %q, want %q", string(data), "version 2")
	}

	// Check operation log
	ops, err := fs.History()
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(ops) != 2 {
		t.Errorf("Expected 2 operations, got %d", len(ops))
	}

	if ops[0].Type != OpCreate {
		t.Errorf("First operation should be create, got %s", ops[0].Type)
	}

	if ops[1].Type != OpWrite {
		t.Errorf("Second operation should be write, got %s", ops[1].Type)
	}
}

func TestTransaction_DeleteFile(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := New(tmpDir, "")
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer fs.Close()

	// Create file
	tx1 := fs.Begin(context.Background())
	if err := tx1.WriteFile("test.txt", []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if err := tx1.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Delete file
	tx2 := fs.Begin(context.Background())
	if err := tx2.DeleteFile("test.txt"); err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}
	if err := tx2.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Verify deletion
	filePath := filepath.Join(tmpDir, "test.txt")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("File should not exist after deletion")
	}
}

func TestTransaction_Mkdir(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := New(tmpDir, "")
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer fs.Close()

	// Create directory
	tx := fs.Begin(context.Background())
	if err := tx.Mkdir("subdir", 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Verify directory exists
	dirPath := filepath.Join(tmpDir, "subdir")
	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("Directory should exist: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("Path should be a directory")
	}
}

func TestTransaction_Rename(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := New(tmpDir, "")
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer fs.Close()

	// Create file
	tx1 := fs.Begin(context.Background())
	if err := tx1.WriteFile("old.txt", []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if err := tx1.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Rename file
	tx2 := fs.Begin(context.Background())
	if err := tx2.Rename("old.txt", "new.txt"); err != nil {
		t.Fatalf("Failed to rename file: %v", err)
	}
	if err := tx2.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Verify rename
	oldPath := filepath.Join(tmpDir, "old.txt")
	newPath := filepath.Join(tmpDir, "new.txt")

	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Errorf("Old file should not exist after rename")
	}

	data, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("New file should exist: %v", err)
	}

	if string(data) != "content" {
		t.Errorf("Content mismatch: got %q, want %q", string(data), "content")
	}
}

func TestTransaction_Rollback(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := New(tmpDir, "")
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer fs.Close()

	// Create and rollback a transaction
	tx := fs.Begin(context.Background())
	if err := tx.WriteFile("test.txt", []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	tx.Rollback()

	// Verify file was not created
	filePath := filepath.Join(tmpDir, "test.txt")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("File should not exist after rollback")
	}

	// Verify no operations in log
	ops, err := fs.History()
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}
	if len(ops) != 0 {
		t.Errorf("Expected 0 operations after rollback, got %d", len(ops))
	}
}

func TestRollbackTo(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := New(tmpDir, "")
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer fs.Close()

	// Create version 1
	tx1 := fs.Begin(context.Background())
	if err := tx1.WriteFile("test.txt", []byte("version 1"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if err := tx1.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Create version 2
	tx2 := fs.Begin(context.Background())
	if err := tx2.WriteFile("test.txt", []byte("version 2"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if err := tx2.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Create version 3
	tx3 := fs.Begin(context.Background())
	if err := tx3.WriteFile("test.txt", []byte("version 3"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if err := tx3.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Rollback to version 1
	ops, err := fs.History()
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(ops) != 3 {
		t.Fatalf("Expected 3 operations, got %d", len(ops))
	}

	// Rollback to after first operation
	if err := fs.RollbackTo(ops[0].ID); err != nil {
		t.Fatalf("Failed to rollback: %v", err)
	}

	// Verify content is version 1
	filePath := filepath.Join(tmpDir, "test.txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(data) != "version 1" {
		t.Errorf("Content mismatch after rollback: got %q, want %q", string(data), "version 1")
	}
}

func TestRollbackTo_DeletedFile(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := New(tmpDir, "")
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer fs.Close()

	// Create file
	tx1 := fs.Begin(context.Background())
	if err := tx1.WriteFile("test.txt", []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if err := tx1.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Delete file
	tx2 := fs.Begin(context.Background())
	if err := tx2.DeleteFile("test.txt"); err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}
	if err := tx2.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Rollback to before deletion
	ops, err := fs.History()
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if err := fs.RollbackTo(ops[0].ID); err != nil {
		t.Fatalf("Failed to rollback: %v", err)
	}

	// Verify file was restored
	filePath := filepath.Join(tmpDir, "test.txt")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("File should exist after rollback: %v", err)
	}

	if string(data) != "content" {
		t.Errorf("Content mismatch: got %q, want %q", string(data), "content")
	}
}

func TestRollbackTo_InitialState(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := New(tmpDir, "")
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer fs.Close()

	// Create file
	tx := fs.Begin(context.Background())
	if err := tx.WriteFile("test.txt", []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Rollback to initial state (before any operations)
	if err := fs.RollbackTo(0); err != nil {
		t.Fatalf("Failed to rollback: %v", err)
	}

	// Verify file was deleted
	filePath := filepath.Join(tmpDir, "test.txt")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("File should not exist after rollback to initial state")
	}
}

func TestOperationDiff(t *testing.T) {
	tests := []struct {
		name     string
		op       *Operation
		expected string
	}{
		{
			name: "create",
			op: &Operation{
				Type:         OpCreate,
				Path:         "test.txt",
				AfterContent: []byte("hello"),
			},
			expected: "Created test.txt (5 bytes)",
		},
		{
			name: "write",
			op: &Operation{
				Type:          OpWrite,
				Path:          "test.txt",
				BeforeContent: []byte("hello"),
				AfterContent:  []byte("hello world"),
			},
			expected: "Modified test.txt (5 -> 11 bytes)",
		},
		{
			name: "delete",
			op: &Operation{
				Type:          OpDelete,
				Path:          "test.txt",
				BeforeContent: []byte("content"),
			},
			expected: "Deleted test.txt (7 bytes)",
		},
		{
			name: "mkdir",
			op: &Operation{
				Type: OpMkdir,
				Path: "subdir",
			},
			expected: "Created directory subdir",
		},
		{
			name: "rename",
			op: &Operation{
				Type: OpRename,
				Path: "new.txt",
				Metadata: map[string]string{
					"old_path": "old.txt",
				},
			},
			expected: "Renamed old.txt -> new.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := tt.op.Diff()
			if diff != tt.expected {
				t.Errorf("Diff mismatch: got %q, want %q", diff, tt.expected)
			}
		})
	}
}

func TestHistory(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := New(tmpDir, "")
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer fs.Close()

	// Perform several operations
	tx1 := fs.Begin(context.Background())
	tx1.WriteFile("file1.txt", []byte("content 1"), 0644)
	tx1.Commit()

	tx2 := fs.Begin(context.Background())
	tx2.WriteFile("file2.txt", []byte("content 2"), 0644)
	tx2.Commit()

	tx3 := fs.Begin(context.Background())
	tx3.DeleteFile("file1.txt")
	tx3.Commit()

	// Check history
	ops, err := fs.History()
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(ops) != 3 {
		t.Errorf("Expected 3 operations, got %d", len(ops))
	}

	// Verify operation order and IDs
	for i, op := range ops {
		expectedID := int64(i + 1)
		if op.ID != expectedID {
			t.Errorf("Operation %d has ID %d, want %d", i, op.ID, expectedID)
		}
	}
}

func TestMultipleOperationsInTransaction(t *testing.T) {
	tmpDir := t.TempDir()
	fs, err := New(tmpDir, "")
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer fs.Close()

	// Create transaction with multiple operations
	tx := fs.Begin(context.Background())
	tx.WriteFile("file1.txt", []byte("content 1"), 0644)
	tx.WriteFile("file2.txt", []byte("content 2"), 0644)
	tx.Mkdir("subdir", 0755)
	tx.WriteFile("subdir/file3.txt", []byte("content 3"), 0644)

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	// Verify all files were created
	files := []string{"file1.txt", "file2.txt", "subdir/file3.txt"}
	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("File %s should exist", file)
		}
	}

	// Verify directory was created
	dirPath := filepath.Join(tmpDir, "subdir")
	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("Directory should exist: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("Path should be a directory")
	}

	// Verify 4 operations in log
	ops, err := fs.History()
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}
	if len(ops) != 4 {
		t.Errorf("Expected 4 operations, got %d", len(ops))
	}
}

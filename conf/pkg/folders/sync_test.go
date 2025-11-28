package folders

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportFile(t *testing.T) {
	// Create temporary directories
	sourceDir, err := os.MkdirTemp("", "source")
	if err != nil {
		t.Fatalf("failed to create source temp dir: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	confDir, err := os.MkdirTemp("", "conf")
	if err != nil {
		t.Fatalf("failed to create conf temp dir: %v", err)
	}
	defer os.RemoveAll(confDir)

	t.Run("import_file", func(t *testing.T) {
		// Create source file
		content := []byte("test content")
		if err := os.WriteFile(filepath.Join(sourceDir, "test.txt"), content, 0o644); err != nil {
			t.Fatalf("failed to write source file: %v", err)
		}

		// Import the file
		if err := ImportFile(sourceDir, confDir, "test.txt"); err != nil {
			t.Fatalf("ImportFile failed: %v", err)
		}

		// Verify file was imported
		importedContent, err := os.ReadFile(filepath.Join(confDir, "test.txt"))
		if err != nil {
			t.Fatalf("failed to read imported file: %v", err)
		}
		if string(importedContent) != string(content) {
			t.Errorf("imported content = %q, want %q", importedContent, content)
		}
	})

	t.Run("import_nested_file", func(t *testing.T) {
		// Create nested source file
		nestedDir := filepath.Join(sourceDir, "subdir")
		if err := os.MkdirAll(nestedDir, 0o755); err != nil {
			t.Fatalf("failed to create nested dir: %v", err)
		}
		content := []byte("nested content")
		if err := os.WriteFile(filepath.Join(nestedDir, "nested.txt"), content, 0o644); err != nil {
			t.Fatalf("failed to write nested file: %v", err)
		}

		// Import the nested file
		if err := ImportFile(sourceDir, confDir, "subdir/nested.txt"); err != nil {
			t.Fatalf("ImportFile failed: %v", err)
		}

		// Verify file was imported
		importedContent, err := os.ReadFile(filepath.Join(confDir, "subdir", "nested.txt"))
		if err != nil {
			t.Fatalf("failed to read imported file: %v", err)
		}
		if string(importedContent) != string(content) {
			t.Errorf("imported content = %q, want %q", importedContent, content)
		}
	})

	t.Run("import_directory", func(t *testing.T) {
		// Create a new directory in source
		newDir := filepath.Join(sourceDir, "newdir")
		if err := os.MkdirAll(newDir, 0o755); err != nil {
			t.Fatalf("failed to create new dir: %v", err)
		}

		// Import the directory
		if err := ImportFile(sourceDir, confDir, "newdir"); err != nil {
			t.Fatalf("ImportFile failed: %v", err)
		}

		// Verify directory was created
		info, err := os.Stat(filepath.Join(confDir, "newdir"))
		if err != nil {
			t.Fatalf("failed to stat imported dir: %v", err)
		}
		if !info.IsDir() {
			t.Errorf("expected directory, got file")
		}
	})
}

func TestApplyFile(t *testing.T) {
	// Create temporary directories
	confDir, err := os.MkdirTemp("", "conf")
	if err != nil {
		t.Fatalf("failed to create conf temp dir: %v", err)
	}
	defer os.RemoveAll(confDir)

	sourceDir, err := os.MkdirTemp("", "source")
	if err != nil {
		t.Fatalf("failed to create source temp dir: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	t.Run("apply_file", func(t *testing.T) {
		// Create conf file
		content := []byte("conf content")
		if err := os.WriteFile(filepath.Join(confDir, "test.txt"), content, 0o644); err != nil {
			t.Fatalf("failed to write conf file: %v", err)
		}

		// Apply the file
		if err := ApplyFile(confDir, sourceDir, "test.txt"); err != nil {
			t.Fatalf("ApplyFile failed: %v", err)
		}

		// Verify file was applied
		appliedContent, err := os.ReadFile(filepath.Join(sourceDir, "test.txt"))
		if err != nil {
			t.Fatalf("failed to read applied file: %v", err)
		}
		if string(appliedContent) != string(content) {
			t.Errorf("applied content = %q, want %q", appliedContent, content)
		}
	})
}

func TestDeleteFile(t *testing.T) {
	// Create temporary directory
	targetDir, err := os.MkdirTemp("", "target")
	if err != nil {
		t.Fatalf("failed to create target temp dir: %v", err)
	}
	defer os.RemoveAll(targetDir)

	t.Run("delete_file", func(t *testing.T) {
		// Create file to delete
		filePath := filepath.Join(targetDir, "delete.txt")
		if err := os.WriteFile(filePath, []byte("delete me"), 0o644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		// Delete the file
		if err := DeleteFile(targetDir, "delete.txt"); err != nil {
			t.Fatalf("DeleteFile failed: %v", err)
		}

		// Verify file was deleted
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Errorf("expected file to be deleted")
		}
	})

	t.Run("delete_directory", func(t *testing.T) {
		// Create directory to delete
		dirPath := filepath.Join(targetDir, "deletedir")
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		// Create a file inside
		if err := os.WriteFile(filepath.Join(dirPath, "file.txt"), []byte("content"), 0o644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		// Delete the directory
		if err := DeleteFile(targetDir, "deletedir"); err != nil {
			t.Fatalf("DeleteFile failed: %v", err)
		}

		// Verify directory was deleted
		if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
			t.Errorf("expected directory to be deleted")
		}
	})
}

func TestImportAll(t *testing.T) {
	// Create temporary directories
	sourceDir, err := os.MkdirTemp("", "source")
	if err != nil {
		t.Fatalf("failed to create source temp dir: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	confDir, err := os.MkdirTemp("", "conf")
	if err != nil {
		t.Fatalf("failed to create conf temp dir: %v", err)
	}
	defer os.RemoveAll(confDir)

	t.Run("import_all_changes", func(t *testing.T) {
		// Create source files
		if err := os.WriteFile(filepath.Join(sourceDir, "added.txt"), []byte("added"), 0o644); err != nil {
			t.Fatalf("failed to write source file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(sourceDir, "modified.txt"), []byte("source"), 0o644); err != nil {
			t.Fatalf("failed to write source file: %v", err)
		}

		// Create conf files
		if err := os.WriteFile(filepath.Join(confDir, "modified.txt"), []byte("conf"), 0o644); err != nil {
			t.Fatalf("failed to write conf file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(confDir, "deleted.txt"), []byte("deleted"), 0o644); err != nil {
			t.Fatalf("failed to write conf file: %v", err)
		}

		drifts := []FileDrift{
			{RelPath: "added.txt", Status: StatusAdded},
			{RelPath: "modified.txt", Status: StatusModified},
			{RelPath: "deleted.txt", Status: StatusDeleted},
		}

		// Import all changes
		if err := ImportAll(sourceDir, confDir, drifts); err != nil {
			t.Fatalf("ImportAll failed: %v", err)
		}

		// Verify added file
		content, err := os.ReadFile(filepath.Join(confDir, "added.txt"))
		if err != nil {
			t.Fatalf("failed to read added file: %v", err)
		}
		if string(content) != "added" {
			t.Errorf("added file content = %q, want %q", content, "added")
		}

		// Verify modified file
		content, err = os.ReadFile(filepath.Join(confDir, "modified.txt"))
		if err != nil {
			t.Fatalf("failed to read modified file: %v", err)
		}
		if string(content) != "source" {
			t.Errorf("modified file content = %q, want %q", content, "source")
		}

		// Verify deleted file
		if _, err := os.Stat(filepath.Join(confDir, "deleted.txt")); !os.IsNotExist(err) {
			t.Errorf("expected deleted file to be removed")
		}
	})
}

func TestApplyAll(t *testing.T) {
	// Create temporary directories
	confDir, err := os.MkdirTemp("", "conf")
	if err != nil {
		t.Fatalf("failed to create conf temp dir: %v", err)
	}
	defer os.RemoveAll(confDir)

	sourceDir, err := os.MkdirTemp("", "source")
	if err != nil {
		t.Fatalf("failed to create source temp dir: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	t.Run("apply_all_changes", func(t *testing.T) {
		// Create conf files
		if err := os.WriteFile(filepath.Join(confDir, "modified.txt"), []byte("conf"), 0o644); err != nil {
			t.Fatalf("failed to write conf file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(confDir, "deleted.txt"), []byte("conf"), 0o644); err != nil {
			t.Fatalf("failed to write conf file: %v", err)
		}

		// Create source files
		if err := os.WriteFile(filepath.Join(sourceDir, "added.txt"), []byte("added"), 0o644); err != nil {
			t.Fatalf("failed to write source file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(sourceDir, "modified.txt"), []byte("source"), 0o644); err != nil {
			t.Fatalf("failed to write source file: %v", err)
		}

		drifts := []FileDrift{
			{RelPath: "added.txt", Status: StatusAdded},       // Added in source, should be deleted from source
			{RelPath: "modified.txt", Status: StatusModified}, // Modified, should be applied from conf
			{RelPath: "deleted.txt", Status: StatusDeleted},   // Deleted from source, should be applied from conf
		}

		// Apply all changes
		if err := ApplyAll(confDir, sourceDir, drifts); err != nil {
			t.Fatalf("ApplyAll failed: %v", err)
		}

		// Verify added file was removed from source
		if _, err := os.Stat(filepath.Join(sourceDir, "added.txt")); !os.IsNotExist(err) {
			t.Errorf("expected added file to be removed from source")
		}

		// Verify modified file was applied from conf
		content, err := os.ReadFile(filepath.Join(sourceDir, "modified.txt"))
		if err != nil {
			t.Fatalf("failed to read modified file: %v", err)
		}
		if string(content) != "conf" {
			t.Errorf("modified file content = %q, want %q", content, "conf")
		}

		// Verify deleted file was applied from conf
		content, err = os.ReadFile(filepath.Join(sourceDir, "deleted.txt"))
		if err != nil {
			t.Fatalf("failed to read deleted file: %v", err)
		}
		if string(content) != "conf" {
			t.Errorf("deleted file content = %q, want %q", content, "conf")
		}
	})
}

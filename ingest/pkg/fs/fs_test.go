package fs

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWalkFilesystemCapturesEntries(t *testing.T) {
	tempDir := t.TempDir()

	smallFilePath := filepath.Join(tempDir, "small.txt")
	smallContent := []byte("hello world")
	if err := os.WriteFile(smallFilePath, smallContent, 0o644); err != nil {
		t.Fatalf("failed to create small file: %v", err)
	}

	subDir := filepath.Join(tempDir, "nested")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("failed to create sub directory: %v", err)
	}

	subFilePath := filepath.Join(subDir, "nested.txt")
	if err := os.WriteFile(subFilePath, []byte("inside nested directory"), 0o600); err != nil {
		t.Fatalf("failed to create nested file: %v", err)
	}

	largeFilePath := filepath.Join(tempDir, "large.bin")
	largeContent := bytes.Repeat([]byte{0x1}, 10*1024*1024+1)
	if err := os.WriteFile(largeFilePath, largeContent, 0o600); err != nil {
		t.Fatalf("failed to create large file: %v", err)
	}

	var progressCalls []int
	entries, err := WalkFilesystem(tempDir, func(count int) {
		progressCalls = append(progressCalls, count)
	})
	if err != nil {
		t.Fatalf("WalkFilesystem returned error: %v", err)
	}

	if len(progressCalls) == 0 || progressCalls[0] != 1 {
		t.Fatalf("expected progress callback to be invoked for first entry, got %v", progressCalls)
	}

	if len(entries) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(entries))
	}

	byPath := map[string]FSEntry{}
	for _, entry := range entries {
		byPath[entry.Path] = entry
	}

	rootEntry, ok := byPath["/"]
	if !ok {
		t.Fatalf("expected root entry '/' to be present")
	}
	if !rootEntry.IsDir {
		t.Errorf("expected root entry to be a directory")
	}

	smallEntry, ok := byPath["small.txt"]
	if !ok {
		t.Fatalf("expected small file entry to be present")
	}
	if smallEntry.IsDir {
		t.Errorf("expected small file to be regular file")
	}
	if !bytes.Equal(smallEntry.Content, smallContent) {
		t.Errorf("expected small file content to be captured")
	}
	expectedHash := sha256.Sum256(smallContent)
	if smallEntry.SHA256Hash != fmt.Sprintf("%x", expectedHash) {
		t.Errorf("expected SHA256 hash %x, got %s", expectedHash, smallEntry.SHA256Hash)
	}

	nestedDirEntry, ok := byPath["nested"]
	if !ok {
		t.Fatalf("expected nested directory entry to be present")
	}
	if !nestedDirEntry.IsDir {
		t.Errorf("expected nested entry to be directory")
	}

	nestedFileEntry, ok := byPath["nested/nested.txt"]
	if !ok {
		t.Fatalf("expected nested file entry to be present")
	}
	if nestedFileEntry.IsDir {
		t.Errorf("expected nested file entry to be regular file")
	}
	if len(nestedFileEntry.Content) == 0 {
		t.Errorf("expected nested file content to be populated")
	}

	largeEntry, ok := byPath["large.bin"]
	if !ok {
		t.Fatalf("expected large file entry to be present")
	}
	if len(largeEntry.Content) != 0 {
		t.Errorf("expected large file content to be omitted for >10MB files")
	}
	if largeEntry.SHA256Hash != "" {
		t.Errorf("expected large file hash to be empty when content omitted, got %q", largeEntry.SHA256Hash)
	}
}

func TestWalkFilesystemMissingPath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope")
	entries, err := WalkFilesystem(missing, nil)
	if err == nil {
		t.Fatal("expected error for missing path, got nil")
	}
	if entries != nil {
		t.Fatalf("expected nil entries on error, got %d entries", len(entries))
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected error to mention missing path, got %v", err)
	}
}

package segment

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/sync"
)

func TestNewSegmentWriter(t *testing.T) {
	sw := NewSegmentWriter("/tmp/test", "personal", "node123", 1, 1024*1024, 60)

	if sw == nil {
		t.Fatal("NewSegmentWriter() returned nil")
	}

	if sw.remotePath != "/tmp/test" {
		t.Errorf("remotePath = %v, want %v", sw.remotePath, "/tmp/test")
	}

	if sw.space != "personal" {
		t.Errorf("space = %v, want %v", sw.space, "personal")
	}

	if sw.node != "node123" {
		t.Errorf("node = %v, want %v", sw.node, "node123")
	}

	if sw.segmentSeq != 1 {
		t.Errorf("segmentSeq = %v, want %v", sw.segmentSeq, 1)
	}

	if sw.maxBytes != 1024*1024 {
		t.Errorf("maxBytes = %v, want %v", sw.maxBytes, 1024*1024)
	}

	if sw.maxAge != 60 {
		t.Errorf("maxAge = %v, want %v", sw.maxAge, 60)
	}

	if len(sw.events) != 0 {
		t.Errorf("events length = %v, want %v", len(sw.events), 0)
	}
}

func TestSegmentWriter_AddEvent(t *testing.T) {
	sw := NewSegmentWriter("/tmp/test", "personal", "node123", 1, 1024*1024, 60)

	event := sync.SegmentEvent{
		Schema:  "v1",
		ID:      "event1",
		Lamport: 1,
		TS:      time.Now().Format(time.RFC3339),
		Node:    "node123",
		Space:   "personal",
		Actor:   "test-user",
		Role:    "human",
		Kind:    "task.created",
		Payload: map[string]any{"title": "Test task"},
	}

	sw.AddEvent(event)

	if len(sw.events) != 1 {
		t.Errorf("events length = %v, want %v", len(sw.events), 1)
	}

	if sw.bytesWritten == 0 {
		t.Error("bytesWritten = 0, expected non-zero")
	}
}

func TestSegmentWriter_ShouldRotate_BySize(t *testing.T) {
	sw := NewSegmentWriter("/tmp/test", "personal", "node123", 1, 100, 3600)

	if sw.ShouldRotate() {
		t.Error("ShouldRotate() = true for empty writer, want false")
	}

	// Add enough events to exceed maxBytes
	for i := range 10 {
		event := sync.SegmentEvent{
			Schema:  "v1",
			ID:      "event" + string(rune(i)),
			Lamport: int64(i),
			TS:      time.Now().Format(time.RFC3339),
			Node:    "node123",
			Space:   "personal",
			Actor:   "test-user",
			Role:    "human",
			Kind:    "task.created",
			Payload: map[string]any{"title": "Test task with some content to make it bigger"},
		}
		sw.AddEvent(event)
	}

	if !sw.ShouldRotate() {
		t.Errorf("ShouldRotate() = false after exceeding maxBytes, want true (bytesWritten=%d)", sw.bytesWritten)
	}
}

func TestSegmentWriter_ShouldRotate_ByAge(t *testing.T) {
	sw := NewSegmentWriter("/tmp/test", "personal", "node123", 1, 1024*1024, 1)

	// Set start time to 2 seconds ago
	sw.startTime = time.Now().Add(-2 * time.Second)

	event := sync.SegmentEvent{
		Schema:  "v1",
		ID:      "event1",
		Lamport: 1,
		TS:      time.Now().Format(time.RFC3339),
		Node:    "node123",
		Space:   "personal",
		Actor:   "test-user",
		Role:    "human",
		Kind:    "task.created",
		Payload: map[string]any{"title": "Test"},
	}
	sw.AddEvent(event)

	if !sw.ShouldRotate() {
		t.Error("ShouldRotate() = false after exceeding maxAge, want true")
	}
}

func TestSegmentWriter_WriteSegment(t *testing.T) {
	tempDir := t.TempDir()
	sw := NewSegmentWriter(tempDir, "personal", "node123", 1, 1024*1024, 3600)

	event := sync.SegmentEvent{
		Schema:  "v1",
		ID:      "event1",
		Lamport: 1,
		TS:      time.Now().Format(time.RFC3339),
		Node:    "node123",
		Space:   "personal",
		Actor:   "test-user",
		Role:    "human",
		Kind:    "task.created",
		Payload: map[string]any{"title": "Test task"},
	}
	sw.AddEvent(event)

	segmentInfo, err := sw.WriteSegment()
	if err != nil {
		t.Fatalf("WriteSegment() error = %v", err)
	}

	if segmentInfo == nil {
		t.Fatal("WriteSegment() returned nil segmentInfo")
	}

	if segmentInfo.Rel == "" {
		t.Error("segmentInfo.Rel is empty")
	}

	if segmentInfo.SHA256 == "" {
		t.Error("segmentInfo.SHA256 is empty")
	}

	if segmentInfo.Size == 0 {
		t.Error("segmentInfo.Size = 0")
	}

	if segmentInfo.MTime == "" {
		t.Error("segmentInfo.MTime is empty")
	}

	// Verify file was created
	fullPath := filepath.Join(tempDir, segmentInfo.Rel)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("segment file not created at %v", fullPath)
	}
}

func TestSegmentWriter_WriteSegment_Empty(t *testing.T) {
	tempDir := t.TempDir()
	sw := NewSegmentWriter(tempDir, "personal", "node123", 1, 1024*1024, 3600)

	// Don't add any events
	segmentInfo, err := sw.WriteSegment()
	if err != nil {
		t.Fatalf("WriteSegment() error = %v", err)
	}

	if segmentInfo != nil {
		t.Error("WriteSegment() should return nil for empty writer")
	}
}

func TestCalculateSHA256(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	sha, err := CalculateSHA256(testFile)
	if err != nil {
		t.Fatalf("CalculateSHA256() error = %v", err)
	}

	if sha == "" {
		t.Error("CalculateSHA256() returned empty hash")
	}

	// SHA256 should be 64 hex characters
	if len(sha) != 64 {
		t.Errorf("CalculateSHA256() hash length = %v, want %v", len(sha), 64)
	}

	// Calculate again to verify consistency
	sha2, err := CalculateSHA256(testFile)
	if err != nil {
		t.Fatalf("CalculateSHA256() second call error = %v", err)
	}

	if sha != sha2 {
		t.Errorf("CalculateSHA256() inconsistent: %v != %v", sha, sha2)
	}
}

func TestCalculateSHA256_NonExistentFile(t *testing.T) {
	_, err := CalculateSHA256("/nonexistent/file")
	if err == nil {
		t.Error("CalculateSHA256() should return error for non-existent file")
	}
}

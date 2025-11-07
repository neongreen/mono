package segment

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/neongreen/mono/tk/internal/sync"
)

// SegmentWriter writes events to segment files with compression
type SegmentWriter struct {
	remotePath   string
	space        string
	node         string
	segmentSeq   int64
	startTime    time.Time
	maxBytes     int64
	maxAge       int
	events       []sync.SegmentEvent
	bytesWritten int64
}

// NewSegmentWriter creates a new segment writer
func NewSegmentWriter(remotePath, space, node string, segmentSeq int64, maxBytes int64, maxAge int) *SegmentWriter {
	return &SegmentWriter{
		remotePath:   remotePath,
		space:        space,
		node:         node,
		segmentSeq:   segmentSeq,
		startTime:    time.Now(),
		maxBytes:     maxBytes,
		maxAge:       maxAge,
		events:       []sync.SegmentEvent{},
		bytesWritten: 0,
	}
}

// AddEvent adds an event to the current segment
func (sw *SegmentWriter) AddEvent(event sync.SegmentEvent) {
	sw.events = append(sw.events, event)
	// Rough estimate of bytes
	eventJSON, _ := json.Marshal(event)
	sw.bytesWritten += int64(len(eventJSON))
}

// ShouldRotate returns true if the segment should be rotated
func (sw *SegmentWriter) ShouldRotate() bool {
	if sw.bytesWritten >= sw.maxBytes {
		return true
	}
	if time.Since(sw.startTime).Seconds() >= float64(sw.maxAge) {
		return true
	}
	return false
}

// WriteSegment writes the current segment to disk and returns the segment info
func (sw *SegmentWriter) WriteSegment() (*sync.SegmentInfo, error) {
	if len(sw.events) == 0 {
		return nil, nil
	}

	// Generate filename: YYYY-MM-DDThh-mm-ssZ_<node>_v1_s<segment_seq>.jsonl.zst
	now := sw.startTime.UTC()
	dateDir := filepath.Join(sw.remotePath, sw.space, "segments",
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()))

	if err := os.MkdirAll(dateDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create segment directory: %w", err)
	}

	filename := fmt.Sprintf("%s_%s_v1_s%06d.jsonl.zst",
		now.Format("2006-01-02T15-04-05Z"),
		sw.node,
		sw.segmentSeq)

	fullPath := filepath.Join(dateDir, filename)
	partialPath := fullPath + ".partial"

	// Write to temporary file
	f, err := os.Create(partialPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create segment file: %w", err)
	}

	// Create zstd writer
	encoder, err := zstd.NewWriter(f, zstd.WithEncoderLevel(zstd.SpeedDefault))
	if err != nil {
		f.Close()
		os.Remove(partialPath)
		return nil, fmt.Errorf("failed to create zstd encoder: %w", err)
	}

	// Write events as JSONL
	for _, event := range sw.events {
		eventJSON, err := json.Marshal(event)
		if err != nil {
			encoder.Close()
			f.Close()
			os.Remove(partialPath)
			return nil, fmt.Errorf("failed to marshal event: %w", err)
		}

		if _, err := encoder.Write(eventJSON); err != nil {
			encoder.Close()
			f.Close()
			os.Remove(partialPath)
			return nil, fmt.Errorf("failed to write event: %w", err)
		}

		if _, err := encoder.Write([]byte("\n")); err != nil {
			encoder.Close()
			f.Close()
			os.Remove(partialPath)
			return nil, fmt.Errorf("failed to write newline: %w", err)
		}
	}

	// Close encoder and file
	if err := encoder.Close(); err != nil {
		f.Close()
		os.Remove(partialPath)
		return nil, fmt.Errorf("failed to close encoder: %w", err)
	}

	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(partialPath)
		return nil, fmt.Errorf("failed to sync file: %w", err)
	}

	if err := f.Close(); err != nil {
		os.Remove(partialPath)
		return nil, fmt.Errorf("failed to close file: %w", err)
	}

	// Get file info for size and mtime
	fileInfo, err := os.Stat(partialPath)
	if err != nil {
		os.Remove(partialPath)
		return nil, fmt.Errorf("failed to stat segment file: %w", err)
	}

	// Calculate SHA256
	sha, err := CalculateSHA256(partialPath)
	if err != nil {
		os.Remove(partialPath)
		return nil, fmt.Errorf("failed to calculate SHA256: %w", err)
	}

	// Atomic rename
	if err := os.Rename(partialPath, fullPath); err != nil {
		os.Remove(partialPath)
		return nil, fmt.Errorf("failed to rename segment file: %w", err)
	}

	// Calculate relative path from remote root
	relPath, err := filepath.Rel(sw.remotePath, fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate relative path: %w", err)
	}

	return &sync.SegmentInfo{
		Rel:    relPath,
		SHA256: sha,
		Size:   fileInfo.Size(),
		MTime:  fileInfo.ModTime().UTC().Format(time.RFC3339),
	}, nil
}

// HasPendingEvents returns true if there are events waiting to be written
func (sw *SegmentWriter) HasPendingEvents() bool {
	return len(sw.events) > 0
}

// CalculateSHA256 calculates the SHA256 hash of a file
func CalculateSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

package segment

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/klauspost/compress/zstd"
)

// SegmentReader reads events from segment files
type SegmentReader struct {
	path string
}

// NewSegmentReader creates a new segment reader
func NewSegmentReader(path string) *SegmentReader {
	return &SegmentReader{path: path}
}

// ReadEvents reads all events from a segment file
func (sr *SegmentReader) ReadEvents() ([]SegmentEvent, error) {
	f, err := os.Open(sr.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open segment file: %w", err)
	}
	defer f.Close()

	// Create zstd decoder
	decoder, err := zstd.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd decoder: %w", err)
	}
	defer decoder.Close()

	var events []SegmentEvent
	scanner := bufio.NewScanner(decoder)

	// Increase buffer size for large lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // 1MB max line size

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		var event SegmentEvent
		if err := json.Unmarshal(line, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event on line %d: %w", lineNum, err)
		}

		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan segment file: %w", err)
	}

	return events, nil
}

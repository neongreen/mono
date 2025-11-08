package segment

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// SegmentFilename represents a parsed segment filename
type SegmentFilename struct {
	Timestamp  time.Time
	NodeID     string
	Version    string
	SequenceID int64
}

// ParseSegmentFilename parses a segment filename into its components
// Format: YYYY-MM-DDThh-mm-ssZ_<nodeID>_v1_s<seq>.jsonl.zst
// Example: 2025-10-29T18-57-58Z_FsnkC8_v1_s000041.jsonl.zst
func ParseSegmentFilename(filename string) (*SegmentFilename, error) {
	// Remove .jsonl.zst extension
	base := filepath.Base(filename)
	if !strings.HasSuffix(base, ".jsonl.zst") {
		return nil, fmt.Errorf("invalid segment filename: must end with .jsonl.zst")
	}
	nameWithoutExt := strings.TrimSuffix(base, ".jsonl.zst")

	// Split by underscore: timestamp_nodeID_version_sequence
	parts := strings.Split(nameWithoutExt, "_")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid segment filename format: expected 4 parts, got %d", len(parts))
	}

	// Parse timestamp
	timestamp, err := time.Parse("2006-01-02T15-04-05Z", parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp in filename: %w", err)
	}

	// Node ID is already a string
	nodeID := parts[1]

	// Version should be v1, v2, etc.
	version := parts[2]
	if !strings.HasPrefix(version, "v") {
		return nil, fmt.Errorf("invalid version format: expected vN, got %s", version)
	}

	// Parse sequence number (format: s000041)
	seqStr := parts[3]
	if !strings.HasPrefix(seqStr, "s") {
		return nil, fmt.Errorf("invalid sequence format: expected sNNNNNN, got %s", seqStr)
	}
	var seq int64
	if _, err := fmt.Sscanf(seqStr, "s%d", &seq); err != nil {
		return nil, fmt.Errorf("failed to parse sequence number: %w", err)
	}

	return &SegmentFilename{
		Timestamp:  timestamp,
		NodeID:     nodeID,
		Version:    version,
		SequenceID: seq,
	}, nil
}

// GenerateSegmentFilename generates a segment filename from components
func GenerateSegmentFilename(timestamp time.Time, nodeID string, version string, sequenceID int64) string {
	return fmt.Sprintf("%s_%s_%s_s%06d.jsonl.zst",
		timestamp.UTC().Format("2006-01-02T15-04-05Z"),
		nodeID,
		version,
		sequenceID)
}

// SegmentBelongsToNode checks if a segment path belongs to the given node ID
func SegmentBelongsToNode(segmentPath, nodeID string) bool {
	parsed, err := ParseSegmentFilename(segmentPath)
	if err != nil {
		return false
	}
	return parsed.NodeID == nodeID
}

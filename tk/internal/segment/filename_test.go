package segment

import (
	"testing"
	"time"
)

func TestParseSegmentFilename(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		wantError bool
		want      *SegmentFilename
	}{
		{
			name:     "valid segment filename",
			filename: "2025-10-29T18-57-58Z_FsnkC8_v1_s000041.jsonl.zst",
			want: &SegmentFilename{
				Timestamp:  time.Date(2025, 10, 29, 18, 57, 58, 0, time.UTC),
				NodeID:     "FsnkC8",
				Version:    "v1",
				SequenceID: 41,
			},
		},
		{
			name:     "valid segment with path",
			filename: "personal/segments/2025/10/29/2025-10-29T18-57-58Z_5L0ktM_v1_s000042.jsonl.zst",
			want: &SegmentFilename{
				Timestamp:  time.Date(2025, 10, 29, 18, 57, 58, 0, time.UTC),
				NodeID:     "5L0ktM",
				Version:    "v1",
				SequenceID: 42,
			},
		},
		{
			name:     "valid segment with large sequence",
			filename: "2025-11-01T01-00-10Z_5L0ktM_v1_s000999.jsonl.zst",
			want: &SegmentFilename{
				Timestamp:  time.Date(2025, 11, 1, 1, 0, 10, 0, time.UTC),
				NodeID:     "5L0ktM",
				Version:    "v1",
				SequenceID: 999,
			},
		},
		{
			name:      "missing extension",
			filename:  "2025-10-29T18-57-58Z_FsnkC8_v1_s000041.jsonl",
			wantError: true,
		},
		{
			name:      "wrong extension",
			filename:  "2025-10-29T18-57-58Z_FsnkC8_v1_s000041.json.gz",
			wantError: true,
		},
		{
			name:      "too few parts",
			filename:  "2025-10-29T18-57-58Z_FsnkC8.jsonl.zst",
			wantError: true,
		},
		{
			name:      "invalid timestamp",
			filename:  "2025-13-99T99-99-99Z_FsnkC8_v1_s000041.jsonl.zst",
			wantError: true,
		},
		{
			name:      "invalid version",
			filename:  "2025-10-29T18-57-58Z_FsnkC8_1_s000041.jsonl.zst",
			wantError: true,
		},
		{
			name:      "invalid sequence",
			filename:  "2025-10-29T18-57-58Z_FsnkC8_v1_000041.jsonl.zst",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSegmentFilename(tt.filename)
			if tt.wantError {
				if err == nil {
					t.Errorf("ParseSegmentFilename() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("ParseSegmentFilename() unexpected error: %v", err)
				return
			}
			if got.NodeID != tt.want.NodeID {
				t.Errorf("NodeID = %v, want %v", got.NodeID, tt.want.NodeID)
			}
			if got.Version != tt.want.Version {
				t.Errorf("Version = %v, want %v", got.Version, tt.want.Version)
			}
			if got.SequenceID != tt.want.SequenceID {
				t.Errorf("SequenceID = %v, want %v", got.SequenceID, tt.want.SequenceID)
			}
			if !got.Timestamp.Equal(tt.want.Timestamp) {
				t.Errorf("Timestamp = %v, want %v", got.Timestamp, tt.want.Timestamp)
			}
		})
	}
}

func TestGenerateSegmentFilename(t *testing.T) {
	tests := []struct {
		name       string
		timestamp  time.Time
		nodeID     string
		version    string
		sequenceID int64
		want       string
	}{
		{
			name:       "basic generation",
			timestamp:  time.Date(2025, 10, 29, 18, 57, 58, 0, time.UTC),
			nodeID:     "FsnkC8",
			version:    "v1",
			sequenceID: 41,
			want:       "2025-10-29T18-57-58Z_FsnkC8_v1_s000041.jsonl.zst",
		},
		{
			name:       "different node",
			timestamp:  time.Date(2025, 11, 1, 1, 0, 10, 0, time.UTC),
			nodeID:     "5L0ktM",
			version:    "v1",
			sequenceID: 999,
			want:       "2025-11-01T01-00-10Z_5L0ktM_v1_s000999.jsonl.zst",
		},
		{
			name:       "zero sequence",
			timestamp:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			nodeID:     "ABC123",
			version:    "v1",
			sequenceID: 0,
			want:       "2025-01-01T00-00-00Z_ABC123_v1_s000000.jsonl.zst",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSegmentFilename(tt.timestamp, tt.nodeID, tt.version, tt.sequenceID)
			if got != tt.want {
				t.Errorf("GenerateSegmentFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSegmentBelongsToNode(t *testing.T) {
	tests := []struct {
		name        string
		segmentPath string
		nodeID      string
		want        bool
	}{
		{
			name:        "matching node ID",
			segmentPath: "personal/segments/2025/10/29/2025-10-29T18-57-58Z_FsnkC8_v1_s000041.jsonl.zst",
			nodeID:      "FsnkC8",
			want:        true,
		},
		{
			name:        "different node ID",
			segmentPath: "personal/segments/2025/10/29/2025-10-29T18-57-58Z_FsnkC8_v1_s000041.jsonl.zst",
			nodeID:      "5L0ktM",
			want:        false,
		},
		{
			name:        "invalid filename",
			segmentPath: "invalid.jsonl.zst",
			nodeID:      "FsnkC8",
			want:        false,
		},
		{
			name:        "just filename",
			segmentPath: "2025-10-29T18-57-58Z_5L0ktM_v1_s000042.jsonl.zst",
			nodeID:      "5L0ktM",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SegmentBelongsToNode(tt.segmentPath, tt.nodeID)
			if got != tt.want {
				t.Errorf("SegmentBelongsToNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that we can generate and parse back to the same values
	timestamp := time.Date(2025, 10, 29, 18, 57, 58, 0, time.UTC)
	nodeID := "FsnkC8"
	version := "v1"
	sequenceID := int64(41)

	filename := GenerateSegmentFilename(timestamp, nodeID, version, sequenceID)
	parsed, err := ParseSegmentFilename(filename)
	if err != nil {
		t.Fatalf("ParseSegmentFilename() error: %v", err)
	}

	if parsed.NodeID != nodeID {
		t.Errorf("NodeID = %v, want %v", parsed.NodeID, nodeID)
	}
	if parsed.Version != version {
		t.Errorf("Version = %v, want %v", parsed.Version, version)
	}
	if parsed.SequenceID != sequenceID {
		t.Errorf("SequenceID = %v, want %v", parsed.SequenceID, sequenceID)
	}
	if !parsed.Timestamp.Equal(timestamp) {
		t.Errorf("Timestamp = %v, want %v", parsed.Timestamp, timestamp)
	}
}

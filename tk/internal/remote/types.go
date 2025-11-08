package remote

import (
	"time"

	"github.com/neongreen/mono/tk/internal/segment"
)

// Re-export segment types for convenience
type SegmentEvent = segment.SegmentEvent
type SegmentContext = segment.SegmentContext
type SegmentInfo = segment.SegmentInfo
type IndexFile = segment.IndexFile

// IngestWatermark tracks the last ingested event per remote/space
type IngestWatermark struct {
	RemoteName  string    `json:"remote_name"`
	Space       string    `json:"space"`
	LastEventID string    `json:"last_event_id"`
	LastLamport int64     `json:"last_lamport"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ExportState tracks export state per remote/space
type ExportState struct {
	RemoteName          string    `json:"remote_name"`
	Space               string    `json:"space"`
	LastExportedEventID string    `json:"last_exported_event_id"`
	SegmentSeq          int64     `json:"segment_seq"`
	UpdatedAt           time.Time `json:"updated_at"`
}

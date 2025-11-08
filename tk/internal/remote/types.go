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

// RemoteConfig represents a configured remote
type RemoteConfig struct {
	Type   string   `json:"type"`   // folder, git, http, s3
	Path   string   `json:"path"`   // for folder type
	Spaces []string `json:"spaces"` // spaces to sync
	Push   bool     `json:"push"`
	Pull   bool     `json:"pull"`
}

// Config represents the tk configuration file
type Config struct {
	Remotes  map[string]RemoteConfig `json:"remotes"`
	Sync     SyncConfig              `json:"sync"`
	Blocking BlockingConfig          `json:"blocking"`
}

// BlockingConfig represents blocking-specific configuration
type BlockingConfig struct {
	BlockingAxis string   `json:"blocking_axis"` // Axis to check for blocked status (e.g., "generic", "code")
	DoneStates   []string `json:"done_states"`   // States that indicate a task is done
}

// SyncConfig represents sync-specific configuration
type SyncConfig struct {
	SegmentMaxBytes int64    `json:"segment_max_bytes"`
	SegmentMaxAge   int      `json:"segment_max_age"`
	Compress        string   `json:"compress"`
	SafeMode        bool     `json:"safe_mode"`
	Spaces          []string `json:"spaces"`
}

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

// DefaultSyncConfig returns the default sync configuration
func DefaultSyncConfig() SyncConfig {
	return SyncConfig{
		SegmentMaxBytes: 2_000_000, // 2 MB
		SegmentMaxAge:   120,       // 120 seconds
		Compress:        "zstd",
		SafeMode:        true,
		Spaces:          []string{"personal"},
	}
}

// DefaultBlockingConfig returns the default blocking configuration
func DefaultBlockingConfig() BlockingConfig {
	return BlockingConfig{
		BlockingAxis: "generic",
		DoneStates:   []string{"done"},
	}
}

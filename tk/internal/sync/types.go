package sync

import "time"

// SegmentEvent represents an event in the v1 segment format
type SegmentEvent struct {
	Schema  string          `json:"schema"`
	ID      string          `json:"id"`
	Lamport int64           `json:"lamport"`
	TS      string          `json:"ts"` // RFC3339 timestamp
	Node    string          `json:"node"`
	Space   string          `json:"space"`
	Actor   string          `json:"actor"`
	Role    string          `json:"role"`
	Kind    string          `json:"kind"`
	Payload any             `json:"payload"`
	Ctx     *SegmentContext `json:"ctx"`
}

// SegmentContext represents the context in a segment event
type SegmentContext struct {
	RepoUUID *string `json:"repo_uuid"`
	Branch   *string `json:"branch"`
	Commit   *string `json:"commit"`
	JJOpID   *string `json:"jj_op_id"`
}

// SegmentInfo represents metadata about a segment file
type SegmentInfo struct {
	Rel    string `json:"rel"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
	MTime  string `json:"mtime"`
}

// IndexFile represents the index.json structure
type IndexFile struct {
	Schema   string        `json:"schema"`
	Space    string        `json:"space"`
	Segments []SegmentInfo `json:"segments"`
}

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

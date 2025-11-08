package segment

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

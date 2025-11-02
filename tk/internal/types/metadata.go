package types

import "encoding/json"

// MetadataClaim represents a metadata value assertion by an actor
type MetadataClaim struct {
	Value     json.RawMessage `json:"value"`     // JSON-encoded value
	Role      string          `json:"role"`      // human / agent / bot / qa / rel
	Tentative bool            `json:"tentative"` // True if overridden by higher authority
	TS        int64           `json:"ts"`        // Lamport timestamp
}

// MetadataStatus represents the MV register for a metadata key
// Uses same authority lattice resolution as AxisStatus
type MetadataStatus struct {
	Effective json.RawMessage `json:"effective"` // Current effective value
	Claims    []MetadataClaim `json:"claims"`    // All competing claims
}

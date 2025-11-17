package types

// Blocker represents a task that blocks another
type Blocker struct {
	TaskUUID      string `json:"uuid"`               // Canonical task UUID
	TaskDisplayID string `json:"display_id"`         // Display ID
	Title         string `json:"title"`              // Task title
	Distance      int    `json:"distance,omitempty"` // Distance in dependency graph
}

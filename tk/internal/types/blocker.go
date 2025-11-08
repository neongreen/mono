package types

// Blocker represents a task that blocks another
type Blocker struct {
	TaskID        string `json:"id"`                 // Canonical task ID
	TaskDisplayID string `json:"display_id"`         // Display ID
	Title         string `json:"title"`              // Task title
	Distance      int    `json:"distance,omitempty"` // Distance in dependency graph
}

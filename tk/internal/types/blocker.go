package types

// Blocker represents a task that blocks another
type Blocker struct {
	TaskID   string `json:"task_id"`
	Title    string `json:"title"`
	Distance int    `json:"distance"` // Distance in dependency graph
}

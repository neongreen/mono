package types

import "time"

// Project represents a project's current state in the reducer
type Project struct {
	ProjectUID  string    `json:"project_uid"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedAtTS int64     `json:"created_at_ts"` // Lamport timestamp for event ordering
	IsSynthetic bool      `json:"is_synthetic"`  // Whether this is a synthetic project (auto-created)
}

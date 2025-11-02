package types

import (
	"time"
)

// Task represents the current state of a task, derived from events
type Task struct {
	TaskUUID  string                    `json:"task_uuid"`         // Canonical immutable UUID
	TaskID    string                    `json:"task_id"`           // Current display ID
	Aliases   []string                  `json:"aliases,omitempty"` // Previous IDs (when task was moved)
	Title     string                    `json:"title"`
	Axes      map[string]AxisStatus     `json:"axes"`
	Metadata  map[string]MetadataStatus `json:"metadata,omitempty"` // Metadata with claims
	Notes     []Note                    `json:"notes"`
	CreatedBy string                    `json:"created_by"`
	CreatedAt time.Time                 `json:"created_at"`
	Relations *Relations                `json:"relations,omitempty"` // Task relations
	Blocked   bool                      `json:"blocked,omitempty"`   // Is this task blocked
	Blockers  []Blocker                 `json:"blockers,omitempty"`  // List of blocking tasks
}

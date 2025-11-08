package types

import (
	"encoding/json"
	"sort"
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

// MarshalJSON provides deterministic JSON output by sorting relation slices
func (t *Task) MarshalJSON() ([]byte, error) {
	// Create an alias type to avoid infinite recursion
	type TaskAlias Task

	// Make a copy with sorted relations
	taskCopy := TaskAlias(*t)
	
	// Sort relations if present
	if taskCopy.Relations != nil {
		taskCopy.Relations = taskCopy.Relations.Sorted()
	}
	
	// Sort blockers by TaskUUID
	if len(taskCopy.Blockers) > 0 {
		blockersCopy := make([]Blocker, len(taskCopy.Blockers))
		copy(blockersCopy, taskCopy.Blockers)
		sort.Slice(blockersCopy, func(i, j int) bool {
			return blockersCopy[i].TaskUUID < blockersCopy[j].TaskUUID
		})
		taskCopy.Blockers = blockersCopy
	}
	
	return json.Marshal(&taskCopy)
}

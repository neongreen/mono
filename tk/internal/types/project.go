package types

import "time"

// Project represents a project in the system
type Project struct {
	UID         ProjectUID  `json:"uid"`
	Type        ProjectType `json:"type"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	CreatedBy   string      `json:"created_by"`
}

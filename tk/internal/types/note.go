package types

import "time"

// Note represents a note on a task
type Note struct {
	Markdown  string    `json:"markdown"`
	Actor     string    `json:"actor"`
	Timestamp time.Time `json:"timestamp"`
}

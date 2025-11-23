package reducer

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/types"
)

func (r *Reducer) applyTaskNoteAdd(e types.Event) error {
	var payload types.TaskNoteAddPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.note.add payload: %w", err)
	}

	// Resolve task ID to UUID
	taskUUID := payload.TaskUUID
	if taskUUID == "" {
		// Legacy event fallback - only for reading old pre-UUID events
		// See tk-190: can be removed after running 'tk migrate compact-remote' on all machines
		var ok bool
		taskUUID, ok = r.taskByID[payload.TaskID]
		if !ok {
			return fmt.Errorf("task not found: %s", payload.TaskID)
		}
	}

	task, ok := r.tasks[taskUUID]
	if !ok {
		return fmt.Errorf("task UUID not found: %s", taskUUID)
	}

	note := types.Note{
		Markdown:  payload.Markdown,
		Actor:     e.Actor,
		Timestamp: e.CreatedAt, // Use actual creation time from event
	}
	task.Notes = append(task.Notes, note)
	task.UpdatedAt = e.CreatedAt

	return nil
}

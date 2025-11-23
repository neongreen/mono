package reducer

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/types"
)

func (r *Reducer) applyTaskAttachmentAdd(e types.Event) error {
	var payload types.TaskAttachmentAddPayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.attachment.add payload: %w", err)
	}

	task, ok := r.tasks[payload.TaskUUID]
	if !ok {
		return fmt.Errorf("task UUID not found: %s", payload.TaskUUID)
	}

	// Add attachment if not already present
	for _, att := range task.Attachments {
		if att.ID == payload.AttachmentID {
			// Already attached, skip
			return nil
		}
	}

	task.Attachments = append(task.Attachments, types.Attachment{
		ID:          payload.AttachmentID,
		Hash:        payload.AttachmentHash,
		Filename:    payload.Filename,
		Description: payload.Description,
		MimeType:    payload.MimeType,
		Size:        payload.Size,
	})
	task.UpdatedAt = e.CreatedAt

	return nil
}

func (r *Reducer) applyTaskAttachmentRemove(e types.Event) error {
	var payload types.TaskAttachmentRemovePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.attachment.remove payload: %w", err)
	}

	task, ok := r.tasks[payload.TaskUUID]
	if !ok {
		return fmt.Errorf("task UUID not found: %s", payload.TaskUUID)
	}

	// Remove attachment by ID
	filtered := make([]types.Attachment, 0, len(task.Attachments))
	for _, att := range task.Attachments {
		if att.ID != payload.AttachmentID {
			filtered = append(filtered, att)
		}
	}
	task.Attachments = filtered
	task.UpdatedAt = e.CreatedAt

	return nil
}

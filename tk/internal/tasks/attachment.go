package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

// AddAttachment attaches a file to a task
func AddAttachment(db *database.DB, taskUUID string, attachmentID string, attachmentHash string, filename string, description string, mimeType string, size int64, actor string, clk clock.Clock) error {
	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return err
	}

	lamportTS, err := db.GetNextLamportTS()
	if err != nil {
		return err
	}

	payload := types.TaskAttachmentAddPayload{
		TaskUUID:       taskUUID,
		AttachmentID:   attachmentID,
		AttachmentHash: attachmentHash,
		Filename:       filename,
		Description:    description,
		MimeType:       mimeType,
		Size:           size,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        lamportTS,
		CreatedAt: clk.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      "task.attachment.add",
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return err
	}

	return nil
}

// RemoveAttachment removes a file attachment from a task
func RemoveAttachment(db *database.DB, taskUUID string, attachmentID string, actor string, clk clock.Clock) error {
	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return err
	}

	lamportTS, err := db.GetNextLamportTS()
	if err != nil {
		return err
	}

	payload := types.TaskAttachmentRemovePayload{
		TaskUUID:     taskUUID,
		AttachmentID: attachmentID,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        lamportTS,
		CreatedAt: clk.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      "task.attachment.remove",
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return err
	}

	return nil
}

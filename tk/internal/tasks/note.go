package tasks

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

// AddNote adds a note to a task
func AddNote(db *database.DB, taskUUID string, markdown string, actor string) error {
	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return err
	}

	lamportTS, err := db.GetNextLamportTS()
	if err != nil {
		return err
	}

	payload := types.TaskNoteAddPayload{
		TaskUUID: taskUUID,
		Markdown: markdown,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        lamportTS,
		CreatedAt: time.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      "task.note.add",
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return err
	}

	return nil
}

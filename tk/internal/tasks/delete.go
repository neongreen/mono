package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

// Delete deletes a task
func Delete(db *database.DB, taskUUID string, actor string, clk clock.Clock) error {
	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return err
	}

	lamportTS, err := db.GetNextLamportTS()
	if err != nil {
		return err
	}

	payload := types.TaskDeletePayload{
		TaskUUID: taskUUID,
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
		Kind:      string(types.EventKindTaskDelete),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return err
	}

	if err := db.ProjectTaskDeleteEvent(event); err != nil {
		return fmt.Errorf("failed to project task.delete event: %w", err)
	}

	return nil
}

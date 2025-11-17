package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

// MarkOptions contains options for marking a task status
type MarkOptions struct {
	Axis  string
	State string // Empty string to unset
	Role  string
}

// Mark sets or unsets the status of a task on a specific axis
func Mark(db *database.DB, taskUUID string, opts MarkOptions, actor string, clk clock.Clock) error {
	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return fmt.Errorf("failed to generate event ID: %w", err)
	}

	lamportTS, err := db.GetNextLamportTS()
	if err != nil {
		return fmt.Errorf("failed to get lamport timestamp: %w", err)
	}

	payload := types.TaskStatusSetPayload{
		TaskUUID: taskUUID,
		TaskID:   "", // Will be filled by database if needed
		Axis:     opts.Axis,
		State:    opts.State,
		Role:     opts.Role,
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
		Role:      opts.Role,
		Kind:      string(types.EventKindTaskStatusSet),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

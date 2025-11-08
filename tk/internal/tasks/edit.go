package tasks

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

// EditField edits a specific field of a task
func EditField(db *database.DB, taskUID, field, value, actor string) error {
	switch field {
	case "number":
		return EditNumber(db, taskUID, value, actor)
	case "title":
		return EditTitle(db, taskUID, value, actor)
	case "status":
		return EditStatus(db, taskUID, value, actor)
	default:
		return fmt.Errorf("unsupported field %s (supported: number, title, status)", field)
	}
}

// EditStatus sets the status of a task on the generic axis
func EditStatus(db *database.DB, taskUID string, value string, actor string) error {
	value = strings.TrimSpace(value)
	// Allow empty status to unset it

	lamport, err := db.GetNextLamportTS()
	if err != nil {
		return fmt.Errorf("failed to get lamport timestamp: %w", err)
	}

	payload := types.TaskStatusSetPayload{
		TaskUUID: taskUID,
		Axis:     "generic",
		State:    value,
		Role:     "human",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task.status.set payload: %w", err)
	}

	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        lamport,
		CreatedAt: time.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindTaskStatusSet),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert task.status.set event: %w", err)
	}

	return nil
}

// EditNumber sets the number of a task
func EditNumber(db *database.DB, taskUID string, value string, actor string) error {
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil || number <= 0 {
		return fmt.Errorf("invalid number %q", value)
	}

	projectUID, err := GetProjectForTask(db, taskUID)
	if err != nil {
		return err
	}

	lamport, err := db.GetNextLamportTS()
	if err != nil {
		return fmt.Errorf("failed to get lamport timestamp: %w", err)
	}

	payload := types.TaskNumberSetPayload{
		TaskUID:    taskUID,
		ProjectUID: projectUID,
		Number:     number,
		Reason:     "edit",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task.number.set payload: %w", err)
	}

	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        lamport,
		CreatedAt: time.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindTaskNumberSet),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert task.number.set event: %w", err)
	}

	if err := db.ProjectTaskNumberSetEvent(event); err != nil {
		return fmt.Errorf("failed to project task.number.set event: %w", err)
	}

	return nil
}

// EditTitle sets the title of a task
func EditTitle(db *database.DB, taskUID string, value string, actor string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("title cannot be empty")
	}

	lamport, err := db.GetNextLamportTS()
	if err != nil {
		return fmt.Errorf("failed to get lamport timestamp: %w", err)
	}

	payload := types.TaskTitleSetPayload{
		TaskUID: taskUID,
		Title:   value,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task.title.set payload: %w", err)
	}

	event := types.Event{
		ID:        string(types.NewEventID()),
		TS:        lamport,
		CreatedAt: time.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindTaskTitleSet),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert task.title.set event: %w", err)
	}

	if err := db.ProjectTaskTitleSetEvent(event); err != nil {
		return fmt.Errorf("failed to project task.title.set event: %w", err)
	}

	return nil
}

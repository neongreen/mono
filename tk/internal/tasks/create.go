package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

// CreateParams holds the parameters for creating a new task
type CreateParams struct {
	ProjectUID types.ProjectUID
	Title      string
	ItemKind   string // Item kind (defaults to "task" if empty)
}

// CreateResult holds the result of creating a task
type CreateResult struct {
	TaskUID   types.TaskUID
	DisplayID string
}

// Create creates a new task in the specified project
func Create(db *database.DB, params CreateParams, actor string, clk clock.Clock) (*CreateResult, error) {
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return nil, err
	}

	taskUID := types.NewTaskUID()

	// Default item kind to "task" if not specified
	itemKind := params.ItemKind
	if itemKind == "" {
		itemKind = "task"
	}

	// Compute proposed number (max + 1)
	var maxNumber int64
	err = db.Db.QueryRow(`
		SELECT COALESCE(MAX(number), 0) FROM task_numbers
		WHERE project_uid = ?
	`, params.ProjectUID.String()).Scan(&maxNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get max number: %w", err)
	}
	proposedNumber := maxNumber + 1

	payload := types.TaskCreatedPayload{
		TaskUID:        string(taskUID),
		ProjectUID:     params.ProjectUID.String(),
		ProposedNumber: proposedNumber,
		CreatedNode:    nodeID,
		Title:          params.Title,
		CreatedBy:      actor,
		ItemKind:       itemKind,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return nil, fmt.Errorf("failed to generate event ID: %w", err)
	}

	ts, err := db.GetNextLamportTS()
	if err != nil {
		return nil, fmt.Errorf("failed to get next lamport timestamp: %w", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: clk.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return nil, err
	}

	numberPayload := types.TaskNumberSetPayload{
		TaskUID:    taskUID,
		ProjectUID: params.ProjectUID,
		Number:     proposedNumber,
		Reason:     "initial",
	}
	numberPayloadJSON, err := json.Marshal(numberPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal number payload: %w", err)
	}

	numberEventID, err := database.GenerateEventID(db)
	if err != nil {
		return nil, fmt.Errorf("failed to generate event ID: %w", err)
	}

	numberTS, err := db.GetNextLamportTS()
	if err != nil {
		return nil, fmt.Errorf("failed to get next lamport timestamp: %w", err)
	}

	numberEvent := types.Event{
		ID:        numberEventID,
		TS:        numberTS,
		CreatedAt: clk.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindTaskNumberSet),
		Payload:   numberPayloadJSON,
	}

	if err := db.InsertEvent(numberEvent); err != nil {
		return nil, fmt.Errorf("failed to insert number event: %w", err)
	}

	// Rebuild projections from events to update database
	// This ensures database consistency with reducer state
	if err := db.RebuildProjections(); err != nil {
		return nil, fmt.Errorf("failed to rebuild projections: %w", err)
	}

	// Get display ID
	displayID, err := database.RenderTaskDisplayID(db, string(taskUID))
	if err != nil {
		return nil, fmt.Errorf("failed to render display ID: %w", err)
	}

	return &CreateResult{
		TaskUID:   taskUID,
		DisplayID: displayID,
	}, nil
}

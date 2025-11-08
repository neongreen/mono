package database

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

// CreateTaskParams holds the parameters for creating a new task.
type CreateTaskParams struct {
	ProjectRef  string // Project reference (UID, alias, or name)
	Title       string
	CurrentUser string
}

// CreateTaskResult holds the result of creating a task.
type CreateTaskResult struct {
	TaskUID   types.TaskUID
	DisplayID string // Format: project-alias-number
}

// CreateTask creates a new task in the specified project.
func CreateTask(db *DB, params CreateTaskParams) (*CreateTaskResult, error) {
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return nil, err
	}

	projectUID, err := ResolveProjectRef(db, types.NewProjectRef(params.ProjectRef))
	if err != nil {
		return nil, fmt.Errorf("project/alias %q not found. Create it first with: tk project create %s", params.ProjectRef, params.ProjectRef)
	}

	taskUID := types.NewTaskUID()

	// Compute proposed number (max + 1)
	var maxNumber int64
	err = db.Db.QueryRow(`
		SELECT COALESCE(MAX(number), 0) FROM task_numbers
		WHERE project_uid = ?
	`, projectUID.String()).Scan(&maxNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get max number: %w", err)
	}
	proposedNumber := maxNumber + 1

	payload := types.TaskCreatedPayload{
		TaskUID:        string(taskUID),
		ProjectUID:     projectUID.String(),
		ProposedNumber: proposedNumber,
		CreatedNode:    nodeID,
		Title:          params.Title,
		CreatedBy:      params.CurrentUser,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	eventID, err := GenerateEventID(db)
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
		CreatedAt: time.Now(),
		Actor:     params.CurrentUser,
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return nil, err
	}

	if err := db.ProjectTaskCreatedEvent(event); err != nil {
		return nil, fmt.Errorf("failed to project task: %w", err)
	}

	numberPayload := types.TaskNumberSetPayload{
		TaskUID:    string(taskUID),
		ProjectUID: projectUID.String(),
		Number:     proposedNumber,
		Reason:     "initial",
	}
	numberPayloadJSON, err := json.Marshal(numberPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal number payload: %w", err)
	}

	numberEventID, err := GenerateEventID(db)
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
		CreatedAt: time.Now(),
		Actor:     params.CurrentUser,
		Role:      "human",
		Kind:      string(types.EventKindTaskNumberSet),
		Payload:   numberPayloadJSON,
	}

	if err := db.InsertEvent(numberEvent); err != nil {
		return nil, fmt.Errorf("failed to insert number event: %w", err)
	}

	if err := db.ProjectTaskNumberSetEvent(numberEvent); err != nil {
		return nil, fmt.Errorf("failed to project task number: %w", err)
	}

	// Get a friendly display name (preferred alias, or project name, or UID as fallback)
	displayPrefix, err := PreferredAliasForProject(db, projectUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get display prefix: %w", err)
	}
	if displayPrefix == "" {
		// No alias found, try to get project name
		var projectName string
		err = db.Db.QueryRow(`SELECT name FROM projects WHERE project_uid = ?`, projectUID.String()).Scan(&projectName)
		if err == nil && projectName != "" {
			displayPrefix = projectName
		} else {
			// Fallback to UID
			displayPrefix = projectUID.String()
		}
	}

	displayID := fmt.Sprintf("%s-%d", displayPrefix, proposedNumber)

	return &CreateTaskResult{
		TaskUID:   taskUID,
		DisplayID: displayID,
	}, nil
}

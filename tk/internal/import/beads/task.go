package beads

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

// ImportBeadsIssue imports a single beads issue as a tk task
func ImportBeadsIssue(db *database.DB, issue BeadsIssue, projectUID string, number int64) (string, error) {
	// Get node ID
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		return "", fmt.Errorf("failed to get node ID: %w", err)
	}

	// Generate task UID
	taskUID := types.NewTaskUID().String()

	// Parse created_at time if available
	createdAt := parseCreatedAt(issue.CreatedAt)

	// Get actor
	actor := "importer"
	if issue.Assignee != "" {
		actor = issue.Assignee
	}

	// Create task.created event
	if err := createTaskCreatedEvent(db, taskUID, projectUID, number, issue.Title, nodeID, actor, createdAt); err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	// Set task number
	if err := createTaskNumberEvent(db, taskUID, projectUID, number, actor, createdAt); err != nil {
		return "", fmt.Errorf("failed to set task number: %w", err)
	}

	// Add description as a note if present
	if issue.Description != "" {
		if err := addTaskNote(db, taskUID, issue.Description, actor, createdAt); err != nil {
			return "", fmt.Errorf("failed to add task note: %w", err)
		}
	}

	// Set status if not open
	if issue.Status != "" && issue.Status != "open" {
		tkStatus := MapBeadsStatus(issue.Status)
		if err := setTaskStatus(db, taskUID, tkStatus, actor, createdAt); err != nil {
			return "", fmt.Errorf("failed to set task status: %w", err)
		}
	}

	// Import metadata: priority
	if issue.Priority >= 0 && issue.Priority <= 4 {
		if err := createMetadataEvent(db, taskUID, "priority", json.RawMessage(fmt.Sprintf("%d", issue.Priority)), actor, createdAt); err != nil {
			return "", fmt.Errorf("failed to create priority metadata: %w", err)
		}
	}

	// Import metadata: labels
	if len(issue.Labels) > 0 {
		labelsJSON, err := json.Marshal(issue.Labels)
		if err != nil {
			return "", fmt.Errorf("failed to marshal labels: %w", err)
		}
		if err := createMetadataEvent(db, taskUID, "labels", json.RawMessage(labelsJSON), actor, createdAt); err != nil {
			return "", fmt.Errorf("failed to create labels metadata: %w", err)
		}
	}

	return taskUID, nil
}

// parseCreatedAt parses the created_at timestamp from beads format
func parseCreatedAt(createdAtStr string) time.Time {
	if createdAtStr == "" {
		return time.Now()
	}

	// Try RFC3339 first
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err == nil {
		return createdAt
	}

	// Try other common formats
	createdAt, err = time.Parse("2006-01-02T15:04:05", createdAtStr)
	if err == nil {
		return createdAt
	}

	// Fallback to current time
	return time.Now()
}

// createTaskCreatedEvent creates a task.created event
func createTaskCreatedEvent(db *database.DB, taskUID, projectUID string, number int64, title, nodeID, actor string, createdAt time.Time) error {
	payload := types.TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: number,
		Title:          title,
		CreatedNode:    nodeID,
		CreatedBy:      actor,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return fmt.Errorf("failed to generate event ID: %w", err)
	}

	ts, err := db.GetNextLamportTS()
	if err != nil {
		return fmt.Errorf("failed to get lamport timestamp: %w", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: createdAt,
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert task created event: %w", err)
	}

	if err := db.ProjectTaskCreatedEvent(event); err != nil {
		return fmt.Errorf("failed to project task created event: %w", err)
	}

	return nil
}

// createTaskNumberEvent creates a task.number.set event
func createTaskNumberEvent(db *database.DB, taskUID, projectUID string, number int64, actor string, createdAt time.Time) error {
	payload := types.TaskNumberSetPayload{
		TaskUID:    types.TaskUID(taskUID),
		ProjectUID: types.ProjectUID(projectUID),
		Number:     number,
		Reason:     "imported",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal number payload: %w", err)
	}

	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return fmt.Errorf("failed to generate event ID: %w", err)
	}

	ts, err := db.GetNextLamportTS()
	if err != nil {
		return fmt.Errorf("failed to get lamport timestamp: %w", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: createdAt,
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindTaskNumberSet),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert number event: %w", err)
	}

	if err := db.ProjectTaskNumberSetEvent(event); err != nil {
		return fmt.Errorf("failed to project task number: %w", err)
	}

	return nil
}

// addTaskNote adds a note to a task
func addTaskNote(db *database.DB, taskUID, markdown, actor string, createdAt time.Time) error {
	payload := types.TaskNoteAddPayload{
		TaskUUID: taskUID,
		TaskID:   "",
		Markdown: markdown,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal note payload: %w", err)
	}

	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return fmt.Errorf("failed to generate event ID: %w", err)
	}

	ts, err := db.GetNextLamportTS()
	if err != nil {
		return fmt.Errorf("failed to get lamport timestamp: %w", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: createdAt,
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindTaskNoteAdd),
		Payload:   payloadJSON,
	}

	return db.InsertEvent(event)
}

// setTaskStatus sets the status of a task
func setTaskStatus(db *database.DB, taskUID, status, actor string, createdAt time.Time) error {
	payload := types.TaskStatusSetPayload{
		TaskUUID: taskUID,
		TaskID:   "",
		Axis:     "generic",
		State:    status,
		Role:     "human",
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal status payload: %w", err)
	}

	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return fmt.Errorf("failed to generate event ID: %w", err)
	}

	ts, err := db.GetNextLamportTS()
	if err != nil {
		return fmt.Errorf("failed to get lamport timestamp: %w", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: createdAt,
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindTaskStatusSet),
		Payload:   payloadJSON,
	}

	return db.InsertEvent(event)
}

// createMetadataEvent creates a task.meta.set event
func createMetadataEvent(db *database.DB, taskUID string, key string, value json.RawMessage, actor string, createdAt time.Time) error {
	payload := types.TaskMetaSetPayload{
		TaskUUID: taskUID,
		TaskID:   "",
		Key:      key,
		Value:    value,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata payload: %w", err)
	}

	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return fmt.Errorf("failed to generate event ID: %w", err)
	}

	ts, err := db.GetNextLamportTS()
	if err != nil {
		return fmt.Errorf("failed to get lamport timestamp: %w", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: createdAt,
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   payloadJSON,
	}

	if err := db.InsertEvent(event); err != nil {
		return fmt.Errorf("failed to insert metadata event: %w", err)
	}

	return nil
}

// AddRenumberNote adds a note to a task explaining it was renumbered during import
func AddRenumberNote(db *database.DB, taskUID, originalID string, newNumber int64) error {
	note := fmt.Sprintf("Note: Original beads ID was %s (non-numeric), renumbered to %d during import",
		originalID, newNumber)

	return addTaskNote(db, taskUID, note, "importer", time.Now())
}

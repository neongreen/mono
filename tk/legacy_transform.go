package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TransformLegacyEvent transforms legacy v1/v2/v3 events into v4 events
// Returns empty slice if event is not a legacy event that needs transformation
// projectResolver is a function that resolves a prefix to a project UID
func TransformLegacyEvent(event Event, projectResolver func(prefix, description, createdBy string, createdAt time.Time, nodeID string) (string, error)) ([]Event, error) {
	switch event.Kind {
	case "prefix.created":
		return TransformPrefixCreated(event, projectResolver)
	case "task.created":
		// Check if it's legacy format (has task_id field, not task_uid)
		var payload TaskCreatedPayload
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return nil, nil // Not a legacy task.created, return empty
		}
		// Legacy format has task_id field with format prefix-number-node
		if payload.TaskID != "" && payload.TaskUUID == "" {
			return TransformLegacyTaskCreated(event, projectResolver)
		}
		return nil, nil // Already v4 format
	case "task.reprefix":
		return TransformTaskReprefix(event, projectResolver)
	default:
		return nil, nil // Not a legacy event that needs transformation
	}
}

// TransformPrefixCreated transforms prefix.created → project.created + project.alias.add
func TransformPrefixCreated(event Event, projectResolver func(prefix, description, createdBy string, createdAt time.Time, nodeID string) (string, error)) ([]Event, error) {
	var payload PrefixCreatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse prefix.created payload: %w", err)
	}

	// Extract node from event ID if possible (format: ev-<number>-<node>)
	nodeID := ""
	parts := strings.Split(event.ID, "-")
	if len(parts) >= 3 {
		nodeID = parts[2]
	}

	// Resolve project UID for this prefix
	projectUID, err := projectResolver(payload.Prefix, payload.Description, payload.CreatedBy, event.CreatedAt, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project for prefix %s: %w", payload.Prefix, err)
	}

	// Create project.created event
	projectCreatedPayload := ProjectCreatedPayload{
		ProjectUID:  projectUID,
		Type:        "local",
		Name:        payload.Prefix,
		Description: payload.Description,
		CreatedBy:   payload.CreatedBy,
	}
	projectCreatedJSON, err := json.Marshal(projectCreatedPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal project.created payload: %w", err)
	}

	projectCreatedEvent := Event{
		ID:        string(NewEventID()),
		TS:        event.TS,
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindProjectCreated),
		Payload:   projectCreatedJSON,
	}

	// Create project.alias.add event
	aliasPayload := ProjectAliasAddPayload{
		ProjectUID: projectUID,
		Alias:      payload.Prefix,
		Node:       nodeID,
		AddedBy:    payload.CreatedBy,
	}
	aliasJSON, err := json.Marshal(aliasPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal project.alias.add payload: %w", err)
	}

	aliasEvent := Event{
		ID:        string(NewEventID()),
		TS:        event.TS + 1, // Slightly after project.created
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindProjectAliasAdd),
		Payload:   aliasJSON,
	}

	return []Event{projectCreatedEvent, aliasEvent}, nil
}

// TransformLegacyTaskCreated transforms legacy task.created → v4 task.created + task.number.set
func TransformLegacyTaskCreated(event Event, projectResolver func(prefix, description, createdBy string, createdAt time.Time, nodeID string) (string, error)) ([]Event, error) {
	var payload TaskCreatedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse legacy task.created payload: %w", err)
	}

	// Parse task_id to extract prefix, number, node
	prefix, number, node, err := ParseTaskIDLegacy(payload.TaskID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse task_id %s: %w", payload.TaskID, err)
	}

	// Resolve project UID for this prefix
	projectUID, err := projectResolver(prefix, "", payload.CreatedBy, event.CreatedAt, node)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project for prefix %s: %w", prefix, err)
	}

	// Generate new task UID
	taskUID := string(NewTaskUID())

	// Create v4 task.created event
	taskCreatedPayload := TaskCreatedV4Payload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: number,
		CreatedNode:    node,
		Title:          payload.Title,
		CreatedBy:      payload.CreatedBy,
	}
	taskCreatedJSON, err := json.Marshal(taskCreatedPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task.created payload: %w", err)
	}

	taskCreatedEvent := Event{
		ID:        string(NewEventID()),
		TS:        event.TS,
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindTaskCreated),
		Payload:   taskCreatedJSON,
	}

	// Create task.number.set event
	numberPayload := TaskNumberSetPayload{
		TaskUID:    taskUID,
		ProjectUID: projectUID,
		Number:     number,
		Reason:     "migrated from legacy",
	}
	numberJSON, err := json.Marshal(numberPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task.number.set payload: %w", err)
	}

	numberEvent := Event{
		ID:        string(NewEventID()),
		TS:        event.TS + 1, // Slightly after task.created
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindTaskNumberSet),
		Payload:   numberJSON,
	}

	return []Event{taskCreatedEvent, numberEvent}, nil
}

// TransformTaskReprefix transforms task.reprefix → task.relocate
func TransformTaskReprefix(event Event, projectResolver func(prefix, description, createdBy string, createdAt time.Time, nodeID string) (string, error)) ([]Event, error) {
	var payload TaskReprefixPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse task.reprefix payload: %w", err)
	}

	// Resolve project UIDs for old and new prefixes
	fromProjectUID, err := projectResolver(payload.OldPrefix, "", event.Actor, event.CreatedAt, payload.OldNode)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project for prefix %s: %w", payload.OldPrefix, err)
	}

	toProjectUID, err := projectResolver(payload.NewPrefix, "", event.Actor, event.CreatedAt, payload.OldNode)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project for prefix %s: %w", payload.NewPrefix, err)
	}

	// Get task UID - need to resolve from legacy task
	// For now, we'll need to resolve it from the task_id or task_uuid
	// This is a simplified version - full implementation would need to track UUID mappings
	taskUID := string(NewTaskUID())

	// Create task.relocate event
	relocatePayload := TaskRelocatePayload{
		TaskUID:        taskUID,
		FromProjectUID: fromProjectUID,
		ToProjectUID:   toProjectUID,
		NumberPolicy: NumberPolicyPayload{
			Mode:   "force",
			Number: payload.NewNumber,
		},
	}
	relocateJSON, err := json.Marshal(relocatePayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task.relocate payload: %w", err)
	}

	relocateEvent := Event{
		ID:        string(NewEventID()),
		TS:        event.TS,
		CreatedAt: event.CreatedAt,
		Actor:     event.Actor,
		Role:      event.Role,
		Kind:      string(EventKindTaskRelocate),
		Payload:   relocateJSON,
	}

	return []Event{relocateEvent}, nil
}

// ParseTaskIDLegacy extracts prefix, number, and node from a legacy task ID
// Format: prefix-number-node
func ParseTaskIDLegacy(taskID string) (prefix string, number int64, node string, err error) {
	parts := strings.Split(taskID, "-")
	if len(parts) != 3 {
		return "", 0, "", fmt.Errorf("invalid task ID format: expected prefix-number-node")
	}

	prefix = parts[0]
	number, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid number in task ID: %w", err)
	}
	node = parts[2]

	return prefix, number, node, nil
}


package testutil

import (
	"encoding/json"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

// Event Factory Functions
//
// This file provides simple helper functions for creating test events.
//
// NOTE: This somewhat duplicates the functionality of the reducer test DSL
// (internal/reducer/test_dsl.go). The DSL provides a fluent, stateful API
// for testing reducer behavior, while these are simple stateless factories
// for quick event creation in database/cmd/tasks tests.
//
// FUTURE IMPROVEMENT: It might be nice to extend the reducer DSL to work
// across all test packages, which would provide a consistent testing interface
// and eliminate this duplication. The DSL would need to be made more flexible
// (allow custom timestamps, event IDs, etc.) to support all use cases.

// CreateProjectEvent creates a project.created event
func CreateProjectEvent(ts int64, projectUID types.ProjectUID, name string) types.Event {
	payload := types.ProjectCreatedPayload{
		ProjectUID:  projectUID,
		Type:        types.ProjectTypeLocal,
		Name:        name,
		Description: "Test project: " + name,
		CreatedBy:   "test",
	}
	payloadJSON, _ := json.Marshal(payload)

	return types.Event{
		ID:        types.NewEventID().String(),
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindProjectCreated),
		Payload:   payloadJSON,
	}
}

// CreateTaskEvent creates a task.created event
func CreateTaskEvent(ts int64, taskUID types.TaskUID, projectUID string, title string) types.Event {
	payload := types.TaskCreatedPayload{
		TaskUID:        string(taskUID),
		ProjectUID:     projectUID,
		ProposedNumber: 1,
		CreatedNode:    string(types.NewNodeID()),
		Title:          title,
		CreatedBy:      "test",
	}
	payloadJSON, _ := json.Marshal(payload)

	return types.Event{
		ID:        types.NewEventID().String(),
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindTaskCreated),
		Payload:   payloadJSON,
	}
}

// SetTaskNumberEvent creates a task.number.set event
func SetTaskNumberEvent(ts int64, taskUID types.TaskUID, projectUID types.ProjectUID, number int64) types.Event {
	payload := types.TaskNumberSetPayload{
		TaskUID:    taskUID,
		ProjectUID: projectUID,
		Number:     number,
		Reason:     "test",
	}
	payloadJSON, _ := json.Marshal(payload)

	return types.Event{
		ID:        types.NewEventID().String(),
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "test",
		Role:      "human",
		Kind:      string(types.EventKindTaskNumberSet),
		Payload:   payloadJSON,
	}
}

// Lowercase aliases for backward compatibility with deleted test code

var createProjectEvent = CreateProjectEvent
var createTaskEvent = CreateTaskEvent
var setTaskNumberEvent = SetTaskNumberEvent
var createTempDB = OpenTempDB

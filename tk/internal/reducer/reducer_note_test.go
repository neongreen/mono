package reducer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

func TestReducer_NoteAdd(t *testing.T) {
	r := NewReducer()

	// Create task
	taskUID := string(types.NewTaskUID())
	projectUID := string(types.NewProjectUID())
	createPayload := types.TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: 1,
		CreatedNode:    string(types.NewNodeID()),
		Title:          "Test task",
		CreatedBy:      "alice",
	}
	createPayloadJSON, _ := json.Marshal(createPayload)

	createEvent := types.Event{
		ID:        "event_01",
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.created",
		Payload:   createPayloadJSON,
	}

	r.Apply(createEvent)

	// Add note
	notePayload := types.TaskNoteAddPayload{
		TaskUUID: taskUID,
		Markdown: "This is a test note",
	}
	notePayloadJSON, _ := json.Marshal(notePayload)

	noteEvent := types.Event{
		ID:        "event_02",
		TS:        2,
		CreatedAt: time.Now(),
		Actor:     "alice",
		Role:      "human",
		Kind:      "task.note.add",
		Payload:   notePayloadJSON,
	}

	err := r.Apply(noteEvent)
	if err != nil {
		t.Fatalf("Failed to apply task.note.add event: %v", err)
	}

	task, _ := r.GetTask(taskUID)

	if len(task.Notes) != 1 {
		t.Fatalf("Expected 1 note, got %d", len(task.Notes))
	}

	if task.Notes[0].Markdown != "This is a test note" {
		t.Errorf("Expected note 'This is a test note', got %s", task.Notes[0].Markdown)
	}

	if task.Notes[0].Actor != "alice" {
		t.Errorf("Expected note actor alice, got %s", task.Notes[0].Actor)
	}
}

package cmd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/reducer"
	"github.com/neongreen/mono/tk/internal/types"
)

func TestMetaIntegration_SetAndGet(t *testing.T) {
	db := openTempDB(t)
	projectUID := seedProject(t, db, "test")
	taskUID := seedTask(t, db, projectUID, "Test metadata task", 1)

	// Set metadata via event
	payload := types.TaskMetaSetPayload{
		TaskUUID: taskUID,
		TaskID:   "",
		Key:      "priority",
		Value:    json.RawMessage(`1`),
	}

	eventID, err := database.GenerateEventID(db)
	if err != nil {
		t.Fatalf("failed to generate event ID: %v", err)
	}

	ts, err := db.GetNextLamportTS()
	if err != nil {
		t.Fatalf("failed to get lamport ts: %v", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        ts,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   mustJSON(t, payload),
	}

	if err := db.InsertEvent(event); err != nil {
		t.Fatalf("failed to insert meta event: %v", err)
	}

	// Rebuild from events and check
	events, err := db.GetEvents()
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}

	// Debug: print event order
	t.Logf("Total events: %d", len(events))
	for i, e := range events {
		t.Logf("Event %d: %s (kind: %s, ts: %d)", i, e.ID, e.Kind, e.TS)
	}

	r := buildReducerFromDB(t, db)
	task, ok := r.GetTask(taskUID)
	if !ok {
		t.Fatalf("task not found: %s", taskUID)
	}

	if task.Metadata == nil {
		t.Fatal("metadata not initialized")
	}

	metaStatus, ok := task.Metadata["priority"]
	if !ok {
		t.Fatal("priority metadata not found")
	}

	if string(metaStatus.Effective) != "1" {
		t.Errorf("Expected priority=1, got %s", metaStatus.Effective)
	}
}

func TestMetaIntegration_AuthorityResolution(t *testing.T) {
	db := openTempDB(t)
	projectUID := seedProject(t, db, "test")
	taskUID := seedTask(t, db, projectUID, "Test authority", 1)

	// Agent sets priority=3
	agentPayload := types.TaskMetaSetPayload{
		TaskUUID: taskUID,
		Key:      "priority",
		Value:    json.RawMessage(`3`),
	}

	agentEventID, _ := database.GenerateEventID(db)
	agentTS, _ := db.GetNextLamportTS()
	agentEvent := types.Event{
		ID:        agentEventID,
		TS:        agentTS,
		CreatedAt: time.Now(),
		Actor:     "claude",
		Role:      "agent",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   mustJSON(t, agentPayload),
	}
	db.InsertEvent(agentEvent)

	// Human sets priority=1
	humanPayload := types.TaskMetaSetPayload{
		TaskUUID: taskUID,
		Key:      "priority",
		Value:    json.RawMessage(`1`),
	}

	humanEventID, _ := database.GenerateEventID(db)
	humanTS, _ := db.GetNextLamportTS()
	humanEvent := types.Event{
		ID:        humanEventID,
		TS:        humanTS,
		CreatedAt: time.Now(),
		Actor:     "emily",
		Role:      "human",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   mustJSON(t, humanPayload),
	}
	db.InsertEvent(humanEvent)

	// Rebuild and check
	events, _ := db.GetEvents()
	r := reducer.NewReducer()
	for _, e := range events {
		if err := r.Apply(e); err != nil {
			t.Fatalf("failed to apply event: %v", err)
		}
	}

	task, _ := r.GetTask(taskUID)
	metaStatus := task.Metadata["priority"]

	// Human should win
	if string(metaStatus.Effective) != "1" {
		t.Errorf("Expected human value (1), got %s", metaStatus.Effective)
	}

	// Agent claim should be tentative
	var agentClaim *types.MetadataClaim
	for i := range metaStatus.Claims {
		if metaStatus.Claims[i].Role == "agent" {
			agentClaim = &metaStatus.Claims[i]
			break
		}
	}

	if agentClaim == nil {
		t.Fatal("agent claim not found")
	}

	if !agentClaim.Tentative {
		t.Error("agent claim should be tentative when human overrides")
	}
}

func TestMetaIntegration_MultipleKeys(t *testing.T) {
	db := openTempDB(t)
	projectUID := seedProject(t, db, "test")
	taskUID := seedTask(t, db, projectUID, "Test multiple keys", 1)

	// Set priority
	priorityPayload := types.TaskMetaSetPayload{
		TaskUUID: taskUID,
		Key:      "priority",
		Value:    json.RawMessage(`2`),
	}
	priorityEventID, _ := database.GenerateEventID(db)
	priorityTS, _ := db.GetNextLamportTS()
	db.InsertEvent(types.Event{
		ID:        priorityEventID,
		TS:        priorityTS,
		CreatedAt: time.Now(),
		Actor:     "emily",
		Role:      "human",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   mustJSON(t, priorityPayload),
	})

	// Set labels
	labelsPayload := types.TaskMetaSetPayload{
		TaskUUID: taskUID,
		Key:      "labels",
		Value:    json.RawMessage(`["bug","urgent"]`),
	}
	labelsEventID, _ := database.GenerateEventID(db)
	labelsTS, _ := db.GetNextLamportTS()
	db.InsertEvent(types.Event{
		ID:        labelsEventID,
		TS:        labelsTS,
		CreatedAt: time.Now(),
		Actor:     "emily",
		Role:      "human",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   mustJSON(t, labelsPayload),
	})

	// Rebuild and check
	r := buildReducerFromDB(t, db)
	task, _ := r.GetTask(taskUID)

	// Check both keys exist
	if task.Metadata == nil {
		t.Fatal("metadata not initialized")
	}

	if _, ok := task.Metadata["priority"]; !ok {
		t.Error("priority metadata not found")
	}

	if _, ok := task.Metadata["labels"]; !ok {
		t.Error("labels metadata not found")
	}

	if string(task.Metadata["priority"].Effective) != "2" {
		t.Errorf("Expected priority=2, got %s", task.Metadata["priority"].Effective)
	}

	if string(task.Metadata["labels"].Effective) != `["bug","urgent"]` {
		t.Errorf("Expected labels array, got %s", task.Metadata["labels"].Effective)
	}
}

package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

func openTempDB(t *testing.T) *DB {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "tk.db")

	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("failed to open temp db: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	if err := db.InitDB(); err != nil {
		t.Fatalf("failed to initialise db: %v", err)
	}
	if err := db.CreateV4Tables(); err != nil {
		t.Fatalf("failed to create v4 tables: %v", err)
	}
	if err := db.SetDBVersion(v4SpecVersion); err != nil {
		t.Fatalf("failed to set v4 version: %v", err)
	}
	if _, err := db.db.Exec(`
		INSERT OR REPLACE INTO metadata (key, value)
		VALUES ('remote_subdir', ?)
	`, v4SegmentSubdir); err != nil {
		t.Fatalf("failed to set remote_subdir: %v", err)
	}
	return db
}

func seedProject(t *testing.T, db *DB, alias string) string {
	t.Helper()
	projectUID := string(NewProjectUID())
	now := time.Now()

	projectPayload := ProjectCreatedPayload{
		ProjectUID:  projectUID,
		Type:        "local",
		Name:        alias,
		Description: alias + " project",
		CreatedBy:   "tester",
	}

	payloadJSON := mustJSON(t, projectPayload)
	event := Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: now,
		Actor:     "tester",
		Role:      "human",
		Kind:      string(EventKindProjectCreated),
		Payload:   payloadJSON,
	}
	if err := db.InsertEvent(event); err != nil {
		t.Fatalf("failed to insert project.created: %v", err)
	}
	if err := db.ProjectProjectCreatedEvent(event); err != nil {
		t.Fatalf("failed to project project.created: %v", err)
	}

	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node id: %v", err)
	}

	aliasPayload := ProjectAliasAddPayload{
		ProjectUID: projectUID,
		Alias:      alias,
		Node:       nodeID,
		AddedBy:    "tester",
	}
	aliasJSON := mustJSON(t, aliasPayload)
	aliasEvent := Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: now,
		Actor:     "tester",
		Role:      "human",
		Kind:      string(EventKindProjectAliasAdd),
		Payload:   aliasJSON,
	}
	if err := db.InsertEvent(aliasEvent); err != nil {
		t.Fatalf("failed to insert project.alias.add: %v", err)
	}
	if err := db.ProjectProjectAliasAddEvent(aliasEvent); err != nil {
		t.Fatalf("failed to project project.alias.add: %v", err)
	}

	return projectUID
}

func seedTask(t *testing.T, db *DB, projectUID string, title string, number int64) string {
	nodeID, err := db.GetOrCreateNodeID()
	if err != nil {
		t.Fatalf("failed to get node id: %v", err)
	}
	return seedTaskWithNode(t, db, projectUID, title, number, nodeID)
}

func seedTaskWithNode(t *testing.T, db *DB, projectUID string, title string, number int64, nodeID string) string {
	t.Helper()
	taskUID := string(NewTaskUID())
	now := time.Now()

	taskPayload := TaskCreatedV4Payload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: number,
		CreatedNode:    nodeID,
		Title:          title,
		CreatedBy:      "tester",
	}

	taskJSON := mustJSON(t, taskPayload)
	taskEvent := Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: now,
		Actor:     "tester",
		Role:      "human",
		Kind:      string(EventKindTaskCreated),
		Payload:   taskJSON,
	}
	if err := db.InsertEvent(taskEvent); err != nil {
		t.Fatalf("failed to insert task.created: %v", err)
	}
	if err := db.ProjectTaskCreatedV4Event(taskEvent); err != nil {
		t.Fatalf("failed to project task.created: %v", err)
	}

	numberPayload := TaskNumberSetPayload{
		TaskUID:    taskUID,
		ProjectUID: projectUID,
		Number:     number,
		Reason:     "seed",
	}
	numberJSON := mustJSON(t, numberPayload)
	numberEvent := Event{
		ID:        string(NewEventID()),
		TS:        0,
		CreatedAt: now,
		Actor:     "tester",
		Role:      "human",
		Kind:      string(EventKindTaskNumberSet),
		Payload:   numberJSON,
	}
	if err := db.InsertEvent(numberEvent); err != nil {
		t.Fatalf("failed to insert task.number.set: %v", err)
	}
	if err := db.ProjectTaskNumberSetEvent(numberEvent); err != nil {
		t.Fatalf("failed to project task.number.set: %v", err)
	}

	return taskUID
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	return json.RawMessage(data)
}

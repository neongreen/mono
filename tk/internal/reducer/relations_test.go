package reducer

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/types"
)

func TestReducer_RelationAdd(t *testing.T) {
	r := NewReducer()

	task1UID := string(types.NewTaskUID())
	task2UID := string(types.NewTaskUID())
	projectUID := string(types.NewProjectUID())

	// Create task 1
	createTask(t, r, task1UID, projectUID, "Task 1")
	// Create task 2
	createTask(t, r, task2UID, projectUID, "Task 2")

	// Add blocks relation
	addRelation(t, r, task1UID, "blocks", task2UID, "")

	r.FinalizeRelations(&config.Config{})

	// Verify relation exists
	task1, _ := r.GetTask(task1UID)
	if len(task1.Relations.Blocks.Out) != 1 {
		t.Fatalf("expected 1 outgoing blocks relation, got %d", len(task1.Relations.Blocks.Out))
	}
	if task1.Relations.Blocks.Out[0].TaskUUID != task2UID {
		t.Errorf("expected task2 as target, got %s", task1.Relations.Blocks.Out[0].TaskUUID)
	}

	// Verify incoming relation on task 2
	task2, _ := r.GetTask(task2UID)
	if len(task2.Relations.Blocks.In) != 1 {
		t.Fatalf("expected 1 incoming blocks relation, got %d", len(task2.Relations.Blocks.In))
	}
	if task2.Relations.Blocks.In[0].TaskUUID != task1UID {
		t.Errorf("expected task1 as source, got %s", task2.Relations.Blocks.In[0].TaskUUID)
	}
}

func TestReducer_RelationSubtask(t *testing.T) {
	r := NewReducer()

	parentUID := string(types.NewTaskUID())
	childUID := string(types.NewTaskUID())
	projectUID := string(types.NewProjectUID())

	createTask(t, r, parentUID, projectUID, "Parent")
	createTask(t, r, childUID, projectUID, "Child")

	// Add subtask relation
	addRelation(t, r, parentUID, "subtask", childUID, "")

	r.FinalizeRelations(&config.Config{})

	// Verify parent has child
	parent, _ := r.GetTask(parentUID)
	if len(parent.Relations.Subtask.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(parent.Relations.Subtask.Children))
	}
	if parent.Relations.Subtask.Children[0] != childUID {
		t.Errorf("expected child UID %s, got %s", childUID, parent.Relations.Subtask.Children[0])
	}

	// Verify child has parent
	child, _ := r.GetTask(childUID)
	if child.Relations.Subtask.Parent != parentUID {
		t.Errorf("expected parent UID %s, got %s", parentUID, child.Relations.Subtask.Parent)
	}
}

func TestReducer_RelationRemove(t *testing.T) {
	r := NewReducer()

	task1UID := string(types.NewTaskUID())
	task2UID := string(types.NewTaskUID())
	projectUID := string(types.NewProjectUID())

	createTask(t, r, task1UID, projectUID, "Task 1")
	createTask(t, r, task2UID, projectUID, "Task 2")

	// Add relation
	addRelation(t, r, task1UID, "blocks", task2UID, "")

	// Remove relation
	removeRelation(t, r, task1UID, "blocks", task2UID)

	r.FinalizeRelations(&config.Config{})

	// Verify relation was removed
	task1, _ := r.GetTask(task1UID)
	if task1.Relations != nil && len(task1.Relations.Blocks.Out) != 0 {
		t.Errorf("expected 0 outgoing blocks relations after removal, got %d", len(task1.Relations.Blocks.Out))
	}

	task2, _ := r.GetTask(task2UID)
	if task2.Relations != nil && len(task2.Relations.Blocks.In) != 0 {
		t.Errorf("expected 0 incoming blocks relations after removal, got %d", len(task2.Relations.Blocks.In))
	}
}

func TestReducer_RelationAddRemoveAdd_ORSet(t *testing.T) {
	r := NewReducer()

	task1UID := string(types.NewTaskUID())
	task2UID := string(types.NewTaskUID())
	projectUID := string(types.NewProjectUID())

	createTask(t, r, task1UID, projectUID, "Task 1")
	createTask(t, r, task2UID, projectUID, "Task 2")

	// Add relation (event 1)
	addRelation(t, r, task1UID, "blocks", task2UID, "")

	// Remove relation (event 2)
	removeRelation(t, r, task1UID, "blocks", task2UID)

	// Add relation again (event 3) - should work due to OR-set semantics
	addRelationWithEventID(t, r, task1UID, "blocks", task2UID, "", "event_03")

	r.FinalizeRelations(&config.Config{})

	// Relation should exist (OR-set: new add wins over old remove)
	task1, _ := r.GetTask(task1UID)
	if len(task1.Relations.Blocks.Out) != 1 {
		t.Errorf("expected 1 blocks relation (OR-set: add after remove), got %d", len(task1.Relations.Blocks.Out))
	}
}

func TestReducer_RelationNote(t *testing.T) {
	r := NewReducer()

	task1UID := string(types.NewTaskUID())
	task2UID := string(types.NewTaskUID())
	projectUID := string(types.NewProjectUID())

	createTask(t, r, task1UID, projectUID, "Task 1")
	createTask(t, r, task2UID, projectUID, "Task 2")

	// Add relation with note
	addRelation(t, r, task1UID, "related", task2UID, "Both work on auth")

	r.FinalizeRelations(&config.Config{})

	// Verify note was stored
	task1, _ := r.GetTask(task1UID)
	if len(task1.Relations.Related.Out) != 1 {
		t.Fatalf("expected 1 related relation, got %d", len(task1.Relations.Related.Out))
	}
	if task1.Relations.Related.Out[0].Note != "Both work on auth" {
		t.Errorf("expected note 'Both work on auth', got %q", task1.Relations.Related.Out[0].Note)
	}
}

// Helper functions

func createTask(t *testing.T, r *Reducer, taskUID string, projectUID string, title string) {
	t.Helper()
	payload := types.TaskCreatedPayload{
		TaskUID:        taskUID,
		ProjectUID:     projectUID,
		ProposedNumber: 1,
		CreatedNode:    string(types.NewNodeID()),
		Title:          title,
		CreatedBy:      "tester",
	}
	payloadJSON, _ := json.Marshal(payload)

	event := types.Event{
		ID:        "create_" + taskUID,
		TS:        1,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      "task.created",
		Payload:   payloadJSON,
	}

	if err := r.Apply(event); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
}

func addRelation(t *testing.T, r *Reducer, src string, relType string, dst string, note string) {
	t.Helper()
	addRelationWithEventID(t, r, src, relType, dst, note, "rel_"+src+"_"+dst)
}

func addRelationWithEventID(t *testing.T, r *Reducer, src string, relType string, dst string, note string, eventID string) {
	t.Helper()
	payload := types.RelationAddPayload{
		Src:  src,
		Type: relType,
		Dst:  dst,
		Note: note,
	}
	payloadJSON, _ := json.Marshal(payload)

	event := types.Event{
		ID:        eventID,
		TS:        100,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      "relation.add",
		Payload:   payloadJSON,
	}

	if err := r.Apply(event); err != nil {
		t.Fatalf("failed to add relation: %v", err)
	}
}

func removeRelation(t *testing.T, r *Reducer, src string, relType string, dst string) {
	t.Helper()
	payload := types.RelationRemovePayload{
		Src:  src,
		Type: relType,
		Dst:  dst,
	}
	payloadJSON, _ := json.Marshal(payload)

	event := types.Event{
		ID:        "remove_" + src + "_" + dst,
		TS:        200,
		CreatedAt: time.Now(),
		Actor:     "tester",
		Role:      "human",
		Kind:      "relation.remove",
		Payload:   payloadJSON,
	}

	if err := r.Apply(event); err != nil {
		t.Fatalf("failed to remove relation: %v", err)
	}
}

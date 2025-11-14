package taskset

import (
	"testing"

	"github.com/neongreen/mono/lib/setlang"
	"github.com/neongreen/mono/tk/internal/types"
)

func TestTaskContext_Status(t *testing.T) {
	tasks := []*types.Task{
		{TaskUUID: "t1", Title: "Task 1", Axes: map[string]types.AxisStatus{
			"generic": {Effective: "wip"},
		}},
		{TaskUUID: "t2", Title: "Task 2", Axes: map[string]types.AxisStatus{
			"generic": {Effective: "done"},
		}},
		{TaskUUID: "t3", Title: "Task 3", Axes: map[string]types.AxisStatus{
			"generic": {Effective: "wip"},
		}},
	}

	ctx := NewTaskContext(tasks, nil)

	result, err := setlang.Eval(ctx, "status(wip)")
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	items := result.Items()
	if len(items) != 2 {
		t.Errorf("Expected 2 wip tasks, got %d", len(items))
	}

	// Check that t1 and t3 are in results
	has := make(map[string]bool)
	for _, item := range items {
		has[item] = true
	}
	if !has["t1"] || !has["t3"] {
		t.Errorf("Expected t1 and t3 in results, got %v", items)
	}
}

func TestTaskContext_Kind(t *testing.T) {
	tasks := []*types.Task{
		{TaskUUID: "t1", ItemKind: "task"},
		{TaskUUID: "t2", ItemKind: "decision"},
		{TaskUUID: "t3", ItemKind: "decision"},
		{TaskUUID: "t4", ItemKind: ""}, // Empty defaults to "task"
	}

	ctx := NewTaskContext(tasks, nil)

	result, err := setlang.Eval(ctx, "kind(decision)")
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	items := result.Items()
	if len(items) != 2 {
		t.Errorf("Expected 2 decisions, got %d", len(items))
	}

	// Test default kind
	result, err = setlang.Eval(ctx, "kind(task)")
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	items = result.Items()
	if len(items) != 2 { // t1 and t4
		t.Errorf("Expected 2 tasks (including default), got %d", len(items))
	}
}

func TestTaskContext_Project(t *testing.T) {
	projectMap := map[string]string{
		"prj-tk":   "tk",
		"prj-mono": "mono",
	}

	tasks := []*types.Task{
		{TaskUUID: "t1", ProjectUUID: "prj-tk"},
		{TaskUUID: "t2", ProjectUUID: "prj-mono"},
		{TaskUUID: "t3", ProjectUUID: "prj-tk"},
	}

	ctx := NewTaskContext(tasks, projectMap)

	result, err := setlang.Eval(ctx, "project(tk)")
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	items := result.Items()
	if len(items) != 2 {
		t.Errorf("Expected 2 tk tasks, got %d", len(items))
	}
}

func TestTaskContext_Blocked(t *testing.T) {
	tasks := []*types.Task{
		{TaskUUID: "t1", Blocked: true},
		{TaskUUID: "t2", Blocked: false},
		{TaskUUID: "t3", Blocked: true},
	}

	ctx := NewTaskContext(tasks, nil)

	result, err := setlang.Eval(ctx, "blocked()")
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	items := result.Items()
	if len(items) != 2 {
		t.Errorf("Expected 2 blocked tasks, got %d", len(items))
	}
}

func TestTaskContext_Author(t *testing.T) {
	tasks := []*types.Task{
		{TaskUUID: "t1", CreatedBy: "alice"},
		{TaskUUID: "t2", CreatedBy: "bob"},
		{TaskUUID: "t3", CreatedBy: "alice"},
	}

	ctx := NewTaskContext(tasks, nil)

	result, err := setlang.Eval(ctx, "author(alice)")
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	items := result.Items()
	if len(items) != 2 {
		t.Errorf("Expected 2 tasks by alice, got %d", len(items))
	}
}

func TestTaskContext_Title(t *testing.T) {
	tasks := []*types.Task{
		{TaskUUID: "t1", Title: "Fix bug in parser"},
		{TaskUUID: "t2", Title: "Add new feature"},
		{TaskUUID: "t3", Title: "Fix another bug"},
	}

	ctx := NewTaskContext(tasks, nil)

	result, err := setlang.Eval(ctx, `title("bug")`)
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	items := result.Items()
	if len(items) != 2 {
		t.Errorf("Expected 2 tasks with 'bug', got %d", len(items))
	}
}

func TestTaskContext_ComplexQuery(t *testing.T) {
	projectMap := map[string]string{
		"prj-tk": "tk",
	}

	tasks := []*types.Task{
		{
			TaskUUID:    "t1",
			ProjectUUID: "prj-tk",
			ItemKind:    "task",
			Axes:        map[string]types.AxisStatus{"generic": {Effective: "wip"}},
		},
		{
			TaskUUID:    "t2",
			ProjectUUID: "prj-tk",
			ItemKind:    "decision",
			Axes:        map[string]types.AxisStatus{"generic": {Effective: "wip"}},
		},
		{
			TaskUUID:    "t3",
			ProjectUUID: "prj-tk",
			ItemKind:    "task",
			Axes:        map[string]types.AxisStatus{"generic": {Effective: "done"}},
		},
	}

	ctx := NewTaskContext(tasks, projectMap)

	// status(wip) & kind(task) & project(tk)
	result, err := setlang.Eval(ctx, "status(wip) & kind(task) & project(tk)")
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	items := result.Items()
	if len(items) != 1 {
		t.Errorf("Expected 1 task, got %d", len(items))
	}
	if len(items) > 0 && items[0] != "t1" {
		t.Errorf("Expected t1, got %s", items[0])
	}
}

func TestTaskContext_All(t *testing.T) {
	tasks := []*types.Task{
		{TaskUUID: "t1"},
		{TaskUUID: "t2"},
		{TaskUUID: "t3"},
	}

	ctx := NewTaskContext(tasks, nil)

	result, err := setlang.Eval(ctx, "all()")
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	items := result.Items()
	if len(items) != 3 {
		t.Errorf("Expected 3 tasks from all(), got %d", len(items))
	}
}

func TestTaskContext_DifferenceOperator(t *testing.T) {
	tasks := []*types.Task{
		{TaskUUID: "t1", Axes: map[string]types.AxisStatus{"generic": {Effective: "wip"}}},
		{TaskUUID: "t2", Axes: map[string]types.AxisStatus{"generic": {Effective: "done"}}},
		{TaskUUID: "t3", Axes: map[string]types.AxisStatus{"generic": {Effective: "next"}}},
	}

	ctx := NewTaskContext(tasks, nil)

	// all() ~ status(done) = everything except done
	result, err := setlang.Eval(ctx, "all() ~ status(done)")
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	items := result.Items()
	if len(items) != 2 { // t1 and t3
		t.Errorf("Expected 2 tasks (not done), got %d", len(items))
	}
}

func TestTaskContext_UnionOperator(t *testing.T) {
	tasks := []*types.Task{
		{TaskUUID: "t1", Axes: map[string]types.AxisStatus{"generic": {Effective: "wip"}}},
		{TaskUUID: "t2", Axes: map[string]types.AxisStatus{"generic": {Effective: "done"}}},
		{TaskUUID: "t3", Axes: map[string]types.AxisStatus{"generic": {Effective: "next"}}},
	}

	ctx := NewTaskContext(tasks, nil)

	// status(wip) | status(next)
	result, err := setlang.Eval(ctx, "status(wip) | status(next)")
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	items := result.Items()
	if len(items) != 2 { // t1 and t3
		t.Errorf("Expected 2 tasks (wip or next), got %d", len(items))
	}
}

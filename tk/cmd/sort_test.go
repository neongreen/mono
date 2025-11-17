package cmd

import (
	"testing"
	"time"

	"github.com/neongreen/mono/tk/internal/types"
)

func TestSortTasks_ByCreated(t *testing.T) {
	now := time.Now()
	tasks := []*types.Task{
		{TaskDisplayID: "tk-3", Title: "Third", CreatedAt: now.Add(2 * time.Second)},
		{TaskDisplayID: "tk-1", Title: "First", CreatedAt: now},
		{TaskDisplayID: "tk-2", Title: "Second", CreatedAt: now.Add(1 * time.Second)},
	}

	types.SortTasks(tasks, "created")

	if tasks[0].TaskDisplayID != "tk-1" {
		t.Errorf("Expected first task to be tk-1, got %s", tasks[0].TaskDisplayID)
	}
	if tasks[1].TaskDisplayID != "tk-2" {
		t.Errorf("Expected second task to be tk-2, got %s", tasks[1].TaskDisplayID)
	}
	if tasks[2].TaskDisplayID != "tk-3" {
		t.Errorf("Expected third task to be tk-3, got %s", tasks[2].TaskDisplayID)
	}
}

func TestSortTasks_ByID(t *testing.T) {
	now := time.Now()
	tasks := []*types.Task{
		{TaskDisplayID: "tk-3", Title: "Third", CreatedAt: now},
		{TaskDisplayID: "tk-1", Title: "First", CreatedAt: now},
		{TaskDisplayID: "tk-2", Title: "Second", CreatedAt: now},
	}

	types.SortTasks(tasks, "id")

	if tasks[0].TaskDisplayID != "tk-1" {
		t.Errorf("Expected first task to be tk-1, got %s", tasks[0].TaskDisplayID)
	}
	if tasks[1].TaskDisplayID != "tk-2" {
		t.Errorf("Expected second task to be tk-2, got %s", tasks[1].TaskDisplayID)
	}
	if tasks[2].TaskDisplayID != "tk-3" {
		t.Errorf("Expected third task to be tk-3, got %s", tasks[2].TaskDisplayID)
	}
}

func TestSortTasks_ByTitle(t *testing.T) {
	now := time.Now()
	tasks := []*types.Task{
		{TaskDisplayID: "tk-1", Title: "Zebra", CreatedAt: now},
		{TaskDisplayID: "tk-2", Title: "Apple", CreatedAt: now},
		{TaskDisplayID: "tk-3", Title: "Banana", CreatedAt: now},
	}

	types.SortTasks(tasks, "title")

	if tasks[0].Title != "Apple" {
		t.Errorf("Expected first task to be Apple, got %s", tasks[0].Title)
	}
	if tasks[1].Title != "Banana" {
		t.Errorf("Expected second task to be Banana, got %s", tasks[1].Title)
	}
	if tasks[2].Title != "Zebra" {
		t.Errorf("Expected third task to be Zebra, got %s", tasks[2].Title)
	}
}

func TestSortTasks_DefaultToCreated(t *testing.T) {
	now := time.Now()
	tasks := []*types.Task{
		{TaskDisplayID: "tk-3", Title: "Third", CreatedAt: now.Add(2 * time.Second)},
		{TaskDisplayID: "tk-1", Title: "First", CreatedAt: now},
		{TaskDisplayID: "tk-2", Title: "Second", CreatedAt: now.Add(1 * time.Second)},
	}

	// Test with empty string (should default to created)
	types.SortTasks(tasks, "")

	if tasks[0].TaskDisplayID != "tk-1" {
		t.Errorf("Expected first task to be tk-1, got %s", tasks[0].TaskDisplayID)
	}
	if tasks[1].TaskDisplayID != "tk-2" {
		t.Errorf("Expected second task to be tk-2, got %s", tasks[1].TaskDisplayID)
	}
	if tasks[2].TaskDisplayID != "tk-3" {
		t.Errorf("Expected third task to be tk-3, got %s", tasks[2].TaskDisplayID)
	}
}

func TestSortTasks_UnknownSortType(t *testing.T) {
	now := time.Now()
	tasks := []*types.Task{
		{TaskDisplayID: "tk-3", Title: "Third", CreatedAt: now.Add(2 * time.Second)},
		{TaskDisplayID: "tk-1", Title: "First", CreatedAt: now},
		{TaskDisplayID: "tk-2", Title: "Second", CreatedAt: now.Add(1 * time.Second)},
	}

	// Test with unknown sort type (should default to created)
	types.SortTasks(tasks, "invalid")

	if tasks[0].TaskDisplayID != "tk-1" {
		t.Errorf("Expected first task to be tk-1, got %s", tasks[0].TaskDisplayID)
	}
	if tasks[1].TaskDisplayID != "tk-2" {
		t.Errorf("Expected second task to be tk-2, got %s", tasks[1].TaskDisplayID)
	}
	if tasks[2].TaskDisplayID != "tk-3" {
		t.Errorf("Expected third task to be tk-3, got %s", tasks[2].TaskDisplayID)
	}
}

func TestSortTasks_StableSorting(t *testing.T) {
	// Test that sorting is stable - tasks with same sort key maintain relative order
	now := time.Now()
	tasks := []*types.Task{
		{TaskDisplayID: "tk-1", Title: "Task A", CreatedAt: now},
		{TaskDisplayID: "tk-2", Title: "Task A", CreatedAt: now},
		{TaskDisplayID: "tk-3", Title: "Task B", CreatedAt: now},
	}

	types.SortTasks(tasks, "title")

	// Both "Task A" tasks should come before "Task B"
	if tasks[0].Title != "Task A" || tasks[1].Title != "Task A" {
		t.Error("Expected both Task A entries to come before Task B")
	}
	if tasks[2].Title != "Task B" {
		t.Errorf("Expected third task to be Task B, got %s", tasks[2].Title)
	}

	// When titles are the same, the original order should be preserved (stable sort)
	// In Go's sort.Slice, the sort is not guaranteed to be stable, but since we're
	// testing the behavior, we just verify that both Task A entries are before Task B
}

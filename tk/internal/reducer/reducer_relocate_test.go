package reducer

import (
	"testing"
)

func TestReducer_TaskRelocate_UpdatesProjectTracking(t *testing.T) {
	test := NewReducerTest(t)

	projectA := test.CreateProject("project-a")
	projectB := test.CreateProject("project-b")

	task := test.CreateTask("alice", "Test task", projectA)
	task.AssertExists().AssertTitle("Test task")

	task.Relocate(projectB)
	task.AssertExists()

	test.DeleteProject(projectB)
	task.AssertDeleted()
}

func TestReducer_TaskRelocate_DeletesFromCorrectProject(t *testing.T) {
	test := NewReducerTest(t)

	projectA := test.CreateProject("project-a")
	projectB := test.CreateProject("project-b")

	task := test.CreateTask("alice", "Test task", projectA)
	task.Relocate(projectB)

	test.DeleteProject(projectA)
	task.AssertExists().AssertTitle("Test task")
}

func TestReducer_TaskRelocate_MultipleRelocations(t *testing.T) {
	test := NewReducerTest(t)

	projectA := test.CreateProject("project-a")
	projectB := test.CreateProject("project-b")
	projectC := test.CreateProject("project-c")

	task := test.CreateTask("alice", "Test task", projectA)
	task.Relocate(projectB)
	task.Relocate(projectC)

	test.DeleteProject(projectA)
	task.AssertExists()

	test.DeleteProject(projectB)
	task.AssertExists()

	test.DeleteProject(projectC)
	task.AssertDeleted()
}

// TestReducer_TaskRelocate_UpdatesTaskProjectUUID is a regression test for the bug where
// task.relocate only updated taskProjects but not the task's ProjectUUID field.
// This caused relocated tasks to still appear under the old project in ls grouping and filtering.
func TestReducer_TaskRelocate_UpdatesTaskProjectUUID(t *testing.T) {
	test := NewReducerTest(t)

	projectA := test.CreateProject("project-a")
	projectB := test.CreateProject("project-b")

	task := test.CreateTask("alice", "Test task", projectA)

	// Verify task starts in project A
	task.AssertInProject(projectA)

	// Relocate to project B
	task.Relocate(projectB)

	// Verify task's ProjectUUID is now project B (regression test - this was the bug)
	task.AssertInProject(projectB)

	// Verify GetProjectNameForTask returns project B's name
	projectName, err := test.reducer.GetProjectNameForTask(task.UID)
	if err != nil {
		t.Fatalf("GetProjectNameForTask failed: %v", err)
	}
	if projectName != "project-b" {
		t.Errorf("GetProjectNameForTask = %s, want project-b", projectName)
	}
}

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

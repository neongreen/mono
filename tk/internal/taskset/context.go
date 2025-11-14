package taskset

import (
	"fmt"

	"github.com/neongreen/mono/lib/setlang"
	"github.com/neongreen/mono/tk/internal/types"
)

// TaskContext implements setlang.Context for task queries
type TaskContext struct {
	tasks      []*types.Task
	tasksByUID map[string]*types.Task
}

// NewTaskContext creates a new task context for query evaluation
func NewTaskContext(tasks []*types.Task) *TaskContext {
	tasksByUID := make(map[string]*types.Task)
	for _, task := range tasks {
		tasksByUID[task.TaskUUID] = task
	}

	return &TaskContext{
		tasks:      tasks,
		tasksByUID: tasksByUID,
	}
}

// LookupIdent resolves named identifiers to task sets
func (tc *TaskContext) LookupIdent(name string) (*setlang.Set[string], error) {
	switch name {
	case "all":
		// Return all task UIDs
		result := setlang.NewSet[string]()
		for _, task := range tc.tasks {
			result.Add(task.TaskUUID)
		}
		return result, nil

	case "none", "empty":
		// Return empty set
		return setlang.NewSet[string](), nil

	default:
		return nil, fmt.Errorf("unknown identifier: %s", name)
	}
}

// CallFunc implements task-specific query functions
func (tc *TaskContext) CallFunc(name string, args []setlang.FuncArg[string]) (*setlang.Set[string], error) {
	switch name {
	case "status":
		return tc.funcStatus(args)
	case "kind":
		return tc.funcKind(args)
	case "project":
		return tc.funcProject(args)
	case "blocked":
		return tc.funcBlocked(args)
	case "unblocked":
		return tc.funcUnblocked(args)
	case "author":
		return tc.funcAuthor(args)
	case "title":
		return tc.funcTitle(args)
	default:
		return nil, fmt.Errorf("unknown function: %s", name)
	}
}

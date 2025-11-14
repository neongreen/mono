package taskset

import (
	"fmt"
	"strings"

	"github.com/neongreen/mono/lib/setlang"
)

// funcAll returns all tasks
// Usage: all()
func (tc *TaskContext) funcAll(args []setlang.FuncArg[string]) (*setlang.Set[string], error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("all() takes no arguments")
	}

	result := setlang.NewSet[string]()
	for _, task := range tc.tasks {
		result.Add(task.TaskUUID)
	}
	return result, nil
}

// funcStatus filters tasks by status
// Usage: status(wip), status(done)
func (tc *TaskContext) funcStatus(args []setlang.FuncArg[string]) (*setlang.Set[string], error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("status() takes 1 argument")
	}

	// Get status name (identifier or string)
	statusName, err := args[0].GetIdent()
	if err != nil {
		statusName, err = args[0].GetString()
		if err != nil {
			return nil, fmt.Errorf("status() argument must be an identifier or string")
		}
	}

	result := setlang.NewSet[string]()
	for _, task := range tc.tasks {
		// Check generic axis
		if axis, ok := task.Axes["generic"]; ok && axis.Effective == statusName {
			result.Add(task.TaskUUID)
		}
	}

	return result, nil
}

// funcKind filters tasks by item kind
// Usage: kind(decision), kind(resource)
func (tc *TaskContext) funcKind(args []setlang.FuncArg[string]) (*setlang.Set[string], error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("kind() takes 1 argument")
	}

	kindName, err := args[0].GetIdent()
	if err != nil {
		kindName, err = args[0].GetString()
		if err != nil {
			return nil, fmt.Errorf("kind() argument must be an identifier or string")
		}
	}

	result := setlang.NewSet[string]()
	for _, task := range tc.tasks {
		// Default to "task" if ItemKind is empty
		itemKind := task.ItemKind
		if itemKind == "" {
			itemKind = "task"
		}
		if itemKind == kindName {
			result.Add(task.TaskUUID)
		}
	}

	return result, nil
}

// funcProject filters tasks by project
// Usage: project(tk), project(mono)
// Note: Matches by project name (needs projectUIDToName map)
func (tc *TaskContext) funcProject(args []setlang.FuncArg[string]) (*setlang.Set[string], error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("project() takes 1 argument")
	}

	projectName, err := args[0].GetIdent()
	if err != nil {
		projectName, err = args[0].GetString()
		if err != nil {
			return nil, fmt.Errorf("project() argument must be an identifier or string")
		}
	}

	result := setlang.NewSet[string]()

	for _, task := range tc.tasks {
		// Look up project name from UUID
		if taskProjectName, ok := tc.projectUIDToName[task.ProjectUUID]; ok {
			if taskProjectName == projectName {
				result.Add(task.TaskUUID)
			}
		}
	}

	return result, nil
}

// funcBlocked returns all blocked tasks
// Usage: blocked()
func (tc *TaskContext) funcBlocked(args []setlang.FuncArg[string]) (*setlang.Set[string], error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("blocked() takes no arguments")
	}

	result := setlang.NewSet[string]()
	for _, task := range tc.tasks {
		if task.Blocked {
			result.Add(task.TaskUUID)
		}
	}

	return result, nil
}

// funcUnblocked returns all unblocked tasks
// Usage: unblocked()
func (tc *TaskContext) funcUnblocked(args []setlang.FuncArg[string]) (*setlang.Set[string], error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("unblocked() takes no arguments")
	}

	result := setlang.NewSet[string]()
	for _, task := range tc.tasks {
		if !task.Blocked {
			result.Add(task.TaskUUID)
		}
	}

	return result, nil
}

// funcAuthor filters tasks by creator
// Usage: author(emily), author(claude)
func (tc *TaskContext) funcAuthor(args []setlang.FuncArg[string]) (*setlang.Set[string], error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("author() takes 1 argument")
	}

	authorName, err := args[0].GetIdent()
	if err != nil {
		authorName, err = args[0].GetString()
		if err != nil {
			return nil, fmt.Errorf("author() argument must be an identifier or string")
		}
	}

	result := setlang.NewSet[string]()
	for _, task := range tc.tasks {
		if task.CreatedBy == authorName {
			result.Add(task.TaskUUID)
		}
	}

	return result, nil
}

// funcTitle filters tasks by title substring match
// Usage: title("bug"), title("implement")
func (tc *TaskContext) funcTitle(args []setlang.FuncArg[string]) (*setlang.Set[string], error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("title() takes 1 argument")
	}

	pattern, err := args[0].GetString()
	if err != nil {
		// Try identifier as fallback
		pattern, err = args[0].GetIdent()
		if err != nil {
			return nil, fmt.Errorf("title() argument must be a string or identifier")
		}
	}

	result := setlang.NewSet[string]()
	for _, task := range tc.tasks {
		// Case-insensitive substring match
		if strings.Contains(strings.ToLower(task.Title), strings.ToLower(pattern)) {
			result.Add(task.TaskUUID)
		}
	}

	return result, nil
}

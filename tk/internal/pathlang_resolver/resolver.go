package pathlang_resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/neongreen/mono/lib/pathlang"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/reducer"
	"github.com/neongreen/mono/tk/internal/types"
)

// NodeType represents the type of node in the tk path hierarchy
type NodeType string

const (
	NodeTypeRoot     NodeType = "root"
	NodeTypeProject  NodeType = "project"
	NodeTypeTask     NodeType = "task"
	NodeTypeSubtasks NodeType = "subtasks"
	NodeTypeBlockers NodeType = "blockers"
	NodeTypeNotes    NodeType = "notes"
)

// Node represents a node in the tk path hierarchy
type Node struct {
	Type       NodeType
	ProjectUID string      // For project and task nodes
	TaskUID    string      // For task nodes
	Task       *types.Task // For task nodes (cached)
}

// TkResolver implements pathlang.Resolver for tk's domain model
type TkResolver struct {
	db      *database.DB
	reducer *reducer.Reducer
}

// NewTkResolver creates a new tk resolver
func NewTkResolver(db *database.DB, reducer *reducer.Reducer) *TkResolver {
	return &TkResolver{
		db:      db,
		reducer: reducer,
	}
}

// Root returns the logical root node for path resolution
func (r *TkResolver) Root(ctx context.Context) (pathlang.Node, error) {
	return &Node{Type: NodeTypeRoot}, nil
}

// Children resolves a segment under a parent node
func (r *TkResolver) Children(ctx context.Context, parent pathlang.Node, seg pathlang.Segment) ([]pathlang.Node, error) {
	node, ok := parent.(*Node)
	if !ok {
		return nil, fmt.Errorf("invalid parent node type")
	}

	switch node.Type {
	case NodeTypeRoot:
		return r.resolveFromRoot(seg)
	case NodeTypeTask:
		return r.resolveFromTask(node, seg)
	default:
		// Projects, subtasks, blockers, notes don't have children
		return nil, nil
	}
}

// resolveFromRoot handles paths from root
// Supports:
// - /project-alias -> project
// - /project-alias-number -> task
func (r *TkResolver) resolveFromRoot(seg pathlang.Segment) ([]pathlang.Node, error) {
	name := seg.Name
	
	// Check if this looks like a display ID (contains a hyphen followed by digits)
	// e.g., "foo-13" or "my-proj-42"
	if r.looksLikeDisplayID(name) {
		// Try to resolve as a task reference
		taskUID, err := database.ResolveTaskReference(r.db, types.NewTaskRef(name))
		if err == nil {
			// Found a task
			task, ok := r.reducer.GetTask(taskUID)
			if !ok {
				return nil, fmt.Errorf("task %s not found", name)
			}
			
			return []pathlang.Node{
				&Node{
					Type:       NodeTypeTask,
					TaskUID:    taskUID,
					ProjectUID: task.ProjectUUID,
					Task:       task,
				},
			}, nil
		}
		// If it failed to resolve as task, try as project alias
	}
	
	// Try to resolve as project alias
	projectUID, err := database.ResolveProjectByAlias(r.db, name)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve %s: not found as project or task", name)
	}
	
	return []pathlang.Node{
		&Node{
			Type:       NodeTypeProject,
			ProjectUID: projectUID,
		},
	}, nil
}

// looksLikeDisplayID checks if a string looks like it could be a task display ID
// A display ID has the format: alias-number or alias-number-nodehint
func (r *TkResolver) looksLikeDisplayID(s string) bool {
	// Must contain at least one hyphen
	if !strings.Contains(s, "-") {
		return false
	}
	
	// Split and check if last part (or second-to-last if there's a node hint) is numeric
	parts := strings.Split(s, "-")
	if len(parts) < 2 {
		return false
	}
	
	// Check if last part is numeric
	lastPart := parts[len(parts)-1]
	for _, c := range lastPart {
		if c < '0' || c > '9' {
			// Not numeric - might be project alias with hyphens
			return false
		}
	}
	
	return true
}

// resolveFromTask handles paths from a task node
// Supports:
// - /task/subtasks -> children tasks
// - /task/blockers -> blocking tasks
// - /task/notes -> task notes (returns a special node)
func (r *TkResolver) resolveFromTask(taskNode *Node, seg pathlang.Segment) ([]pathlang.Node, error) {
	if taskNode.Task == nil {
		return nil, fmt.Errorf("task node has no task data")
	}
	
	switch seg.Name {
	case "subtasks":
		return r.getSubtasks(taskNode)
	case "blockers":
		return r.getBlockers(taskNode)
	case "notes":
		// Return a special node representing notes
		return []pathlang.Node{
			&Node{
				Type:       NodeTypeNotes,
				TaskUID:    taskNode.TaskUID,
				ProjectUID: taskNode.ProjectUID,
				Task:       taskNode.Task,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown task child: %s", seg.Name)
	}
}

// getSubtasks returns child tasks
func (r *TkResolver) getSubtasks(taskNode *Node) ([]pathlang.Node, error) {
	if taskNode.Task.Relations == nil || len(taskNode.Task.Relations.Subtask.Children) == 0 {
		return nil, nil
	}
	
	var nodes []pathlang.Node
	for _, childUID := range taskNode.Task.Relations.Subtask.Children {
		childTask, ok := r.reducer.GetTask(childUID)
		if !ok {
			continue
		}
		
		nodes = append(nodes, &Node{
			Type:       NodeTypeTask,
			TaskUID:    childUID,
			ProjectUID: childTask.ProjectUUID,
			Task:       childTask,
		})
	}
	
	return nodes, nil
}

// getBlockers returns tasks that block this task
func (r *TkResolver) getBlockers(taskNode *Node) ([]pathlang.Node, error) {
	if len(taskNode.Task.Blockers) == 0 {
		return nil, nil
	}
	
	var nodes []pathlang.Node
	for _, blocker := range taskNode.Task.Blockers {
		blockerTask, ok := r.reducer.GetTask(blocker.TaskUUID)
		if !ok {
			continue
		}
		
		nodes = append(nodes, &Node{
			Type:       NodeTypeTask,
			TaskUID:    blocker.TaskUUID,
			ProjectUID: blockerTask.ProjectUUID,
			Task:       blockerTask,
		})
	}
	
	return nodes, nil
}

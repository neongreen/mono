package types

import "sort"

// Relations represents all relations for a task
type Relations struct {
	Blocks     RelationSet `json:"blocks,omitzero"`     // Tasks this task blocks
	Subtask    RelationSet `json:"subtask,omitzero"`    // Parent/children for subtasks
	Related    RelationSet `json:"related,omitzero"`    // Related tasks
	Duplicate  RelationSet `json:"duplicate,omitzero"`  // Duplicate tasks
	Supersedes RelationSet `json:"supersedes,omitzero"` // Tasks this supersedes
}

// Sorted returns a deep copy of Relations with all slices sorted for deterministic JSON output
func (r *Relations) Sorted() *Relations {
	if r == nil {
		return nil
	}
	
	return &Relations{
		Blocks:     r.Blocks.Sorted(),
		Subtask:    r.Subtask.Sorted(),
		Related:    r.Related.Sorted(),
		Duplicate:  r.Duplicate.Sorted(),
		Supersedes: r.Supersedes.Sorted(),
	}
}

// RelationSet represents directional relations
type RelationSet struct {
	Out      []RelationTarget `json:"out,omitempty"`      // Outgoing edges (this task -> others)
	In       []RelationTarget `json:"in,omitempty"`       // Incoming edges (others -> this task)
	Children []string         `json:"children,omitempty"` // For subtask relations
	Parent   string           `json:"parent,omitempty"`   // For subtask relations
}

// Sorted returns a copy of RelationSet with all slices sorted
func (rs RelationSet) Sorted() RelationSet {
	result := RelationSet{
		Parent: rs.Parent,
	}
	
	// Sort Out by TaskUUID
	if len(rs.Out) > 0 {
		result.Out = make([]RelationTarget, len(rs.Out))
		copy(result.Out, rs.Out)
		sort.Slice(result.Out, func(i, j int) bool {
			return result.Out[i].TaskUUID < result.Out[j].TaskUUID
		})
	}
	
	// Sort In by TaskUUID
	if len(rs.In) > 0 {
		result.In = make([]RelationTarget, len(rs.In))
		copy(result.In, rs.In)
		sort.Slice(result.In, func(i, j int) bool {
			return result.In[i].TaskUUID < result.In[j].TaskUUID
		})
	}
	
	// Sort Children
	if len(rs.Children) > 0 {
		result.Children = make([]string, len(rs.Children))
		copy(result.Children, rs.Children)
		sort.Strings(result.Children)
	}
	
	return result
}

// RelationTarget represents a relation target
type RelationTarget struct {
	TaskUUID string `json:"dst"` // Destination task UUID
	Note     string `json:"note,omitempty"`
}

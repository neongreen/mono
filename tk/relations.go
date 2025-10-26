package main

import (
	"fmt"
)

// RelationEdge represents an edge in the relation graph with OR-set semantics
type RelationEdge struct {
	Src       string          // Source task UUID
	Type      string          // Relation type
	Dst       string          // Destination task UUID
	Note      string          // Optional note
	Adds      map[string]bool // Set of add tags (event_id -> exists)
	Removals  map[string]bool // Set of removed tags (event_id -> exists)
	EventID   string          // Last event ID that modified this edge
	CreatedAt int64           // Lamport timestamp when created
}

// RelationsGraph stores all relation edges with OR-set CRDT semantics
type RelationsGraph struct {
	edges map[string]*RelationEdge // Key: src:type:dst
}

// NewRelationsGraph creates a new relations graph
func NewRelationsGraph() *RelationsGraph {
	return &RelationsGraph{
		edges: make(map[string]*RelationEdge),
	}
}

// AddRelation adds a relation edge with OR-set semantics
func (g *RelationsGraph) AddRelation(src, relType, dst, note, eventID, node string, ts int64) {
	key := makeEdgeKey(src, relType, dst)

	edge, exists := g.edges[key]
	if !exists {
		edge = &RelationEdge{
			Src:       src,
			Type:      relType,
			Dst:       dst,
			Note:      note,
			Adds:      make(map[string]bool),
			Removals:  make(map[string]bool),
			EventID:   eventID,
			CreatedAt: ts,
		}
		g.edges[key] = edge
	}

	// Add this event's tag to the adds set
	edge.Adds[eventID] = true
	edge.EventID = eventID

	// Update note if provided
	if note != "" {
		edge.Note = note
	}
}

// RemoveRelation removes a relation edge with OR-set semantics (remove-wins)
func (g *RelationsGraph) RemoveRelation(src, relType, dst, eventID, node string, ts int64) {
	key := makeEdgeKey(src, relType, dst)

	edge, exists := g.edges[key]
	if !exists {
		// Create edge structure to track the removal
		edge = &RelationEdge{
			Src:       src,
			Type:      relType,
			Dst:       dst,
			Adds:      make(map[string]bool),
			Removals:  make(map[string]bool),
			EventID:   eventID,
			CreatedAt: ts,
		}
		g.edges[key] = edge
	}

	// Tombstone all currently observed add tags
	for addTag := range edge.Adds {
		edge.Removals[addTag] = true
	}

	edge.EventID = eventID
}

// isEdgePresent checks if an edge is present according to observed-remove semantics
// An edge exists if there's at least one add tag that hasn't been removed
func (g *RelationsGraph) isEdgePresent(edge *RelationEdge) bool {
	for addTag := range edge.Adds {
		if !edge.Removals[addTag] {
			return true
		}
	}
	return false
}

// SetRelationNote sets a note on a relation
func (g *RelationsGraph) SetRelationNote(src, relType, dst, note string) {
	key := makeEdgeKey(src, relType, dst)
	if edge, exists := g.edges[key]; exists && g.isEdgePresent(edge) {
		edge.Note = note
	}
}

// GetOutgoingRelations returns all outgoing relations for a task by type
func (g *RelationsGraph) GetOutgoingRelations(taskUUID, relType string) []RelationTarget {
	var targets []RelationTarget
	for _, edge := range g.edges {
		if edge.Src == taskUUID && edge.Type == relType && g.isEdgePresent(edge) {
			targets = append(targets, RelationTarget{
				TaskUUID: edge.Dst,
				Note:     edge.Note,
			})
		}
	}
	return targets
}

// GetIncomingRelations returns all incoming relations for a task by type
func (g *RelationsGraph) GetIncomingRelations(taskUUID, relType string) []RelationTarget {
	var targets []RelationTarget
	for _, edge := range g.edges {
		if edge.Dst == taskUUID && edge.Type == relType && g.isEdgePresent(edge) {
			targets = append(targets, RelationTarget{
				TaskUUID: edge.Src,
				Note:     edge.Note,
			})
		}
	}
	return targets
}

// DetectCycles detects cycles in the relation graph for a given relation type
func (g *RelationsGraph) DetectCycles(relType string) [][]string {
	var cycles [][]string

	// Build adjacency list for this relation type
	adj := make(map[string][]string)
	for _, edge := range g.edges {
		if edge.Type == relType && g.isEdgePresent(edge) {
			adj[edge.Src] = append(adj[edge.Src], edge.Dst)
		}
	}

	// Track visited nodes and recursion stack
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var detectCycle func(node string) bool
	detectCycle = func(node string) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, neighbor := range adj[node] {
			if !visited[neighbor] {
				if detectCycle(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				// Found a cycle - extract it from path
				cycleStart := -1
				for i, n := range path {
					if n == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := make([]string, len(path)-cycleStart)
					copy(cycle, path[cycleStart:])
					cycles = append(cycles, cycle)
				}
				return true
			}
		}

		recStack[node] = false
		path = path[:len(path)-1]
		return false
	}

	// Check all nodes for cycles
	for node := range adj {
		if !visited[node] {
			detectCycle(node)
		}
	}

	return cycles
}

// ComputeBlocked computes which tasks are blocked based on the blocks relation
// and the blocking axis configuration
func (g *RelationsGraph) ComputeBlocked(tasks map[string]*Task, blockingAxis string, doneStates []string) {
	// Build set of done states for quick lookup
	doneSet := make(map[string]bool)
	for _, state := range doneStates {
		doneSet[state] = true
	}

	// For each task, check if any of its blockers are not done
	for uuid, task := range tasks {
		blockers := g.GetIncomingRelations(uuid, "blocks")
		task.Blocked = false
		task.Blockers = nil

		if len(blockers) == 0 {
			continue
		}

		// Check each blocker
		var activeBlockers []Blocker
		for _, blocker := range blockers {
			blockerTask, ok := tasks[blocker.TaskUUID]
			if !ok {
				// Blocker task not found - treat as blocking
				activeBlockers = append(activeBlockers, Blocker{
					TaskID:   blocker.TaskUUID,
					Title:    "(unknown task)",
					Distance: 1,
				})
				continue
			}

			// Check if blocker is done according to the blocking axis
			isDone := false
			if axis, ok := blockerTask.Axes[blockingAxis]; ok {
				if doneSet[axis.Effective] {
					isDone = true
				}
			}

			if !isDone {
				activeBlockers = append(activeBlockers, Blocker{
					TaskID:   blockerTask.TaskID,
					Title:    blockerTask.Title,
					Distance: 1,
				})
			}
		}

		if len(activeBlockers) > 0 {
			task.Blocked = true
			task.Blockers = activeBlockers
		}
	}
}

// GetTransitiveBlockers returns all transitive blockers for a task
func (g *RelationsGraph) GetTransitiveBlockers(taskUUID string, tasks map[string]*Task, blockingAxis string, doneStates []string, maxDepth int) []Blocker {
	// Build set of done states for quick lookup
	doneSet := make(map[string]bool)
	for _, state := range doneStates {
		doneSet[state] = true
	}

	visited := make(map[string]bool)
	var result []Blocker

	var dfs func(uuid string, distance int)
	dfs = func(uuid string, distance int) {
		if distance > maxDepth || visited[uuid] {
			return
		}
		visited[uuid] = true

		blockers := g.GetIncomingRelations(uuid, "blocks")
		for _, blocker := range blockers {
			blockerTask, ok := tasks[blocker.TaskUUID]
			if !ok {
				result = append(result, Blocker{
					TaskID:   blocker.TaskUUID,
					Title:    "(unknown task)",
					Distance: distance,
				})
				continue
			}

			// Check if blocker is done
			isDone := false
			if axis, ok := blockerTask.Axes[blockingAxis]; ok {
				if doneSet[axis.Effective] {
					isDone = true
				}
			}

			if !isDone {
				result = append(result, Blocker{
					TaskID:   blockerTask.TaskID,
					Title:    blockerTask.Title,
					Distance: distance,
				})
				// Recurse to find transitive blockers
				dfs(blocker.TaskUUID, distance+1)
			}
		}
	}

	dfs(taskUUID, 1)
	return result
}

// BuildTaskRelations builds the Relations structure for a task
func (g *RelationsGraph) BuildTaskRelations(taskUUID string) *Relations {
	rel := &Relations{}

	// Blocks relations
	blocksOut := g.GetOutgoingRelations(taskUUID, "blocks")
	blocksIn := g.GetIncomingRelations(taskUUID, "blocks")
	if len(blocksOut) > 0 || len(blocksIn) > 0 {
		rel.Blocks = RelationSet{
			Out: blocksOut,
			In:  blocksIn,
		}
	}

	// Subtask relations
	subtaskOut := g.GetOutgoingRelations(taskUUID, "subtask")
	subtaskIn := g.GetIncomingRelations(taskUUID, "subtask")
	if len(subtaskOut) > 0 || len(subtaskIn) > 0 {
		// For subtasks: out = children, in = parent
		rel.Subtask = RelationSet{}
		if len(subtaskOut) > 0 {
			children := make([]string, len(subtaskOut))
			for i, t := range subtaskOut {
				children[i] = t.TaskUUID
			}
			rel.Subtask.Children = children
		}
		if len(subtaskIn) > 0 {
			rel.Subtask.Parent = subtaskIn[0].TaskUUID
		}
	}

	// Related relations
	relatedOut := g.GetOutgoingRelations(taskUUID, "related")
	relatedIn := g.GetIncomingRelations(taskUUID, "related")
	if len(relatedOut) > 0 || len(relatedIn) > 0 {
		rel.Related = RelationSet{
			Out: relatedOut,
			In:  relatedIn,
		}
	}

	// Duplicate relations
	duplicateOut := g.GetOutgoingRelations(taskUUID, "duplicate_of")
	duplicateIn := g.GetIncomingRelations(taskUUID, "duplicate_of")
	if len(duplicateOut) > 0 || len(duplicateIn) > 0 {
		rel.Duplicate = RelationSet{
			Out: duplicateOut,
			In:  duplicateIn,
		}
	}

	// Supersedes relations
	supersedesOut := g.GetOutgoingRelations(taskUUID, "supersedes")
	supersedesIn := g.GetIncomingRelations(taskUUID, "supersedes")
	if len(supersedesOut) > 0 || len(supersedesIn) > 0 {
		rel.Supersedes = RelationSet{
			Out: supersedesOut,
			In:  supersedesIn,
		}
	}

	// Only return Relations if there are any relations
	if rel.Blocks.Out == nil && rel.Blocks.In == nil &&
		rel.Subtask.Children == nil && rel.Subtask.Parent == "" &&
		rel.Related.Out == nil && rel.Related.In == nil &&
		rel.Duplicate.Out == nil && rel.Duplicate.In == nil &&
		rel.Supersedes.Out == nil && rel.Supersedes.In == nil {
		return nil
	}

	return rel
}

// makeEdgeKey creates a unique key for an edge
func makeEdgeKey(src, relType, dst string) string {
	return fmt.Sprintf("%s:%s:%s", src, relType, dst)
}

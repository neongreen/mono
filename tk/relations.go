package main

import (
	"github.com/neongreen/mono/tk/internal/relations"
	"github.com/neongreen/mono/tk/internal/types"
)

// RelationEdge represents an edge in the relation graph with OR-set semantics

// Source task UUID
// Relation type
// Destination task UUID
// Optional note
// Set of add tags (event_id -> exists)
// Set of removed tags (event_id -> exists)
// Last event ID that modified this edge
// Lamport timestamp when created

// RelationsGraph stores all relation edges with OR-set CRDT semantics

// Key: src:type:dst

// ComputeBlocked computes which tasks are blocked based on the blocks relation
// and the blocking axis configuration
func ComputeBlocked(g *relations.RelationsGraph, tasks map[string]*types.Task, blockingAxis string, doneStates []string) {
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
		var activeBlockers []types.Blocker
		for _, blocker := range blockers {
			blockerTask, ok := tasks[blocker.TaskUUID]
			if !ok {
				// Blocker task not found - treat as blocking
				activeBlockers = append(activeBlockers, types.Blocker{
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
				activeBlockers = append(activeBlockers, types.Blocker{
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
func GetTransitiveBlockers(g *relations.RelationsGraph, taskUUID string, tasks map[string]*types.Task, blockingAxis string, doneStates []string, maxDepth int) []types.Blocker {
	// Build set of done states for quick lookup
	doneSet := make(map[string]bool)
	for _, state := range doneStates {
		doneSet[state] = true
	}

	visited := make(map[string]bool)
	var result []types.Blocker

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
				result = append(result, types.Blocker{
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
				result = append(result, types.Blocker{
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

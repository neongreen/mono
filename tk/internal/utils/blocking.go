package utils

import (
	"github.com/neongreen/mono/tk/internal/relations"
	"github.com/neongreen/mono/tk/internal/types"
)

// ComputeBlocked computes which tasks are blocked based on the blocks relation
// and the blocking axis configuration
func ComputeBlocked(g *relations.RelationsGraph, tasks map[string]*types.Task, blockingAxis string, doneStates []string) {
	doneSet := make(map[string]bool)
	for _, state := range doneStates {
		doneSet[state] = true
	}

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
			blockerTask, ok := tasks[blocker.TaskID]
			if !ok {

				activeBlockers = append(activeBlockers, types.Blocker{
					TaskID:        blocker.TaskID,
					TaskDisplayID: blocker.TaskID,
					Title:         "(unknown task)",
					Distance:      1,
				})
				continue
			}

			isDone := false
			if axis, ok := blockerTask.Axes[blockingAxis]; ok {
				if doneSet[axis.Effective] {
					isDone = true
				}
			}

			if !isDone {
				activeBlockers = append(activeBlockers, types.Blocker{
					TaskID:        blockerTask.TaskID,
					TaskDisplayID: blockerTask.TaskDisplayID,
					Title:         blockerTask.Title,
					Distance:      1,
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
			blockerTask, ok := tasks[blocker.TaskID]
			if !ok {
				result = append(result, types.Blocker{
					TaskID:        blocker.TaskID,
					TaskDisplayID: blocker.TaskID,
					Title:         "(unknown task)",
					Distance:      distance,
				})
				continue
			}

			isDone := false
			if axis, ok := blockerTask.Axes[blockingAxis]; ok {
				if doneSet[axis.Effective] {
					isDone = true
				}
			}

			if !isDone {
				result = append(result, types.Blocker{
					TaskID:        blockerTask.TaskID,
					TaskDisplayID: blockerTask.TaskDisplayID,
					Title:         blockerTask.Title,
					Distance:      distance,
				})

				dfs(blocker.TaskID, distance+1)
			}
		}
	}

	dfs(taskUUID, 1)
	return result
}

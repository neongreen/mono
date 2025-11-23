package reducer

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/types"
)

func (r *Reducer) applyTaskDelete(e types.Event) error {
	var payload types.TaskDeletePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task.delete payload: %w", err)
	}

	r.removeTaskFromMaps(payload.TaskUUID)
	return nil
}

// removeTaskFromMaps removes a task from all internal maps and relations
func (r *Reducer) removeTaskFromMaps(taskUUID string) {
	// Remove task from tasks map
	delete(r.tasks, taskUUID)

	// Remove all task ID mappings that point to this UUID
	for taskID, uuid := range r.taskByID {
		if uuid == taskUUID {
			delete(r.taskByID, taskID)
		}
	}

	// Remove project association
	delete(r.taskProjects, taskUUID)

	// Remove all relations involving this task
	r.relations.RemoveTaskRelations(taskUUID)
}

func (r *Reducer) applyProjectDelete(e types.Event) error {
	var payload types.ProjectDeletePayload
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal project.delete payload: %w", err)
	}

	projectUID := payload.ProjectUID

	// Find all tasks in this project and delete them
	tasksToDelete := make([]string, 0)
	for taskUUID, taskProjectUID := range r.taskProjects {
		if taskProjectUID == projectUID.String() {
			tasksToDelete = append(tasksToDelete, taskUUID)
		}
	}

	// Delete each task using the helper
	for _, taskUUID := range tasksToDelete {
		r.removeTaskFromMaps(taskUUID)
	}

	return nil
}

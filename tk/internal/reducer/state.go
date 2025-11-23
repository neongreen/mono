package reducer

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/relations"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
)

// Reducer reconstructs task state from events, mirroring projection results while keeping
// in-memory conveniences for the CLI. It tracks tasks, projects, and relations so the CLI
// can answer queries without hitting the database during command execution.
// Reducer reconstructs task state from events
//
//lion:reducer section="Reducer overview"
type Reducer struct {
	tasks        map[string]*types.Task    // Key: task UUID
	taskByID     map[string]string         // Key: task ID (current or alias) -> Value: task UUID
	taskProjects map[string]string         // Key: task UUID -> Value: project UID
	projects     map[string]*types.Project // Key: project UID (includes synthetic placeholders for compatibility)
	deletedProj  map[string]bool           // Set of project UIDs that were deleted (tombstones)
	relations    *relations.RelationsGraph // Relations graph
}

// NewReducer creates a new reducer
func NewReducer() *Reducer {
	return &Reducer{
		tasks:        make(map[string]*types.Task),
		taskByID:     make(map[string]string),
		taskProjects: make(map[string]string),
		projects:     make(map[string]*types.Project),
		deletedProj:  make(map[string]bool),
		relations:    relations.NewRelationsGraph(),
	}
}

// Relations returns the relations graph
func (r *Reducer) Relations() *relations.RelationsGraph {
	return r.relations
}

// Tasks returns the tasks map
func (r *Reducer) Tasks() map[string]*types.Task {
	return r.tasks
}

// GetTask returns the current state of a task by ID (supports task ID or UUID)
func (r *Reducer) GetTask(idOrUUID string) (*types.Task, bool) {
	// Try direct UUID lookup first
	task, ok := r.tasks[idOrUUID]
	if ok {
		return task, true
	}

	// Try looking up by task ID
	taskUUID, ok := r.taskByID[idOrUUID]
	if !ok {
		return nil, false
	}

	task, ok = r.tasks[taskUUID]
	return task, ok
}

// GetAllTasks returns all tasks
func (r *Reducer) GetAllTasks() []*types.Task {
	tasks := make([]*types.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// GetProject returns a project by UID
func (r *Reducer) GetProject(projectUID string) (*types.Project, bool) {
	project, ok := r.projects[projectUID]
	return project, ok
}

// GetAllProjects returns all projects
func (r *Reducer) GetAllProjects() []*types.Project {
	projects := make([]*types.Project, 0, len(r.projects))
	for _, project := range r.projects {
		projects = append(projects, project)
	}
	return projects
}

// GetProjectUIDToNameMap returns a map of project UIDs to their names
func (r *Reducer) GetProjectUIDToNameMap() map[string]string {
	result := make(map[string]string)
	for _, project := range r.projects {
		result[project.ProjectUID] = project.Name
	}
	return result
}

// GetProjectNameForTask returns the project name for a task
func (r *Reducer) GetProjectNameForTask(taskUID string) (string, error) {
	// Get the task from reducer
	task, exists := r.GetTask(taskUID)
	if !exists {
		return "", fmt.Errorf("task %s not found", taskUID)
	}

	// Get project from reducer
	project, exists := r.GetProject(task.ProjectUUID)
	if !exists {
		// Fall back to project UID if project not found
		return task.ProjectUUID, nil
	}

	return project.Name, nil
}

// FinalizeRelations builds relations for all tasks and computes blocked status
func (r *Reducer) FinalizeRelations(config *config.Config) {
	// Build relations for all tasks
	for uuid, task := range r.tasks {
		task.Relations = r.relations.BuildTaskRelations(uuid)
	}

	// Compute blocked status
	utils.ComputeBlocked(r.relations, r.tasks, config.Blocking.BlockingAxis, config.Blocking.DoneStates)
}

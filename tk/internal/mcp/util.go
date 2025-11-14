package mcp

import (
	"fmt"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
	"github.com/neongreen/mono/tk/internal/utils"
)

// GetActor returns the current user, falling back to "mcp-server" if unavailable
func GetActor() string {
	user, err := utils.GetCurrentUser()
	if err != nil {
		return "mcp-server"
	}
	return user
}

// GetDisplayID returns a human-readable display ID for a task
func GetDisplayID(db *database.DB, taskUUID string) string {
	displayID, err := database.RenderTaskDisplayID(db, taskUUID)
	if err != nil {
		return taskUUID
	}
	return displayID
}

// ResolveProject resolves a project reference to a UID, using "tk" as default
func ResolveProject(db *database.DB, projectRef string) (string, error) {
	if projectRef == "" {
		projectRef = "tk"
	}

	projectUID, err := database.ResolveProjectRef(db, types.NewProjectRef(projectRef))
	if err != nil {
		return "", fmt.Errorf("project %q not found - create it first with: tk project create %s", projectRef, projectRef)
	}
	return projectUID.String(), nil
}

package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/tasks"
	"github.com/neongreen/mono/tk/internal/types"
)

func MoveTaskTool(db *database.DB) func(context.Context, *sdk.CallToolRequest, MoveTaskArgs) (*sdk.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdk.CallToolRequest, args MoveTaskArgs) (*sdk.CallToolResult, any, error) {
		if args.TaskID == "" {
			return nil, nil, fmt.Errorf("task_id is required")
		}
		if args.ToProject == "" {
			return nil, nil, fmt.Errorf("to_project is required")
		}

		// Resolve task
		taskUID, err := database.ResolveTaskReference(db, types.NewTaskRef(args.TaskID))
		if err != nil {
			return nil, nil, fmt.Errorf("task not found: %w", err)
		}

		// Resolve destination project
		toProjectUID, err := database.ResolveProjectRef(db, types.NewProjectRef(args.ToProject))
		if err != nil {
			return nil, nil, fmt.Errorf("destination project not found: %w", err)
		}

		displayID := GetDisplayID(db, taskUID)
		actor := GetActor()

		// Move task (use "auto" mode for number assignment - safest option)
		opts := tasks.MoveOptions{
			Mode:        "auto",
			OnCollision: "auto",
		}

		if err := tasks.Move(db, taskUID, toProjectUID.String(), opts, actor, &clock.RealClock{}); err != nil {
			return nil, nil, fmt.Errorf("failed to move task: %w", err)
		}

		// Get new display ID after move
		newDisplayID := GetDisplayID(db, taskUID)

		return nil, map[string]any{
			"task_id":        displayID,
			"new_display_id": newDisplayID,
			"to_project":     args.ToProject,
			"moved":          true,
			"message":        fmt.Sprintf("Moved task %s to project %s (now %s)", displayID, args.ToProject, newDisplayID),
		}, nil
	}
}

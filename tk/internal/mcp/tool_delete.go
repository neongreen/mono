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

func DeleteTaskTool(db *database.DB) func(context.Context, *sdk.CallToolRequest, DeleteTaskArgs) (*sdk.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdk.CallToolRequest, args DeleteTaskArgs) (*sdk.CallToolResult, any, error) {
		if args.TaskID == "" {
			return nil, nil, fmt.Errorf("task_id is required")
		}

		// Resolve task reference
		taskUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(args.TaskID))
		if err != nil {
			return nil, nil, fmt.Errorf("task not found: %w", err)
		}

		displayID := GetDisplayID(db, taskUUID)
		actor := GetActor()

		// Delete task
		if err := tasks.Delete(db, taskUUID, actor, &clock.RealClock{}); err != nil {
			return nil, nil, fmt.Errorf("failed to delete task: %w", err)
		}

		return nil, map[string]any{
			"task_id":    displayID,
			"deleted":    true,
			"message":    fmt.Sprintf("Deleted task %s", displayID),
		}, nil
	}
}

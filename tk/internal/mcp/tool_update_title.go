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

func UpdateTitleTool(db *database.DB) func(context.Context, *sdk.CallToolRequest, UpdateTitleArgs) (*sdk.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdk.CallToolRequest, args UpdateTitleArgs) (*sdk.CallToolResult, any, error) {
		if args.TaskID == "" {
			return nil, nil, fmt.Errorf("task_id is required")
		}
		if args.Title == "" {
			return nil, nil, fmt.Errorf("title is required")
		}

		// Resolve task reference
		taskUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(args.TaskID))
		if err != nil {
			return nil, nil, fmt.Errorf("task not found: %w", err)
		}

		displayID := GetDisplayID(db, taskUUID)
		actor := GetActor()

		// Update title
		if err := tasks.EditTitle(db, taskUUID, args.Title, actor, &clock.RealClock{}); err != nil {
			return nil, nil, fmt.Errorf("failed to update title: %w", err)
		}

		return nil, map[string]any{
			"task_id":    displayID,
			"new_title":  args.Title,
			"updated":    true,
			"message":    fmt.Sprintf("Updated title for %s", displayID),
		}, nil
	}
}

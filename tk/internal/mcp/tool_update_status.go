package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/tasks"
	"github.com/neongreen/mono/tk/internal/types"
)

func UpdateStatusTool(db *database.DB) func(context.Context, *sdk.CallToolRequest, UpdateStatusArgs) (*sdk.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdk.CallToolRequest, args UpdateStatusArgs) (*sdk.CallToolResult, any, error) {
		actor := GetActor()

		// Default axis to "generic"
		axis := args.Axis
		if axis == "" {
			axis = "generic"
		}

		// Default role to "human"
		role := args.Role
		if role == "" {
			role = "human"
		}

		// Resolve task ID to UUID
		taskUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(args.TaskID))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve task ID: %w", err)
		}

		// Update status
		err = tasks.Mark(db, taskUUID, tasks.MarkOptions{
			Axis:  axis,
			State: args.Status,
			Role:  role,
		}, actor, &clock.RealClock{})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to update status: %w", err)
		}

		displayID := GetDisplayID(db, taskUUID)

		// Return JSON response
		response := map[string]interface{}{
			"uuid":       taskUUID,
			"display_id": displayID,
			"status":     args.Status,
			"axis":       axis,
		}

		responseJSON, err := json.Marshal(response)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: string(responseJSON)},
			},
		}, nil, nil
	}
}

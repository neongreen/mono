package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/query"
	"github.com/neongreen/mono/tk/internal/types"
)

func ListTasksTool(db *database.DB) func(context.Context, *sdk.CallToolRequest, ListTasksArgs) (*sdk.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdk.CallToolRequest, args ListTasksArgs) (*sdk.CallToolResult, any, error) {
		cfg, err := LoadConfig()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load config: %w", err)
		}

		reducer, err := db.GetCachedReducerWithConfig(cfg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to build reducer: %w", err)
		}

		// Get all tasks
		allTasks := reducer.GetAllTasks()

		// Build filter options
		opts := query.FilterOptions{
			BlockedOnly: args.Blocked,
		}

		// Add item kind filter if specified
		if args.Kind != "" {
			opts.ItemKinds = []string{args.Kind}
		}

		// Filter by project if specified
		var taskUIDSet map[string]bool
		if args.Project != "" {
			projectUID, err := database.ResolveProjectRef(db, types.NewProjectRef(args.Project))
			if err != nil {
				return nil, nil, fmt.Errorf("failed to resolve project: %w", err)
			}
			taskUIDSet = make(map[string]bool)
			for _, task := range allTasks {
				if task.ProjectUUID == projectUID.String() {
					taskUIDSet[task.TaskUUID] = true
				}
			}
		}

		// Filter by status if specified
		if args.Status != "" {
			opts.AxisFilter = "generic:" + args.Status
		}

		// Apply filters
		filteredTasks := query.FilterTasks(allTasks, taskUIDSet, opts)

		// Apply limit
		limit := args.Limit
		if limit <= 0 {
			limit = 50
		}
		if len(filteredTasks) > limit {
			filteredTasks = filteredTasks[:limit]
		}

		// Format tasks as JSON
		var taskList []map[string]any
		for _, task := range filteredTasks {
			statusStr := "unknown"
			if axis, ok := task.Axes["generic"]; ok && axis.Effective != "" {
				statusStr = axis.Effective
			}

			displayID := GetDisplayID(db, task.TaskUUID)

			taskInfo := map[string]any{
				"uuid":       task.TaskUUID,
				"display_id": displayID,
				"project":    task.ProjectUUID,
				"title":      task.Title,
				"status":     statusStr,
			}
			taskList = append(taskList, taskInfo)
		}

		result, err := json.Marshal(taskList)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal tasks: %w", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: string(result)},
			},
		}, nil, nil
	}
}

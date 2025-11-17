package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

func GetTaskTool(db *database.DB) func(context.Context, *sdk.CallToolRequest, GetTaskArgs) (*sdk.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdk.CallToolRequest, args GetTaskArgs) (*sdk.CallToolResult, any, error) {
		cfg, err := LoadConfig()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load config: %w", err)
		}

		reducer, err := db.GetCachedReducerWithConfig(cfg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get reducer: %w", err)
		}

		taskUID, err := database.ResolveTaskReference(db, types.NewTaskRef(args.TaskID))
		if err != nil {
			return nil, nil, fmt.Errorf("task not found: %w", err)
		}

		task, ok := reducer.GetTask(taskUID)
		if !ok {
			return nil, nil, fmt.Errorf("task not found: %s", args.TaskID)
		}

		// Get status
		statusStr := "unknown"
		if axis, ok := task.Axes["generic"]; ok && axis.Effective != "" {
			statusStr = axis.Effective
		}

		displayID := GetDisplayID(db, taskUID)

		// Build comprehensive task info
		taskInfo := map[string]any{
			"uuid":       task.TaskUUID,
			"display_id": displayID,
			"project":    task.ProjectUUID,
			"title":      task.Title,
			"status":     statusStr,
		}

		// Add metadata if present
		if len(task.Metadata) > 0 {
			metadata := make(map[string]string)
			for key, meta := range task.Metadata {
				// Unmarshal the Effective value from json.RawMessage
				var value string
				if err := json.Unmarshal(meta.Effective, &value); err == nil {
					metadata[key] = value
				}
			}
			if len(metadata) > 0 {
				taskInfo["metadata"] = metadata
			}
		}

		// Add notes if present
		if len(task.Notes) > 0 {
			notes := make([]map[string]any, 0, len(task.Notes))
			for _, note := range task.Notes {
				notes = append(notes, map[string]any{
					"markdown":  note.Markdown,
					"actor":     note.Actor,
					"timestamp": note.Timestamp,
				})
			}
			taskInfo["notes"] = notes
		}

		// Add relations if present
		if task.Relations != nil {
			relations := make(map[string]any)

			// Subtasks
			if len(task.Relations.Subtask.Children) > 0 {
				subtasks := make([]map[string]string, 0, len(task.Relations.Subtask.Children))
				for _, childUID := range task.Relations.Subtask.Children {
					if childTask, ok := reducer.GetTask(childUID); ok {
						subtasks = append(subtasks, map[string]string{
							"uuid":       childTask.TaskUUID,
							"display_id": childTask.TaskDisplayID,
							"title":      childTask.Title,
						})
					}
				}
				relations["subtasks"] = subtasks
			}

			// Parent
			if task.Relations.Subtask.Parent != "" {
				if parentTask, ok := reducer.GetTask(task.Relations.Subtask.Parent); ok {
					relations["parent"] = map[string]string{
						"uuid":       parentTask.TaskUUID,
						"display_id": parentTask.TaskDisplayID,
						"title":      parentTask.Title,
					}
				}
			}

			if len(relations) > 0 {
				taskInfo["relations"] = relations
			}
		}

		result, err := json.Marshal(taskInfo)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal task: %w", err)
		}

		return &sdk.CallToolResult{
			Content: []sdk.Content{
				&sdk.TextContent{Text: string(result)},
			},
		}, nil, nil
	}
}

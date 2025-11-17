package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

func SetMetadataTool(db *database.DB) func(context.Context, *sdk.CallToolRequest, SetMetadataArgs) (*sdk.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdk.CallToolRequest, args SetMetadataArgs) (*sdk.CallToolResult, any, error) {
		if args.TaskID == "" {
			return nil, nil, fmt.Errorf("task_id is required")
		}
		if args.Key == "" {
			return nil, nil, fmt.Errorf("key is required")
		}
		if args.Value == nil {
			return nil, nil, fmt.Errorf("value is required")
		}

		// Resolve task reference
		taskUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(args.TaskID))
		if err != nil {
			return nil, nil, fmt.Errorf("task not found: %w", err)
		}

		displayID := GetDisplayID(db, taskUUID)
		actor := GetActor()

		// Default role to human if not specified
		role := args.Role
		if role == "" {
			role = "human"
		}

		// Marshal value to JSON
		valueJSON, err := json.Marshal(args.Value)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal value: %w", err)
		}

		// Generate event
		eventID, err := database.GenerateEventID(db)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate event ID: %w", err)
		}

		ts, err := db.GetNextLamportTS()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get timestamp: %w", err)
		}

		// Create task.meta.set event
		payload := types.TaskMetaSetPayload{
			TaskUUID: taskUUID,
			TaskID:   args.TaskID,
			Key:      args.Key,
			Value:    json.RawMessage(valueJSON),
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal payload: %w", err)
		}

		event := types.Event{
			ID:        eventID,
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      role,
			Kind:      string(types.EventKindTaskMetaSet),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return nil, nil, fmt.Errorf("failed to insert event: %w", err)
		}

		return nil, map[string]any{
			"task_id": displayID,
			"key":     args.Key,
			"value":   args.Value,
			"role":    role,
			"message": fmt.Sprintf("Set %s for task %s", args.Key, displayID),
		}, nil
	}
}

package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

func RelateTasksTool(db *database.DB) func(context.Context, *sdk.CallToolRequest, RelateTasksArgs) (*sdk.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdk.CallToolRequest, args RelateTasksArgs) (*sdk.CallToolResult, any, error) {
		actor := GetActor()

		// Resolve parent task ID
		parentUID, err := database.ResolveTaskReference(db, types.NewTaskRef(args.ParentID))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve parent task: %w", err)
		}

		// Resolve child task ID
		childUID, err := database.ResolveTaskReference(db, types.NewTaskRef(args.ChildID))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve child task: %w", err)
		}

		// Generate event ID and timestamp
		eventID, err := database.GenerateEventID(db)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate event ID: %w", err)
		}

		lamportTS, err := db.GetNextLamportTS()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get lamport timestamp: %w", err)
		}

		// Create relation payload
		payload := types.RelationAddPayload{
			TaskUUID:   parentUID,
			RelType:    "subtask",
			TargetTask: childUID,
			Direction:  "out",
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal payload: %w", err)
		}

		// Create event
		event := types.Event{
			ID:        eventID,
			TS:        lamportTS,
			CreatedAt: clock.RealClock{}.Now(),
			Actor:     actor,
			Role:      "human",
			Kind:      string(types.EventKindRelationAdd),
			Payload:   payloadJSON,
		}

		// Insert event
		if err := db.InsertEvent(event); err != nil {
			return nil, nil, fmt.Errorf("failed to create relation: %w", err)
		}

		parentDisplayID := GetDisplayID(db, parentUID)
		childDisplayID := GetDisplayID(db, childUID)

		// Return JSON response
		response := map[string]interface{}{
			"parent": map[string]string{
				"uuid":       parentUID,
				"display_id": parentDisplayID,
			},
			"child": map[string]string{
				"uuid":       childUID,
				"display_id": childDisplayID,
			},
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

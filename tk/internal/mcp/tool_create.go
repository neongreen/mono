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

func CreateTaskTool(db *database.DB) func(context.Context, *sdk.CallToolRequest, CreateTaskArgs) (*sdk.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdk.CallToolRequest, args CreateTaskArgs) (*sdk.CallToolResult, any, error) {
		if args.Title == "" {
			return nil, nil, fmt.Errorf("title is required")
		}

		actor := GetActor()

		// Resolve project
		projectUID, err := ResolveProject(db, args.Project)
		if err != nil {
			return nil, nil, err
		}

		// Create task
		result, err := tasks.Create(db, tasks.CreateParams{
			ProjectUID: types.ProjectUID(projectUID),
			Title:      args.Title,
		}, actor, &clock.RealClock{})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create task: %w", err)
		}

		taskUUID := string(result.TaskUID)

		// Set initial status if provided
		if args.Status != "" {
			cfg, err := LoadConfig()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to load config: %w", err)
			}
			if err := tasks.Mark(db, taskUUID, tasks.MarkOptions{
				Axis:  cfg.Axes.Blocking,
				State: args.Status,
				Role:  "human",
			}, actor, &clock.RealClock{}); err != nil {
				return nil, nil, fmt.Errorf("failed to set status: %w", err)
			}
		}

		// Set metadata if provided
		for key, value := range args.Metadata {
			if err := setTaskMetadata(db, taskUUID, key, value, actor); err != nil {
				return nil, nil, fmt.Errorf("failed to set metadata %q: %w", key, err)
			}
		}

		// Return JSON response
		response := map[string]interface{}{
			"uuid":       taskUUID,
			"display_id": result.DisplayID,
			"title":      args.Title,
			"project":    projectUID,
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

func setTaskMetadata(db *database.DB, taskUUID, key, value, actor string) error {
	eventID, err := database.GenerateEventID(db)
	if err != nil {
		return fmt.Errorf("failed to generate event ID: %w", err)
	}

	lamportTS, err := db.GetNextLamportTS()
	if err != nil {
		return fmt.Errorf("failed to get lamport timestamp: %w", err)
	}

	payload := types.TaskMetaSetPayload{
		TaskUUID: taskUUID,
		Key:      key,
		Value:    value,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	event := types.Event{
		ID:        eventID,
		TS:        lamportTS,
		CreatedAt: clock.RealClock{}.Now(),
		Actor:     actor,
		Role:      "human",
		Kind:      string(types.EventKindTaskMetaSet),
		Payload:   payloadJSON,
	}

	return db.InsertEvent(event)
}

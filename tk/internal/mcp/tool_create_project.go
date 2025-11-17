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

func CreateProjectTool(db *database.DB) func(context.Context, *sdk.CallToolRequest, CreateProjectArgs) (*sdk.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdk.CallToolRequest, args CreateProjectArgs) (*sdk.CallToolResult, any, error) {
		if args.Name == "" {
			return nil, nil, fmt.Errorf("name is required")
		}

		// Validate project name
		if err := types.ValidateProjectName(args.Name); err != nil {
			return nil, nil, fmt.Errorf("invalid project name: %w", err)
		}

		actor := GetActor()

		// Generate new project UID
		projectUID := types.NewProjectUID()

		// Create project.created event
		payload := types.ProjectCreatedPayload{
			ProjectUID:  projectUID,
			Type:        types.ProjectTypeLocal,
			Name:        args.Name,
			Description: args.Description,
			CreatedBy:   actor,
		}

		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal payload: %w", err)
		}

		eventID, err := database.GenerateEventID(db)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate event ID: %w", err)
		}

		ts, err := db.GetNextLamportTS()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get timestamp: %w", err)
		}

		event := types.Event{
			ID:        eventID,
			TS:        ts,
			CreatedAt: time.Now(),
			Actor:     actor,
			Role:      "human",
			Kind:      string(types.EventKindProjectCreated),
			Payload:   payloadJSON,
		}

		if err := db.InsertEvent(event); err != nil {
			return nil, nil, fmt.Errorf("failed to insert event: %w", err)
		}

		// Project the event
		if err := db.ProjectProjectCreatedEvent(event); err != nil {
			return nil, nil, fmt.Errorf("failed to project event: %w", err)
		}

		return nil, map[string]any{
			"project_uid": string(projectUID),
			"name":        args.Name,
			"description": args.Description,
			"created":     true,
			"message":     fmt.Sprintf("Created project: %s", args.Name),
		}, nil
	}
}

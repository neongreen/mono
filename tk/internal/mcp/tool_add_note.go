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

func AddNoteTool(db *database.DB) func(context.Context, *sdk.CallToolRequest, AddNoteArgs) (*sdk.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdk.CallToolRequest, args AddNoteArgs) (*sdk.CallToolResult, any, error) {
		actor := GetActor()

		// Resolve task ID to UUID
		taskUUID, err := database.ResolveTaskReference(db, types.NewTaskRef(args.TaskID))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve task ID: %w", err)
		}

		// Add note
		err = tasks.AddNote(db, taskUUID, args.Note, actor, &clock.RealClock{})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to add note: %w", err)
		}

		displayID := GetDisplayID(db, taskUUID)

		// Return JSON response
		response := map[string]any{
			"uuid":       taskUUID,
			"display_id": displayID,
			"note":       args.Note,
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

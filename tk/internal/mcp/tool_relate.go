package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

func RelateTasksTool(db *database.DB) func(context.Context, *sdk.CallToolRequest, RelateTasksArgs) (*sdk.CallToolResult, any, error) {
	return func(ctx context.Context, req *sdk.CallToolRequest, args RelateTasksArgs) (*sdk.CallToolResult, any, error) {
		actor := GetActor()

		// Handle backward compatibility (parent_id/child_id -> source_id/target_id)
		sourceID := args.SourceID
		targetID := args.TargetID
		if sourceID == "" && args.ParentID != "" {
			sourceID = args.ParentID
		}
		if targetID == "" && args.ChildID != "" {
			targetID = args.ChildID
		}

		if sourceID == "" {
			return nil, nil, fmt.Errorf("source_id is required")
		}
		if targetID == "" {
			return nil, nil, fmt.Errorf("target_id is required")
		}

		// Validate relation type
		validTypes := []string{"blocks", "blocked_by", "subtask", "parent", "related", "duplicate_of", "supersedes"}
		relationType := args.RelationType
		if relationType == "" {
			relationType = "subtask" // Default for backward compatibility
		}

		isValid := slices.Contains(validTypes, relationType)
		if !isValid {
			return nil, nil, fmt.Errorf("invalid relation_type %q, must be one of: blocks, blocked_by, subtask, parent, related, duplicate_of, supersedes", relationType)
		}

		// Resolve source task ID
		srcUID, err := database.ResolveTaskReference(db, types.NewTaskRef(sourceID))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve source task: %w", err)
		}

		// Resolve target task ID
		dstUID, err := database.ResolveTaskReference(db, types.NewTaskRef(targetID))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve target task: %w", err)
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
			Src:  srcUID,
			Type: relationType,
			Dst:  dstUID,
			Note: args.Note,
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal payload: %w", err)
		}

		// Create event
		event := types.Event{
			ID:        eventID,
			TS:        lamportTS,
			CreatedAt: (&clock.RealClock{}).Now(),
			Actor:     actor,
			Role:      "human",
			Kind:      string(types.EventKindRelationAdd),
			Payload:   json.RawMessage(payloadJSON),
		}

		// Insert event
		if err := db.InsertEvent(event); err != nil {
			return nil, nil, fmt.Errorf("failed to create relation: %w", err)
		}

		srcDisplayID := GetDisplayID(db, srcUID)
		dstDisplayID := GetDisplayID(db, dstUID)

		// Return JSON response
		response := map[string]any{
			"source": map[string]string{
				"uuid":       srcUID,
				"display_id": srcDisplayID,
			},
			"target": map[string]string{
				"uuid":       dstUID,
				"display_id": dstDisplayID,
			},
			"relation_type": relationType,
			"note":          args.Note,
			"message":       fmt.Sprintf("Added relation: %s %s %s", srcDisplayID, relationType, dstDisplayID),
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

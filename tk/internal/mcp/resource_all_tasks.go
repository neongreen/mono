package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/neongreen/mono/tk/internal/database"
)

func AllTasksResource(db *database.DB) func(context.Context, *sdk.ReadResourceRequest) (*sdk.ReadResourceResult, error) {
	return func(ctx context.Context, req *sdk.ReadResourceRequest) (*sdk.ReadResourceResult, error) {
		cfg, err := LoadConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}

		reducer, err := db.GetCachedReducerWithConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to get reducer: %w", err)
		}

		tasks := reducer.GetAllTasks()

		// Convert tasks to JSON
		tasksJSON, err := json.Marshal(tasks)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tasks: %w", err)
		}

		return &sdk.ReadResourceResult{
			Contents: []*sdk.ResourceContents{
				{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     string(tasksJSON),
				},
			},
		}, nil
	}
}

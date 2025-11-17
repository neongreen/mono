package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	config_pkg "github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/types"
)

func TaskResource(db *database.DB) func(context.Context, *sdk.ReadResourceRequest) (*sdk.ReadResourceResult, error) {
	return func(ctx context.Context, req *sdk.ReadResourceRequest) (*sdk.ReadResourceResult, error) {
		// Extract task ID from URI (task://{id})
		taskID := req.Params.URI[7:] // Remove "task://"

		cfg, err := config_pkg.LoadConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}

		reducer, err := db.GetCachedReducerWithConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to get reducer: %w", err)
		}

		taskUID, err := database.ResolveTaskReference(db, types.NewTaskRef(taskID))
		if err != nil {
			return nil, fmt.Errorf("task not found: %w", err)
		}

		task, ok := reducer.GetTask(taskUID)
		if !ok {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}

		// Convert task to JSON
		taskJSON, err := json.Marshal(task)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal task: %w", err)
		}

		return &sdk.ReadResourceResult{
			Contents: []*sdk.ResourceContents{
				{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     string(taskJSON),
				},
			},
		}, nil
	}
}

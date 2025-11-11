package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/clock"
	"github.com/neongreen/mono/tk/internal/config"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/query"
	"github.com/neongreen/mono/tk/internal/tasks"
	"github.com/neongreen/mono/tk/internal/types"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps a tk database and exposes it via MCP
type Server struct {
	db     *database.DB
	server *sdkmcp.Server
}

// NewServer creates a new MCP server for tk
func NewServer(db *database.DB) *Server {
	impl := &sdkmcp.Implementation{
		Name:    "tk-mcp-server",
		Version: "1.0.0",
	}

	server := sdkmcp.NewServer(impl, nil)
	s := &Server{
		db:     db,
		server: server,
	}

	// Register tools
	s.registerListTasksTool()
	s.registerShowTaskTool()
	s.registerCreateTaskTool()
	s.registerUpdateTaskStatusTool()
	s.registerAddNoteTool()

	return s
}

// Run starts the MCP server on the given transport
func (s *Server) Run(ctx context.Context, transport sdkmcp.Transport) error {
	return s.server.Run(ctx, transport)
}

// listTasksArgs defines the arguments for list_tasks tool
type listTasksArgs struct {
	Project string `json:"project,omitempty" jsonschema:"description=Filter by project alias or UID"`
	Status  string `json:"status,omitempty" jsonschema:"description=Filter by status (e.g., 'done', 'in_progress')"`
	Blocked bool   `json:"blocked,omitempty" jsonschema:"description=Show only blocked tasks"`
}

func (s *Server) registerListTasksTool() {
	sdkmcp.AddTool(s.server, &sdkmcp.Tool{
		Name:        "list_tasks",
		Description: "List tasks in the tk database",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, args listTasksArgs) (*sdkmcp.CallToolResult, any, error) {
		// Build reducer to query tasks
		cfg, err := config.LoadConfig()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load config: %w", err)
		}

		reducer, err := s.db.GetCachedReducerWithConfig(cfg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to build reducer: %w", err)
		}

		// Get all tasks
		allTasks := reducer.GetAllTasks()

		// Build filter options
		opts := query.FilterOptions{
			BlockedOnly:   args.Blocked,
		}

		// Filter by project if specified
		var taskUIDSet map[string]bool
		if args.Project != "" {
			projectUID, err := database.ResolveProjectRef(s.db, types.NewProjectRef(args.Project))
			if err != nil {
				return nil, nil, fmt.Errorf("failed to resolve project: %w", err)
			}
			// Get all tasks for this project
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

		// Format tasks for output
		var taskList []map[string]interface{}
		for _, task := range filteredTasks {
			statusStr := "unknown"
			if axis, ok := task.Axes["generic"]; ok && axis.Effective != "" {
				statusStr = axis.Effective
			}

			// Get display ID
			displayID, err := database.RenderTaskDisplayID(s.db, task.TaskUUID)
			if err != nil {
				displayID = task.TaskUUID
			}

			taskInfo := map[string]interface{}{
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

		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{
				&sdkmcp.TextContent{Text: string(result)},
			},
		}, nil, nil
	})
}

// showTaskArgs defines the arguments for show_task tool
type showTaskArgs struct {
	TaskID string `json:"task_id" jsonschema:"required,description=Task UID or display ID (e.g., 'tk-1')"`
}

func (s *Server) registerShowTaskTool() {
	sdkmcp.AddTool(s.server, &sdkmcp.Tool{
		Name:        "show_task",
		Description: "Show detailed information about a specific task",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, args showTaskArgs) (*sdkmcp.CallToolResult, any, error) {
		cfg, err := config.LoadConfig()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load config: %w", err)
		}

		reducer, err := s.db.GetCachedReducerWithConfig(cfg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to build reducer: %w", err)
		}

		// Resolve task ID to UUID
		taskUUID, err := database.ResolveTaskReference(s.db, types.NewTaskRef(args.TaskID))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve task ID: %w", err)
		}

		// Get task
		task, ok := reducer.GetTask(taskUUID)
		if !ok {
			return nil, nil, fmt.Errorf("task not found: %s", args.TaskID)
		}

		// Get status
		statusStr := "unknown"
		if axis, ok := task.Axes["generic"]; ok && axis.Effective != "" {
			statusStr = axis.Effective
		}

		// Get display ID
		displayID, err := database.RenderTaskDisplayID(s.db, taskUUID)
		if err != nil {
			displayID = taskUUID
		}

		taskInfo := map[string]interface{}{
			"uuid":       task.TaskUUID,
			"display_id": displayID,
			"project":    task.ProjectUUID,
			"title":      task.Title,
			"status":     statusStr,
			"notes":      task.Notes,
		}

		result, err := json.Marshal(taskInfo)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal task: %w", err)
		}

		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{
				&sdkmcp.TextContent{Text: string(result)},
			},
		}, nil, nil
	})
}

// createTaskArgs defines the arguments for create_task tool
type createTaskArgs struct {
	Title   string `json:"title" jsonschema:"required,description=Task title"`
	Project string `json:"project,omitempty" jsonschema:"description=Project alias or UID (default: 'tk')"`
}

func (s *Server) registerCreateTaskTool() {
	sdkmcp.AddTool(s.server, &sdkmcp.Tool{
		Name:        "create_task",
		Description: "Create a new task",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, args createTaskArgs) (*sdkmcp.CallToolResult, any, error) {
		project := args.Project
		if project == "" {
			project = "tk"
		}

		// Resolve project reference to UID
		projectUID, err := database.ResolveProjectRef(s.db, types.NewProjectRef(project))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve project: %w", err)
		}

		// Create task
		result, err := tasks.Create(s.db, tasks.CreateParams{
			ProjectUID: projectUID,
			Title:      args.Title,
		}, "mcp-server", &clock.RealClock{})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create task: %w", err)
		}

		resultText := fmt.Sprintf("Created task: %s", result.DisplayID)

		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{
				&sdkmcp.TextContent{Text: resultText},
			},
		}, nil, nil
	})
}

// updateTaskStatusArgs defines the arguments for update_task_status tool
type updateTaskStatusArgs struct {
	TaskID string `json:"task_id" jsonschema:"required,description=Task UID or display ID"`
	Status string `json:"status" jsonschema:"required,description=New status (e.g., 'in_progress', 'done')"`
	Axis   string `json:"axis,omitempty" jsonschema:"description=Status axis (default: 'generic')"`
	Role   string `json:"role,omitempty" jsonschema:"description=Role making the claim (default: 'agent')"`
}

func (s *Server) registerUpdateTaskStatusTool() {
	sdkmcp.AddTool(s.server, &sdkmcp.Tool{
		Name:        "update_task_status",
		Description: "Update the status of a task",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, args updateTaskStatusArgs) (*sdkmcp.CallToolResult, any, error) {
		axis := args.Axis
		if axis == "" {
			axis = "generic"
		}

		role := args.Role
		if role == "" {
			role = "agent"
		}

		// Resolve task ID to UUID
		taskUUID, err := database.ResolveTaskReference(s.db, types.NewTaskRef(args.TaskID))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve task ID: %w", err)
		}

		// Update status
		err = tasks.Mark(s.db, taskUUID, tasks.MarkOptions{
			Axis:  axis,
			State: args.Status,
			Role:  role,
		}, "mcp-server", &clock.RealClock{})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to update status: %w", err)
		}

		result := fmt.Sprintf("Updated task %s status to %s", args.TaskID, args.Status)

		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{
				&sdkmcp.TextContent{Text: result},
			},
		}, nil, nil
	})
}

// addNoteArgs defines the arguments for add_note tool
type addNoteArgs struct {
	TaskID string `json:"task_id" jsonschema:"required,description=Task UID or display ID"`
	Note   string `json:"note" jsonschema:"required,description=Note text to add"`
}

func (s *Server) registerAddNoteTool() {
	sdkmcp.AddTool(s.server, &sdkmcp.Tool{
		Name:        "add_note",
		Description: "Add a note to a task",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, args addNoteArgs) (*sdkmcp.CallToolResult, any, error) {
		// Resolve task ID to UUID
		taskUUID, err := database.ResolveTaskReference(s.db, types.NewTaskRef(args.TaskID))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to resolve task ID: %w", err)
		}

		// Add note
		err = tasks.AddNote(s.db, taskUUID, args.Note, "mcp-server", &clock.RealClock{})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to add note: %w", err)
		}

		result := fmt.Sprintf("Added note to task %s", args.TaskID)

		return &sdkmcp.CallToolResult{
			Content: []sdkmcp.Content{
				&sdkmcp.TextContent{Text: result},
			},
		}, nil, nil
	})
}

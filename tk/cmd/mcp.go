package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/mcp"
	"github.com/spf13/cobra"
)

var (
	mcpPort       string
	mcpHost       string
	mcpStdio      bool
	mcpStreamable bool
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run tk as an MCP server",
	Long: `Run tk as a Model Context Protocol (MCP) server.

The MCP server exposes tk's functionality through a standardized protocol
that can be used by AI assistants and other tools.

The server supports multiple transport modes:
  - HTTP with SSE (default): Server-Sent Events for unidirectional streaming
  - HTTP Streamable (--streamable): Bidirectional HTTP streaming (more efficient)
  - stdio (--stdio): Direct stdin/stdout communication

Available tools:
  - create_task: Create a new task with optional status and metadata
  - list_tasks: List tasks with filtering by project, status, or blocked state
  - get_task: Get detailed information about a specific task
  - update_status: Update task status on any axis
  - add_note: Add a markdown note to a task
  - relate_tasks: Create parent/subtask relationships

Available resources:
  - task://{id}: Get individual task data as JSON
  - tasks://all: Get all tasks as JSON array

Example usage:
  tk mcp                     # Start HTTP server with SSE on localhost:8080
  tk mcp --streamable        # Start HTTP server with streamable transport
  tk mcp --port 9000         # Start HTTP server on custom port
  tk mcp --stdio             # Start in stdio mode for MCP client`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Open database
		db, err := database.OpenExistingDB()
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		// Create MCP server
		server := sdk.NewServer(&sdk.Implementation{
			Name:    "tk",
			Version: "1.0.0",
		}, &sdk.ServerOptions{
			Instructions: "tk task management system. Use the provided tools to create, list, update, and manage tasks. All responses are in JSON format for easy parsing.",
		})

		// Register tools
		sdk.AddTool(server, &sdk.Tool{
			Name:        "create_task",
			Description: "Create a new task with optional status and metadata",
		}, mcp.CreateTaskTool(db))

		sdk.AddTool(server, &sdk.Tool{
			Name:        "list_tasks",
			Description: "List tasks with optional filtering by project, status, or blocked state",
		}, mcp.ListTasksTool(db))

		sdk.AddTool(server, &sdk.Tool{
			Name:        "get_task",
			Description: "Get detailed information about a task including metadata, notes, and relations",
		}, mcp.GetTaskTool(db))

		sdk.AddTool(server, &sdk.Tool{
			Name:        "update_status",
			Description: "Update the status of a task on a specific axis",
		}, mcp.UpdateStatusTool(db))

		sdk.AddTool(server, &sdk.Tool{
			Name:        "add_note",
			Description: "Add a markdown note to a task",
		}, mcp.AddNoteTool(db))

		sdk.AddTool(server, &sdk.Tool{
			Name:        "relate_tasks",
			Description: "Create a parent/subtask relationship between two tasks",
		}, mcp.RelateTasksTool(db))

		// Register resources
		server.AddResource(&sdk.Resource{
			Name:        "task",
			URI:         "task://{id}",
			Description: "Individual task data as JSON",
			MIMEType:    "application/json",
		}, mcp.TaskResource(db))

		server.AddResource(&sdk.Resource{
			Name:        "all-tasks",
			URI:         "tasks://all",
			Description: "All tasks in the system as JSON array",
			MIMEType:    "application/json",
		}, mcp.AllTasksResource(db))

		// Start server based on mode
		if mcpStdio {
			log.Println("Starting MCP server in stdio mode...")
			transport := &sdk.LoggingTransport{
				Transport: &sdk.StdioTransport{},
				Writer:    os.Stderr,
			}
			if err := server.Run(ctx, transport); err != nil {
				return fmt.Errorf("server failed: %w", err)
			}
		} else {
			addr := fmt.Sprintf("%s:%s", mcpHost, mcpPort)
			var handler http.Handler

			if mcpStreamable {
				// Use streamable HTTP transport (bidirectional)
				handler = sdk.NewStreamableHTTPHandler(func(*http.Request) *sdk.Server {
					return server
				}, nil)
				log.Printf("Starting MCP server with streamable HTTP transport at http://%s\n", addr)
				log.Printf("Connect MCP clients to: http://%s\n", addr)
			} else {
				// Use SSE transport (default, unidirectional)
				handler = sdk.NewSSEHandler(func(*http.Request) *sdk.Server {
					return server
				}, nil)
				log.Printf("Starting MCP server with SSE transport at http://%s\n", addr)
				log.Printf("Connect MCP clients to: http://%s\n", addr)
			}

			if err := http.ListenAndServe(addr, handler); err != nil {
				return fmt.Errorf("HTTP server failed: %w", err)
			}
		}

		return nil
	},
}

func init() {
	mcpCmd.Flags().StringVar(&mcpPort, "port", "8080", "Port to listen on (HTTP mode)")
	mcpCmd.Flags().StringVar(&mcpHost, "host", "localhost", "Host to bind to (HTTP mode)")
	mcpCmd.Flags().BoolVar(&mcpStdio, "stdio", false, "Use stdio transport instead of HTTP")
	mcpCmd.Flags().BoolVar(&mcpStreamable, "streamable", false, "Use streamable HTTP transport (more efficient than SSE)")
	RootCmd.AddCommand(mcpCmd)
}

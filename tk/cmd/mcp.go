package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/neongreen/mono/tk/internal/database"
	"github.com/neongreen/mono/tk/internal/mcpserver"
	"github.com/spf13/cobra"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run tk as an MCP server",
	Long: `Run tk as a Model Context Protocol (MCP) server.

The MCP server exposes tk's functionality through a standardized protocol
that can be used by AI assistants and other tools.

The server runs on stdio by default and provides the following tools:
- list_tasks: List tasks with optional filtering
- show_task: Show detailed information about a task
- create_task: Create a new task
- update_task_status: Update task status
- add_note: Add a note to a task

Example usage:
  tk mcp

The server will run until interrupted (Ctrl+C).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get database path
		dbPath, err := database.GetDBPath()
		if err != nil {
			return fmt.Errorf("failed to get database path: %w", err)
		}

		// Open database
		db, err := database.OpenDB(dbPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		// Initialize database if needed
		if err := db.InitDB(); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}

		// Create MCP server
		server := mcpserver.NewServer(db)

		// Create stdio transport
		transport := &sdkmcp.StdioTransport{}

		// Run the server
		log.Printf("Starting tk MCP server on stdio...")
		if err := server.Run(context.Background(), transport); err != nil {
			return fmt.Errorf("MCP server failed: %w", err)
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(mcpCmd)
}

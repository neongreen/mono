package log

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/invlog"
	"github.com/spf13/cobra"
)

var SearchCmd = &cobra.Command{
	Use:   "search <pattern>",
	Short: "Search invocation logs for a pattern",
	Long: `Search invocation logs for a pattern in any field.

Requires DuckDB CLI to be installed (https://duckdb.org/docs/installation/).

This is a convenience command that searches across all fields in the log entries.
For more precise control, use 'tk log query' with a SQL WHERE clause.

Examples:
  # Find entries mentioning "error"
  tk log search error

  # Find entries with specific project
  tk log search "project ls"

  # Find debug commands
  tk log search "debug"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		limit, _ := cmd.Flags().GetInt("limit")

		// Build parameterized query to avoid SQL injection
		pattern := args[0]
		query := `
			SELECT * FROM logs
			WHERE CAST(logs AS VARCHAR) LIKE ?
			ORDER BY timestamp DESC
			LIMIT ?
		`

		// Wrap pattern with % wildcards for substring matching
		searchPattern := "%" + pattern + "%"

		results, err := invlog.Query(query, searchPattern, limit)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		if jsonOutput {
			output, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal results: %w", err)
			}
			fmt.Println(string(output))
		} else {
			// Pretty print results
			if len(results) == 0 {
				fmt.Printf("No results found for pattern: %s\n", pattern)
				return nil
			}

			for i, result := range results {
				fmt.Printf("--- Result %d ---\n", i+1)

				// Show timestamp, command, args first
				if ts, ok := result["timestamp"]; ok {
					fmt.Printf("timestamp: %v\n", ts)
				}
				if cmd, ok := result["command"]; ok {
					fmt.Printf("command: %v\n", cmd)
				}
				if args, ok := result["args"]; ok {
					fmt.Printf("args: %v\n", args)
				}
				if success, ok := result["success"]; ok {
					fmt.Printf("success: %v\n", success)
				}

				// Show other fields
				for k, v := range result {
					if k != "timestamp" && k != "command" && k != "args" && k != "success" {
						fmt.Printf("%s: %v\n", k, v)
					}
				}
				fmt.Println()
			}

			fmt.Printf("Total: %d result(s)\n", len(results))
		}

		return nil
	},
}

func init() {
	SearchCmd.Flags().Bool("json", false, "Output as JSON")
	SearchCmd.Flags().Int("limit", 100, "Maximum number of results")
}

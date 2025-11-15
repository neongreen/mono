package log

import (
	"encoding/json"
	"fmt"

	"github.com/neongreen/mono/tk/internal/invlog"
	"github.com/spf13/cobra"
)

var QueryCmd = &cobra.Command{
	Use:   "query <sql>",
	Short: "Query invocation logs with SQL",
	Long: `Execute arbitrary SQL queries against invocation logs using DuckDB.

Requires DuckDB CLI to be installed (https://duckdb.org/docs/installation/).

The logs are available in a table called 'logs'. All JSONL files (compressed
and uncompressed) are automatically loaded.

Examples:
  # Find failed commands
  tk log query "SELECT * FROM logs WHERE success = false LIMIT 10"

  # Find slow commands
  tk log query "SELECT command, args, duration_ms FROM logs WHERE duration_ms > 1000"

  # Search in output
  tk log query "SELECT * FROM logs WHERE stdout LIKE '%error%'"

  # Recent invocations
  tk log query "SELECT * FROM logs ORDER BY timestamp DESC LIMIT 20"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")

		results, err := invlog.Query(args[0])
		if err != nil {
			return fmt.Errorf("query failed: %w", err)
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
				fmt.Println("No results")
				return nil
			}

			for i, result := range results {
				fmt.Printf("--- Result %d ---\n", i+1)
				for k, v := range result {
					fmt.Printf("%s: %v\n", k, v)
				}
				fmt.Println()
			}

			fmt.Printf("Total: %d result(s)\n", len(results))
		}

		return nil
	},
}

func init() {
	QueryCmd.Flags().Bool("json", false, "Output as JSON")
}

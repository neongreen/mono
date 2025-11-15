package rotatinglog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// QueryResult represents a single row from a query
type QueryResult map[string]any

// checkDuckDBAvailable checks if the duckdb CLI is available
func checkDuckDBAvailable() error {
	_, err := exec.LookPath("duckdb")
	if err != nil {
		return fmt.Errorf("duckdb CLI not found in PATH. Please install DuckDB from https://duckdb.org/docs/installation/")
	}
	return nil
}

// Query executes an arbitrary SQL query against all log files in the directory.
// Uses DuckDB CLI to read JSONL files (both .jsonl and .jsonl.zst).
//
// The SQL query should reference the table as 'logs'. Example:
//
//	SELECT * FROM logs WHERE field = ?
//
// Use ? placeholders for parameters to avoid SQL injection.
// Parameters are passed to DuckDB as prepared statement arguments.
// DuckDB automatically detects and decompresses .zst files.
func Query(dir string, sqlQuery string, args ...any) ([]QueryResult, error) {
	// Check if duckdb is available
	if err := checkDuckDBAvailable(); err != nil {
		return nil, err
	}

	// Create a view that reads all JSONL files
	// DuckDB's read_json supports both .jsonl and .jsonl.zst (auto-detects compression)
	// union_by_name=true allows optional fields to appear in different records
	// maximum_object_size increased from default 16MB to 256MB for large log files
	pattern := filepath.Join(dir, "*.jsonl*")
	createView := fmt.Sprintf(`
		CREATE VIEW logs AS
		SELECT * FROM read_json('%s', auto_detect=true, format='newline_delimited', union_by_name=true, maximum_object_size=268435456);
	`, pattern)

	// Build SQL script with prepared statement if there are parameters
	var fullSQL string
	if len(args) > 0 {
		// Convert ? placeholders to DuckDB's $1, $2, etc. format
		preparedQuery := sqlQuery
		for i := 1; i <= len(args); i++ {
			preparedQuery = strings.Replace(preparedQuery, "?", fmt.Sprintf("$%d", i), 1)
		}

		// Build EXECUTE statement with parameters
		var paramList []string
		for _, arg := range args {
			// Convert Go values to SQL literals for EXECUTE
			switch v := arg.(type) {
			case string:
				// Escape single quotes for SQL string literal
				escaped := strings.ReplaceAll(v, "'", "''")
				paramList = append(paramList, fmt.Sprintf("'%s'", escaped))
			case int, int64, float64, bool:
				paramList = append(paramList, fmt.Sprintf("%v", v))
			case nil:
				paramList = append(paramList, "NULL")
			default:
				// Fallback to string representation
				escaped := strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''")
				paramList = append(paramList, fmt.Sprintf("'%s'", escaped))
			}
		}

		fullSQL = fmt.Sprintf("%s\nPREPARE query AS %s;\nEXECUTE query(%s);",
			createView, preparedQuery, strings.Join(paramList, ", "))
	} else {
		// No parameters, just execute the query directly
		fullSQL = createView + "\n" + sqlQuery
	}

	// Execute the query using DuckDB CLI with JSON output
	// Use stdin to pass the SQL to avoid interactive mode issues with -cmd
	cmd := exec.Command("duckdb", ":memory:", "-json")
	cmd.Stdin = strings.NewReader(fullSQL)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		if stderrStr != "" {
			return nil, fmt.Errorf("duckdb query failed: %s", stderrStr)
		}
		return nil, fmt.Errorf("duckdb query failed: %w", err)
	}

	// Parse JSON output
	output := stdout.String()
	if output == "" {
		return []QueryResult{}, nil
	}

	// DuckDB returns a JSON array when using -json flag
	var results []QueryResult
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w", err)
	}

	return results, nil
}

// Search is a convenience function that searches for a pattern in all fields.
// It's equivalent to Query with a LIKE clause on common text fields.
//
// The pattern is used with SQL LIKE and automatically wrapped with % wildcards
// for substring matching. Uses prepared statements to prevent SQL injection.
func Search(dir string, pattern string) ([]QueryResult, error) {
	query := `
		SELECT * FROM logs
		WHERE CAST(logs AS VARCHAR) LIKE ?
		ORDER BY timestamp DESC
		LIMIT 100
	`

	// Wrap pattern with % wildcards for substring matching
	searchPattern := "%" + pattern + "%"

	return Query(dir, query, searchPattern)
}

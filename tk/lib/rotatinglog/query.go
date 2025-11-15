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
//	SELECT * FROM logs WHERE field = 'value'
//
// Note: Parameterized queries (?) are not supported when using the CLI.
// Ensure user input is properly sanitized to avoid SQL injection.
// DuckDB automatically detects and decompresses .zst files.
func Query(dir string, sqlQuery string, args ...any) ([]QueryResult, error) {
	// Check if duckdb is available
	if err := checkDuckDBAvailable(); err != nil {
		return nil, err
	}

	// For backwards compatibility, replace ? placeholders with actual values
	// This is less safe but maintains the API contract
	finalQuery := sqlQuery
	for _, arg := range args {
		// Simple replacement - wrap strings in quotes, convert others to string
		var replacement string
		switch v := arg.(type) {
		case string:
			// Escape single quotes in the string
			escaped := strings.ReplaceAll(v, "'", "''")
			replacement = fmt.Sprintf("'%s'", escaped)
		case int, int64, float64, bool:
			replacement = fmt.Sprintf("%v", v)
		default:
			replacement = fmt.Sprintf("'%v'", v)
		}
		// Replace first occurrence of ?
		finalQuery = strings.Replace(finalQuery, "?", replacement, 1)
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

	// Combine the view creation and the query
	fullSQL := createView + "\n" + finalQuery

	// Execute the query using DuckDB CLI with JSON output
	cmd := exec.Command("duckdb", ":memory:", "-json", "-cmd", fullSQL)
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
// for substring matching.
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

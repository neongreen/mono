package rotatinglog

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/duckdb/duckdb-go/v2"
)

// QueryResult represents a single row from a query
type QueryResult map[string]any

// Query executes an arbitrary SQL query against all log files in the directory.
// Uses DuckDB to read JSONL files (both .jsonl and .jsonl.zst).
//
// The SQL query should reference the table as 'logs'. Example:
//
//	SELECT * FROM logs WHERE field = ?
//
// Use ? placeholders for parameters to avoid SQL injection.
// DuckDB automatically detects and decompresses .zst files.
func Query(dir string, sqlQuery string, args ...interface{}) ([]QueryResult, error) {
	// Create in-memory DuckDB instance
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("failed to open duckdb: %w", err)
	}
	defer db.Close()

	// Create a view that reads all JSONL files
	// DuckDB's read_json supports both .jsonl and .jsonl.zst (auto-detects compression)
	// union_by_name=true allows optional fields to appear in different records
	pattern := filepath.Join(dir, "*.jsonl*")
	createView := fmt.Sprintf(`
		CREATE VIEW logs AS
		SELECT * FROM read_json('%s', auto_detect=true, format='newline_delimited', union_by_name=true)
	`, pattern)

	if _, err := db.Exec(createView); err != nil {
		return nil, fmt.Errorf("failed to create view: %w", err)
	}

	// Execute user query with parameters
	rows, err := db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Read all results
	var results []QueryResult
	for rows.Next() {
		// Create a slice of interface{} to hold column values
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Build result map
		result := make(QueryResult)
		for i, col := range columns {
			result[col] = values[i]
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return results, nil
}

// Search is a convenience function that searches for a pattern in all fields.
// It's equivalent to Query with a LIKE clause on common text fields.
//
// The pattern is used with SQL LIKE and automatically wrapped with % wildcards
// for substring matching. Uses parameterized queries to prevent SQL injection.
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

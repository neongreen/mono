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
//	SELECT * FROM logs WHERE field = 'value'
//
// DuckDB automatically detects and decompresses .zst files.
func Query(dir string, sqlQuery string) ([]QueryResult, error) {
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

	// Execute user query
	rows, err := db.Query(sqlQuery)
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
// Note: This requires knowing which fields to search. For generic search,
// you might want to customize the query based on your schema.
func Search(dir string, pattern string) ([]QueryResult, error) {
	// For generic search, we search all text columns
	// This is a simple implementation - customize based on your schema
	query := fmt.Sprintf(`
		SELECT * FROM logs
		WHERE CAST(logs AS VARCHAR) LIKE '%%%s%%'
		ORDER BY timestamp DESC
		LIMIT 100
	`, pattern)

	return Query(dir, query)
}

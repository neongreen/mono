// Package rotatinglog provides a rotating, compressed log writer with search capabilities.
//
// This library implements a rotating log system that:
//   - Writes JSONL entries to an active log file
//   - Rotates files when they exceed a size threshold
//   - Compresses rotated files with zstd
//   - Handles concurrent writes with file locking
//   - Provides search via DuckDB queries
//
// # Architecture
//
// Directory structure:
//
//	logdir/
//	  current.jsonl              # Active log (uncompressed)
//	  current-2.jsonl            # Fallback if current is locked
//	  current-3.jsonl            # Further fallback
//	  2025-11-11-123456.jsonl.zst # Compressed archives
//	  2025-11-10-091234.jsonl.zst
//
// # Rotation
//
// When the active log file reaches the size threshold:
//  1. Rename to timestamped filename (YYYY-MM-DD-HHMMSS.jsonl)
//  2. Compress with zstd to .jsonl.zst
//  3. Delete uncompressed file
//  4. Start new current.jsonl
//
// Rotation happens in the background on the next write after threshold is exceeded.
//
// # Concurrent Writes
//
// Uses flock (Unix) or LockFileEx (Windows) for file locking:
//  1. Try to lock current.jsonl
//  2. If locked, try current-2.jsonl
//  3. Continue with current-N.jsonl until lock acquired
//  4. All current-*.jsonl files are compressed on rotation
//
// This ensures concurrent writers don't corrupt files while avoiding write failures.
//
// # Search
//
// Search is powered by DuckDB reading JSONL files:
//   - DuckDB auto-detects zstd compression
//   - read_json() handles both .jsonl and .jsonl.zst files
//   - Filter pushdown for efficient queries
//   - No persistent database needed
//
// Example:
//
//	writer, _ := rotatinglog.NewWriter("/path/to/logs", 10*1024*1024) // 10MB
//	writer.Append([]byte(`{"field": "value"}`))
//	writer.Close()
//
//	results, _ := rotatinglog.Query("/path/to/logs", "SELECT * FROM logs WHERE field = 'value'")
//
// # Design
//
// This library is generic and not tied to any specific log format. Users provide:
//   - Directory for log files
//   - Size threshold for rotation
//   - JSONL data to append
//
// The library handles all rotation, compression, and search mechanics.
package rotatinglog

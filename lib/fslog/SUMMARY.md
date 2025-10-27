# fslog Library Summary

## Overview

`fslog` is a Go library providing ACID-ish semantics over filesystem operations with an immutable operation log and point-in-time rollback support.

## Core Design

The library implements a layered architecture:

1. **FileSystem** - Main entry point managing the operation log and providing transaction support
2. **Transaction** - Groups multiple operations applied atomically with all-or-nothing semantics
3. **Operation** - Single filesystem change with complete before/after state capture
4. **OperationLog** - Persistent, append-only log in JSON Lines format

## Key Features

### Atomic Transactions

Operations are grouped into transactions that either fully succeed or fully fail:

```go
tx := fs.Begin(context.Background())
tx.WriteFile("config.toml", newConfig, 0644)
tx.WriteFile("state.json", newState, 0644)
tx.Commit() // Both files written atomically
```

### Immutable Operation Log

Every operation is recorded with complete state information:
- Unique sequential ID
- Operation type (create, write, delete, mkdir, rename)
- Target path
- Before state (content, mode, existence)
- After state (content, mode)
- Timestamp
- Optional metadata

The log is append-only and never modified, providing a complete audit trail.

### Point-in-Time Rollback

Restore the filesystem to any previous state:

```go
// Get operation history
ops, _ := fs.History()

// Rollback to state after operation 5
fs.RollbackTo(ops[4].ID)

// Rollback to initial state
fs.RollbackTo(0)
```

Rollback works by reversing operations in reverse chronological order, restoring the exact before state.

### Human-Readable Diffs

Each operation provides a simple diff summary:
- "Created config.txt (42 bytes)"
- "Modified data.json (128 -> 256 bytes)"
- "Deleted temp.log (1024 bytes)"
- "Created directory logs"
- "Renamed old.txt -> new.txt"

## Implementation Details

### Storage Format

Operations are stored in JSON Lines format (`operations.jsonl`):
```jsonl
{"id":1,"type":"create","path":"config.txt","before_exists":false,"after_content":"...","after_mode":420,"timestamp":"..."}
{"id":2,"type":"write","path":"config.txt","before_exists":true,"before_content":"...","after_content":"...","timestamp":"..."}
```

### File Layout

```
/path/to/data/
  .fslog/
    operations.jsonl  # Operation log
  config.txt          # Actual files
  data/
    file.txt
```

### Concurrency

- Read-write mutex protects the operation log
- Transactions can be created concurrently
- Commits are serialized
- Safe for single-process concurrent use

### Safety Guarantees

1. **Atomicity**: All operations in a transaction succeed or none do
2. **Durability**: Operations are synced to disk immediately after commit
3. **Consistency**: Before/after state is always captured completely
4. **Rollback Safety**: Can always restore to any logged state

## Supported Operations

- `WriteFile(path, content, mode)` - Write/create file
- `DeleteFile(path)` - Delete file
- `Mkdir(path, mode)` - Create directory
- `Rename(oldPath, newPath)` - Rename/move file or directory

## Use Cases

### Configuration Management

Tools like `conf` that modify user configuration files can use fslog to:
- Never lose user data even with bugs
- Provide undo/redo capabilities
- Show complete history of changes
- Rollback bad configuration updates

### File Processing

Scripts that transform files can:
- Try operations speculatively
- Rollback on errors
- Maintain audit trail
- Ensure data safety

### Development Tools

IDEs and build tools can:
- Implement safe refactoring
- Provide file history without git
- Support experimental changes with easy rollback

## Limitations

- **File Size**: Stores full content in log, not suitable for large files (>10MB)
- **Throughput**: Not designed for high-frequency operations
- **Log Growth**: No automatic log compaction or cleanup
- **Concurrency**: Single process only, no distributed locking
- **Binary Files**: Works but increases log size significantly

## Design Decisions

### Why Capture Full State?

Storing complete before/after content in each operation:
- **Pro**: Guarantees perfect rollback without relying on filesystem state
- **Pro**: Self-contained audit trail
- **Con**: Log grows proportional to data size
- **Decision**: Acceptable for target use case (small config files)

### Why JSON Lines?

JSON Lines format for the operation log:
- **Pro**: Human-readable
- **Pro**: Easy to debug
- **Pro**: Simple to parse and append
- **Pro**: Tool-friendly (grep, jq, etc.)
- **Con**: Larger than binary format
- **Decision**: Simplicity and debuggability over efficiency

### Why No Log Compaction?

The operation log grows indefinitely:
- **Reason**: Preserves complete history
- **Reason**: Simple implementation
- **Future**: Could add optional compaction/archival
- **Mitigation**: Target use case has low data volume

## Testing

All core functionality is tested:
- File creation and modification
- File deletion and directory creation
- File rename/move operations
- Transaction commit and rollback
- State rollback to any operation
- Rollback to initial state
- Multiple operations per transaction
- Operation diff generation
- History retrieval

See `fslog_test.go` for 13 comprehensive test cases.

## Example

See `example/main.go` for a complete working demonstration showing:
- Creating and modifying files
- Creating directories
- Viewing operation history
- Rolling back to previous states
- Transaction rollback

## Performance Characteristics

- **Transaction Begin**: O(1) - just allocates structure
- **Add Operation**: O(1) - appends to in-memory list
- **Transaction Commit**: O(n) where n = operations in transaction
- **Rollback**: O(m) where m = operations to reverse
- **History**: O(k) where k = total operations in log

## Future Enhancements

Potential additions (not currently implemented):
- Log compaction and archival
- Binary format option for efficiency
- Streaming large files (chunks)
- Multi-process locking
- Remote operation log storage
- Incremental snapshots
- Diff visualization tools
- Integration with version control systems

## Status

Pre-alpha. API may change. Not recommended for production use without thorough testing for your specific use case.

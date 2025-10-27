# fslog - ACID-ish Filesystem Operations Library

`fslog` is a Go library that provides ACID-ish semantics over filesystem operations. It maintains an immutable operation log and supports rollback to any past state, making it safe to use for tools that modify user data.

## Features

- **Atomic Transactions**: Group multiple filesystem operations that are applied atomically
- **Immutable Operation Log**: Append-only log of all filesystem operations with before/after state
- **Rollback Support**: Restore filesystem to any previous state
- **Operation Diffs**: Human-readable diffs showing what changed in each operation
- **Data Safety**: Copy-on-write semantics prevent data loss even in the presence of bugs

## Use Cases

This library is designed for tooling (like configuration managers, file processors, etc.) that:
- Reads and writes a small amount of data
- Requires absolute certainty that user data won't be lost
- Needs audit trails of all changes
- Benefits from rollback capabilities

## Installation

```bash
go get github.com/neongreen/mono/lib/fslog
```

## Usage

### Basic Example

```go
package main

import (
    "context"
    "log"
    "github.com/neongreen/mono/lib/fslog"
)

func main() {
    // Create a new filesystem that operates on /path/to/data
    // Logs will be stored in /path/to/data/.fslog by default
    fs, err := fslog.New("/path/to/data", "")
    if err != nil {
        log.Fatal(err)
    }
    defer fs.Close()

    // Begin a transaction
    tx := fs.Begin(context.Background())

    // Perform operations
    if err := tx.WriteFile("config.toml", []byte("key = value"), 0644); err != nil {
        log.Fatal(err)
    }

    if err := tx.Mkdir("subdir", 0755); err != nil {
        log.Fatal(err)
    }

    // Commit the transaction (applies all operations atomically)
    if err := tx.Commit(); err != nil {
        log.Fatal(err)
    }
}
```

### Supported Operations

- `WriteFile(path, content, mode)` - Write to a file (creates if doesn't exist)
- `DeleteFile(path)` - Delete a file
- `Mkdir(path, mode)` - Create a directory
- `Rename(oldPath, newPath)` - Rename/move a file

### Viewing History

```go
// Get all operations in chronological order
ops, err := fs.History()
if err != nil {
    log.Fatal(err)
}

for _, op := range ops {
    fmt.Printf("%d: %s at %s\n", op.ID, op.Diff(), op.Timestamp)
}
```

### Rolling Back

```go
// Get operation history
ops, err := fs.History()
if err != nil {
    log.Fatal(err)
}

// Rollback to the state after operation 5
if err := fs.RollbackTo(ops[4].ID); err != nil {
    log.Fatal(err)
}

// Rollback to initial state (before any operations)
if err := fs.RollbackTo(0); err != nil {
    log.Fatal(err)
}
```

### Transaction Rollback

```go
// Start a transaction
tx := fs.Begin(context.Background())

// Perform operations
tx.WriteFile("test.txt", []byte("content"), 0644)

// Decide not to commit
tx.Rollback()

// File won't be created
```

## Architecture

### FileSystem

The main entry point. Manages the operation log and provides transaction support.

### Transaction

Groups multiple operations that are applied atomically. Operations are only written to disk when `Commit()` is called.

### Operation

Represents a single filesystem change with complete before/after state:
- `OpCreate` - File creation
- `OpWrite` - File modification  
- `OpDelete` - File deletion
- `OpMkdir` - Directory creation
- `OpRename` - File/directory rename

### OperationLog

Append-only, immutable log stored in SQLite database. Each operation includes:
- Unique ID (auto-incremented)
- Operation type
- Target path
- Before state (content, mode, existence)
- After state (content, mode)
- Timestamp
- Metadata

## Design Principles

1. **Immutability**: The operation log is append-only and never modified
2. **Complete State**: Each operation captures full before/after state for reliable rollback
3. **Atomicity**: Transactions apply all operations or none
4. **Durability**: Operations are synced to disk immediately after commit
5. **Safety First**: Designed to prevent data loss even when the calling application has bugs

## Limitations

- Not designed for large files or high-throughput scenarios
- Operation log grows over time (no automatic cleanup)
- No built-in conflict resolution for concurrent access
- Stores full file contents in operation log (not suitable for binary files or large files)

## Status

This library is pre-alpha. The API may change without notice.

## Testing

```bash
# Run tests
mise run //lib/fslog:test

# Or use go directly
cd lib/fslog
go test -v ./...
```

## Example

See [example/main.go](example/main.go) for a complete demonstration of the library's features.

Run the example:
```bash
cd lib/fslog/example
go run main.go
```

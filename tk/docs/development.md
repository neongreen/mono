# tk Development Guide

This guide covers how to work on tk itself.

## Development vs Installation

tk has two ways to run the binary:

### Production Binary (`tk`)

The globally installed binary, typically installed via:
- Homebrew: `brew install tk`
- Want: `want mono tk@latest`
- Manual installation from releases

This is what you use for normal task tracking.

### Development Binary (`tk-dev`)

The development binary that automatically builds from source. Located at `/path/to/mono/bin/tk-dev`.

**How it works:**
- Just run `tk-dev <command>` - it automatically builds from source on each invocation
- Always use `tk-dev` when testing local changes to avoid conflicts with your global installation
- No need to run `go build` manually

**Example:**
```bash
# Make changes to tk code
vim internal/database/db.go

# Test your changes
tk-dev ls  # Automatically builds and runs from source

# Compare with production
tk ls      # Uses the globally installed binary
```

## Building Manually

If you want to build manually instead of using `tk-dev`:

```bash
# Using mise
mise tk:build  # Builds to _build/tk

# Using go directly
cd tk
go build -o tk .
```

## Testing

Run all tests:

```bash
go test ./...
```

Run specific tests:

```bash
# Test a specific package
go test ./internal/database

# Test a specific function
go test ./internal/database -run TestTaskResolver
```

## Code Organization

```
tk/
├── cmd/              # CLI commands (using cobra)
│   ├── new.go       # tk new command
│   ├── ls.go        # tk ls command
│   ├── project/     # tk project subcommands
│   └── ...
├── internal/        # Internal packages
│   ├── database/    # Database layer (SQLite)
│   ├── tasks/       # Task creation and management
│   ├── types/       # Type definitions and validation
│   ├── reducer/     # Event sourcing reducer
│   └── ...
├── docs/            # Documentation
├── specs/           # Technical specifications
└── main.go          # Entry point
```

## Key Concepts for Contributors

1. **Event Sourcing**: All changes are recorded as immutable events before updating state tables
2. **Type Safety**: Use the typed wrappers (`TaskUID`, `ProjectUID`) instead of raw strings
3. **Validation**: Always validate input using the `Validate()` methods on types
4. **Testing**: Write tests for new features, especially around event handling and sync

## See Also

- [concepts.md](concepts.md) - Understand tk's architecture
- [../specs/](../specs/) - Technical specifications
- [IMPLEMENTATION.md](../IMPLEMENTATION.md) - Implementation details

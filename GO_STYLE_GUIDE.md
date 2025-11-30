# Go Style Guide

This document defines coding standards for Go code in this monorepo. These patterns are based on existing code in the repository but are standardized here for consistency.

## CLI Framework

### Use Cobra

All CLI applications must use [Cobra](https://github.com/spf13/cobra) for command-line argument parsing and command structure.

**Examples:**
- `conf/main.go` - Full-featured CLI with subcommands (uses cmd/ package for command logic)
- `tk/main.go` - Simple CLI with RunE handlers
- `dissect/main.go` - CLI with subcommands

**Exceptions:**
- `prrun/main.go` - Uses manual flag parsing due to simplicity (single command with minimal flags)
- `want/main.go` - Uses `flag` package for simple flag parsing

### Command Structure

Use this structure for Cobra commands:

```go
var rootCmd = &cobra.Command{
    Use:   "tool-name",
    Short: "Brief description",
    Long:  `Detailed description with examples.`,
}

var subCmd = &cobra.Command{
    Use:   "subcommand [args]",
    Short: "Brief description",
    Long:  `Detailed description.`,
    Args:  cobra.ExactArgs(1), // or MinimumNArgs, RangeArgs, etc.
    RunE:  runSubCommand,
}

func init() {
    rootCmd.AddCommand(subCmd)
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Command Flags

**Prefer using `init()` to define command flags** rather than defining them in `root.go` or elsewhere. This keeps flag definitions close to the command they belong to, making the code more maintainable and self-contained.

**Good:**
```go
var lsCmd = &cobra.Command{
    Use:   "ls",
    Short: "List items",
    RunE: func(cmd *cobra.Command, args []string) error {
        status, _ := cmd.Flags().GetString("status")
        jsonOutput, _ := cmd.Flags().GetBool("json")
        // ... implementation
        return nil
    },
}

func init() {
    lsCmd.Flags().String("status", "", "Filter by status")
    lsCmd.Flags().Bool("json", false, "Output as JSON")
}
```

**Acceptable (for commands with many flags):**
```go
// In root.go, for commands with 10+ flags where centralization aids discoverability
lsCmd.Flags().String("status", "", "Filter by status")
lsCmd.Flags().String("kind", "", "Filter by kind")
// ... many more flags
RootCmd.AddCommand(lsCmd)
```

### Use RunE for Command Handlers

**Always use `RunE` instead of `Run` for command handlers.** This allows errors to be returned and handled consistently in main.

**Good:**
```go
var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize database",
    RunE: func(cmd *cobra.Command, args []string) error {
        db, err := openDB()
        if err != nil {
            return fmt.Errorf("failed to open database: %w", err)
        }
        defer db.Close()
        return db.Init()
    },
}
```

**Bad:**
```go
var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize database",
    Run: func(cmd *cobra.Command, args []string) {
        db, err := openDB()
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
        defer db.Close()
        if err := db.Init(); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    },
}
```

**Why RunE is better:**
- More testable (can test error cases)
- Consistent with Go error handling philosophy
- Centralizes exit handling in main
- Better separation of concerns

### Flags

Use persistent flags for flags shared across commands:

```go
var dryRun bool

func init() {
    rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without making modifications")
}
```

Use command-specific flags for flags unique to a command:

```go
var outputDir string

func init() {
    extractCmd.Flags().StringVarP(&outputDir, "output", "o", "./output", "Output directory")
}
```

## Error Handling

### Error Wrapping

Always use `%w` verb when wrapping errors to preserve the error chain. This is already documented in `AGENTS.md` but repeated here for completeness.

**Good:**
```go
if err != nil {
    return fmt.Errorf("failed to fetch release %s/%s tag %s: %w", owner, repo, tag, err)
}
```

**Bad:**
```go
if err != nil {
    return fmt.Errorf("failed to fetch release: %v", err)  // Loses error chain
}
```

### Error Context

Always include relevant context in error messages:
- File paths for file operations
- URLs for HTTP requests
- Resource identifiers (project names, tag names, etc.)
- What operation was being performed

### Error Output in Main

Standardize error output format in main:

```go
func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

Use `fmt.Fprintf(os.Stderr, "Error: %v\n", err)` consistently. The error returned from RunE handlers should already include context.

## Project Structure

### CLI Applications

**Single-application projects must have `main.go` in the project root.**

The `/cmd/{app}/main.go` pattern is for repositories that produce **multiple different binaries** (e.g., kubernetes/kubernetes with kubelet, kube-apiserver, kubectl). Since each project in this monorepo is a single application, `main.go` belongs at the project root.

```
project/
├── main.go            # CLI entry point (package main)
├── cmd/               # Command logic (optional, for Cobra commands)
│   ├── root.go
│   ├── list.go
│   └── ...
├── pkg/
│   └── ...            # Package code
├── internal/
│   └── ...            # Internal packages
├── go.mod
└── mise.toml
```

**Pattern used in all projects:**
- `tk/main.go`, `want/main.go`, `conf/main.go`, `dissect/main.go`, `ingest/main.go`, `printpdf/main.go`, `prrun/main.go`, `markdown-format/main.go`, `jj-run/main.go`

**cmd/ subdirectory (optional):**
- For projects using Cobra, the `cmd/` subdirectory may contain command implementation files (e.g., `tk/cmd/root.go`, `conf/cmd/jj.go`)
- These are **not** package main - they're part of a `cmd` package imported by `main.go`
- This provides clean separation between entry point and command logic

### Libraries

Libraries should use `pkg/` subdirectory:

```
lib/
└── project/
    ├── pkg/
    │   └── ...        # Library code
    ├── go.mod
    └── mise.toml
```

**Examples:**
- `lib/ghclient/`
- `lib/ghrelease/`
- `lib/toml/`

## Code Organization

### Import Organization

Organize imports in three groups, separated by blank lines:

1. Standard library packages
2. Third-party packages
3. Local packages (from this monorepo)

```go
import (
    "fmt"
    "os"
    "strings"

    "github.com/spf13/cobra"

    "github.com/neongreen/mono/conf/pkg/config"
    "github.com/neongreen/mono/conf/pkg/diff"
)
```

**Example:** See `conf/main.go` and `tk/main.go` for reference.

### Package Naming

- Use lowercase, single-word package names
- No underscores or mixedCaps
- Short, concise names: `config`, `diff`, `sync`

### File Organization

Split code into multiple files when:
- A single file exceeds ~500 lines
- Related functionality groups naturally (e.g., `types.go`, `parser.go`, `renderer.go`)
- Commands are split into separate files (e.g., `cmd/init.go`, `cmd/list.go`)

## Testing

### Test File Naming

Test files must end with `_test.go` and be in the same directory as the code they test.

```
pkg/
├── parser.go
├── parser_test.go      # Tests for parser.go
└── types.go
```

### Test Organization

- Use `testdata/` directory for test fixtures and golden files
- Use table-driven tests for multiple similar test cases
- Reference existing testing documentation:
  - `dissect/TESTING.md` - File-based integration testing
  - `markdown-format/TESTING.md` - Golden file testing

### Test Package Names

- Use `package main` for testing main package (white-box testing)
- Use `package package_test` for black-box testing of public API only

## Documentation Comments

- Export all public symbols with documentation comments
- Use complete sentences
- Start with the name of the symbol being described

```go
// DownloadReleaseAsset downloads a release asset for the specified platform.
// It returns an error if the release or asset is not found.
func DownloadReleaseAsset(owner, repo, tag, projectName, destPath string) error {
    // ...
}
```

## Examples

### Complete CLI Example

```go
package main

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "mytool",
    Short: "A tool that does something",
    Long:  `A tool that does something useful with examples.`,
}

var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize configuration",
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg, err := loadConfig()
        if err != nil {
            return fmt.Errorf("failed to load config: %w", err)
        }
        return cfg.Init()
    },
}

func init() {
    rootCmd.AddCommand(initCmd)
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Library Example

```go
package config

import (
    "fmt"
    "os"
)

// Config represents application configuration.
type Config struct {
    Path string
}

// Load loads configuration from the default location.
// Returns an error if the configuration file cannot be read.
func Load() (*Config, error) {
    path, err := defaultConfigPath()
    if err != nil {
        return nil, fmt.Errorf("failed to determine config path: %w", err)
    }
    
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
    }
    
    return parseConfig(data)
}
```

## Migration Notes

Some existing code does not follow these standards yet:

- `conf/main.go` - Uses `Run` instead of `RunE` (migration recommended)
- `dissect/main.go` - Uses `Run` instead of `RunE` (migration recommended)
- `ingest/main.go` - Uses `Run` instead of `RunE` (migration recommended)
- `printpdf/main.go` - Uses `Run` instead of `RunE` (migration recommended)
- `tk/main.go` - Main error output uses `fmt.Fprintln` instead of `fmt.Fprintf` (minor inconsistency)

New code should follow these standards. Existing code should be migrated when convenient.

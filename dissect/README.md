# dissect

A Go tool that automatically refactors Go code by extracting top-level functions into separate files, following Go's best practices for code organization.

## Overview

`dissect` helps break down large Go files with multiple functions into smaller, more focused files where each function lives in its own file. It uses `gopls` (the Go language server) to perform intelligent refactoring that preserves imports, handles methods correctly, and maintains code correctness.

## Features

- **Automatic function extraction** - Extracts each top-level function from a Go file into its own file
- **Selective moving** - Move specific functions, types, and interfaces to target files
- **Intelligent naming** - Creates descriptive file names based on function names (e.g., `util_foo.go` for `foo()`, `mystruct_baz.go` for `(*MyStruct).Baz()`)
- **Import management** - Automatically handles imports using `gopls` and `goimports`
- **Method support** - Correctly handles methods and groups them with their receiver type
- **Test helper extraction** - Moves test helper functions to `internal/testutils` package
- **Build verification** - Ensures the code still builds after refactoring

## Supported Declarations

The `move` command supports the following Go declarations:

| Declaration Type | Status | Notes |
|-----------------|--------|-------|
| Functions | ✅ Implemented | Regular functions and methods |
| Types | ✅ Implemented | Struct types, type aliases |
| Interfaces | ✅ Implemented | Interface definitions |
| Constants | ✅ Implemented | Const declarations |
| Variables | ✅ Implemented | Var declarations |

The `explode` command currently only extracts **functions** automatically. Support for automatically splitting types and other declarations is not yet implemented.

## Installation

### With Go

```bash
cd dissect
go build -o dissect ./cmd
```

### With mise

Install using [mise](https://mise.jdx.dev/) with the Go backend:

```bash
mise use -g go:github.com/neongreen/mono/dissect@main
```

Or add to your `.mise.toml`:

```toml
[tools]
"go:github.com/neongreen/mono/dissect" = "main"
```

## Requirements

- Go 1.24.4 or later
- `gopls` (Go language server) must be installed and available in PATH

## Usage

`dissect` has two main commands:

1. **explode** - Automatically extracts all functions from files (WARNING: usually you want `move` instead)
2. **move** - Selectively moves specific functions to a target file

### Explode Command

**WARNING:** The `explode` command puts each function into its own separate file. For most refactoring tasks, you should use the `move` command instead to selectively move specific functions to target files.

The `explode` command automatically extracts each function from a file into its own separate file.

Extract functions from a single file:

```bash
dissect explode path/to/file.go
```

Extract functions from multiple files:

```bash
dissect explode file1.go file2.go file3.go
```

Extract functions from all Go files in a directory:

```bash
dissect explode path/to/directory
```

### Move Command

The `move` command allows you to selectively move specific declarations (functions, types, interfaces) to a target file.

Move a single function to a new file:

```bash
dissect move source.go:FunctionName target.go
```

Move types and interfaces:

```bash
dissect move source.go:MyType,MyInterface target.go
```

Move multiple declarations using comma-separated list:

```bash
dissect move source.go:Foo,Bar,Baz target.go
```

Move declarations from different files to the same target:

```bash
dissect move file1.go:Foo file2.go:Bar target.go
```

#### What Can Be Moved

The `move` command supports moving:

- **Functions** - Regular functions and methods
- **Types** - Struct types, type aliases (e.g., `type MyInt int`)
- **Interfaces** - Interface definitions
- **Consts** - Constant declarations
- **Vars** - Variable declarations

#### Glob Pattern Support

Both file paths and identifier names support glob patterns:

```bash
# Move all functions named "Helper" from any .go file
dissect move *.go:Helper target.go

# Move all functions starting with "Test" from files in pkg/
dissect move pkg/**/*.go:Test* target.go

# Move all identifiers ending with "Helper" or starting with "Util"
dissect move file.go:*Helper,Util* target.go

# Move all types matching a pattern
dissect move source.go:*Type target.go
```

**Glob Behavior:**
- If a file doesn't contain a matching identifier, it's silently skipped (no error)
- An error is only shown if no identifiers match across all files
- File globs are expanded first, then identifier globs match within each file

The target file will be created if it doesn't exist, or declarations will be appended if it does exist.

### What Gets Extracted

For a file like this:

```go
package main

import "fmt"

func main() {
    foo()
    barQuix(2)
    s := MyStruct{}
    s.Baz()
}

func foo() {
    fmt.Println("This is foo")
}

func barQuix(x int) int {
    return x * 2
}

type MyStruct struct{}

func (s *MyStruct) Baz() {
    fmt.Println("This is Baz")
}
```

`dissect` will create:

- `util_foo.go` - Contains `foo()` function
- `util_bar_quix.go` - Contains `barQuix()` function
- `mystruct_baz.go` - Contains `(*MyStruct).Baz()` method
- The original file keeps `main()` and type definitions

### Test Files

For test files (`*_test.go`), helper functions are extracted to an `internal/testutils` package to follow Go's testing best practices.

## How It Works

1. **Parse** - Parses the Go file to identify top-level functions and methods
2. **Extract** - Uses `gopls refactor.extract.toNewFile` to move each function to its own file
3. **Clean up** - Runs `goimports` to organize imports
4. **Verify** - Optionally runs `go build` to ensure the code still compiles

## Examples

### Before

```
mypackage/
└── main.go (200 lines with 10 functions)
```

### After

```
mypackage/
├── main.go (main function and types)
├── util_foo.go
├── util_bar.go
├── util_baz.go
├── processor_process.go
├── validator_validate.go
└── ...
```

## Configuration

The tool respects `.gitignore` patterns and won't process ignored files.

## Integration

`dissect` works well with:

- **gopls** - Uses gopls for all refactoring operations
- **goimports** - Automatically organizes imports after extraction
- **go build** - Verifies code still compiles

## Known Limitations

- Functions with the same name in the same package may have naming conflicts (gopls handles this)
- Very large files may take longer to process
- Grouped declarations (e.g., `const (...)`) move as a complete block
- Moving a type doesn't move its methods (they remain in the source file)
- Dot imports are not validated upfront

For detailed information about limitations and edge cases, see [DESIGN.md](DESIGN.md).

## Development

See [ARCHITECTURE.md](ARCHITECTURE.md) for details about the internal structure.

See [TESTING.md](TESTING.md) for information about running tests.

See [DESIGN.md](DESIGN.md) for implementation approach and design decisions.

## Design Documents

- [DESIGN.md](DESIGN.md) - Implementation approach, technical analysis, and limitations
- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture and component structure
- [TESTING.md](TESTING.md) - Testing approach and guidelines
- [plan.md](plan.md) - Original development plan
- [TODO.md](TODO.md) - Known issues and future improvements
- [docs/gopls/](docs/gopls/) - Documentation about gopls integration
- [docs/logic/](docs/logic/) - Logic and design decisions

## License

See [LICENSE](../LICENSE) in the repository root.

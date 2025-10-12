# dissect

A Go tool that automatically refactors Go code by extracting top-level functions into separate files, following Go's best practices for code organization.

## Overview

`dissect` helps break down large Go files with multiple functions into smaller, more focused files where each function lives in its own file. It uses `gopls` (the Go language server) to perform intelligent refactoring that preserves imports, handles methods correctly, and maintains code correctness.

## Features

- **Automatic function extraction** - Extracts each top-level function from a Go file into its own file
- **Intelligent naming** - Creates descriptive file names based on function names (e.g., `util_foo.go` for `foo()`, `mystruct_baz.go` for `(*MyStruct).Baz()`)
- **Import management** - Automatically handles imports using `gopls` and `goimports`
- **Method support** - Correctly handles methods and groups them with their receiver type
- **Test helper extraction** - Moves test helper functions to `internal/testutils` package
- **Build verification** - Ensures the code still builds after refactoring

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

1. **split** - Automatically extracts all functions from files
2. **move** - Selectively moves specific functions to a target file

### Split Command

The `split` command automatically extracts each function from a file into its own separate file.

Extract functions from a single file:

```bash
dissect split path/to/file.go
```

Extract functions from multiple files:

```bash
dissect split file1.go file2.go file3.go
```

Extract functions from all Go files in a directory:

```bash
dissect split path/to/directory
```

### Move Command

The `move` command allows you to selectively move specific functions to a target file.

Move a single function to a new file:

```bash
dissect move source.go:FunctionName target.go
```

Move multiple functions using comma-separated list:

```bash
dissect move source.go:Foo,Bar,Baz target.go
```

Move functions from different files to the same target:

```bash
dissect move file1.go:Foo file2.go:Bar target.go
```

Move functions from files matching a glob pattern:

```bash
dissect move *.go:Helper target.go
dissect move pkg/**/*.go:Utility target.go
```

The target file will be created if it doesn't exist, or functions will be appended if it does exist.

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
- Some edge cases with complex type definitions (see [TODO.md](TODO.md))

## Development

See [ARCHITECTURE.md](ARCHITECTURE.md) for details about the internal structure.

See [TESTING.md](TESTING.md) for information about running tests.

## Design Documents

- [plan.md](plan.md) - Original development plan and design decisions
- [TODO.md](TODO.md) - Known issues and future improvements
- [docs/gopls/](docs/gopls/) - Documentation about gopls integration
- [docs/logic/](docs/logic/) - Logic and design decisions

## License

See [LICENSE](../LICENSE) in the repository root.

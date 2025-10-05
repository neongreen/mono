# dissect Architecture

This document explains the internal architecture and design of the dissect tool.

## Overview

The dissect tool is a Go refactoring tool that extracts functions from monolithic Go files into separate, focused files. It leverages `gopls` (Go language server) to perform safe, semantic-aware refactoring operations.

## Design Philosophy

1. **Use existing tools** - Rely on `gopls` for refactoring rather than implementing AST transformations from scratch
2. **Iterative processing** - Extract one function at a time, re-parsing between extractions to maintain correct line numbers
3. **Safe refactoring** - Let `gopls` handle imports, type checking, and semantic correctness
4. **Test-aware** - Special handling for test files and test helper functions

## Architecture Flow

```
Input Files → Parse & Validate → Extract Functions → Import Cleanup → Verification
     ↓              ↓                    ↓                  ↓              ↓
  .go files    go/parser         gopls extract      goimports     go build (optional)
                go/ast          refactor.extract
                                 .toNewFile
```

## Directory Structure

```
dissect/
├── cmd/                      # Main entry point
│   ├── main.go              # CLI with cobra
│   ├── main_test.go         # Integration tests
│   ├── process_file.go      # Core file processing logic
│   └── util_cwd_rel_path.go # Path utilities
├── pkg/
│   ├── commands/            # Go toolchain commands
│   │   ├── find_go_files.go
│   │   ├── find_go_module_root.go
│   │   ├── get_full_import_path.go
│   │   ├── run_go_build.go
│   │   ├── run_go_mod_tidy.go
│   │   ├── run_goimports_*.go
│   │   └── ...
│   ├── gopls/               # gopls integration
│   │   ├── extract_to_new_file.go  # Main refactoring command
│   │   ├── guess_extracted_file_name.go
│   │   ├── add_import.go
│   │   ├── add_dot_import.go
│   │   └── rename.go
│   ├── goutils/             # Go AST and parsing utilities
│   │   ├── find_func.go
│   │   ├── get_package_declaration.go
│   │   ├── get_receiver_type_name.go
│   │   ├── is_test_file.go
│   │   ├── is_test_function.go
│   │   ├── normalize_imports.go
│   │   ├── read_go_file.go
│   │   ├── should_refactor.go
│   │   ├── update_package_declaration.go
│   │   └── write_go_file.go
│   ├── refactor/            # Refactoring logic
│   │   └── determine_extraction_target.go
│   ├── testutils/           # Testing utilities
│   └── utils/               # General utilities
│       ├── capitalize_first_letter.go
│       ├── delete_file.go
│       ├── hash_string.go
│       ├── is_lower.go
│       └── move_file.go
├── tests/                   # Integration test cases
│   ├── simple.toml
│   ├── with_test.toml
│   ├── internal_extraction.toml
│   └── ...
└── docs/                    # Design documentation
    ├── gopls/
    └── logic/
```

## Core Components

### 1. Main Entry Point (`cmd/main.go`)

The CLI uses `cobra` for command-line parsing. It:
- Accepts file or directory paths
- Discovers Go files in directories
- Calls `ProcessFile` for each file
- Reports success/failure statistics

### 2. File Processing (`cmd/process_file.go`)

The `ProcessFile` function is the core processing loop:

```go
func ProcessFile(absPath string) (status RefactorStatus, exclusionReason string, err error)
```

**Processing steps:**

1. **Validation** - Check if file should be refactored (`goutils.ShouldRefactor`)
2. **Iterative extraction** - Loop until no more functions to extract:
   - Parse file to find functions
   - Determine extraction target
   - Call `gopls.ExtractToNewFile`
   - Re-parse to get updated AST
3. **Import cleanup** - Run `goimports` on all modified files
4. **Verification** - Optionally run `go build`

**Key insight:** The loop re-parses the file after each extraction because line numbers change when functions are removed.

### 3. gopls Integration (`pkg/gopls/`)

The `gopls` package wraps the Go language server's refactoring capabilities:

#### `extract_to_new_file.go`

The primary refactoring command:

```go
func ExtractToNewFile(filePath string, line, column int) (extractedFileName string, err error)
```

This executes:
```bash
gopls codeaction -kind=refactor.extract.toNewFile -exec -w <file>:<line>:<column>
```

**What gopls does:**
- Extracts the function at the given position
- Creates a new file with an appropriate name
- Moves the function to the new file
- Adds necessary imports to the new file
- Removes the function from the original file
- Removes unused imports from the original file

#### Other gopls operations

- `add_import.go` - Add imports when needed
- `rename.go` - Rename files or functions
- `guess_extracted_file_name.go` - Predict what gopls will name the extracted file

### 4. Go Utilities (`pkg/goutils/`)

AST parsing and Go-specific utilities:

- **`find_func.go`** - Locate functions in the AST
- **`is_test_file.go`** - Detect `*_test.go` files
- **`is_test_function.go`** - Detect test/benchmark functions
- **`should_refactor.go`** - Determine if a file should be processed
- **`get_receiver_type_name.go`** - Extract receiver type from methods
- **`normalize_imports.go`** - Standardize import formatting
- **`read_go_file.go` / `write_go_file.go`** - File I/O with AST parsing

### 5. Commands (`pkg/commands/`)

Wrappers for Go toolchain commands:

- **`find_go_module_root.go`** - Locate the `go.mod` file
- **`run_goimports_*.go`** - Format imports
- **`run_go_build.go`** - Verify compilation
- **`run_go_mod_tidy.go`** - Clean up dependencies

### 6. Refactoring Logic (`pkg/refactor/`)

High-level refactoring decisions:

- **`determine_extraction_target.go`** - Decide where to extract test helpers (same package vs `internal/testutils`)

## Key Algorithms

### Function Extraction Loop

```go
for {
    // 1. Parse the file
    fset, node := goutils.ReadGoFile(absPath)
    
    // 2. Find extractable functions
    funcs := findExtractableFunctions(node)
    
    if len(funcs) == 0 {
        break  // Done!
    }
    
    // 3. Extract first function
    funcDecl := funcs[0]
    line := fset.Position(funcDecl.Name.Pos()).Line
    column := fset.Position(funcDecl.Name.Pos()).Column
    
    extractedFile := gopls.ExtractToNewFile(absPath, line, column)
    
    // 4. Track changed files for cleanup
    changedFiles[extractedFile] = struct{}{}
    
    // Loop continues with fresh parse
}
```

**Why iterative?** Line numbers change after each extraction, so we must re-parse.

### Test Helper Extraction

Test helper functions need special handling:

1. **Detect test files** - Check for `_test.go` suffix
2. **Identify test helpers** - Functions that aren't `Test*`, `Benchmark*`, or `Example*`
3. **Extract to correct location**:
   - If helper uses `_test` package → extract to `internal/testutils`
   - If helper uses main package → extract normally

**Reasoning:** Go's package rules require test files in `package foo_test` to be in the same directory, but extracted helpers would violate the "one package per directory" rule. Solution: move to `internal/testutils`.

See [docs/logic/test-files.md](docs/logic/test-files.md) for details.

### File Naming

gopls determines file names automatically:

- Regular functions: `util_<snake_case_name>.go`
- Methods: `<receiver_type_snake_case>_<method_snake_case>.go`

Example:
- `foo()` → `util_foo.go`
- `(*MyStruct).Baz()` → `mystruct_baz.go`

## Testing

### Integration Tests (`cmd/main_test.go`)

Tests use a file-based approach with TOML configurations:

```toml
[files_in]
"go.mod" = '''...'''
"main.go" = '''...'''

[files_out]
"main.go" = '''...'''
"util_foo.go" = '''...'''
```

Each test:
1. Creates a temporary Go module
2. Writes input files
3. Runs dissect
4. Compares output against expected files

See [TESTING.md](TESTING.md) for details.

### Test Cases

- `simple.toml` - Basic function extraction
- `with_test.toml` - Test file handling
- `internal_extraction.toml` - Test helper extraction
- `no_refactor.toml` - Files that shouldn't be refactored
- `ignore_file.toml` - Respecting `.gitignore`

## Performance Considerations

- **gopls speed** - gopls is fast but each extraction is a separate invocation
- **Re-parsing** - Files are re-parsed after each extraction (necessary for correctness)
- **Parallelization** - Files are processed sequentially to avoid gopls conflicts

## Error Handling

The tool follows a fail-fast approach:
- Parsing errors → stop processing file
- gopls errors → stop processing file
- Build errors → optional, can be disabled

## Future Enhancements

See [TODO.md](TODO.md) for planned improvements:
- Handling type dependencies in extractions
- Mutually recursive function support
- Cleaning up empty leftover files
- Single-function file renaming

## Dependencies

- **`github.com/spf13/cobra`** - CLI framework
- **`github.com/golang-cz/devslog`** - Structured logging
- **`github.com/iancoleman/strcase`** - String case conversion
- **`github.com/pelletier/go-toml/v2`** - TOML parsing for tests
- **Go standard library** - `go/parser`, `go/ast`, `go/token`

## Design Documents

Additional documentation:
- [docs/gopls/commands.md](docs/gopls/commands.md) - gopls command reference
- [docs/gopls/to-new-file.md](docs/gopls/to-new-file.md) - Extract-to-new-file details
- [docs/logic/test-files.md](docs/logic/test-files.md) - Test file handling logic
- [plan.md](plan.md) - Original design plan and research notes

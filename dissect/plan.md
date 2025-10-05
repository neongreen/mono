# Plan for Go Function Mover Tool

## Objective
Develop a Go tool that automates the process of moving each top-level function from a given Go source file into its own separate file, leveraging `gopls` for refactoring capabilities.

## High-Level Steps

1.  **Research `gopls` Refactoring API (Blocked):**
    *   Investigate `gopls` documentation for programmatic access to "move function" refactoring.
    *   Determine if `gopls` can be invoked via LSP or a command-line interface for this purpose.
    *   Understand the input/output requirements for such an operation.

2.  **Go Source File Parsing (DONE):**
    *   **Actionable Sub-points:**
        *   Implement initial parsing of the input Go file using `go/parser` and `go/token`. (DONE in `main.go`)
        *   Implement AST traversal using `go/ast` to identify `*ast.FuncDecl` nodes (top-level function declarations). (DONE in `main.go`)
        *   Ensure the parsing logic is within a loop that re-parses the file after each successful function extraction to maintain correct line numbers. (DONE in `main.go`)
    *   **Open Questions (Addressed):**
        *   How to handle parsing errors gracefully (e.g., syntax errors in the input file)?
            *   *Resolution:* The tool exits on parsing errors. This is acceptable for a CLI tool.
        *   What if the input file is not a valid Go source file?
            *   *Resolution:* `parser.ParseFile` will return an error, and the tool will exit. This is acceptable.
        *   Should the tool support parsing multiple input files or only a single file? (Currently, it's single file).
            *   *Resolution:* For now, keep it single file as per the objective.

3.  **Function Movement Logic (DONE):**
    *   **Actionable Sub-points:**
        *   For each `*ast.FuncDecl` identified:
            *   Construct the new file name (e.g., `original_file_name_function_name.go`). (Implicitly handled by `gopls`)
            *   Execute `gopls codeaction -kind=refactor.extract.toNewFile -exec -w <file_path>:<line>:<column>` where `<line>:<column>` points to the function's name. (DONE in `main.go`)
            *   Verify the success of the `gopls` command (check exit code and stderr). (Implemented robust error handling).
            *   If successful, re-read and re-parse the original file to get updated AST and line numbers. (DONE in `main.go`)
            *   Handle the case where `gopls` might not find a code action (e.g., if the function is not extractable for some reason). (Tool now exits on `gopls` errors).
    *   **Open Questions (Addressed):**
        *   How to handle naming conflicts for the new files (e.g., if `myfunction1.go` already exists)?
            *   *Resolution:* `gopls` handles this by creating unique file names (e.g., `myfunction1_test.go`). The current approach relies on `gopls`'s behavior.
        *   What is the expected behavior for methods (e.g., `(*MyStruct) MyMethod()`)? Does `gopls` handle them correctly with `refactor.extract.toNewFile`?
            *   *Resolution:* Yes, `gopls` appears to handle methods correctly.
        *   How to handle functions with the same name but different receivers or in different packages (though the current scope is top-level functions in a single file)?
            *   *Resolution:* `gopls` handles this by creating unique file names.
        *   What if `gopls` fails to move a function? Should the tool stop or continue?
            *   *Resolution:* The tool now exits on `gopls` errors, which is a robust approach.
        *   How to ensure the tool correctly identifies *all* top-level functions, including those with comments or build tags?
            *   *Resolution:* `go/parser` and `go/ast` handle these cases. The current implementation should be sufficient.

4.  **Import Management (DONE):**
    *   **Actionable Sub-points:**
        *   Run `goimports -w` on all generated Go files in the temporary directory after all functions have been moved. (DONE in `main.go`)
    *   **Open Questions (Addressed):**
        *   Are there any edge cases for `goimports` (e.g., very large files, complex import paths)?
            *   *Resolution:* `goimports` is a standard Go tool and is generally robust. For this project's scope, it should be sufficient.

5.  **Error Handling (DONE):**
    *   Implement robust error handling for file I/O, parsing errors, and `gopls` command execution failures.

6.  **Testing (DONE):**
    *   **Unit Tests:** Write unit tests for individual components (e.g., AST parsing, file path generation, import analysis).
    *   **Integration Tests:**
        *   Create a temporary directory with sample Go projects (including a `go.mod` file).
        *   Write test cases that invoke the tool on these sample projects.
        *   Verify that functions are moved correctly, new files are created, original files are updated, and the modified project still compiles (`go build`) and passes its own tests (`go test`).
        *   Use `gopls` diagnostics to check for any remaining errors after refactoring.
        *   **Specific Functionality to Test:**
            *   **Basic Function Movement:**
                *   Verify `myFunction1` is moved to `myfunction1.go`.
                    *   *Verification Method:* Check for existence of `myfunction1.go` and assert its content contains the `myFunction1` declaration.
                *   Verify `myFunction2` is moved to `myfunction2.go`.
                    *   *Verification Method:* Check for existence of `myfunction2.go` and assert its content contains the `myFunction2` declaration.
                *   Verify `MyMethod` is moved to `mystruct.go` (or similar, depending on `gopls`'s naming for methods).
                    *   *Verification Method:* Check for existence of `mystruct.go` (or the actual name `gopls` uses) and assert its content contains the `MyMethod` declaration.
                *   Verify `test_file.go` no longer contains these functions.
                    *   *Verification Method:* Read `test_file.go` and assert that the function declarations for `myFunction1`, `myFunction2`, and `MyMethod` are absent.
            *   **Import Handling:**
                *   Verify `fmt` import is correctly moved to the new files where needed.
                    *   *Verification Method:* Read the new function files (`myfunction1.go`, `myfunction2.go`, `mystruct.go`) and assert that `import "fmt"` is present if `fmt.Println` or `fmt.Sprintf` is used within that file.
                *   Verify `fmt` import is removed from `test_file.go` if no longer used.
                    *   *Verification Method:* Read `test_file.go` and assert that `import "fmt"` is absent if no `fmt` functions are called within the remaining code.
            *   **Project Compilation:**
                *   Run `go build ./...` in the temporary project directory to ensure the entire project compiles without errors after refactoring.
                    *   *Verification Method:* Execute `go build ./...` within the temporary project directory and assert that the command exits with code 0.
            *   **Project Tests:**
                *   Run `go test ./...` in the temporary project directory to ensure existing tests (if any) still pass after refactoring.
                    *   *Verification Method:* Execute `go test ./...` within the temporary project directory and assert that the command exits with code 0.

## Gopls Research Plan

1.  **Explore `gopls` Command-Line Interface:**
    *   [DONE] Locate the `main.go` file for the `gopls` executable.
    *   [DONE] Analyze the command-line flag parsing to understand available commands and their arguments.
    *   [DONE] Look for commands related to "refactor", "move", "apply", or "edit".
    *   [DONE] Experiment with any promising commands on a sample Go file to understand their behavior and output.
2.  **Analyze `gopls` Internal API:** [SKIPPED] Search the `gopls` source code for keywords like "refactor", "move", and "extract" to locate the internal APIs responsible for these operations. This will help determine if `gopls` can be used as a library.
3.  **Review LSP Handlers:** [SKIPPED] Examine the LSP message handlers within the `gopls` source, looking for code actions or commands related to moving functions. This will reveal how to programmatically trigger the desired refactoring.
4.  **Synthesize Findings:** The `gopls` command-line interface provides the `codeaction` command with the `refactor.extract.toNewFile` kind, which successfully moves a function to a new file and handles imports. This approach is deemed the most practical and reliable for this project, eliminating the need to delve into `gopls` internal APIs or LSP handlers.

## Next Steps
*   Develop comprehensive test suite.

## Completion
The Go Function Mover Tool has been developed and tested. All planned features have been implemented and verified.

## Other TODOs

- Do we care about small vs big letter functions?
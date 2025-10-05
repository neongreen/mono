# Plan for Refactoring with External Gopls and Custom File Management

This plan outlines a strategy to leverage `gopls` for its robust code extraction capabilities while maintaining fine-grained control over the extracted file's path and package declaration. This approach avoids reimplementing complex `gopls` logic or patching its source code.

## Goal

To correctly extract a Go function to a new file, place it at a custom path, and update its package declaration, ensuring all imports are correctly managed and the Go module remains buildable.

## Steps

### Step 1: Execute `gopls` for Function Extraction

1.  **Identify Target:** Determine the source file path and the exact byte range (start and end offsets) of the function to be extracted.
2.  **Construct `gopls` Command:** Formulate a command to execute `gopls` for the `refactor.extract.toNewFile` action. This command will typically involve:
    *   `gopls codeaction -kind=refactor.extract.toNewFile -exec -w`
    *   Parameters for the source file URI and the selection range (e.g., `<file_path>:<line>:<column>`).
    *   Example (conceptual): `gopls codeaction -kind=refactor.extract.toNewFile -exec -w file:///path/to/original.go:10:5`
3.  **Execute Command:** Run the `gopls` command. This command will directly modify the original file and create the new file in the same directory.

### Step 1.5: Remove Extracted Function from Original File

1.  **Identify and Remove:** After `gopls` extracts the function to a new file, the original function declaration remains in the source file. `dissect` must remove this function from the original file.
2.  **Constraint:** Due to the constraint of not implementing AST parsing logic directly, `dissect` cannot programmatically remove the function by manipulating the AST. This step currently represents a limitation that needs to be addressed by a future `gopls` feature or a different approach that adheres to the constraints.

### Step 2: Move Extracted File

1.  **Locate New File:** After `gopls` execution, locate the newly created file. `gopls` typically creates the new file in the same directory as the original file, with a name derived from the extracted function.
3.  **Determine Custom Destination:** Decide on the desired custom path for the new file (e.g., `internal/myutils/extracted_func.go`).
4.  **Write New File to Custom Path:** Write the extracted content to the determined custom path. This step gives full control over the file's location, overriding `gopls`'s suggested path.

### Step 3: Update Package Declaration

1.  **Read New File:** Read the content of the newly created file from its custom path.
2.  **Parse AST:** Parse the file's content into an `go/ast` representation.
3.  **Modify Package Name:** Access the `ast.File.Name` field and update it to the correct package name corresponding to the new file's directory (e.g., if moved to `internal/myutils`, change to `package myutils`).
4.  **Format and Write Back:** Use `go/format` to format the modified AST and write the updated content back to the file.

### Step 4: Post-Processing and Module Synchronization

1.  **Run `goimports`:** Execute `goimports` on the directory containing the newly moved and modified file. This ensures that imports are correctly organized and resolved for the new package context.
2.  **Run `go mod tidy`:** Execute `go mod tidy` at the root of the Go module. This is crucial to:
    *   Add any new module dependencies that might have become necessary due to the refactoring (e.g., if the extracted function now imports a package that wasn't previously imported in the original file's context).
    *   Remove unused dependencies.
    *   Synchronize the `go.mod` file with the actual import paths used in the code, especially if `goimports` modified any `/vX` suffixes.
3.  **Verify Build:** Run `go build ./...` or `go test ./...` to ensure the entire module still builds and passes tests after the refactoring.

## Considerations

*   **Error Handling:** Implement robust error handling at each step, especially for external command execution and file operations.
*   **Dependency Management:** Be mindful of dependencies. If the extracted function relies on unexported symbols from its original package, or if it calls functions that are now in a different package, further refactoring (e.g., making symbols exported, moving more code, or qualifying calls) might be necessary.
*   **Temporary Directories:** When testing this process, ensure that temporary directories are properly managed and cleaned up.
*   **`gopls` Version:** Ensure compatibility with the `gopls` version being used.

## Files to be Modified for Implementation

To implement this plan, the following files within the `dissect` project would likely need to be modified:

1.  **`pkg/refactor/gopls.go`**: This file would be the primary location for implementing the interaction with `gopls`. This includes constructing and executing the `gopls` command for function extraction.

2.  **`pkg/fileutils/fileutils.go`**: This file might need new functions or modifications to existing ones to handle writing the content of the newly extracted file to a custom path, as well as reading its content for package declaration updates.

3.  **`pkg/refactor/package.go`**: This file might be involved in the logic for updating the package declaration of the newly created file (Step 3), especially if it contains utilities for parsing and manipulating Go package information.

4.  **`cmd/main.go`**: This is the main entry point of the `dissect` tool. The overall workflow for orchestrating the steps (calling `gopls`, applying changes, moving files, updating package, running `goimports`, and `go mod tidy`) would be integrated here or in a new package called from here.

5.  **`cmd/main_test.go`**: New test cases would need to be added to this file to thoroughly verify the correct implementation of the function extraction, file movement, package update, and post-processing steps.

Additionally, the `go.mod` file would be affected by the `go mod tidy` command, but this is an external tool modifying the file, not a direct code modification within `dissect`. The original source file and the newly extracted file are dynamic and depend on the specific refactoring operation being performed.

## Open Questions

*   **Updating References to Moved Identifiers:** When a function is extracted to a new file and its package is changed (as per Step 3), `gopls`'s `refactor.extract.toNewFile` action handles updating references within the *original* file. However, any other files in the codebase that previously imported the *original* package and used the *extracted* identifier will now need to be updated to import the *new* package and reference the identifier from there. `goimports` and `go mod tidy` (Step 4) primarily handle import path resolution and module dependencies, but they do not automatically update all call sites across the entire codebase when a symbol's package changes. Therefore, a separate mechanism (e.g., a targeted search and replace, or a more comprehensive refactoring tool if available) would be necessary to update all such references. This is a manual step not fully covered by the current plan's automated steps.
